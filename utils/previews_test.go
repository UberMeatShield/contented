package utils

import (
    "os"
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
    pLoc, err := GetImagePreview(srcDir, testFile, dstDir, 10000)
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
    pLoc, err := GetImagePreview(srcDir, testFile, dstDir, 10000)

    if err != nil {
        t.Errorf("Failed to get a preview %v", err)
    }
    expectDst := filepath.Join(dstDir, "preview_" + testFile)
    if expectDst != pLoc {
        t.Errorf("Failed to find the expected location %s was %s", expectDst, pLoc)
    }
}

func Test_ShouldCreate(t *testing.T) {
    srcDir := filepath.Join(testDir, "dir1")
    testFile := "this_is_p_ng"

    filename := filepath.Join(srcDir, testFile)
    srcImg, fErr := os.Open(filename)

    if fErr != nil {
        t.Errorf("This file cannot be opened %s with err %s", filename, fErr)
    }

    preview := ShouldCreatePreview(srcImg, 3000)
    if preview == true {
        t.Errorf("This preview should not be created")
    }
    preview_yes := ShouldCreatePreview(srcImg, 1000)
    if preview_yes != true {
        t.Errorf("At this size it should create a preview")
    }
}
