// List of directories under the main element
package utils

import (
    "errors"
	"strconv"
    "strings"
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
    "github.com/gobuffalo/envy"
	"github.com/gobuffalo/nulls"
	"github.com/gofrs/uuid"
)

const sniffLen = 512  // How many bytes to read in a file when trying to determine mime type
const DefaultLimit int = 10000 // The max limit set by environment variable
const DefaultPreviewCount int = 8
const DefaultUseDatabase bool = false

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
	Limit           int    // The absolute max you can load in a single operation
    UseDatabase     bool   // Should it use the database or an in memory version
    CoreCount       int    // How many cores are likely available (used in creating multithread workers / previews)
    StaticResourcePath string  // The location where compiled js and css is hosted (container vs dev server)
	Initialized     bool   // Has the configuration actually be initialized properly
    PreviewCount    int    // How many files should be listed for a preview (todo: USE)
    PreviewOverSize int64  // Over how many bytes should previews be created for the file
    PreviewVideoType string // This will be either gif or png based on config

    // Matchers that will determine which media elements to be included or excluded
    IncFiles MediaMatcher
    ExcFiles MediaMatcher
}

 // https://medium.com/@TobiasSchmidt89/the-singleton-object-oriented-design-pattern-in-golang-9f6ce75c21f7
 var appCfg DirConfigEntry = GetCfgDefaults()


 // TODO: Manager has a config as does utils, this seems sketchy
 func GetCfg() *DirConfigEntry {
     return &appCfg
 }
 func SetCfg(cfg DirConfigEntry) {
     appCfg = cfg
 }

func GetCfgDefaults() DirConfigEntry {
   return DirConfigEntry{
       Initialized:  false,
       UseDatabase: true,
       Dir: "",
       CoreCount: 4,
       Limit: DefaultLimit,
       PreviewCount: DefaultPreviewCount,
       PreviewOverSize: 1024000,
       PreviewVideoType: "png",

       // Just grab all files by default
       IncFiles: IncludeAllFiles,
       ExcFiles: ExcludeNoFiles,
   }
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
    SetupConfigMatchers(cfg, "", "", "", "")
	return cfg
}

// Should I move this into the config itself?
func InitConfigEnvy(cfg *DirConfigEntry) *DirConfigEntry {
    var err error
    dir := envy.Get("DIR", "")
    if dir == "" {
        dir, err = envy.MustGet("CONTENT_DIR")  // From the .env file
    }
    if !strings.HasSuffix(dir, "/") {
        dir = dir + "/"
    }
    log.Printf("Setting up the content directory with %s", dir)

    staticDir := envy.Get("STATIC_RESOURCE_PATH", "./public/build")
    limitCount, limErr := strconv.Atoi(envy.Get("LIMIT", strconv.Itoa(DefaultLimit)))
    previewCount, previewErr := strconv.Atoi(envy.Get("PREVIEW", strconv.Itoa(DefaultPreviewCount)))
    useDatabase, connErr := strconv.ParseBool(envy.Get("USE_DATABASE", strconv.FormatBool(DefaultUseDatabase)))
    coreCount, coreErr := strconv.Atoi(envy.Get("CORE_COUNT", "4"))

    psize, perr := strconv.ParseInt(envy.Get("CREATE_PREVIEW_SIZE", "1024000"), 10, 64)
    previewType := envy.Get("PREVIEW_VIDEO_TYPE", "png")

    if err != nil {
        panic(err)
    } else if limErr != nil {
        panic(limErr)
    } else if previewErr != nil {
        panic(previewErr)
    } else if _, noDirErr := os.Stat(dir); os.IsNotExist(noDirErr) {
        panic(noDirErr)
    } else if connErr != nil {
        panic(connErr)
    } else if coreErr != nil {
        panic(coreErr)
    } else if perr != nil {
        panic(perr)
    }
    if !(previewType == "png" || previewType == "gif") {
        panic(errors.New("The video preview type is not png or gif"))
    }

	cfg.Dir = dir
    cfg.UseDatabase = useDatabase
    cfg.StaticResourcePath = staticDir
    cfg.Limit = limitCount
    cfg.CoreCount = coreCount
	cfg.PreviewCount = previewCount
    cfg.PreviewVideoType = previewType
    cfg.PreviewOverSize = psize

    SetupConfigMatchers(
        cfg,
        envy.Get("INCLUDE_FILES_MATCH", ""),
        envy.Get("INCLUDE_TYPES_MATCH", ""),
        envy.Get("EXCLUDE_FILES_MATCH", ""),
        envy.Get("EXCLUDE_TYPES_MATCH", ""),
    )
	cfg.Initialized = true
	return cfg
}


// Setup the matchers on the configuration, these are used to determine which media elments should match
// yes filename matches, yes mime matches, no if the filename matches, no if the mime matches.
func SetupConfigMatchers(cfg *DirConfigEntry, y_fn string, y_mime string, n_fn string, n_mime string) {

    //To include media only if it matches the filename or mime type
    if y_fn != "" || y_mime != "" {
        cfg.IncFiles = CreateMatcher(y_fn, y_mime)
    } else {
        cfg.IncFiles = IncludeAllFiles
    }

    // If you do not specify exclusion regexes it will just include everything
    if n_fn != "" || n_mime != "" {
        cfg.ExcFiles = CreateMatcher(n_fn, n_mime)
    } else {
        cfg.ExcFiles = ExcludeNoFiles
    }
}

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

    cfg := GetCfg()
    fqDirPath := filepath.Join(cfg.Dir, cnt.Name)
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

// Given a container ID and the src of a file in there, get a path and check if it exists
func GetFilePathInContainer(src string, dir_name string) (string, error) {
    cfg := GetCfg() // TODO: Potentially this should be passed in
    path := filepath.Join(cfg.Dir, dir_name)
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
