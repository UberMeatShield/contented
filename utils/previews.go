package utils

import (
	"errors"
	"image"
	"log"
	"os"
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

//  No need to check
func PreviewExists(filename string, dstPath string) (string, error) {
	// Does the preview already exist, return that
	dstFile := filepath.Join(dstPath, filename)
	if _, e_err := os.Stat(dstFile); os.IsNotExist(e_err) {
		return dstFile, nil
	}
	return "", errors.New("Preview Already Exists: " + dstFile)
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

	// HMMM (if exists do not do anything)
	dstFile, dErr := PreviewExists(filename, dstPath)
	if dErr != nil {
		return dstFile, dErr
	}

	// Try and determine the content type (required for doing encoding and decoding)
	contentType, tErr := GetMimeType(path, filename)
	if tErr != nil {
		log.Fatal(tErr)
		return "Could not determine img type", tErr
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
	if ShouldCreatePreview(srcImg, pIfSize) == true {
		return CreateImagePreview(srcImg, dstFile, contentType)
	}
    // No Preview is required
	return "", nil
}


// TODO: Validate if this works
func ExampleReadFrameAsJpeg(inFileName string, frameNum int) io.Reader {
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

func jpeg_from_movie(movie_path string) {
    reader := ExampleReadFrameAsJpeg("./sample_data/in1.mp4", 5)
    img, err := imaging.Decode(reader)
    if err != nil {
        log.Fatal(err)
    }
    err = imaging.Save(img, "./sample_data/out1.jpeg")
    if err != nil {
        log.Fatal(err)
    }
}

func create_gif_from_video(input_file_fq_path string, output_file_fq_path string) error{
    err := ffmpeg.Input("./sample_data/in1.mp4", ffmpeg.KwArgs{"ss": "1"}).
        Output("./sample_data/out1.gif", ffmpeg.KwArgs{"s": "320x240", "pix_fmt": "rgb24", "t": "3", "r": "3"}).
        OverWriteOutput().Run()
    return err
}
