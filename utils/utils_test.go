package utils

import "testing"
import "time"

var testDir = "./../mocks/content"

func TestGetDirContents(t *testing.T) {
    lookup := GetDirectoriesLookup(testDir)

    var dirs = []string{"dir1", "dir2", "dir3", "screens"}
    count := 0
    for _, dir := range dirs {
        if !lookup[dir] {
            t.Errorf("Failed to get a lookup for this dir %s", dir)
        } else {
            count++
        }
    }
    if (count != 4) {
        t.Error("Failed to actually test the API")
    }
}


func TestGetSpecificDir(t *testing.T) {
	var count = 2
	files := GetDirContents(testDir + "/dir3", 2, 0, "mocks")
	if (len(files.Contents) != 2) {
		t.Errorf("Did not limit the directory length, wanted %d found %d", count, len(files.Contents))
	}

	files = GetDirContents(testDir + "/dir3", 10, 0, "mocks")
	if (len(files.Contents) < 3) {
		t.Error("There are more test files in this directory than 3")
	}

    start_offset := 4
    offset_files := GetDirContents(testDir + "/dir3", 3, start_offset, "mocks")
    len_contents := len(offset_files.Contents)
    if (len_contents != 1 ) {
		t.Errorf("With the offset we should have only have 1 %d", len_contents)
    }
    if (offset_files.Total != 5) {
		t.Error("There should be exactly 5 images in the dir")
    }

    first_file := offset_files.Contents[0]
    if (first_file.Id != 4) {
		t.Errorf("Offset should change the initial id %d", first_file.Id)
    }
    if (first_file.Src != "hkacMG4.jpg") {
		t.Errorf("Offset should change the initial file %s", first_file.Src)
    }
    
}


func example(sleep int, msg string, reply chan string) {
    sleepTime := time.Duration(sleep) * time.Millisecond
    time.Sleep(sleepTime)
    // fmt.Printf("Done sleeping %d with msg %s \n", sleep, msg)
    reply <- msg
}

func TestChannels(t *testing.T) {
    learn := make(chan string)

    // Timeouts should mean the first returned value is going to be derp
    go example(1000, "wtf", learn)
    go example(500, "derp", learn)
    a, b := <-learn, <-learn
    if a != "derp" {
        t.Errorf("Should return wtf %s", a)
    }
    if b != "wtf" {
        t.Errorf("What the actual fuck %s", b)
    }
}
