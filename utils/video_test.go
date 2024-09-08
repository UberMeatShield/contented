package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tidwall/gjson"
	ffmpeg "github.com/u2takey/ffmpeg-go"
)

func nukeFile(dstFile string) {
	// This test should be marked as slow probably
	if _, err := os.Stat(dstFile); !os.IsNotExist(err) {
		os.Remove(dstFile)
	}
}

func Test_VideoMeta(t *testing.T) {
	srcDir, _, testFile := Get_VideoAndSetupPaths()
	srcFile := filepath.Join(srcDir, testFile)

	id := int64(666)
	finfo, err := os.Stat(srcFile)
	if err != nil {
		t.Errorf("Failed to stat %s err: %s", srcFile, err)
	}
	c := GetContent(id, finfo, srcDir)
	if c.Corrupt {
		t.Errorf("Failure, file is corrupt %s", c.Meta)
	}
	codec := gjson.Get(c.Meta, "streams.0.codec_name").String()
	if codec != "h264" {
		t.Errorf("Codec format incorrect %s wanted h264 found %s", c.Meta, codec)
	}

}

func Test_VideoEncoding(t *testing.T) {
	srcDir, _, testFile := Get_VideoAndSetupPaths()
	srcFile := filepath.Join(srcDir, testFile)
	dstFile := fmt.Sprintf("%s.%s", srcFile, "[h265].mp4")

	nukeFile(dstFile)
	// Check if the dstFile exists and delete it if it does.
	msg, err, encoded := ConvertVideoToH265(srcFile, dstFile)
	if err != nil {
		t.Errorf("Failed to convert %s", err)
	}
	if msg == "" {
		t.Errorf("We should have a success message %s", err)
	}
	if _, err := os.Stat(dstFile); os.IsNotExist(err) {
		t.Errorf("We should have a file called %s", dstFile)
	}
	if encoded == false {
		t.Errorf("It should have encoded but something went wrong: %s, err: %s", msg, err)
	}

	cfg := GetCfg()
	vidInfo, err := ffmpeg.Probe(dstFile)

	totalTimeSrc, _, _ := GetTotalVideoLength(srcFile)
	totalTimeDst, _, _ := GetTotalVideoLength(dstFile)
	if totalTimeSrc != totalTimeDst {
		t.Errorf("Invalid output times are different %f vs %f", totalTimeSrc, totalTimeDst)
	}
	codecName := gjson.Get(vidInfo, "streams.0.codec_name").String()
	if codecName != "hevc" {
		t.Errorf("Failed encoding %s dstFile: %s was not hevc but %s", cfg.CodecForConversion, dstFile, codecName)
	}

	// Now check if we think the srcFile is a duplicate
	isDuplicate, dupeErr := IsDuplicateVideo(dstFile, srcFile)
	if dupeErr != nil {
		t.Errorf("The srcFile had an error when determining if it was a dupe %s", dupeErr)
	}
	if isDuplicate == false {
		t.Errorf("The srcFile was not detected as a duplicate and it should be a candidate for removal")
	}

	shouldNotEncodeTwice := dstFile + "ShouldNotEncodeAlreadyDone.mp4"
	checkMsg, err, encoded := ConvertVideoToH265(dstFile, shouldNotEncodeTwice)
	if !strings.Contains(checkMsg, "ignored because it matched") || err != nil {
		t.Errorf("This should be encoded as hevc and shouldn't work %s err: %s", checkMsg, err)
	}
	if encoded == true {
		t.Errorf("It should not consider an already encoded file as something converted")
	}
	nukeFile(dstFile)
	nukeFile(shouldNotEncodeTwice) // If it did encode it should error out blow it up anyway
}

func Test_VideoEncodingNotMatching(t *testing.T) {
	srcDir, _, testFile := Get_VideoAndSetupPaths()
	srcFile := filepath.Join(srcDir, testFile)
	dstFile := fmt.Sprintf("%s.%s", srcFile, "[h265].mp4")

	nukeFile(dstFile) // Ensure a previous test fail doesn't leave files
	cfg := GetCfg()
	cfg.CodecsToConvert = "windows_trash|quicktime" // Shouldn't match
	SetCfg(*cfg)

	checkMsg, checkErr, encoded := ConvertVideoToH265(srcFile, dstFile)
	if _, err := os.Stat(dstFile); !os.IsNotExist(err) {
		t.Errorf("We should NOT have a file called %s", dstFile)
	}
	if checkErr != nil || !strings.Contains(checkMsg, "Not on the conversion list") {
		t.Errorf("It should not try and convert a codec that doesn't match")
	}
	if encoded == true {
		t.Errorf("The state set should show we did NOT successfully encode")
	}
}

func Test_VideoImageDiff(t *testing.T) {
	dir := MustGetEnvString("DIR")
	srcDir := filepath.Join(dir, "test_encoding")

	encodedFile := filepath.Join(srcDir, "SampleVideo_1280x720_1mb_h265.mp4")
	duplicateFile := filepath.Join(srcDir, "SampleVideo_1280x720_1mb.mp4")

	isDupe, err := VideoDiffFrames(encodedFile, duplicateFile, 1)
	if err != nil {
		t.Errorf("Failed to get frame with error %s", err)
	}
	if !isDupe {
		t.Errorf("This should be a duplicate file but was not?")
	}
	isDuplicate, noErr := IsDuplicateVideo(encodedFile, duplicateFile)
	if noErr != nil {
		t.Errorf("The files failed to compare %s", noErr)
	}
	if isDuplicate == false {
		t.Errorf("These files should be the same")
	}
}

func Test_VideosAreDifferent(t *testing.T) {
	dir := MustGetEnvString("DIR")
	srcDir := filepath.Join(dir, "test_encoding")

	encodedFile := filepath.Join(srcDir, "SampleVideo_1280x720_1mb_h265.mp4")

	donutDir, _, testFile := Get_VideoAndSetupPaths()
	donutFile := filepath.Join(donutDir, testFile)
	isNotDupe, dupeErr := VideoDiffFrames(encodedFile, donutFile, 3)
	if isNotDupe == true {
		t.Errorf("This is not a duplicate %s with %s", encodedFile, donutFile)
	}
	if dupeErr != nil {
		t.Errorf("There should not be an error %s", dupeErr)
	}
}
