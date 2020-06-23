package utils

import (
	"os"
	"bufio"
    "io/ioutil"
	"log"
)

type MediaContainer struct{
	Id int `json:"id"`
    Src string `json:"src"`
	Type string `json:"type"`
}

type DirContents struct{
	Total int `json:"total"`
	Contents []MediaContainer `json:"contents"`
	Path string `json:"path"`
	Id string `json:"id"`
}

/**
 *  Check if a directory is a legal thing to view
 */
func GetDirectoriesLookup(legal string) map[string]bool {
    var listings = make(map[string]bool)
    files, _ := ioutil.ReadDir(legal)
    for _, f := range files {
        if f.IsDir() {
            listings[f.Name()] = true
        }
    }
    return listings
}

/**
 * Grab a small preview list of all items in the directory.
 */
func ListDirs(dir string, previewCount int) []DirContents {
	// Get the current listings, check they passed in a legal key
	var listings []DirContents
    files, _ := ioutil.ReadDir(dir)
    for _, f := range files {
        if f.IsDir() {
			id := f.Name()
            listings = append(listings, GetDirContents(dir + id, previewCount, 0, id))
        }
    }
	log.Println("Reading from: ", dir, " With preview count", previewCount)
    return listings
}

/**
 * Return a reader for the file contents
 */
func GetFileContents(dir string, filename string) *bufio.Reader {
	f, err := os.Open(dir + "/" + filename)
	if err != nil {
		panic(err)
	}
    return bufio.NewReader(f)
}


/**
 *  Get all the content in a particular directory.
 */
func GetDirContents(dir string, limit int, start_offset int, id string) DirContents {
    var arr = []MediaContainer{}
    imgs, _ := ioutil.ReadDir(dir)

	total := 0
    for idx, img := range imgs {
        if !img.IsDir() && len(arr) < limit && idx >= start_offset {

            // TODO: Actually try and determine the type of the file
            media := MediaContainer{
                Id: idx,
                Src: img.Name(),
                Type: "image/jpg",
            }
            arr = append(arr, media)
        }
		total++
    }
    log.Println("Limit for content dir was.", dir, " with limit", limit, " offset: ", start_offset)
	return DirContents{
		Total: total,
		Contents: arr,
		Path: "static/" + id,
		Id: id,
	}
}
