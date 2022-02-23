package utils

import (
	"os"
//    "errors"
	"path/filepath"
	"testing"
 //   "contented/models"
	"github.com/gobuffalo/envy"
)

// Possibly make this some sort of global test helper function (harder to do in GoLang?)
func Test_JpegPreview(t *testing.T) {
	var testDir, _ = envy.MustGet("DIR")
	srcDir := filepath.Join(testDir, "dir1")
	dstDir := GetPreviewDst(srcDir)
	testFile := "this_is_jp_eg"

	ResetPreviewDir(dstDir)

	var size int64 = 20000

	checkFile, _ := os.Open(filepath.Join(srcDir, testFile))
	if ShouldCreatePreview(checkFile, size) == true {
		st, _ := checkFile.Stat()
		t.Errorf("Error, this should be too small file size was: %d", st.Size())
	}

	expectNoPreview, err := GetImagePreview(srcDir, testFile, dstDir, size)
	if err != nil {
		t.Errorf("Failed to get a preview %v", err)
	}
	if expectNoPreview != testFile && expectNoPreview != "" {
		t.Errorf("File too small for psize found  %s and expected %s", expectNoPreview, testFile)
	}

	pLoc, err := GetImagePreview(srcDir, testFile, dstDir, 10)
	if err != nil {
		t.Errorf("Error occurred creating preview %v", err)
	}
	expectDst := filepath.Join(dstDir, testFile)
	if expectDst != pLoc {
		t.Errorf("Failed to find the expected file location %s had %s", expectDst, pLoc)
	}
}

// Does it work when there is a png
func Test_PngPreview(t *testing.T) {
	var testDir, _ = envy.MustGet("DIR")
	srcDir := filepath.Join(testDir, "dir1")
	dstDir := GetPreviewDst(srcDir)
	testFile := "this_is_p_ng"

	// Add a before each to nuke the dstDir and create it
	ResetPreviewDir(dstDir)
	pLoc, err := GetImagePreview(srcDir, testFile, dstDir, 10)
	if err != nil {
		t.Errorf("Failed to get a preview %v", err)
	}
	expectDst := filepath.Join(dstDir, testFile)
	if expectDst != pLoc {
		t.Errorf("Failed to find the expected location %s was %s", expectDst, pLoc)
	}
}

// We know this file is 10.08 seconds long
func Test_VideoLength(t *testing.T) {
	var testDir, _ = envy.MustGet("DIR")
	srcDir := filepath.Join(testDir, "dir2")
	testFile := "donut.mp4"

    srcFile := filepath.Join(srcDir, testFile)
    checkLen, fps, err := GetTotalVideoLength(srcFile)
    if err != nil {
        t.Errorf("Failed to load length %s", err)
    }
    if (checkLen != 10.08) {
        t.Errorf("Could not get the length correctly %f", checkLen)
    }
    if fps != 30 {
        t.Errorf("Couldn't get the right FPS from the video %d", fps)
    }
}

// Test MultiScreen
func Test_MultiScreen(t *testing.T) {
	var testDir, _ = envy.MustGet("DIR")
	srcDir := filepath.Join(testDir, "dir2")
	testFile := "donut.mp4"
    dstDir := GetPreviewDst(srcDir)
	ResetPreviewDir(dstDir)

    destFile := filepath.Join(dstDir, "donut.png")
    srcFile := filepath.Join(srcDir, testFile)
    dstFile, err := CreateScreensFromVideo(srcFile, destFile)
    if err != nil {
        t.Errorf("Failed to create a set of screens %s", err)
    }
    if dstFile == "" {
        t.Errorf("Did not get a valid destination file.")
    }

    // TODO: Check that the files we expect are ACTUALLY created
}


func Test_BrokenImagePreview(t *testing.T) {
	var testDir, _ = envy.MustGet("DIR")
	srcDir := filepath.Join(testDir, "dir3")
	dstDir := GetPreviewDst(srcDir)
	testFile := "nature-corrupted-free-use.jpg"
	ResetPreviewDir(dstDir)

    // TODO: This needs to be made into a better place around previews
	pLoc, err := GetImagePreview(srcDir, testFile, dstDir, 0)
    if err == nil {
        t.Errorf("This file should definitely cause an error")
    }
    if pLoc != "" {
        t.Errorf("And it absolutely does not have a preview")
    }
}

// Makes it so that the preview is generated
func Test_VideoPreviewPNG(t *testing.T) {
	var testDir, _ = envy.MustGet("DIR")
	srcDir := filepath.Join(testDir, "dir2")
	dstDir := GetPreviewDst(srcDir)
	testFile := "donut.mp4"

	// Add a before each to nuke the dstDir and create it
	ResetPreviewDir(dstDir)
    expectDst, dErr := ErrorOnPreviewExists(testFile, dstDir, "video/hack")
    if dErr != nil {
        t.Errorf("The dest file already exists %s\n", expectDst)
    }

	pLoc, err := GetImagePreview(srcDir, testFile, dstDir, 10)
	if err != nil {
		t.Errorf("Failed to get Video preview %v", err)
	}
	if expectDst != pLoc {
		t.Errorf("Failed to find the expected location %s was %s", expectDst, pLoc)
	}
    // TODO: Figure out sime sizing constraints
    _, no_file_err := os.Stat(pLoc); 
    if no_file_err != nil {
        t.Errorf("We had no error but the file is not on disk %s", pLoc)
    }
    // TODO: Should probably check the size as well
}


func Test_FileExistsError(t *testing.T) {
	var testDir, _ = envy.MustGet("DIR")
	srcDir := filepath.Join(testDir, "dir1")
	dstDir := GetPreviewDst(srcDir)
    knownFile := "0_LargeScreen.png"

	ResetPreviewDir(dstDir)
    fqPath := GetPreviewPathDestination(knownFile, dstDir, "image/png")
    f, err := os.Create(fqPath)
    if err != nil {
        t.Errorf("Could not create the file at %s", fqPath)
    }
    _, wErr := f.WriteString("Now something exists in the file")
    if wErr != nil {
        t.Errorf("Could not write to the file at %s", fqPath)
    }
    f.Sync()

    dstCheck, exists := ErrorOnPreviewExists(knownFile, dstDir, "image/png")
    if exists == nil {
        t.Errorf("This file should exist now, so we should have a preview conflict")
    }
    if dstCheck != fqPath {
        t.Errorf("The destination check %s was not == to what we wrote in the test %s", dstCheck, fqPath)
    }
    //Write minimal content to file
}

func Test_VideoPreviewGif(t *testing.T) {
	var testDir, _ = envy.MustGet("DIR")
	srcDir := filepath.Join(testDir, "dir2")
	dstDir := GetPreviewDst(srcDir)
	testFile := "donut.mp4"

	ResetPreviewDir(dstDir)
    expectDst, dErr := ErrorOnPreviewExists(testFile, dstDir, "video/hack")
    if dErr != nil {
        t.Errorf("The dest file already exists %s\n", expectDst)
    }
    
    vidFile := filepath.Join(srcDir, testFile)
    destFile := filepath.Join(dstDir, testFile + ".gif")

    _, err := CreateGifFromVideo(vidFile, destFile)
    if err != nil {
        t.Errorf("Failed to create a gif preview %s", err)
    }
    _, no_file_err := os.Stat(destFile); 
    if no_file_err != nil {
        t.Errorf("We had no error but the file is not on disk %s", destFile)
    }
    // TODO: Should probably check the size as well
}

func Test_ShouldCreate(t *testing.T) {
    var testDir, _ = envy.MustGet("DIR")
	srcDir := filepath.Join(testDir, "dir1")
	testFile := "this_is_p_ng"

	filename := filepath.Join(srcDir, testFile)
	srcImg, fErr := os.Open(filename)

	if fErr != nil {
		t.Errorf("This file cannot be opened %s with err %s", filename, fErr)
	}

	preview_no := ShouldCreatePreview(srcImg, 30000)
	if preview_no != false {
		t.Errorf("This preview should not be created")
	}
	preview_yes := ShouldCreatePreview(srcImg, 1000)
	if preview_yes != true {
		t.Errorf("At this size it should create a preview")
	}
}
