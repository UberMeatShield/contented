package utils

/**
 * These functions deal with using ffmpeg to encode video to new formats (h265)
 */
import (
	"contented/pkg/config"
	"contented/pkg/models"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/tidwall/gjson"
	ffmpeg "github.com/u2takey/ffmpeg-go"
	"github.com/vitali-fedulov/images4"
)

// Used in the case of async processing when encoding.
// Pretty much identical to previews but I might have to tweak this a lot
type EncodingResult struct {
	C_ID     int64
	MC_ID    int64
	NewVideo string
	Err      error

	// We should ensure that there are some stats gathered around this.
	InitialSize int64
	EncodedSize int64
}

type EncodingRequest struct {
	C       *models.Container
	Mc      *models.Content
	SrcFile string
	DstFile string
	Out     chan EncodingResult
}

type EncodingWorker struct {
	Id int
	In chan EncodingRequest
}
type EncodingRequests []EncodingRequest
type EncodingResults []EncodingResult

// Returns a message about the video codec, size of it, or if it is invalid.
func IsValidVideo(srcFile string) (string, int64, error, string) {
	srcStat, statErr := os.Stat(srcFile)
	fileSize := int64(0)
	codecName := ""
	if statErr != nil {
		return codecName, fileSize, statErr, ""
	}
	fileSize = int64(srcStat.Size())
	if fileSize == 0 {
		return codecName, fileSize, errors.New(fmt.Sprintf("%s was empty or small %d", srcFile, fileSize)), ""
	}

	filename := filepath.Base(srcFile)
	path := filepath.Dir(srcFile)
	contentType, tErr := GetMimeType(path, filename)
	if tErr != nil {
		return codecName, fileSize, tErr, ""
	}
	if !strings.Contains(contentType, "video") {
		msg := fmt.Sprintf("Not a video file so not converting %s", contentType)
		return codecName, fileSize, errors.New(msg), ""
	}

	vidInfo, probeErr := GetVideoInfo(srcFile)
	if probeErr != nil {
		log.Printf("Couldn't probe %s", srcFile)
		return codecName, fileSize, probeErr, ""
	}
	codecName = gjson.Get(vidInfo, "streams.0.codec_name").String()
	return codecName, fileSize, nil, vidInfo
}

// Expand this into something that can trim down the video info to the little bits we care about
func GetVideoInfo(srcFile string) (string, error) {
	return ffmpeg.Probe(srcFile)
}

/**
 *  This returns a message about what is happening with the encoding (should it do it etc).
 */
func ShouldEncodeVideo(srcFile string, dstFile string) (string, error, bool) {
	codecName, _, videoInvalid, srcInfo := IsValidVideo(srcFile)
	if videoInvalid != nil {
		return fmt.Sprintf("Not valid video %s", videoInvalid), nil, false
	}

	// Check that the converted file doesn't exist
	cfg := config.GetCfg()
	log.Printf("The current src codec %s checking if we should convert to %s", codecName, cfg.CodecForConversion)

	// The CodecForConversion is NOT the name of the thing you convert to...
	if cfg.CodecForConversion == codecName {
		okMsg := fmt.Sprintf("%s Already in the desired codec %s", srcFile, cfg.CodecForConversion)
		return okMsg, nil, false
	}

	// TODO: Config setting where if the filesize is too small we should _NOT_ reencode
	matcher := regexp.MustCompile(cfg.CodecsToConvert)
	if !matcher.MatchString(codecName) {
		ignoreMsg := fmt.Sprintf("%s Not on the conversion list %s", srcFile, cfg.CodecsToConvert)
		return ignoreMsg, nil, false
	}
	ignore := regexp.MustCompile(cfg.CodecsToIgnore)
	if ignore.MatchString(codecName) {
		ignoreMsg := fmt.Sprintf("%s ignored because it matched %s", srcFile, cfg.CodecsToIgnore)
		return ignoreMsg, nil, false
	}

	// Now checks that the video is ACTUALLY proper or at least the same time
	_, statErr := os.Stat(dstFile)
	if !os.IsNotExist(statErr) {
		dstInfo, err := GetVideoInfo(dstFile)
		if err != nil {
			// This should fail a test (might need to dump junk in the file)
			msg := fmt.Sprintf("File %s exists but cannot probe dst video info (likely corrupt so re-encode) err: %s", dstFile, err)
			log.Print(msg)
			return msg, nil, true
		} else {
			// Check the times (the name maybe the same but the content might differ if the time is off then re-encode)
			existsMsg := fmt.Sprintf("Destination file already exists %s checking if it is valid", dstFile)
			log.Print(existsMsg)

			srcDuration := gjson.Get(srcInfo, "format.duration").Float()
			dstDuration := gjson.Get(dstInfo, "format.duration").Float()

			// TODO: could also check we are in the actually requested codec as well.
			if int(srcDuration) == int(dstDuration) { // Within minimal amount length?
				existsMsg = fmt.Sprintf("Done %s exists and has source duration %d", dstFile, int(srcDuration))
				log.Print(existsMsg)
				return existsMsg, nil, false
			} else {
				msg := fmt.Sprintf("%s Existed but did NOT have the same duration src(%f) vs dst(%f)", dstFile, srcDuration, dstDuration)
				log.Print(msg)
				return msg, nil, true
			}
		}
	}
	msg := fmt.Sprintf("File will be converted from %s to %s\nOld File: %s\nNew File: %s", codecName, cfg.CodecForConversion, srcFile, dstFile)
	return msg, nil, true
}

// Just rename to something with the previous extension stripped
func GetVideoConversionName(srcFile string) string {
	path := filepath.Dir(srcFile)
	filename := filepath.Base(srcFile)
	ext := filepath.Ext(filename)

	cfg := config.GetCfg()
	if cfg.EncodingDestination != "" {
		path = cfg.EncodingDestination
	}
	stripExtension := regexp.MustCompile(fmt.Sprintf("%s$", ext))
	extension := fmt.Sprintf("%s.mp4", cfg.EncodingFilenameModifier)
	newFilename := stripExtension.ReplaceAllString(filename, extension)
	return filepath.Join(path, newFilename)
}

// This will check if we should convert the source file then run the ffmpeg converter
// Returns
//   - (msg: string) : What happened in human readable form
//   - (err: error) : did we hit a full error state
//   - (encoded: bool) : Did actual encoding take place vs just 'should not do it (ie: already encoded)'
func ConvertVideoToH265(srcFile string, dstFile string) (string, error, bool) {
	reason, err, shouldConvert := ShouldEncodeVideo(srcFile, dstFile)
	if !shouldConvert {
		log.Printf("Not converting %s", reason)
		return reason, err, shouldConvert
	}

	// If codec in list of conversion codecs, then do it
	cfg := config.GetCfg()
	log.Printf("About to convert %s to codec %s", reason, cfg.CodecForConversion)

	// ffmpeg
	//	-hwaccel cuvid
	//	-hwaccel_output_format cuda
	//	-c:v hevc_nvenc -preset slow -movflags faststart output.mp4
	// Might need a new version of the ffmpeg-go library

	var encode_err error
	if cfg.CodecForConversion == "hevc_nvenc" {
		kwArgs := ffmpeg.KwArgs{"c:v": cfg.CodecForConversion, "tag:v": "hvc1", "preset": "slow", "movflags": "faststart"}

		encode_err = ffmpeg.Input(srcFile).
			Output(dstFile, kwArgs).
			GlobalArgs("-hwaccel", "cuda").
			GlobalArgs("-hwaccel_device", fmt.Sprintf("%d", 0)).
			GlobalArgs("-hwaccel_output_format", "cuda").
			GlobalArgs("-loglevel", "quiet").
			OverWriteOutput().ErrorToStdOut().Run()
	} else {
		encode_err = ffmpeg.Input(srcFile).
			Output(dstFile, ffmpeg.KwArgs{"c:v": cfg.CodecForConversion, "tag:v": "hvc1"}).
			GlobalArgs("-loglevel", "quiet").
			OverWriteOutput().ErrorToStdOut().Run()
	}

	if encode_err != nil {
		log.Printf("Encoding error when actually running ffmpeg  %s", encode_err)
		return "", encode_err, false
	}
	return "Success: " + reason, nil, true
}

/*
 * Given two video files see if they are likely the same video
 */
func IsDuplicateVideo(encodedFile string, dupeFile string) (bool, error) {

	// Probe srcFile
	encodedCodec, encodedSize, encodedErr, encodedMeta := IsValidVideo(encodedFile)
	if encodedErr != nil {
		return false, encodedErr
	}
	encodedDuration := gjson.Get(encodedMeta, "format.duration").Float()
	log.Printf("Src %s had codec %s, size %d and runtime %f", encodedFile, encodedCodec, encodedSize, encodedDuration)

	cfg := config.GetCfg()
	if encodedCodec != cfg.CodecForConversionName {
		msg := fmt.Sprintf("Encoded File %s was not what we want %s", encodedFile, encodedCodec)
		return false, errors.New(msg)
	}

	// Probe dstFile
	dupeCodec, dupeSize, dupeErr, dupeMeta := IsValidVideo(dupeFile)
	if dupeErr != nil {
		return false, dupeErr
	}
	dupeDuration := gjson.Get(dupeMeta, "format.duration").Float()
	log.Printf("Dst %s had codec %s, size %d and runtime %f", dupeFile, dupeCodec, dupeSize, dupeDuration)

	// Within minimal amount length?
	if int(encodedDuration) != int(dupeDuration) {
		log.Printf("Files had different durations %f and %f", encodedDuration, dupeDuration)
		return false, nil
	}
	/*
		_, fpsEncoded, _ := GetTotalVideoLengthFromMeta(encodedMeta, encodedFile)
		_, fpsDupe, _ := GetTotalVideoLengthFromMeta(dupeMeta, dupeFile)
	*/

	// Testing a few frames to see if the files seem to have the same content.
	// We never _REALLY_ know what fps is so be conservative
	testFrames := []float64{1.0, 2.0, 4.0, 5.0, 8.0}
	for _, val := range testFrames {
		screenTime := encodedDuration * (val / 10.0)
		same, err := VideoDiffFrames(encodedFile, dupeFile, int(screenTime))
		if !same || err != nil {
			return same, err
		}
	}
	return true, nil
}

func VideoDiffFrames(encodedFile string, dupeFile string, screenTime int) (bool, error) {
	img1Buffer, _ := ReadSeekScreen(encodedFile, screenTime)
	img1, err1 := imaging.Decode(img1Buffer)
	if err1 != nil {
		log.Printf("Failed to get a jpeg from encoded %s", err1)
		return false, err1
	}
	img2Buffer, _ := ReadSeekScreen(dupeFile, screenTime)
	img2, err2 := imaging.Decode(img2Buffer)
	if err2 != nil {
		log.Printf("Failed to get a jpeg from dupe %s", err2)
		return false, err2
	}
	// I should create N screens at time diff and then
	// Icons are compact hash-like image representations.
	icon1 := images4.Icon(img1)
	icon2 := images4.Icon(img2)

	// Comparison. Images are not used directly.
	// Use func CustomSimilar for different precision.
	if images4.Similar(icon1, icon2) {
		// TODO: We can remove this after a little more experimentation
		// log.Printf("Images are similar at time %d", screenTime)
		return true, nil
	} else {
		// log.Printf("Images are different at time %d", screenTime)
		return false, nil
	}
}
