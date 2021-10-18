package utils

import (
	"os"
	"errors"
	"image"
    "strings"
	"log"
	"bytes"
	"fmt"
	"io"
	"image/jpeg"
	"image/png"
	"path/filepath"
    "contented/models"
	"github.com/nfnt/resize"
    "github.com/gofrs/uuid"
    "github.com/disintegration/imaging"
    "github.com/tidwall/gjson"
	ffmpeg "github.com/u2takey/ffmpeg-go"
)

// Used in the case of async processing when creating Preview results
type PreviewResult struct {
    C_ID uuid.UUID
    MC_ID uuid.UUID
    Preview string
    Err error
}

type PreviewRequest struct {
    C *models.Container
    Mc *models.MediaContainer
    Out chan PreviewResult
    Size int64
}

type PreviewWorker struct {
    Id int
    In chan PreviewRequest
}

func MakePreviewPath(dstPath string) error {
	if _, err := os.Stat(dstPath); os.IsNotExist(err) {
		return os.MkdirAll(dstPath, 0740)
	}
	return nil
}

func ResetPreviewDir(dstDir string) error {
    os.RemoveAll(dstDir + "/")
    return MakePreviewPath(dstDir)
}

func ShouldCreatePreview(f *os.File, fsize int64) bool {
	finfo, err := f.Stat()
	if err != nil {
		log.Printf("Error determining file stat for %s", err)
		return false
	}

	if finfo.Size() > fsize {
		// log.Printf("How big was the size %d", finfo.Size())
		return true
	}
	return false
}

// TODO: make the preview directory name configurable
// Make it use the container Path instead of the name?
func GetPreviewDst(fqDir string) string {
    return filepath.Join(fqDir, "container_previews")
}

// Get the relative path for a preview
func GetRelativePreviewPath(fqPath string, cntPath string) string {
    return strings.ReplaceAll(fqPath, cntPath, "")
}

// In most cases it is currently considered an error if you are trying to create a preview and
// one already exists (mostly this is still debugging)
func ErrorOnPreviewExists(filename string, dstPath string, contentType string) (string, error) {
    dstFile := GetPreviewPathDestination(filename, dstPath, contentType)
	if _, e_err := os.Stat(dstFile); os.IsNotExist(e_err) {
		return dstFile, nil
	}
	return dstFile, errors.New("Preview Already Exists: " + dstFile)
}

// Get the location of what we expect a preview to be called for a filename modified by the original
// content type.
func GetPreviewPathDestination(filename string, dstPath string, contentType string) string {
    dstFilename := filename
    if strings.Contains(contentType, "video") {
        // The image library for video previews sets the output by ext (not a video)
        dstFilename +=  ".png"  
    }
	return filepath.Join(dstPath, dstFilename)
}

// Break this down better using just a file object?
func CreateImagePreview(srcImg *os.File, dstFile string, contentType string) (string, error) {
	// Attempt to create previews for different media types
	var img image.Image
	var dErr error

	// HOW TO DO THIS in a sane extensible fashion?
	if contentType == "image/png" {
		img, dErr = png.Decode(srcImg)
	} else if contentType == "image/jpeg" {
		img, dErr = jpeg.Decode(srcImg)
    } else {
		log.Printf("No provided method for this file type %s", contentType)
		fname, _ := srcImg.Stat()
		return "", errors.New("Cannot handle type for file: " + fname.Name())
	}
	if dErr != nil {
		log.Fatal(dErr)
	}

	// Now creat the preview image
	dstImg := resize.Resize(640, 0, img, resize.Lanczos3)
	previewImg, errCreate := os.Create(dstFile)
	if errCreate != nil {
		log.Fatal(errCreate)
		return "Error" + dstFile, errCreate
	}

	// All previews should then be jpeg (change file extensioni?)
	jpeg.Encode(previewImg, dstImg, nil)
	return dstFile, previewImg.Close()
}

// Make sure dstPath already exists before you call this (MakePreviewPath)
func GetImagePreview(path string, filename string, dstPath string, pIfSize int64) (string, error) {

	// Try and determine the content type (required for doing encoding and decoding)
	contentType, tErr := GetMimeType(path, filename)
	if tErr != nil {
		log.Fatal(tErr)
		return "Could not determine img type", tErr
	}

    // Determine the preview based on content type 
	dstFile, dErr := ErrorOnPreviewExists(filename, dstPath, contentType)
	if dErr != nil {
        log.Printf("The preview image already exists %s", dstFile)
		return dstFile, dErr
	}

	// The file we are going to check about making a preview of
	fqFile := filepath.Join(path, filename)
	srcImg, fErr := os.Open(fqFile)
	if fErr != nil {
		log.Fatal(fErr)
		return "Error Generating Preview", fErr
	}
	defer srcImg.Close()

	// Check to see if the image is ACTUALLY over a certain size to be worth previewing
    if strings.Contains(contentType, "video") {
        return CreateVideoPreview(fqFile, dstFile, contentType)
    }
	if strings.Contains(contentType, "image") && ShouldCreatePreview(srcImg, pIfSize) == true {
		return CreateImagePreview(srcImg, dstFile, contentType)
	}
    // No Preview is required
	return "", nil
}

// TODO: Determine if the gif preview just works.
// TODO: Determine how the heck to check length and output a composite or a few screens.
func ReadFrameAsJpeg(inFileName string, frameNum int) io.Reader {
	buf := bytes.NewBuffer(nil)
	err := ffmpeg.Input(inFileName).
		Filter("select", ffmpeg.Args{fmt.Sprintf("gte(n,%d)", frameNum)}).
		Output("pipe:", ffmpeg.KwArgs{"vframes": 1, "format": "image2", "vcodec": "mjpeg"}).
		WithOutput(buf, os.Stdout).
		Run()
	if err != nil {
		panic(err)
	}
	return buf
}

// TODO: The preview it creates is pretty large, enable lossy compression
func CreateVideoPreview(srcFile string, dstFile string, contentType string) (string, error) {
    reader := ReadFrameAsJpeg(srcFile, 2)
    img, err := imaging.Decode(reader)
    if err != nil {
        log.Fatalf("Failed to decode the image from the processing %s\n", err)
        return "", err
    }

    // TODO: Make it so the 640 is a config setting
    resizedImg := imaging.Resize(img, 640, 0, imaging.Lanczos)
    err = imaging.Save(resizedImg, dstFile)

    if err != nil {
        log.Fatalf("Could not save the image %s with error %s\n", dstFile, err)
        return "", err
    }
    return dstFile, err
}

// What is up with gjson?
func GetTotalVideoLength(srcFile string) float64 {
	a, err := ffmpeg.Probe(srcFile)
	if err != nil {
		panic(err)
	}
	return gjson.Get(a, "format.duration").Float()
}

// Oh gods this is a lot https://engineering.giphy.com/how-to-make-gifs-with-ffmpeg/
// There is a way to setup a palette file (maybe not via the lib)
func CreateGifVideo(srcFile string, dstFile string) (string, error) {
    // Check the total time of the file
    // Calculate a framerate that will work
    // Calculate a max -t based on frame + total time
    // Base it on config ?
    // Produce a gif if size > X
    total := GetTotalVideoLength(srcFile)
    framerate := "0.5"
    vframes := 10
    if int(2 * total) < vframes {
        vframes = 5 
    }
    if total > (60 * 5) {
        vframes = 100
        framerate = fmt.Sprintf("%f", (total / float64(vframes)))
    }
    time_to_encode := fmt.Sprintf("%f", total - 3)

    framerate = "0.5"
    err := ffmpeg.Input(srcFile, ffmpeg.KwArgs{"ss": "2"}).
        Output(dstFile, ffmpeg.KwArgs{
            "s": "320x240",
            "pix_fmt": "yuvj422p",
            "t": time_to_encode,
            "vframes": vframes,
            "r": framerate,
        }).OverWriteOutput().Run()
        
    if err != nil {
        log.Fatalf("Failed to create the gif output %s\n with err: %s\n", dstFile, err)
    }
    return dstFile, err
}
