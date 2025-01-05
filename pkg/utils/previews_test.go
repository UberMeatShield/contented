package utils

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	//    "errors"
	"contented/pkg/config"
	"contented/pkg/models"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Helper for a common block of video test code (duplicated in internals)
func Get_VideoAndSetupPaths() (string, string, string) {
	cfg := config.GetCfgDefaults()

	// The video we use is only 10.08 seconds long.
	cfg.PreviewFirstScreenOffset = 2
	cfg.PreviewNumberOfScreens = 4
	config.SetCfg(cfg)

	testDir := config.MustGetEnvString("DIR")
	srcDir := filepath.Join(testDir, "dir2")
	dstDir := GetPreviewDst(srcDir)
	testFile := "donut_[special( gunk.mp4"

	// Ensure that the preview destination directory is clean
	ResetPreviewDir(dstDir)
	return srcDir, dstDir, testFile
}

// Should probably toss this into internals but cannot because it is all the way down in utils
func WriteScreenFile(dstPath string, fileName string, count int) (string, error) {
	screenName := fmt.Sprintf("%s.screens.00%d.jpg", fileName, count)
	fqPath := filepath.Join(dstPath, screenName)
	f, err := os.Create(fqPath)
	if err != nil {
		return "", err
	}
	_, wErr := f.WriteString("Now something exists in the file")
	if wErr != nil {
		return "", wErr
	}
	return screenName, nil
}

func CreateTestPreviewsContainerDirectory(t *testing.T) (string, string) {
	testDir := config.MustGetEnvString("DIR")
	_, dstDir, _ := Get_VideoAndSetupPaths()

	// Check we can write to the video destination directory (probably not needed)
	ok, err := PathIsOk(dstDir, "", testDir)
	if err != nil {
		t.Errorf("Failed to check path %s", err)
	}
	if !ok {
		t.Errorf("Path was not ok %s", dstDir)
	}

	containerPreviews := filepath.Join(dstDir, "container_previews")
	if _, err := os.Stat(containerPreviews); os.IsNotExist(err) {
		err := os.MkdirAll(containerPreviews, 0755)
		if err != nil {
			t.Fatalf("Failed to create container_previews directory: %v", err)
		}
	}
	ok, err = PathIsOk(containerPreviews, "", testDir)
	if err != nil {
		t.Errorf("Failed to check container path %s", err)
	}
	if !ok {
		t.Errorf("container Path was not ok %s", containerPreviews)
	}
	return containerPreviews, testDir
}

func Test_RemoveFile(t *testing.T) {
	// Create a temp file to test removal
	containerPreviews, testDir := CreateTestPreviewsContainerDirectory(t)
	screenName, err := WriteScreenFile(containerPreviews, "ScreenTestRemove", 1)
	assert.NoError(t, err, "Failed to write screen file: %v", err)

	ok, noErr := RemoveFileIsOk(containerPreviews, screenName, testDir)
	assert.NoError(t, noErr, "This file is ok to remove %s", err)
	assert.True(t, ok, "This is a valid file  to remove %s", screenName)

	bad, err := RemoveFileIsOk("~/.ssh", screenName, testDir)
	assert.Error(t, err, "This file is not ok to remove %s", err)
	assert.False(t, bad, "It was not a valid file %s", "~/.ssh")

	bad, err = RemoveFileIsOk("../../etc/passwd", screenName, testDir)
	assert.Error(t, err, "This file is not ok to remove %s", err)
	assert.False(t, bad, "It was not a valid file %s", "../../etc/passwd")

	bad, err = RemoveFileIsOk(containerPreviews, fmt.Sprintf("../../%s", screenName), testDir)
	assert.Error(t, err, "This file is not ok to remove %s", err)
	log.Printf("Bad %t", bad)
	assert.False(t, bad, "File should not allow going up a directory %s", fmt.Sprintf("./../%s", screenName))

	fqPath := filepath.Join(containerPreviews, screenName)
	_, err = os.Stat(fqPath)
	assert.NoError(t, err, "Failed to stat file %s", err)

	removed, err := RemoveFile(containerPreviews, screenName, testDir)
	assert.NoError(t, err, "Failed to remove file %s", err)
	assert.True(t, removed, "File was removed %s", screenName)

	_, goneErr := os.Stat(fqPath)
	assert.Error(t, goneErr, "File should be deleted %s", goneErr)

	stillSuccess, err := RemoveFileIsOk(containerPreviews, screenName, testDir)
	assert.ErrorIs(t, err, os.ErrNotExist, "This file is should be removed %s", err)
	assert.True(t, stillSuccess, "The file is removed so it is a success %s", fqPath)
}

func Test_ImageMetaLookup(t *testing.T) {
	testDir := config.MustGetEnvString("DIR")
	srcDir := filepath.Join(testDir, "dir2")
	testFile := "typescript_nginx_ci_dir2.png"

	srcFile := filepath.Join(srcDir, testFile)
	meta, corrupt := GetImageMeta(srcFile)
	if corrupt == true {
		t.Errorf("This file should generate acceptable meta %s", meta)
	}

	if !strings.Contains(meta, "600") || !strings.Contains(meta, "400") {
		t.Errorf("We should have valid meta %s", meta)
	}
}

// Check that handling bad inputs behaves in an expected fashion
func Test_BrokenImagePreview(t *testing.T) {
	testDir := config.MustGetEnvString("DIR")
	srcDir := filepath.Join(testDir, "dir3")
	dstDir := GetPreviewDst(srcDir)
	testFile := "nature-corrupted-free-use.jpg"
	ResetPreviewDir(dstDir)

	srcFile := filepath.Join(srcDir, testFile)
	meta, corrupt := GetImageMeta(srcFile)
	if corrupt == false {
		t.Errorf("This should cause an error %s", meta)
	}

	// TODO: This needs to be made into a better place around previews
	pLoc, err := GetImagePreview(srcDir, testFile, dstDir, 0)
	if err == nil {
		t.Errorf("This file should definitely cause an error")
	}
	if pLoc != "" {
		t.Errorf("And it absolutely does not have a preview")
	}
}

// Should it Create a preview based on size of the file
func Test_ShouldCreate(t *testing.T) {
	testDir := config.MustGetEnvString("DIR")
	srcDir := filepath.Join(testDir, "dir1")
	testFile := "this_is_p_ng"

	filename := filepath.Join(srcDir, testFile)
	srcImg, fErr := os.Open(filename)
	if fErr != nil {
		t.Errorf("This file cannot be opened %s with err %s", filename, fErr)
	}
	defer srcImg.Close()

	preview_no := ShouldCreatePreview(srcImg, 30000)
	if preview_no != false {
		t.Errorf("This preview should not be created")
	}
	preview_yes := ShouldCreatePreview(srcImg, 1000)
	if preview_yes != true {
		t.Errorf("At this size it should create a preview")
	}
}

func Test_FileExistsError(t *testing.T) {
	testDir := config.MustGetEnvString("DIR")
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
	// Write minimal content to file?
}

// Possibly make this some sort of global test helper function (harder to do in GoLang?)
func Test_JpegPreview(t *testing.T) {
	testDir := config.MustGetEnvString("DIR")
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
	testDir := config.MustGetEnvString("DIR")
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
	srcDir, _, testFile := Get_VideoAndSetupPaths()

	srcFile := filepath.Join(srcDir, testFile)
	checkLen, fps, err := GetTotalVideoLength(srcFile)
	if err != nil {
		t.Errorf("Failed to load length %s", err)
	}
	if checkLen != 10.08 {
		t.Errorf("Could not get the length correctly %f", checkLen)
	}
	if fps != 25 {
		t.Errorf("Couldn't get the right FPS from the video %d", fps)
	}
}

func Test_WebpFromVideo(t *testing.T) {
	srcDir, dstDir, testFile := Get_VideoAndSetupPaths()
	cfg := config.GetCfgDefaults()
	cfg.ScreensOverSize = 1024
	cfg.PreviewVideoType = "screens"
	config.SetCfg(cfg)

	// It will tack on .webp
	dstFile := GetPreviewPathDestination(testFile, dstDir, "video")
	srcFile := filepath.Join(srcDir, testFile)

	// Uses the cfg size to create incremental screens
	previewFile, err := CreateWebpFromVideo(srcFile, dstFile)
	if err != nil {
		t.Errorf("Couldn't create preview from %s err: %s", srcFile, err)
	}
	webpStat, noWebp := os.Stat(previewFile)
	assert.Nil(t, noWebp, "Did not create a preview from screens %s", previewFile)
	assert.LessOrEqual(t, webpStat.Size(), int64(700*1024), "Webp has too much chonk %d", webpStat.Size())

	// Check that if we use a screens version it will work as a preview
	// using memory storage
	c := &models.Container{
		Path: filepath.Dir(srcDir),
		Name: filepath.Base(srcDir),
	}
	mc := &models.Content{
		ContentType: "video/mp4",
		Src:         testFile,
	}
	screens := AssignScreensIfExists(c, mc)
	assert.Equal(t, 10, len(*screens), "Screens should be the same")

	checkFile := AssignPreviewIfExists(c, mc)
	assert.Equal(t, previewFile, checkFile, "Check not set to Expected\n check(%s) \n previewFile(%s)", checkFile, previewFile)
	assert.Contains(t, checkFile, mc.Preview, "mc.Preview (%s) not contained in check(%s)", mc.Preview, checkFile)
}

func Test_AssignScreensWithEscapeChars(t *testing.T) {
	srcDir, dstDir, _ := Get_VideoAndSetupPaths()

	badFilename := "Bad(a))_something_darkside.mp4"
	_, err := WriteScreenFile(dstDir, badFilename, 1)
	if err != nil {
		t.Errorf("Failed to setup the test screen %s", err)
	}
	WriteScreenFile(dstDir, badFilename, 2)
	WriteScreenFile(dstDir, "ShouldNotMatch", 1)

	c := &models.Container{
		Path: filepath.Dir(srcDir),
		Name: filepath.Base(srcDir),
	}
	files, f_err := GetPotentialScreens(c)
	if f_err != nil {
		t.Errorf("Did not find any screens in the preview dir %s", f_err)
	}
	if len(*files) != 3 {
		t.Errorf("We should have looked up all potential files %s", files)
	}

	mc := &models.Content{
		ContentType: "video/mp4",
		Src:         badFilename,
	}
	ps := AssignScreensIfExists(c, mc)

	if ps == nil {
		t.Errorf("We did not find matching screens")
		return
	}
	if len(*ps) != 2 {
		t.Errorf("We did not find the correct number of files %d", len(*ps))
	}
}

// Test Generating screens using the sampling method vs seeking.
func TestVideoSelectScreens(t *testing.T) {
	srcDir, dstDir, testFile := Get_VideoAndSetupPaths()

	empty_check, _ := os.ReadDir(dstDir)
	if len(empty_check) > 0 {
		t.Errorf("The destination directory was not empty %s", empty_check)
	}

	cfg := config.GetCfgDefaults()
	cfg.PreviewNumberOfScreens = 10
	config.SetCfg(cfg)

	destFile := filepath.Join(dstDir, "donut.mp4.webp")
	srcFile := filepath.Join(srcDir, testFile)
	screensSrc, err := CreateScreensFromVideoSized(srcFile, destFile, 1024*300000)
	if err != nil {
		t.Errorf("Failed to create a set of screens %s", err)
	}
	if screensSrc == "" {
		t.Errorf("Did not get a valid destination file.")
	}
	screens_check, _ := os.ReadDir(dstDir)

	expected := 10
	assert.GreaterOrEqual(t, len(screens_check), expected, fmt.Sprintf("Screens %s", screens_check))

	// TODO: Really need to fix the dest file info
	globMatch := GetScreensOutputGlob(destFile)
	webpFile, err := CreateWebpFromScreens(globMatch, destFile)
	if err != nil {
		t.Errorf("Failed to create preview %s", err)
	}
	webpStat, noWebp := os.Stat(webpFile)
	if noWebp != nil {
		t.Errorf("Did not create a preview from screens %s", webpFile)
	}
	if webpStat.Size() > (700 * 1024) {
		t.Errorf("Webp has too much chonk %d", webpStat.Size())
	}
}

func Test_ScreenNumber(t *testing.T) {
	totalTime := 10.08
	maxScreens := 11
	frameOffset := 2
	totalScreenTime, totalScreens, frameOffset := GetScreenNumber(totalTime, maxScreens, frameOffset)
	if totalScreenTime != 10 {
		t.Errorf("Total screen time was not 10 %d", totalScreenTime)
	}
	if totalScreens != 10 {
		t.Errorf("Total screens was not 10 %d", totalScreens)
	}
	if frameOffset != 0 {
		t.Errorf("Frame offset was not 0 %d", frameOffset)
	}

	shortTime := 5.6
	frameOffset = 1
	totalScreenTime, totalScreens, frameOffset = GetScreenNumber(shortTime, maxScreens, frameOffset)
	if totalScreenTime != 5 {
		t.Errorf("Total screen time was not 5 %d", totalScreenTime)
	}
	if totalScreens != 5 {
		t.Errorf("Total screens was not 5 %d", totalScreens)
	}
	if frameOffset != 0 {
		t.Errorf("Frame offset was not 0 %d", frameOffset)
	}
}

// TODO: Make a damn helper for this type of thing
func Test_VideoCreateSeekScreens(t *testing.T) {
	srcDir, dstDir, testFile := Get_VideoAndSetupPaths()
	// With bigger files ~ 100mb it is much faster to do 10 seek time screens
	// instead of using a single operation.  The small donut file is faster with
	// a single operation with a frame selection.
	cfg := config.GetCfgDefaults()
	cfg.ScreensOverSize = 1024
	cfg.PreviewVideoType = "screens"
	config.SetCfg(cfg)

	previewName := filepath.Join(dstDir, testFile+".webp")
	srcFile := filepath.Join(srcDir, testFile)

	err := CreateSeekScreen(srcFile, previewName+".jpeg", 10)
	if err != nil {
		t.Errorf("Screen seek failed %s", err)
	}

	count := cfg.PreviewNumberOfScreens
	offset := cfg.PreviewFirstScreenOffset
	startMulti := time.Now()
	screens, screenPtrn, multiErr := CreateSeekScreens(srcFile, previewName, count, offset)
	if multiErr != nil {
		t.Errorf("Failed creating multiple screens %s", multiErr)
	}
	fmt.Printf("Screen Multi timing %s\n", time.Since(startMulti))
	assert.Equal(t, 10, len(screens), "Didn't find enough screens %d", len(screens))
	assert.Contains(t, screenPtrn, "screens", "We should have a pattern to match against %s", screenPtrn)

	// Check to ensure you can create a gif from the seek screens
	globMatch := GetScreensOutputGlob(previewName)
	webp, webpErr := CreateWebpFromScreens(globMatch, previewName)
	if webpErr != nil {
		t.Errorf("Failed to create a webp screen collection %s", webpErr)
	}
	webInfo, werr := os.Stat(webp)
	if werr != nil {
		t.Errorf("The webp file doesn't exist at %s pattern %s", webp, screenPtrn)
	}
	size := webInfo.Size()
	if size > (1024*700) || size < 1000 {
		t.Errorf("The webp Preview was either less than 1000 or too big %d", size)
	}

	// TODO: Check file size and determine the faster way to create a gif
	singleScreen := time.Now()
	_, screenErr := CreateScreensFromVideo(srcFile, previewName)
	assert.Nil(t, screenErr, "Couldn't create screens all at once %s", screenErr)
	fmt.Printf("Screen single execution %s\n", time.Since(singleScreen))
}

func Test_VideoCreatePaletteFile(t *testing.T) {
	srcDir, dstDir, testFile := Get_VideoAndSetupPaths()

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

	// Ensure that we can cleanup a file that is a palette
	killErr := CleanPaletteFile(paletteFile)
	if killErr != nil {
		t.Errorf("Didn't cleanup the paletteFile %s", killErr)
	}
	_, noPalNow := os.Stat(paletteFile)
	if noPalNow == nil {
		t.Errorf("Now the palette file should be dead")
	}
	// Deny cleanup of non Palette files
	denyErr := CleanPaletteFile(srcFile)
	if denyErr == nil {
		t.Errorf("better restore the donut file somehow it thought it was a palette")
	}
}

// Makes it so that the preview is generated
func Test_VideoPreviewPNG(t *testing.T) {
	srcDir, dstDir, testFile := Get_VideoAndSetupPaths()

	// Add a before each to nuke the dstDir and create it
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
	_, noFileErr := os.Stat(pLoc)
	if noFileErr != nil {
		t.Errorf("We had no error but the file is not on disk %s", pLoc)
	}
	// TODO: Should probably check the size as well
}

func Test_VideoPreviewGif(t *testing.T) {
	srcDir, dstDir, testFile := Get_VideoAndSetupPaths()

	expectDst, dErr := ErrorOnPreviewExists(testFile, dstDir, "video/hack")
	if dErr != nil {
		t.Errorf("The dest file already exists %s\n", expectDst)
	}

	vidFile := filepath.Join(srcDir, testFile)
	vFile, _ := os.Stat(vidFile)
	destFile := filepath.Join(dstDir, testFile+".gif")

	_, err := CreateGifFromVideo(vidFile, destFile)
	if err != nil {
		t.Errorf("Failed to create a gif preview %s", err)
	}
	fCheck, noFileErr := os.Stat(destFile)
	if noFileErr != nil {
		t.Errorf("We had no error but the file is not on disk %s", destFile)
	}
	if fCheck.Size() > vFile.Size() {
		t.Errorf("Preview was bigger than video %d > %d", fCheck.Size(), vFile.Size())
	}
	ResetPreviewDir(dstDir)
}

func Test_ScreenOutputPatterns(t *testing.T) {
	badFilename := "Bad(a)).jpg"
	_, err := GetScreensMatcherRE(badFilename)
	if err != nil {
		t.Errorf("Error trying to compile a re match for %s %s", badFilename, err)
	}

	badTwo := "../Bad[b)).jpg"
	_, err2 := GetScreensMatcherRE(badTwo)
	if err2 != nil {
		t.Errorf("Error trying to compile a re match for %s %s", badTwo, err2)
	}
}
