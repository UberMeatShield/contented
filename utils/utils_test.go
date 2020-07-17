package utils

import (
  "testing"
  "time"
  "github.com/gobuffalo/envy"
)


// Possibly make this some sort of global test helper function (harder to do in GoLang?)
var testDir, _ = envy.MustGet("DIR")

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

func TestGetFileRefById(t *testing.T) {
    fq_dir := testDir + "/dir1"
	dir_c := GetDirContents(fq_dir, 10, 0, "mocks")
    if len(dir_c.Contents) < 8 {
        t.Errorf("There should be contents inside of this test dir")
    }
    contents := dir_c.Contents
    entry_1 := contents[1]

    f1, err := GetFileRefById(fq_dir, entry_1.Id)
    if err != nil || f1 == nil {
        t.Errorf("Failed to lookup %s found err %s", entry_1.Id, err)
    }
    if f1.Name() != entry_1.Src {
        t.Errorf("Looked up id %s and expected %s but found %s", entry_1.Id, entry_1.Src, f1.Name())
    }

    entry_3 := contents[3]
    f3, err3 := GetFileRefById(fq_dir, entry_3.Id)
    if err3 != nil || f3 == nil {
        t.Errorf("Failed to lookup %s found err %s", entry_3.Id, err)
    }
    if f3.Name() != entry_3.Src {
        t.Errorf("Looked up id %s and expected %s but found %s", entry_3.Id, entry_3.Src, f3.Name())
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
    if (first_file.Id != "4") {
		t.Errorf("Offset should change the initial id %s", first_file.Id)
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
