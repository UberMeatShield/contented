package utils

import (
  "os"
  "fmt"
  "testing"
  "path/filepath"
  "github.com/tidwall/gjson"
  ffmpeg "github.com/u2takey/ffmpeg-go"
)


func nukeFile(dstFile string) {
    // This test should be marked as slow probably
    if _, err := os.Stat(dstFile); !os.IsNotExist(err) {
        os.Remove(dstFile)
    }
}


func Test_VideoEncoding(t *testing.T) {
    srcDir, _, testFile := Get_VideoAndSetupPaths()
    srcFile := filepath.Join(srcDir, testFile)
    dstFile := fmt.Sprintf("%s.%s", srcFile, "[h256].mp4")

    nukeFile(dstFile)
    // Check if the dstFile exists and delete it if it does.
    msg, err := ConvertVideoToH256(srcFile, dstFile)
    if err != nil {
        t.Errorf("Failed to convert %s", err)
    }
    if msg == "" {
        t.Errorf("We should have a success message %s", err)
    }
    if _, err := os.Stat(dstFile); os.IsNotExist(err) {
        t.Errorf("We should have a file called %s", dstFile)
        t.Fail()
    }

    cfg := GetCfg()
    vidInfo, err := ffmpeg.Probe(dstFile)

    totalTimeSrc, _, _ := GetTotalVideoLength(srcFile)
    totalTimeDst, _, _ := GetTotalVideoLength(dstFile)
    if totalTimeSrc != totalTimeDst {
        t.Errorf("Failed to create a valid output times are different %f vs %f", totalTimeSrc, totalTimeDst)
    }
    // Cleanup after the test
    nukeFile(dstFile)

    codecName := gjson.Get(vidInfo, "streams.0.codec_name").String()
    if codecName != "hevc" {
        t.Errorf("Failed to encode with %s dstFile: %s was not hevc but %s", cfg.CodecForConversion, dstFile, codecName)
        t.Fail()
    }

    // Test should check the ffmpeg.Probe of both files and check length
    // Test should validate the new file uses a new codec
    // There should be an option to ignore vs nuke the test file
}
