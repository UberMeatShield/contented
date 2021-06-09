package utils

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/gobuffalo/envy"
)

// Possibly make this some sort of global test helper function (harder to do in GoLang?)
var testDir, _ = envy.MustGet("DIR")

func TestFindContainers(t *testing.T) {
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

func TestFindMedia(t *testing.T) {
    containers := FindContainers(testDir)
    for _, c := range containers {
        media := FindMedia(c, 42, 0)
        if len(media) == 0 {
            t.Errorf("Failed to lookup media in container %s", c.Name)
        }
    }
}

func Test_MediaMatcher(t *testing.T) {
    containers := FindContainers(testDir)

    FailAll := func(filename string, content_type string) bool {
        return true // nothing should match
    }

    // The positive include all cases handled by using FindMedia tests (default include all matches)
    for _, cnt := range containers {
        media := FindMediaMatcher(cnt, 0, 20, IncludeAllFiles, FailAll)
        if len(media) != 0 {
            t.Errorf("All Files should be excluded")
        }
        inc_test := FindMediaMatcher(cnt, 0, 20, FailAll, ExcludeNoFiles)
        if len(inc_test) != 0 {
            t.Errorf("None of these should be included")
        }
    }

}

func Test_ContentType(t *testing.T) {
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

// Eh, kinda like some form of hashing
func Test_DirId(t *testing.T) {
	id1 := GetDirId("dir1")
	if id1 != "4c1f6165302b81fd587e79db729a5a05ea130ea35602a76dcf0dd96a2366f33c" {
		t.Errorf("Failed to hash correctly %s", id1)
	}
}

func TestFindMediaOffset(t *testing.T) {
    containers := FindContainers(testDir)

    expect_dir := false
    expect_total := 6
    for _, c := range containers {
        if c.Name == "dir3" {
            expect_dir = true
            media := FindMedia(c, 2, 0)
            if len(media) > 3 {
		        t.Error("Limit failed to restrict contents")
            }
            allm := FindMedia(c, 42, 0)
            total := len(allm)

            if total != expect_total {
		        t.Errorf("There should be exactly n(%d) found but returned %d", expect_total, total)
            }

            offset := FindMedia(c, 6, 2)
            if len(offset) != 4 {
                t.Errorf("The offset should lower the total returned but we found %d in %s", len(offset), c.Name)
            }
        }
    }

    if expect_dir == false {
        t.Fatal("The test directory dir3 was not found")
    }
}

func Test_CreateMatcher(t *testing.T) {
    matcher := CreateMatcher(".jpg|.png|.gif", "image")

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
