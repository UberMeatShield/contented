package utils

/**
*  These are helper functions around creating Content preview information for large images
* and video content.   It can be configured to generate a single image or a gif for a video.
 */
import (
	"bytes"
	"contented/pkg/config"
	"contented/pkg/models"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/nfnt/resize"
	"github.com/tidwall/gjson"
	ffmpeg "github.com/u2takey/ffmpeg-go"
)

// Used in the case of async processing when creating Preview results
type PreviewResult struct {
	C_ID    int64
	MC_ID   int64
	Preview string
	Err     error
}

type PreviewRequest struct {
	C    *models.Container
	Mc   *models.Content
	Out  chan PreviewResult
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
	if strings.Contains(dstDir, config.PREVIEW_DIRECTORY) {
		dstDir = filepath.Join(dstDir, "/")
		os.RemoveAll(dstDir)
	}
	return MakePreviewPath(dstDir)
}

// Check the file status and if over a size create a preview.
func ShouldCreatePreview(f *os.File, fsize int64) bool {
	finfo, err := f.Stat()
	if err != nil {
		log.Printf("Error determining file stat for %s", err)
		return false
	}
	if finfo.Size() > fsize {
		return true
	}
	return false
}

func FileOverSize(srcFile string, fsize int64) bool {
	finfo, e_err := os.Stat(srcFile)
	if e_err == nil {
		if finfo.Size() > fsize {
			return true
		}
	}
	return false
}

// TODO: make the preview directory name configurable
// Make it use the container Path instead of the name?
func GetPreviewDst(fqDir string) string {
	return filepath.Join(fqDir, config.PREVIEW_DIRECTORY)
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
		previewType := config.GetCfg().PreviewVideoType
		if previewType == "screens" {
			dstFilename += ".webp"
		} else {
			dstFilename += ("." + previewType)
		}
	}
	return filepath.Join(dstPath, dstFilename)
}

// Break this down better using just a file object?
func CreateImagePreview(srcImg *os.File, dstFile string, contentType string) (string, error) {
	// Attempt to create previews for different content types
	var img image.Image
	var dErr error

	// HOW TO DO THIS in a sane extensible fashion?
	if contentType == "image/png" {
		img, dErr = png.Decode(srcImg)
	} else if contentType == "image/jpeg" {
		img, dErr = jpeg.Decode(srcImg)
	} else if contentType == "image/webp" || contentType == "image/gif" {
		log.Printf("No provided method for this file type %s", contentType)
		return "", nil
	} else {
		log.Printf("No provided method for this file type %s", contentType)
		fname, _ := srcImg.Stat()
		return "", fmt.Errorf("cannot handle preview for file: %s contentType %s", fname.Name(), contentType)
	}
	if dErr != nil {
		log.Printf("Failed to determine image type %s for %s", dstFile, dErr)
		return "", dErr
	}

	// Now creat the preview image
	dstImg := resize.Resize(640, 0, img, resize.Lanczos3)
	previewImg, errCreate := os.Create(dstFile)
	if errCreate != nil {
		log.Printf("Failed to create a preview %s for %s img", errCreate, dstFile)
		return "", errCreate
	}

	// All previews should then be jpeg (change file extension)?
	jpeg.Encode(previewImg, dstImg, nil)
	return dstFile, previewImg.Close()
}

func ClearContainerPreviews(c *models.Container) error {
	dst := GetContainerPreviewDst(c)
	if _, err := os.Stat(dst); os.IsNotExist(err) {
		return nil
	}
	r_err := os.RemoveAll(dst)
	if r_err != nil {
		log.Fatal(r_err)
		return r_err
	}
	return nil
}

// Create unit test
func CleanPaletteFile(paletteFile string) error {
	if _, err := os.Stat(paletteFile); os.IsNotExist(err) {
		return nil
	}
	// Not perfect but "good enough"
	if strings.Contains(paletteFile, "palette") && strings.Contains(paletteFile, config.PREVIEW_DIRECTORY) {
		os.Remove(paletteFile)
	} else {
		return errors.New("Unwilling to remove non-paletteFile: " + paletteFile)
	}
	return nil
}

// Strip off the PNG, we are just going to dump out some jpegs and this format works for
// ffmpeg outputs.  Note the output pattern has to work for ffmpeg import params so it
// has some pretty odd restrictions.
func GetScreensOutputPattern(dstFile string) string {
	stripExtension := regexp.MustCompile(".png$|.gif$|.webp$")
	dstFile = stripExtension.ReplaceAllString(dstFile, "")
	cfg := config.GetCfg()
	if cfg.PreviewVideoType == "screens" {
		return fmt.Sprintf("%s%s", dstFile, ".screens.%03dss%05d.jpg")
	} else {
		return fmt.Sprintf("%s%s", dstFile, ".screens.%03dss00000.jpg")
	}
}

// Holy crap this is terrible, at some point the default must have been swapped to sequence
// in ffmpeg from glob_sequence so the same matchers just don't work right.
func GetScreensOutputGlob(dstFile string) string {
	stripExtension := regexp.MustCompile(".png$|.gif$|.webp$")
	dstFile = stripExtension.ReplaceAllString(dstFile, "")
	dstFile = regexp.QuoteMeta(dstFile)
	// The destination filename must be properly escaped for the glob pattern
	// check for (_ [] & etc).  Create unit test around this
	return fmt.Sprintf("%s%s", dstFile, ".screens.*ss*.jpg")
}

// Used to search for a matched screen.   I am using ss to denote that the screen is at
// a specific time, this is the most common case but I haven't found a clever way to have
// smaller video files set their second time using a sel(n, framemod) format.
func GetScreensMatcherRE(dstFile string) (*regexp.Regexp, error) {
	stripExtension := regexp.MustCompile(".png$|.jpeg$|.jpg$")
	dstFile = stripExtension.ReplaceAllString(dstFile, "")
	dstFile = regexp.QuoteMeta(dstFile)

	// Check if there is a screens option and modify the screen time
	cfg := config.GetCfg()

	// This can be changed to use ffmpeg -pattern_type glob -i 'name.ss*.jpg' which is BETTER on linux
	// but seemingly would never work on windows which is annoying.
	if cfg.PreviewVideoType == "screens" {
		return regexp.Compile(fmt.Sprintf("%s%s", dstFile, ".screens.[0-9]+ss[0-9]+.jpg"))
	} else {
		return regexp.Compile(fmt.Sprintf("%s%s", dstFile, ".screens.[0-9]+.jpg"))
	}
}

// TODO: Move to utils or make it wrapped for some reason?
func GetContainerPreviewDst(c *models.Container) string {

	// Might need to make this use cfg directory location as a CENTRAL
	// Preview location somehow.
	return GetPreviewDst(c.GetFqPath())
}

// Make sure dstPath already exists before you call this (MakePreviewPath)
func GetImagePreview(path string, filename string, dstPath string, pIfSize int64) (string, error) {
	// Try and determine the content type (required for doing encoding and decoding)
	contentType, tErr := GetMimeType(path, filename)
	if tErr != nil {
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
		log.Printf("Could not open %s err %s", fqFile, fErr)
		return "Error opening file to to create preview", fErr
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

// Hmmm, how do I make it load the right type?
func CreateVideoPreview(srcFile string, dstFile string, contentType string) (string, error) {
	// Split based on the environment variable () - Need to set the location
	// And probably the name of the thing better based on the type.
	cfg := config.GetCfg()

	if cfg.PreviewVideoType == "gif" {
		return CreateGifFromVideo(srcFile, dstFile)
	} else if cfg.PreviewVideoType == "screens" {
		// This creates screens and also a webp file
		return CreateWebpFromVideo(srcFile, dstFile)
	} else {
		return CreatePngFromVideo(srcFile, dstFile)
	}
}

/* https://ottverse.com/how-to-create-gif-from-images-using-ffmpeg/
* https://ottverse.com/thumbnails-screenshots-using-ffmpeg/#ScreenshotThumbnail_every_10_seconds
* ffmpeg -i donut.mp4 -vf "select='not(mod(n,10))',setpts='N/(30*TB)'" -f image2 thumbnail%03d.jpg
*
* Creates a set of preview files in the preview directory from the source image.
 */
func CreateScreensFromVideo(srcFile string, dstFile string) (string, error) {
	cfg := config.GetCfg()
	return CreateScreensFromVideoSized(srcFile, dstFile, cfg.ScreensOverSize)
}

func CreateScreensFromVideoSized(srcFile string, dstFile string, previewScreensOverSize int64) (string, error) {

	cfg := config.GetCfg()
	frameOffsetSeconds := cfg.PreviewFirstScreenOffset
	totalScreens := cfg.PreviewNumberOfScreens

	if FileOverSize(srcFile, previewScreensOverSize) {
		log.Printf("File size is large for %s using SEEK screen", srcFile)

		// Currently I get a list of screens but don't do anything with it.
		_, screenFmt, err := CreateSeekScreens(srcFile, dstFile, totalScreens, frameOffsetSeconds)
		return screenFmt, err
	} else {
		log.Printf("File size is small %s using SELECT filter", srcFile)
		return CreateSelectFilterScreens(srcFile, dstFile, totalScreens, frameOffsetSeconds)
	}
}

// Create screens as needed a palette file and return the image
func CreateWebpFromVideo(srcFile string, dstFile string) (string, error) {
	// screensSrc, err := CreateScreensFromVideo(srcFile, dstFile)
	_, err := CreateScreensFromVideo(srcFile, dstFile)
	if err != nil {
		log.Printf("Couldn't create screens for the %s err: %s", srcFile, err)
		return "", err
	}
	globMatch := GetScreensOutputGlob(dstFile)
	return CreateWebpFromScreens(globMatch, dstFile)
}

func CreateSelectFilterScreens(srcFile string, dstFile string, maxScreens int, frameOffsetSeconds int) (string, error) {
	totalTime, fps, err := GetTotalVideoLength(srcFile)
	if err != nil {
		log.Printf("Error creating screens for %s err: %s", srcFile, err)
	}
	msg := fmt.Sprintf("%s Total time was %f with %d as the fps max screens %d frame offset %d", srcFile, totalTime, fps, maxScreens, frameOffsetSeconds)
	log.Print(msg)
	if int(fps) == 0 || int(totalTime) == 0 {
		return "", errors.New(msg + " Invalid duration or fps")
	}

	totalScreenTime, totalScreens, _ := GetScreenNumber(totalTime, maxScreens, frameOffsetSeconds)
	frameNum := (totalScreenTime * fps) / totalScreens

	screensDst := GetScreensOutputPattern(dstFile)
	filter := fmt.Sprintf("select='not(mod(n,%d))',setpts='N/(30*TB)'", frameNum)
	screenErr := ffmpeg.Input(srcFile, ffmpeg.KwArgs{}).
		Output(screensDst, ffmpeg.KwArgs{"format": "image2", "vf": filter}).
		OverWriteOutput().Run()
	if screenErr != nil {
		log.Printf("Failed to write multiple screens out %s", screenErr)
	}
	// Rename the dstFile with Indexing information (replace.png with info)
	return screensDst, err
}

func GetScreenNumber(totalTime float64, maxScreens int, frameOffsetSeconds int) (int, int, int) {
	totalScreenTime := int(totalTime) - frameOffsetSeconds
	if totalScreenTime <= maxScreens {
		return int(totalTime), int(totalTime), 0
	}
	return totalScreenTime, maxScreens, frameOffsetSeconds
}

// Need to do timing test with this then a timing test with a much bigger file.
// IMPORTANT if this is > 4 it will break ffmpeg finding the screens.
func CreateSeekScreens(srcFile string, dstFile string, maxScreens int, frameOffsetSeconds int) ([]string, string, error) {
	totalTime, fps, err := GetTotalVideoLength(srcFile)
	if err != nil {
		log.Printf("Error creating screens for %s err: %s", srcFile, err)
	}
	msg := fmt.Sprintf("%s Total time was %f with %d as the fps", srcFile, totalTime, fps)
	log.Print(msg)
	if int(totalTime) == 0 {
		return []string{}, "", errors.New(msg + " Invalid duration or fps")
	}

	// This is ugly enough that maybe it should be a method small files cause
	// surprising numbers of problems.
	// For a very short video (testing we don't want to skip or take a lot of screens)
	// or even do a frame skip, so reassign to something more sensible.
	totalScreenTime, totalScreens, frameOffsetSeconds := GetScreenNumber(totalTime, maxScreens, frameOffsetSeconds)
	timeSkip := int(totalScreenTime) / totalScreens
	log.Printf("Setting up screens (%d) with timeSkip (%d) dstFile (%s)", totalScreens, timeSkip, dstFile)

	screenFiles := []string{}
	screenFmt := GetScreensOutputPattern(dstFile)

	// Screen file can be modified to take a second format which is the time skip
	for idx := 0; idx < totalScreens; idx++ {
		ss := (idx * timeSkip) + frameOffsetSeconds
		// screenFile := fmt.Sprintf(screenFmt, ss)
		screenFile := fmt.Sprintf(screenFmt, idx+1, ss)
		// screenFile := fmt.Sprintf(screenFmt, idx)
		err := CreateSeekScreen(srcFile, screenFile, ss)
		if err != nil {
			log.Printf("Error creating a seek screen %s", err)
			break
		} else {
			screenFiles = append(screenFiles, screenFile)
		}
	}
	return screenFiles, screenFmt, err
}

// This can be much faster to do multiple seek screens vs a filter over about a 50mb
// video file.  Then creating a palette and using these screens that makes for smaller webp.
func CreateSeekScreen(srcFile string, dstFile string, screenTime int) error {
	screenErr := ffmpeg.Input(srcFile, ffmpeg.KwArgs{"ss": screenTime}).
		Output(dstFile, ffmpeg.KwArgs{"format": "image2", "vframes": 1}).
		OverWriteOutput().Run()
	return screenErr
}

// Note a src can b either a set of images with a %d00 or a video link
func PaletteGen(paletteSrc string, dstFile string) (string, error) {
	// TODO: Make this into a palette method
	paletteFile := fmt.Sprintf("%s.palette.png", dstFile)

	// A single file will fail if you give it a glob, even if the glob SHOULD match
	paletteArgs := ffmpeg.KwArgs{}
	if strings.Contains(paletteSrc, "*") {
		paletteArgs = ffmpeg.KwArgs{
			"pattern_type": "glob",
		}
	}
	outputArgs := ffmpeg.KwArgs{
		"update":   "true",
		"frames:v": 1,
		"vf":       "palettegen",
	}
	paletteErr := ffmpeg.Input(paletteSrc, paletteArgs).
		Output(paletteFile, outputArgs).
		OverWriteOutput().Run()
	//   OverWriteOutput().ErrorToStdOut().Run()

	if paletteErr != nil {
		log.Printf("Failed to create a palette %s", paletteErr)
		return "", paletteErr
	}
	return paletteFile, nil
}

// This does not seem to be much faster, but the gif/Webp might be a better toggle.
func CreateWebpFromScreens(screensSrc string, dstFile string) (string, error) {
	stripExtension := regexp.MustCompile(".png$")
	dstFile = stripExtension.ReplaceAllString(dstFile, "")

	// Need a function that determines the preview output filename and takes in the config
	// for the preview type name...
	log.Printf("What is the screens %s vs dstFile %s", screensSrc, dstFile)
	paletteFile, palErr := PaletteGen(screensSrc, dstFile)
	if palErr != nil {
		return "", palErr
	}

	// Should scale based on a probe of the size maybe?  No need to make something
	// tiny even smaller. This seems to produce a "decent" output.
	filter := "paletteuse,setpts=25*PTS,scale=iw*.5:ih*.5"
	screenErr := ffmpeg.Input(screensSrc, ffmpeg.KwArgs{
		"pattern_type": "glob",
	}).Output(dstFile, ffmpeg.KwArgs{
		"i":              paletteFile,
		"filter_complex": filter,
		"loop":           0,
	}).OverWriteOutput().Run()

	if screenErr != nil {
		return dstFile, screenErr
	}

	// Probably want to strip the paths but for now this is ok
	if f, err := os.Stat(dstFile); err == nil {
		if f.Size() > 0 {
			return dstFile, nil
		}
		return dstFile, fmt.Errorf("file %s exists on disk but was too small", dstFile)
	} else {
		return dstFile, fmt.Errorf("%s doesn't seem to exist err %s", dstFile, err)
	}
}

func CreatePngFromVideo(srcFile string, dstFile string) (string, error) {
	reader := ReadFrameAsJpeg(srcFile, 20) // Determine how to get a better frame
	img, err := imaging.Decode(reader)
	if err != nil {
		log.Printf("Failed to decode the image from the processing %s\n", err)
		return "", err
	}

	// TODO: Get a full resolution image based on the stream resolution?
	resizedImg := imaging.Resize(img, 640, 0, imaging.Lanczos)
	err = imaging.Save(resizedImg, dstFile)
	if err != nil {
		log.Printf("Could not save the image %s with error %s\n", dstFile, err)
		return "", err
	}
	return dstFile, nil
}

// TODO: Determine if the gif preview just works.  ffmpeg is bloody complicated.
// TODO: Determine how the heck to check length and output a composite or a few screens.
func ReadFrameAsJpeg(inFileName string, frameNum int) io.Reader {
	buf := bytes.NewBuffer(nil)
	err := ffmpeg.Input(inFileName).
		Filter("select", ffmpeg.Args{fmt.Sprintf("gte(n,%d)", frameNum)}).
		Output("pipe:", ffmpeg.KwArgs{"vframes": 1, "format": "image2", "vcodec": "mjpeg", "pix_fmt": "yuvj422p", "update": true}).
		WithOutput(buf, os.Stdout).
		Run()
	if err != nil {
		panic(err)
	}
	return buf
}

/**
 * Use a faster read method instead of trying to grab a frame via gte()
 */
func ReadSeekScreen(srcFile string, screenTime int) (io.Reader, error) {
	buf := bytes.NewBuffer(nil)
	screenErr := ffmpeg.Input(srcFile, ffmpeg.KwArgs{"ss": screenTime}).
		Output("pipe:", ffmpeg.KwArgs{"format": "image2", "vframes": 1, "update": true}).
		WithOutput(buf, os.Stdout).
		Run()
	return buf, screenErr
}

// What is up with gjson vs normal processing (this does seem easier to use)?
// TODO: Consider moving more logic into a ContentHelper (size, rez, probe etc)
// TODO: Rename this as a helper around the Probe
func GetTotalVideoLength(srcFile string) (float64, int, error) {
	vidInfo, err := ffmpeg.Probe(srcFile)
	if err != nil {
		return 0, 0, err
	}
	return GetTotalVideoLengthFromMeta(vidInfo, srcFile)

}

func GetTotalVideoLengthFromMeta(vidInfo string, srcFile string) (float64, int, error) {
	//log.Printf("What the heck %s", vidInfo) could also get out the resolution
	duration := gjson.Get(vidInfo, "format.duration").Float()
	r_frame_rate := gjson.Get(vidInfo, "streams.0.r_frame_rate").String()
	if r_frame_rate == "" {
		log.Printf("Video Information for %s info %s", srcFile, vidInfo)
	}
	// log.Printf("Video Information for %s info %s", srcFile, vidInfo)
	// Note that even the r_frame rate is kinda approximate
	real_frame_re := regexp.MustCompile("/.*")
	fps_str := real_frame_re.ReplaceAllString(r_frame_rate, "")
	fps := 30 // GUESS
	if fps_str != "" {
		fps_parsed, fps_err := strconv.ParseInt(fps_str, 10, 64)
		if fps_err != nil {
			log.Printf("Error determining frame rate %s with vidInfo %s", fps_err, vidInfo)
		} else {
			fps = int(fps_parsed)
		}
	}
	//return duration, fps, resolution, nil
	return duration, fps, nil
}

// Oh gods this is a lot https://engineering.giphy.com/how-to-make-gifs-with-ffmpeg/
// There is a way to setup a palette file (maybe not via the lib)
func CreateGifFromVideo(srcFile string, dstFile string) (string, error) {
	// Check the total time of the file
	// Calculate a framerate that will work
	// Calculate a max -t based on frame + total time
	// Base it on config ?
	// Produce a gif if size > X
	total, _, err := GetTotalVideoLength(srcFile)
	//log.Printf("Video Format %s", vidInfo)
	if err != nil {
		return "", err
	}

	cfg := config.GetCfg()
	framerate := "0.5"
	vframes := cfg.PreviewNumberOfScreens
	filter_v := "setpts=PTS/2"
	if int(2*total) < vframes {
		vframes = 5
	}
	skipSeconds := cfg.PreviewFirstScreenOffset
	if skipSeconds >= int(total) {
		skipSeconds = 2
	}

	// This whole mess makes a relatively decent preview gif for a full movie
	// But I could still probably cut it down for size (or config tweak it)
	if total > (60 * 5) {
		vframes = 30
		speedup := int(total / float64(vframes))
		filter_v = fmt.Sprintf("setpts=PTS/%d", speedup)
		framerate = fmt.Sprintf("%f", (float64(vframes) / (total - 3)))
	}
	time_to_encode := fmt.Sprintf("%f", total-3)
	log.Printf("Gif total time %s framerate %s speedup %s", time_to_encode, framerate, filter_v)

	// Framerate vframes
	gif_err := ffmpeg.Input(srcFile, ffmpeg.KwArgs{"ss": skipSeconds}).
		Output(dstFile, ffmpeg.KwArgs{
			"s":        "640x480",
			"pix_fmt":  "yuvj422p",
			"t":        time_to_encode,
			"vframes":  vframes,
			"r":        framerate,
			"filter:v": filter_v,
		}).OverWriteOutput().Run()
	if gif_err != nil {
		log.Printf("Failed to create the gif output %s\n with err: %s\n", dstFile, gif_err)
	}
	return dstFile, gif_err
}

// This might not need to be a fatal on an error, but is nice for debugging now
// Unit test is in helper_test...
func CreateContentPreview(c *models.Container, mc *models.Content) (string, error) {
	cfg := config.GetCfg()
	cntPath := filepath.Join(c.Path, c.Name)
	dstPath := GetContainerPreviewDst(c)

	dstFqPath, err := GetImagePreview(cntPath, mc.Src, dstPath, cfg.PreviewOverSize)
	if err != nil {
		log.Printf("Failed to create a preview in %s for content %d err: %s", dstPath, mc.ID, err)
		if cfg.PreviewCreateFailIsFatal {
			log.Fatal(err)
		}
	}
	return GetRelativePreviewPath(dstFqPath, cntPath), err
}

func AssignScreensFromSet(c *models.Container, mc *models.Content, maybeScreens *[]os.FileInfo) *models.Screens {
	if !strings.Contains(mc.ContentType, "video") {
		// log.Printf("Content is not of type video, no screens likely")
		return nil
	}
	screenRe, reErr := GetScreensMatcherRE(mc.Src)
	if reErr != nil {
		log.Printf("Error trying to compile re match for %s", mc.Src)
		return nil
	}

	// Could probably just go with FileInfo references
	previewPath := GetPreviewDst(c.GetFqPath())
	// ie: 1000 episodes of One Piece * (15 screens  + 1 webp) in a loop running
	// the regex against them all over and over...
	previewScreens := models.Screens{}

	for idx, fRef := range *maybeScreens {
		name := fRef.Name()
		if screenRe.MatchString(name) {
			ps := models.Screen{
				Path:      previewPath,
				Src:       name,
				ContentID: mc.ID,
				Idx:       idx,
				SizeBytes: fRef.Size(),
			}
			previewScreens = append(previewScreens, ps)
		}
	}
	mc.Screens = previewScreens
	return &previewScreens
}

func GetPotentialScreens(c *models.Container) (*[]os.FileInfo, error) {
	previewPath := GetPreviewDst(c.GetFqPath())

	if _, e_err := os.Stat(previewPath); os.IsNotExist(e_err) {
		empty := []os.FileInfo{}
		return &empty, nil
	}
	dirEntries, err := os.ReadDir(previewPath)
	if err != nil {
		log.Printf("Couldn't list for path %s err %s", previewPath, err)
		return nil, err
	}
	maybeScreens := []os.FileInfo{}
	for _, fRef := range dirEntries {
		if !fRef.IsDir() { // Quick check to ensure screens is in the filename?
			info, _ := fRef.Info()
			maybeScreens = append(maybeScreens, info)
		}
	}
	return &maybeScreens, nil
}

func AssignScreensIfExists(c *models.Container, mc *models.Content) *models.Screens {
	if !strings.Contains(mc.ContentType, "video") {
		// log.Printf("Content is not of type video, no screens likely")
		return nil
	}
	maybeScreens, err := GetPotentialScreens(c)
	if err != nil {
		return nil
	}
	return AssignScreensFromSet(c, mc, maybeScreens)
}

func AssignPreviewIfExists(c *models.Container, mc *models.Content) string {
	// This check is normally to determine if we didn't clear out old previews.
	// For memory only managers it will just consider that a bonus and use the preview.
	previewPath := GetPreviewDst(c.GetFqPath())
	previewFile, exists := ErrorOnPreviewExists(mc.Src, previewPath, mc.ContentType)
	if exists != nil {
		mc.Preview = GetRelativePreviewPath(previewFile, c.GetFqPath())
		log.Printf("Added a preview to content %s", mc.Preview)
	}
	return previewFile
}
