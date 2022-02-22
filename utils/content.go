// List of directories under the main element
package utils

import (
    "os"
    //"errors"
    "io"
    "io/ioutil"
    "bufio"
    "log"
    "mime"
    "net/http"
    "contented/models"
    "crypto/sha256"
    "encoding/hex"
    "path/filepath"
    "github.com/gobuffalo/nulls"
    "github.com/gofrs/uuid"
)

// Content Information is used as an array to dfs from a directory
type ContentInformation struct {
    Cnt models.Container
    Media models.MediaContainers
}
type ContentTree []ContentInformation

/**
 * Grab a small preview list of all items in the directory.
 */
func FindContainers(dir_root string) models.Containers {
    // Get the current listings, check they passed in a legal key
    log.Printf("FindContainers Reading from: %s", dir_root)

    var listings = models.Containers{}
    files, err := ioutil.ReadDir(dir_root)
    if err != nil {
        log.Printf("Could not read from the directory %s", dir_root)
    }

    for _, f := range files {
        if f.IsDir() {
            dir_name := f.Name()
            id, _  := uuid.NewV4()
            c := models.Container{
                ID:   id,
                Name: dir_name,
                Path: dir_root,
                Active: true,
            }
            listings = append(listings, c)
        }
    }
    return listings
}

/**
 *  Get all the content in a particular directory (would be good to filter down to certain file types?)
 */
func FindMedia(cnt models.Container, limit int, start_offset int) models.MediaContainers {
    return FindMediaMatcher(cnt, limit, start_offset, IncludeAllFiles, ExcludeNoFiles)
}

// func yup(string, string) bool is a required positive check on the filename and content type (default .*)
// func nope(string, string) bool is a negative check (ie no zip files) default (everything is fine)
func FindMediaMatcher(cnt models.Container, limit int, start_offset int, yup MediaMatcher, nope MediaMatcher) models.MediaContainers {
    var arr = models.MediaContainers{}

    fqDirPath := filepath.Join(cnt.Path, cnt.Name)
    maybe_media, _ := ioutil.ReadDir(fqDirPath)

    // TODO: Move away from "img" into something else
    total := 0
    imgs := []os.FileInfo{}  // To get indexing 'right' you have to exlcude directories
    for _, img := range maybe_media {
        if !img.IsDir() {
            imgs = append(imgs, img)
        }
    }

    for idx, img := range imgs {
        if !img.IsDir() {
            if len(arr) < limit && idx >= start_offset {
                id, _  := uuid.NewV4()
                media := getMediaContainer(id, img, fqDirPath)
                media.ContainerID = nulls.NewUUID(cnt.ID)
                media.Idx = idx

                if yup(media.Src, media.ContentType) && !nope(media.Src, media.ContentType) {
                    arr = append(arr, media)
                }
            }
            total++ // Only add a total for non-directory files (exclude other types?)
        }
    }
    log.Printf("Finished reading from %s and found %d media", fqDirPath, len(arr))
    return arr
}


/**
 * Return a reader for the file contents
 */
func GetFileContents(dir string, filename string) *bufio.Reader {
    return GetFileContentsByFqName(filepath.Join(dir, filename))
}

// Given a container ID and the src of a file in there, get a path and check if it exists
func GetFilePathInContainer(src string, path string) (string, error) {
    //TODO: Potentially I should look at cfg.Dir, cnt.Path and src
    //cfg := GetCfg() 
    fq_path := filepath.Join(path, src)
    if _, os_err := os.Stat(fq_path); os_err != nil {
        return fq_path, os_err
    }
    return fq_path, nil
}

func GetFileContentsByFqName(fq_name string) *bufio.Reader {
    f, err := os.Open(fq_name)
    if err != nil {
        panic(err)
    }
    return bufio.NewReader(f)
}

func GetDirId(name string) string {
    h := sha256.New()
    h.Write([]byte(name))
    return hex.EncodeToString(h.Sum(nil))
}

// Make a guess at the content type of the file (might be wrong based on file extension)
func GetMimeType(path string, filename string) (string, error) {
    name := filepath.Join(path, filename)
    ctype := mime.TypeByExtension(filepath.Ext(name))

    if ctype == "" {
        // read a chunk to decide between utf-8 text and binary
        content, errOpen := os.Open(name)
        if errOpen != nil {
            return "Error Reading File", errOpen
        }
        defer content.Close()
        return SniffFileType(content)
    }
    return ctype, nil
}

func SniffFileType(content *os.File) (string, error) {
    var buf [sniffLen]byte
    n, _ := io.ReadFull(content, buf[:])
    ctype := http.DetectContentType(buf[:n])
    _, err := content.Seek(0, io.SeekStart) // rewind to output whole file
    if err != nil {
        return "error", err
    }
    return ctype, nil
}

func getMediaContainer(id uuid.UUID, fileInfo os.FileInfo, path string) models.MediaContainer {
    // https://golangcode.com/get-the-content-type-of-file/
    contentType, err := GetMimeType(path, fileInfo.Name())
    if err != nil {
        log.Printf("Failed to determine contentType: %s", err)
        contentType = "application/unknown"
    }


    // I could do an ffmpeg.Probe(srcFile) to determine encoding and resolution
    // For images I could try and probe the encoding & resolution

    // TODO: Need to add the unique ID for each media (are they uniq?)
    // TODO: Should I get a Hash onto the media as well?
    media := models.MediaContainer{
        ID:   id,
        Src:  fileInfo.Name(),
        SizeBytes: fileInfo.Size(),
        ContentType: contentType,
    }
    return media
}


// Write a recurse method for getting all the data up to depth N
func CreateStructure(dir string, cfg *DirConfigEntry, results *ContentTree, depth int) (*ContentTree, error) {
    //log.Printf("Looking in directory %s set have results %d depth %d", dir, len(*results), depth)
    if depth > cfg.MaxSearchDepth {
        return results, nil
    }

    // Find all the containers under the specified directory (is directory)
    cnts := FindContainers(dir)
    for _, cnt := range cnts {
        if cnt.Name == "container_previews" {
            continue
        }
        media := FindMediaMatcher(cnt, cfg.MaxMediaPerContainer, 0, cfg.IncFiles, cfg.ExcFiles)
        cnt.Total = len(media)
        cTree := ContentInformation{
            Cnt: cnt,
            Media: media,
        } 
        tree := append(*results, cTree) 
        subDir := filepath.Join(dir, cnt.Name)
        //log.Printf("SubDir %s and depth is currently %d count of containers %d", subDir, depth, len(tree))

        mergeTree, err := CreateStructure(subDir, cfg, &tree, depth+1)
        if err != nil {
            log.Printf("Error searching down the subTree %s with error %s", subDir, err)
            return results, err
        }
        results = mergeTree
    }
    return results, nil
}
