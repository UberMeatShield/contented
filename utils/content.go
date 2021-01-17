// List of directories under the main element
package utils

import (
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
	"strconv"

	"github.com/gobuffalo/nulls"
	"github.com/gofrs/uuid"
)

const sniffLen = 512

type MediaContainer struct {
	Id      string `json:"id"`
	Src     string `json:"src"`
	Type    string `json:"type"`
	Preview string `json:"preview"`
}

type DirContents struct {
	Total    int              `json:"total"`
	Path     string           `json:"path"`
	Id       string           `json:"id"`
	Name     string           `json:"name"`
	Contents []MediaContainer `json:"contents"`
}

// TODO: this might be useful to add into the utils
type DirConfigEntry struct {
	Dir             string // The root of our loading (path to top level container directory)
	PreviewCount    int    // How many files should be listed for a preview
	Limit           int    // The absolute max you can load in a single operation
	Initialized     bool
	ValidDirs       map[string]os.FileInfo // List of directories under the main element
	ValidFiles      models.MediaMap        // TODO: Eventually remove this in favor of a startup + sqllite load
	ValidContainers models.ContainerMap
    UseDatabase     bool
}

/*
 * Build out a valid configuration given the directory etc.
 *
 * Note we do not create a new instance, we are updating the overall app config.
 * TODO: Figure out how to do this "right" for a Buffalo app.
 */
func InitConfig(dir_root string, cfg *DirConfigEntry) *DirConfigEntry {
	dir_lookup := GetDirectoriesLookup(dir_root)
	containers, files := PopulatePreviews(dir_root, dir_lookup)
	cfg.ValidDirs = dir_lookup
	cfg.ValidContainers = containers
	cfg.ValidFiles = files

	cfg.Dir = dir_root  // Always Common
	cfg.PreviewCount = 8
	cfg.Limit = 16
	cfg.Initialized = true
    cfg.UseDatabase = true  // TODO: Make it so we can read all the data from memory

	return cfg
}

/**
 *  TODO:  Require the number to preview and will be Memory only supported.
 */
func PopulatePreviews(dir_root string, valid_dirs map[string]os.FileInfo) (models.ContainerMap, models.MediaMap) {
	containers := models.ContainerMap{}
	files := models.MediaMap{}

	for _, f := range valid_dirs {
		// Need to make this just return a container
		dir_name := filepath.Join(dir_root + f.Name())
		dc := GetDirContents(dir_name, 20, 0, f.Name())

		log.Printf("Searching in %s", dir_name)

		c_id, _ := uuid.NewV4()
		c := models.Container{
			ID:   c_id,
			Name: dc.Name,
			Path: dc.Path,
		}
		containers[c.ID] = c
		for _, todo_mc := range dc.Contents {
			mc_id, _ := uuid.NewV4()
			mc := models.MediaContainer{
				ID:          mc_id,
				Src:         todo_mc.Src,
				Type:        todo_mc.Type,
				ContainerID: nulls.NewUUID(c.ID),
			}
			files[mc.ID] = mc
		}
	}
	return containers, files
}

/**
 *  Builds a lookup of all the valid sub directories under our root / file host.
 */
func GetDirectoriesLookup(rootDir string) map[string]os.FileInfo {
	var listings = make(map[string]os.FileInfo)
	files, err := ioutil.ReadDir(rootDir)
	if err != nil {
		panic("The main directory could not be read: " + rootDir)
	}

	// TODO: This needs to probably paginate as well and should just return the Container
	for _, f := range files {
		if f.IsDir() {
			name := f.Name()
			id := GetDirId(name)
			// listings[name] = f
			listings[id] = f
		}
	}
	return listings
}

/**
 * Grab a small preview list of all items in the directory.
 */
func ListDirs(dir string, previewCount int) []DirContents {
	// Get the current listings, check they passed in a legal key
	log.Printf("ListDirs Reading from: %s with preview count %d", dir, previewCount)

	var listings []DirContents
	files, _ := ioutil.ReadDir(dir)
	for _, f := range files {
		if f.IsDir() {
			id := f.Name() // This should definitely be some other ID format => Lookup
			listings = append(listings, GetDirContents(dir+id, previewCount, 0, id))
		}
	}
	return listings
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

/**
 * Get the file we want to lookup by ID (eventually this should be DB or just memory)
 */
func GetFileRefById(dir string, file_id_str string) (os.FileInfo, error) {
	imgs, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	file_id, ferr := strconv.Atoi(file_id_str)
	if ferr != nil {
		return nil, ferr
	}
	if file_id > len(imgs) || file_id < 0 {
		return nil, nil
	}
	return imgs[file_id], nil
}

/**
 *  Get all the content in a particular directory (would be good to filter down to certain file types?)
 */
func GetDirContents(fqDirPath string, limit int, start_offset int, dirname string) DirContents {
	var arr = []MediaContainer{}
	imgs, _ := ioutil.ReadDir(fqDirPath)

	total := 0
	for idx, img := range imgs {
		if !img.IsDir() {
			if len(arr) < limit && idx >= start_offset {
				media := getMediaContainer(strconv.Itoa(idx), img, fqDirPath)
				arr = append(arr, media)
			}
			total++ // Only add a total for non-directory files (exclude other types?)
		}
	}
	// log.Println("Limit for content dir was.", fqDirPath, " with limit", limit, " offset: ", start_offset)
	id := GetDirId(dirname)
	return DirContents{
		Total:    total,
		Contents: arr,
		Path:     "view/" + id, // from env.DIR. static/ is a configured FileServer for all content
		Id:       id,
		Name:     dirname,
	}
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

func getMediaContainer(id string, fileInfo os.FileInfo, path string) MediaContainer {
	contentType, err := GetMimeType(path, fileInfo.Name())
	if err != nil {
		log.Printf("Failed to determine contentType: %s", err)
		contentType = "image/jpeg"
	}

	// TODO: https://golangcode.com/get-the-content-type-of-file/
	// TODO: Need to cache this data (Loading all the file directory on preview is probably dumb)
	// TODO: Need to add the unique ID for each dir (are they uniq?)
	media := MediaContainer{
		Id:   id,
		Src:  fileInfo.Name(),
		Type: contentType,
	}
	return media
}
