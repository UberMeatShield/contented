package utils

import (
  "os"
  "fmt"
  "testing"
  "strings"
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
    checkMsg, err := ConvertVideoToH256(dstFile, dstFile + "FAIL")
    if !strings.Contains(checkMsg, "ignored because it matched") || err != nil {
        t.Errorf("This should be encoded as hevc and shouldn't work %s err: %s", checkMsg, err)
    }
    nukeFile(dstFile)
    nukeFile(dstFile + "FAIL")
}


func Test_VideoEncodingNotMatching(t *testing.T) {
    srcDir, _, testFile := Get_VideoAndSetupPaths()
    srcFile := filepath.Join(srcDir, testFile)
    dstFile := fmt.Sprintf("%s.%s", srcFile, "[h256].mp4")

    nukeFile(dstFile) // Ensure a previous test fail doesn't leave files
    cfg := GetCfg()
    cfg.CodecsToConvert = "windows_trash|quicktime"  // Shouldn't match
    SetCfg(*cfg)

    checkMsg, checkErr := ConvertVideoToH256(srcFile, dstFile)
    if _, err := os.Stat(dstFile); !os.IsNotExist(err) {
        t.Errorf("We should NOT have a file called %s", dstFile)
    }
    if checkErr != nil || !strings.Contains(checkMsg, "Not on the conversion list") {
        t.Errorf("It should not try and convert a codec that doesn't match")
    }
}