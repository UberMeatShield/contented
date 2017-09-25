package utils

import (
    "io/ioutil"
)

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
func ListDirs(dir string, previewCount int) map[string][]string {
    var listings = make(map[string][]string)
    files, _ := ioutil.ReadDir(dir)
    for _, f := range files {
        if f.IsDir() {
            listings[f.Name()] = GetDirContents(dir + f.Name(), previewCount)
        }
    }
    return listings
}

/**
 *  Get all the content in a particular directory.
 */
func GetDirContents(dir string, limit int) []string {
    var arr = []string{}
    imgs, _ := ioutil.ReadDir(dir)

    for _, img := range imgs {
        if !img.IsDir() && len(arr) < limit + 1 {
            arr = append(arr, img.Name())
        }
    }
    return arr
}
