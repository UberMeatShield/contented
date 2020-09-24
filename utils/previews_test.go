package utils

import (
	"path/filepath"
	"testing"
	"github.com/gobuffalo/envy"
)


// func Test Create preview path...

// Possibly make this some sort of global test helper function (harder to do in GoLang?)
func TestJpegPreview(t *testing.T) {

    var testDir, _ = envy.MustGet("DIR")
    srcDir := filepath.Join(testDir, "dir1")
    dstDir := filepath.Join(srcDir, "previews_dir")
    testFile := "this_is_jp_eg"

    MakePreviewPath(dstDir)
    pLoc, err := GetImagePreview(srcDir, testFile, dstDir)

    if err != nil {
        t.Errorf("Failed to get a preview %v", err)
    }

    expectDst := filepath.Join(dstDir, "preview_" + testFile)
    if expectDst != pLoc {
        t.Errorf("Failed to find the expected file location %s was %s", expectDst, pLoc)
    }
}
