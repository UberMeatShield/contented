package utils

import (
	"path/filepath"
	"testing"
	"github.com/gobuffalo/envy"
)


// func Test Create preview path...

// Possibly make this some sort of global test helper function (harder to do in GoLang?)
func Test_JpegPreview(t *testing.T) {

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

// Does it work when there is a png
func Test_PngPreview(t *testing.T) {
    var testDir, _ = envy.MustGet("DIR")
    srcDir := filepath.Join(testDir, "dir1")
    dstDir := filepath.Join(srcDir, "previews_dir")
    testFile := "this_is_p_ng"

    // Add a before each to nuke the dstDir and create it
    MakePreviewPath(dstDir)
    pLoc, err := GetImagePreview(srcDir, testFile, dstDir)

    if err != nil {
        t.Errorf("Failed to get a preview %v", err)
    }
    expectDst := filepath.Join(dstDir, "preview_" + testFile)
    if expectDst != pLoc {
        t.Errorf("Failed to find the expected location %s was %s", expectDst, pLoc)
    }
}
