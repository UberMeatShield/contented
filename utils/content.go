// List of directories under the main element
package utils

import (
    "regexp"
	"bufio"
	"contented/models"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"github.com/gobuffalo/nulls"
	"github.com/gofrs/uuid"
)

const sniffLen = 512

// Matchers that determine if you want to include specific filenames/content types
type MediaMatcher func(string, string) bool

func ExcludeNoFiles(filename string, content_type string) bool {
    return false
}

func IncludeAllFiles(filename string, content_type string) bool {
    return true
}

func CreateMatcher(filenameStrRE string, typesStrRE string) MediaMatcher {
    filenameRE := regexp.MustCompile(filenameStrRE) 
    typeRE := regexp.MustCompile(typesStrRE)

    return func(filename string, content_type string) bool {
        return filenameRE.MatchString(filename) && typeRE.MatchString(content_type)
    }
}

// TODO: this might be useful to add into the utils
type DirConfigEntry struct {
	Dir             string // The root of our loading (path to top level container directory)
	PreviewCount    int    // How many files should be listed for a preview
	Limit           int    // The absolute max you can load in a single operation
	Initialized     bool
    UseDatabase     bool
    CoreCount       int
    StaticResourcePath string

    // Matchers that will determine which media elements tend to be included
    IncFiles MediaMatcher
    ExcFiles MediaMatcher
}

/*
 * Build out a valid configuration given the directory etc.
 *
 * Note we do not create a new instance, we are updating the overall app config.
 * TODO: Figure out how to do this "right" for a Buffalo app.
 */
func InitConfig(dir_root string, cfg *DirConfigEntry) *DirConfigEntry {
	cfg.Dir = dir_root  // Always Common
	cfg.Initialized = true
    cfg.CoreCount = 4
    cfg.IncFiles = IncludeAllFiles
    cfg.ExcFiles = ExcludeNoFiles
	return cfg
}

/**
 * Grab a small preview list of all items in the directory.
 */
func FindContainers(dir_root string) models.Containers {
	// Get the current listings, check they passed in a legal key
	log.Printf("FindContainers Reading from: %s", dir_root)

	var listings = models.Containers{}
	files, _ := ioutil.ReadDir(dir_root)
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
    return arr
}


/**
 * Return a reader for the file contents
 */
func GetFileContents(dir string, filename string) *bufio.Reader {
	return GetFileContentsByFqName(filepath.Join(dir, filename))
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
		contentType = "image/jpeg"
	}

	// TODO: Need to add the unique ID for each dir (are they uniq?)
	media := models.MediaContainer{
		ID:   id,
		Src:  fileInfo.Name(),
		ContentType: contentType,
	}
	return media
}
