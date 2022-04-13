package utils

import (
    "os"
    "time"
    "io/ioutil"
    "fmt"
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
    if fps != 25 {
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

    empty_check, _ := ioutil.ReadDir(dstDir)
    if len(empty_check) > 0 {
        t.Errorf("The destination directory was not empty %s", empty_check)
    }

    destFile := filepath.Join(dstDir, "donut.png")
    srcFile := filepath.Join(srcDir, testFile)
    screensSrc, err := CreateScreensFromVideo(srcFile, destFile)
    if err != nil {
        t.Errorf("Failed to create a set of screens %s", err)
    }
    if screensSrc == "" {
        t.Errorf("Did not get a valid destination file.")
    }

    screens_check, _ := ioutil.ReadDir(dstDir)
    expected := 11
    if len(screens_check) != expected {
        t.Errorf("Not enough screens created %d vs expected %d", len(screens_check), expected)
    }

    // TODO: Really need to fix the dest file info
    gifFile, err := CreateGifFromScreens(screensSrc, destFile)
    if err != nil {
        t.Errorf("Failed to create a gif from screens %s", err)
    }

    gifStat, noGif := os.Stat(gifFile)
    if noGif != nil {
        t.Errorf("Did not create a gif from screens %s", gifFile)
    }
    if gifStat.Size() > 702114 {
        t.Errorf("Gif has too much chonk %d", gifStat.Size())
    }
}

// TODO: Make a damn helper for this type of thing
func Test_CreateSeekScreens(t *testing.T) {
    var testDir, _ = envy.MustGet("DIR")
    testFile := "donut.mp4"

    srcDir := filepath.Join(testDir, "dir2")
    srcFile := filepath.Join(srcDir, testFile)

    dstDir := GetPreviewDst(srcDir)
    previewName := filepath.Join(dstDir, testFile)
    ResetPreviewDir(dstDir)

    err := CreateSeekScreen(srcFile, previewName, 10)
    if err != nil {
        t.Errorf("Screen seek failed %s", err)
    }


    // TODO: Need to get a bigger file test
    startMulti := time.Now()
    _, multiErr := CreateSeekScreens(srcFile, previewName)
    if multiErr != nil {
        t.Errorf("Failed creating multiple screens %s", multiErr)
    }
    fmt.Printf("Screen Multi timing %s\n", time.Since(startMulti))
    
    singleScreen := time.Now()
    _, screenErr := CreateScreensFromVideo(srcFile, previewName)
    if screenErr != nil {
        t.Errorf("Couldn't create screens all at once %s", screenErr)
    }
    fmt.Printf("Screen single execution %s\n", time.Since(singleScreen))
}

func Test_CreatePaletteFile(t *testing.T) {
    var testDir, _ = envy.MustGet("DIR")
    srcDir := filepath.Join(testDir, "dir2")
    testFile := "donut.mp4"
    dstDir := GetPreviewDst(srcDir)
    ResetPreviewDir(dstDir)

    previewName := filepath.Join(dstDir, testFile)
    srcFile := filepath.Join(srcDir, testFile)

    paletteFile, err := PaletteGen(srcFile, previewName)
    if err != nil {
        t.Errorf("Couldn't create a palette for %s err %s", srcFile, err)
    }
    palStat, noPal := os.Stat(paletteFile)
    if noPal != nil {
        t.Errorf("Did not create a palette from a movie %s", paletteFile)
    }
    if palStat.Size() <= 0 {
        t.Errorf("The palette was created empty %s", palStat)
    }
    killErr := CleanPaletteFile(paletteFile)
    if killErr != nil {
        t.Errorf("Didn't cleanup the paletteFile %s", killErr)
    }
    _, noPalNow := os.Stat(paletteFile)
    if noPalNow == nil {
        t.Errorf("Now the palette file should be dead")
    }

    // Deny cleanup of other files
    denyErr := CleanPaletteFile(srcFile)
    if denyErr == nil {
        t.Errorf("better restore the donut file somehow it thought it was a palette")
    }
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
    _, noFileErr := os.Stat(pLoc); 
    if noFileErr != nil {
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
    vFile, _ := os.Stat(vidFile)
    destFile := filepath.Join(dstDir, testFile + ".gif")

    _, err := CreateGifFromVideo(vidFile, destFile)
    if err != nil {
        t.Errorf("Failed to create a gif preview %s", err)
    }
    fCheck, noFileErr := os.Stat(destFile); 
    if noFileErr != nil {
        t.Errorf("We had no error but the file is not on disk %s", destFile)
    }
    if fCheck.Size() > vFile.Size()  {
        t.Errorf("Preview was bigger than video %d > %d", fCheck.Size(), vFile.Size())
    }
    ResetPreviewDir(dstDir)
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
