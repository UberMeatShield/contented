package utils

import (
  "fmt"
  "testing"
  "path/filepath"
)

func Test_VideoEncoding(t *testing.T) {
    srcDir, _, testFile := Get_VideoAndSetupPaths()
    srcFile := filepath.Join(srcDir, testFile)
    dstFile := fmt.Sprintf("%s.%s", srcFile, "[h256].mp4")

    // Check if the dstFile exists and delete it if it does.

    ConvertVideoToH256(srcFile, dstFile)

    // Test should check the ffmpeg.Probe of both files and check length
    // Test should validate the new file uses a new codec
    // There should be an option to ignore vs nuke the test file
    t.Fail()
}
