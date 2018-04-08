package utils

import "testing"
import "time"
import "fmt"

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
	files := GetDirContents(testDir + "/dir3", 2, "mocks")
	if (len(files.Contents) != 2) {
		t.Errorf("Did not limit the directory length, wanted %d found %d", count, len(files.Contents))
	}

	files = GetDirContents(testDir + "/dir3", 10, "mocks")
	if (len(files.Contents) < 3) {
		t.Error("There are more test files in this directory than 3")
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
