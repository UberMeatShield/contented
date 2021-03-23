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

func Test_DirId(t *testing.T) {
	id1 := GetDirId("dir1")
	if id1 != "4c1f6165302b81fd587e79db729a5a05ea130ea35602a76dcf0dd96a2366f33c" {
		t.Errorf("Failed to hash correctly %s", id1)
	}
}

func Test_GetSpecificDir(t *testing.T) {
	var count = 2
	files := GetDirContents(testDir+"/dir3", 2, 0, "mocks")
	if len(files.Contents) != 2 {
		t.Errorf("Did not limit the directory length, wanted %d found %d", count, len(files.Contents))
	}

	files = GetDirContents(testDir+"/dir3", 10, 0, "mocks")
	if len(files.Contents) < 3 {
		t.Error("There are more test files in this directory than 3")
	}

	start_offset := 4
	offset_files := GetDirContents(testDir+"/dir3", 3, start_offset, "mocks")
	len_contents := len(offset_files.Contents)
	if len_contents != 2 {
		t.Errorf("With the offset we should have only have 2 %d", len_contents)
	}
	if offset_files.Total != 6 {
		t.Error("There should be exactly 6 images in the dir")
	}

	first_file := offset_files.Contents[0]
	if first_file.Src != "fff&text=04-dir3.png" {
		t.Errorf("Offset should change the initial filename %s", first_file.Src)
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
