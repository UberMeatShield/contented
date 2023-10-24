package utils

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/gobuffalo/envy"
)

// Possibly make this some sort of global test helper function (harder to do in GoLang?)
func TestFindContainers(t *testing.T) {
	var testDir, _ = envy.MustGet("DIR")
	containers := FindContainers(testDir)
	if len(containers) != 4 {
		t.Fatal("There should be 4 containers in the mock")
	}

	var known_dirs = map[string]bool{"dir1": true, "dir2": true, "dir3": true, "screens": true}
	count := 0
	for _, c := range containers {
		if _, ok := known_dirs[c.Name]; ok {
			count++
		} else {
			t.Errorf("Failed to get a lookup for this dir %s", c.Name)
		}
	}
	if count != 4 {
		t.Error("Failed to pull in the known / expected directories")
	}
}

func Test_FindContent(t *testing.T) {
	var testDir, _ = envy.MustGet("DIR")
	cfg := GetCfg()
	cfg.Dir = testDir
	containers := FindContainers(testDir)
	for _, c := range containers {
		content := FindContent(c, 42, 0)
		if len(content) == 0 {
			t.Errorf("Failed to lookup content in container %s", c.Name)
		}
	}
}

func Test_SetupContainerMatchers(t *testing.T) {
	var testDir, _ = envy.MustGet("DIR")
	cfg := DirConfigEntry{}
	cfg.Dir = testDir
	SetupContainerMatchers(&cfg, "dir1|screens", "DS_Store")

	containers := FindContainersMatcher(testDir, cfg.IncContainer, cfg.ExcContainer)
	if len(containers) != 2 {
		t.Errorf("We should only have matched two directories")
	}
	all_containers := FindContainersMatcher(testDir, IncludeAllContainers, cfg.ExcContainer)
	if len(all_containers) != 4 {
		t.Errorf("We should have excluded the .DS_Store")
	}
	exclude_none := FindContainersMatcher(testDir, IncludeAllContainers, ExcludeNoContainers)
	if len(exclude_none) != 5 { // Prove we excluded some containers
		t.Errorf("The DS_Store exists, if nothing is excluded then it should count")
	}
}

// Exclude the movie by name
func Test_SetupContentMatchers(t *testing.T) {
	var testDir, _ = envy.MustGet("DIR")
	cfg := DirConfigEntry{}
	cfg.Dir = testDir
	SetupContentMatchers(&cfg, "", "", "donut|.DS_Store", "")

	containers := FindContainers(testDir)
	found := false
	for _, c := range containers {
		content := FindContentMatcher(c, 42, 0, cfg.IncContent, cfg.ExcContent)
		if c.Name == "dir2" {
			found = true
			if len(content) != 3 {
				t.Errorf("It did not exclude the movie by partial name match %s", content)
			}
		}
	}
	if found != true {
		t.Errorf("The test did not find the container that would include the movie")
	}
}

func Test_ContentMatcher(t *testing.T) {
	var testDir, _ = envy.MustGet("DIR")
	containers := FindContainers(testDir)

	FailAll := func(filename string, content_type string) bool {
		return true // nothing should match
	}

	// The positive include all cases handled by using FindContent tests (default include all matches)
	for _, cnt := range containers {
		content := FindContentMatcher(cnt, 0, 20, IncludeAllFiles, FailAll)
		if len(content) != 0 {
			t.Errorf("All Files should be excluded")
		}
		inc_test := FindContentMatcher(cnt, 0, 20, FailAll, ExcludeNoFiles)
		if len(inc_test) != 0 {
			t.Errorf("None of these should be included")
		}
	}

}

func Test_ContentType(t *testing.T) {
	var testDir, _ = envy.MustGet("DIR")
	imgName := "this_is_jp_eg"
	dirPath := filepath.Join(testDir, "dir1")

	// Test out determining content type from content (this is a jpg)
	contentType, err := GetMimeType(dirPath, imgName)
	if err != nil {
		t.Errorf("There should be a valid mime type %s", err)
	}
	if contentType != "image/jpeg" {
		t.Errorf("The content type returned was %s", contentType)
	}

	// Next test out a PNG type
	pngName := "this_is_p_ng"
	pngType, pngErr := GetMimeType(dirPath, pngName)
	if pngErr != nil {
		t.Errorf("Failed to determine png type %s", pngErr)
	}
	if pngType != "image/png" {
		t.Errorf("Failed to determine content type %s", pngType)
	}
}

func TestCreateStructure(t *testing.T) {
	var testDir, _ = envy.MustGet("DIR")
	cfg := GetCfg()
	cfg.MaxSearchDepth = 1
	cfg.Dir = testDir
	SetupContentMatchers(cfg, "", "", "DS_Store", "")

	cTree := ContentTree{}
	tree, err := CreateStructure(testDir, cfg, &cTree, 0)
	if err != nil {
		t.Errorf("Could not create a proper tree %s", err)
	}
	if tree == nil {
		t.Errorf("Container tree was set to nil")
	}
	lenTree := len(*tree)
	if lenTree != 5 {
		t.Errorf("The Tree should have 5 containers %d", lenTree)
	}

	cfg.MaxSearchDepth = 0
	restricted := ContentTree{}
	restrictTree, err := CreateStructure(testDir, cfg, &restricted, 0)
	lenRestricted := len(*restrictTree)
	if lenRestricted != 4 {
		t.Errorf("We should have restricted it to top level 4 vs %d", lenRestricted)
	}
}

// Eh, kinda like some form of hashing
func Test_DirId(t *testing.T) {
	id1 := GetDirId("dir1")
	if id1 != "4c1f6165302b81fd587e79db729a5a05ea130ea35602a76dcf0dd96a2366f33c" {
		t.Errorf("Failed to hash correctly %s", id1)
	}
}

func Test_FindContentOffset(t *testing.T) {
	var testDir, _ = envy.MustGet("DIR")
	containers := FindContainers(testDir)

	expect_dir := false
	expect_total := 10
	for _, c := range containers {
		if c.Name == "dir3" {
			expect_dir = true
			content := FindContent(c, 2, 0)
			if len(content) > 3 {
				t.Error("Limit failed to restrict contents")
			}
			allm := FindContent(c, 42, 0)
			total := len(allm)
			if total != expect_total {
				t.Errorf("There should be exactly n(%d) found but returned %d", expect_total, total)
			}

			offset := FindContent(c, 6, 4)
			if len(offset) != 6 {
				t.Errorf("The offset should lower the total returned but we found %d in %s", len(offset), c.Name)
			}
		}
	}

	if expect_dir == false {
		t.Fatal("The test directory dir3 was not found")
	}
}

func Test_CreateContentMatcher(t *testing.T) {
	matcher := CreateContentMatcher(".jpg|.png|.gif", "image", "AND")

	valid_fn := "derp.jpg"
	valid_mime := "image/jpeg"
	valid := matcher(valid_fn, valid_mime)
	if !valid {
		t.Errorf("The matcher should have been ok with fn %s and mime %s", valid_fn, valid_mime)
	}

	invalid_fn := "zugzug.zip"
	invalid := matcher(invalid_fn, valid_mime)
	if invalid {
		t.Errorf("Does not match fn %s and mime %s", invalid_fn, valid_mime)
	}

	wild_matcher := CreateContentMatcher("", ".*", "AND")
	// empty_or_wild := matcher(empty_fn, wild_mime)
	empty_or_wild := wild_matcher(invalid_fn, "application/x-bzip")
	if !empty_or_wild {
		t.Errorf("This should be a wildcard or empty match")
	}
}

func example(sleep int, msg string, reply chan string) {
	sleepTime := time.Duration(sleep) * time.Millisecond
	time.Sleep(sleepTime)
	// fmt.Printf("Done sleeping %d with msg %s \n", sleep, msg)
	reply <- msg
}

func Test_Channels(t *testing.T) {
	learn := make(chan string)

	// Timeouts should mean the first returned value is going to be derp
	go example(1000, "wtf", learn)
	go example(500, "derp", learn)
	a, b := <-learn, <-learn
	if a != "derp" {
		t.Errorf("Should return derp %s", a)
	}
	if b != "wtf" {
		t.Errorf("What the actual fuck %s", b)
	}
}

func Test_TagFileRead(t *testing.T) {
	var testDir, _ = envy.MustGet("DIR")
	tagFile := filepath.Join(testDir, "dir2", "tags.txt")

	tags, err := ReadTagsFromFile(tagFile)
	if err != nil {
		t.Errorf("It should not have an issue reading the tag file %s", err)
	}
	if tags == nil {
		t.Errorf("No tags were found %s", err)
	}
	if len(*tags) < 5 {
		t.Errorf("There were not enough tags found")
	}

	actorCount := 0
	actorExpect := 2
	typeKeywordCount := 0
	typeKeywordExpect := 7
	keywordCount := 0
	keywordExpect := 40

	for _, tag := range *tags {
		if tag.TagType == "keywords" {
			keywordCount += 1
		}
		if tag.TagType == "typeKeywords" {
			typeKeywordCount += 1
		}
		if tag.TagType == "actors" {
			actorCount += 1
		}
	}
	if actorExpect != actorCount {
		t.Errorf("Actor count should be %d but was %d", actorExpect, actorCount)
	}
	if typeKeywordExpect != typeKeywordCount {
		t.Errorf("typeKeywords count should be %d but was %d", typeKeywordExpect, typeKeywordCount)
	}
	if keywordExpect != keywordCount {
		t.Errorf("keywords count should be %d but was %d", keywordExpect, keywordCount)
	}

}
