package utils

import "testing"

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
