// List of directories under the main element
package utils

/**
* These utilities are how initial content is looked up and provides other simple
* helpers around checking filetypes and building out the base data model.
 */
import (
	"errors"
	"fmt"
	"image"
	"io"
	"os"
	"strings"

	//"errors"
	"bufio"
	"contented/pkg/config"
	"contented/pkg/models"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"mime"
	"net/http"
	"path/filepath"

	"github.com/tidwall/gjson"
	"golang.org/x/exp/maps"
)

// Content Information is used as an array to dfs from a directory
type ContentInformation struct {
	Cnt     models.Container
	Content models.Contents
}
type ContentTree []ContentInformation

/**
 * Grab a small preview list of all items in the directory.
 */
func FindContainers(dir_root string) models.Containers {
	return FindContainersMatcher(dir_root, config.IncludeAllContainers, config.ExcludeContainerDefault)
}

func FindContainersMatcher(dir_root string, incCnt config.ContainerMatcher, excCnt config.ContainerMatcher) models.Containers {
	//log.Printf("FindContainers Reading from: %s", dir_root)
	var listings = models.Containers{}
	files, err := os.ReadDir(dir_root)
	if err != nil {
		log.Printf("Could not read from the directory %s", dir_root)
	}

	for _, f := range files {
		if f.IsDir() {
			dir_name := f.Name()

			if !incCnt(dir_name) || excCnt(dir_name) {
				continue
			}
			id := AssignNumerical(0, "containers")
			c := models.Container{
				ID:     id,
				Name:   dir_name,
				Path:   dir_root,
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
func FindContent(cnt models.Container, limit int64, start_offset int) models.Contents {
	return FindContentMatcher(cnt, limit, start_offset, config.IncludeAllFiles, config.ExcludeNoFiles)
}

// func yup(string, string) bool is a required positive check on the filename and content type (default .*)
// func nope(string, string) bool is a negative check (ie no zip files) default (everything is fine)
func FindContentMatcher(cnt models.Container, limit int64, start_offset int, yup config.ContentMatcher, nope config.ContentMatcher) models.Contents {
	var arr = models.Contents{}

	fqDirPath := filepath.Join(cnt.Path, cnt.Name)
	maybe_content, _ := os.ReadDir(fqDirPath)

	// TODO: Move away from "img" into something else
	total := 0
	imgs := []os.FileInfo{} // To get indexing 'right' you have to exclude directories
	for _, img := range maybe_content {
		if !img.IsDir() {
			info, _ := img.Info()
			imgs = append(imgs, info)
		}
	}

	for idx, img := range imgs {
		if !img.IsDir() {
			if int64(len(arr)) < limit && idx >= start_offset {
				id := AssignNumerical(0, "contents")
				content := GetContent(id, img, fqDirPath)
				content.ContainerID = &cnt.ID
				content.Idx = idx

				if yup(content.Src, content.ContentType) && !nope(content.Src, content.ContentType) {
					arr = append(arr, content)
				}
			}
			total++ // Only add a total for non-directory files (exclude other types?)
		}
	}
	//log.Printf("Finished reading from %s and found %d content", fqDirPath, len(arr))
	return arr
}

/**
 * Return a reader for the file contents
 */
func GetFileContents(dir string, filename string) *bufio.Reader {
	return GetFileContentsByFqName(filepath.Join(dir, filename))
}

func GetFilePathInContainer(src string, path string) (string, error) {
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

	var buf [config.SniffLen]byte
	n, _ := io.ReadFull(content, buf[:])
	ctype := http.DetectContentType(buf[:n])
	_, err := content.Seek(0, io.SeekStart) // rewind to output whole file
	if err != nil {
		return "error", err
	}
	return ctype, nil
}

// This is a little slow using the memory library so there needs to be some optimizations
func GetContent(id int64, fileInfo os.FileInfo, path string) models.Content {
	return GetContentOptionalMetadata(id, fileInfo, path, true)
}

/**
 * Get a content object with optional metadata. If you have a LOT of content loading the image and video metadata
 * can take a very long time so this allows you to load the content without the metadata.
 */
func GetContentOptionalMetadata(id int64, fileInfo os.FileInfo, path string, withMetadata bool) models.Content {
	// https://golangcode.com/get-the-content-type-of-file/
	contentType, err := GetMimeType(path, fileInfo.Name())
	if err != nil {
		log.Printf("Failed to determine contentType: %s", err)
		contentType = "application/unknown"
	}

	// I could do an ffmpeg.Probe(srcFile) to determine encoding and resolution
	// For images I could try and probe the encoding & resolution
	meta := ""
	encoding := ""
	corrupt := false
	duration := 0.0
	srcFile := filepath.Join(path, fileInfo.Name())

	//
	if withMetadata {
		if strings.Contains(contentType, "image") {
			// TODO: Determine if we can use the image library to get some information about the file.
			meta, corrupt = GetImageMeta(srcFile)
		} else if strings.Contains(contentType, "video") {
			vidInfo, probeErr := GetVideoInfo(srcFile)
			if probeErr == nil {
				meta = vidInfo
				encoding = gjson.Get(meta, "streams.0.codec_name").String() // hate
				duration = gjson.Get(meta, "format.duration").Float()
			} else {
				meta = fmt.Sprintf("Failed to probe video %s", probeErr)
				corrupt = true
			}
		}
	}
	id = AssignNumerical(id, "contents")
	content := models.Content{
		ID:          id,
		Src:         fileInfo.Name(),
		SizeBytes:   int64(fileInfo.Size()),
		ContentType: contentType,
		Meta:        meta,
		Duration:    duration,
		Corrupt:     corrupt,
		Encoding:    encoding,
		CreatedAt:   fileInfo.ModTime(),
		UpdatedAt:   fileInfo.ModTime(),
	}
	return content
}

func GetImageMeta(srcFile string) (string, bool) {
	corrupt := false
	meta := ""

	reader, err := os.Open(srcFile)
	if err != nil {
		return fmt.Sprintf("No access to source file %s err %s", srcFile, err), true
	}
	defer reader.Close()

	// TODO: this is faster but we have a lot less indication of if it is config
	// Potentially the preview generation woul dbe the way to go.
	m, _, i_err := image.DecodeConfig(bufio.NewReader(reader))
	if i_err != nil {
		meta = "Couldn't Determine Image info"
		corrupt = true
	} else {
		// A Full image Decode is way too slow so instead we are just looking at the config for now.
		// m, _, i_err := image.Decode(bufio.NewReader(reader))
		// bounds := m.Bounds()
		// w := bounds.Dx()
		// h := bounds.Dy()
		w := m.Width
		h := m.Height
		meta = fmt.Sprintf("{\"width\": %d, \"height\": %d}", w, h)
	}
	return meta, corrupt
}

// Write a recurse method for getting all the data up to depth N
func CreateStructure(dir string, cfg *config.DirConfigEntry, results *ContentTree, depth int) (*ContentTree, error) {
	// log.Printf("Looking in directory %s set have results %d depth %d", dir, len(*results), depth)
	if depth > cfg.MaxSearchDepth {
		return results, nil
	}

	log.Printf("Starting search in %s", dir)
	// Find all the containers under the specified directory (is directory)
	// Could specify the cfg to use with the matching?
	cnts := FindContainersMatcher(dir, cfg.IncContainer, cfg.ExcContainer)
	for _, cnt := range cnts {
		content := FindContentMatcher(cnt, cfg.MaxContentPerContainer, 0, cfg.IncContent, cfg.ExcContent)
		cnt.Total = len(content)
		cTree := ContentInformation{
			Cnt:     cnt,
			Content: content,
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

// From StackOverflow Ensure that something is UNDER the content directory
func SubPath(parent string, sub string) (bool, error) {
	up := ".." + string(os.PathSeparator)

	// path-comparisons using filepath.Abs don't work reliably according to docs (no unique representation).
	rel, err := filepath.Rel(parent, sub)
	if err != nil {
		return false, err
	}
	if !strings.HasPrefix(rel, up) && rel != ".." {
		return true, nil
	}
	return false, nil
}

func HasContent(src string, path string) (bool, error) {
	up := ".." + string(os.PathSeparator)
	if strings.Contains(src, up) || strings.Contains(src, "~") {
		return false, errors.New("filename includes possible up directory access, denied")
	}
	_, err := GetFilePathInContainer(src, path)
	if err != nil {
		return false, err
	}
	return true, nil
}

/*
 * Check if the path is under a valid directory and doesn't try and escape up
 */
func PathIsOk(path string, name string, ensureUnder string) (bool, error) {
	dest := filepath.Join(path, name)
	if HasUpwardTraversal(dest) {
		return false, errors.New("path includes possible up directory access, denied")
	}

	// Optional, ensure the path is under some configured root.
	if ensureUnder != "" && path != ensureUnder {
		ok, nope := SubPath(ensureUnder, path)
		if !ok || nope != nil {
			return ok, nope
		}
	}
	// Now check it is a directory
	f, err := os.Stat(dest)
	if err != nil {
		return false, err // No access potentially
	}
	if f.IsDir() {
		return true, nil
	}
	return false, fmt.Errorf("%s is not a directory under the path", name)
}

/*
 * Remove a file from the disk if it exists and is under a safe directory.
 */
func RemoveFileIsOk(path string, name string, underDirectory string) (bool, error) {

	// NOTE if you join these paths using filepath.Join it will remove the traversal attempts
	// for you already... which is not great when trying to prevent that kind of access.
	if HasUpwardTraversal(name) {
		return false, fmt.Errorf("filename includes possible up directory access, denied")
	}
	if HasUpwardTraversal(path) {
		return false, fmt.Errorf("path includes possible up directory access, denied")
	}

	if ok, err := PathIsOk(path, "", underDirectory); err != nil {
		return false, err
	} else if !ok {
		return false, fmt.Errorf("path is not under root %s", path)
	}

	fqPath := filepath.Join(path, name)

	// Check if file exists
	fileInfo, err := os.Stat(fqPath)
	if errors.Is(err, os.ErrNotExist) {
		return true, err
	}
	if err != nil {
		return false, err
	}
	// Verify it's not a directory
	if fileInfo.IsDir() {
		return false, fmt.Errorf("this points to a directory: %s", fqPath)
	}
	return true, nil
}

func RemoveFile(path string, name string, underDirectory string) (bool, error) {
	ok, err := RemoveFileIsOk(path, name, underDirectory)

	// It is ok if the file was already removed by another process or call
	if errors.Is(err, os.ErrNotExist) {
		log.Printf("File already removed %s", err)
		return true, err
	}
	// Otherwise all errors will be considered fatal
	if err != nil {
		return false, err
	}
	if !ok {
		return false, fmt.Errorf("path is not ok %s", path)
	}

	// Remove the file
	fqPath := filepath.Join(path, name)
	if err := os.Remove(fqPath); err != nil {
		return false, err
	}
	return true, nil
}

/*
 * Detects if a filename is attempting to traverse up directories (Cursor wrote it impressive)
 * Returns true if the filename contains directory traversal attempts
 */
func HasUpwardTraversal(filename string) bool {
	// Check for ../ or ..\ patterns
	up := ".." + string(os.PathSeparator)
	if strings.Contains(filename, up) {
		return true
	}

	// Check for encoded variants
	encodedUp := "%2e%2e%2f"    // ../
	encodedUpAlt := "%2e%2e%5c" // ..\
	if strings.Contains(strings.ToLower(filename), encodedUp) ||
		strings.Contains(strings.ToLower(filename), encodedUpAlt) {
		return true
	}

	// Check for double-encoded variants
	doubleEncodedUp := "%252e%252e%252f"    // ../
	doubleEncodedUpAlt := "%252e%252e%255c" // ..\
	if strings.Contains(strings.ToLower(filename), doubleEncodedUp) ||
		strings.Contains(strings.ToLower(filename), doubleEncodedUpAlt) {
		return true
	}

	return false
}

func ReadTagsFromFile(tagFile string) (*models.Tags, error) {
	tags := models.Tags{}
	if tagFile == "" {
		return &tags, nil

	}
	log.Printf("Processing Tags Attempting to read tags from %s", tagFile)
	if _, err := os.Stat(tagFile); !os.IsNotExist(err) {
		f, fErr := os.OpenFile(tagFile, os.O_RDONLY, os.ModePerm)
		if fErr == nil {
			defer f.Close()
		}
		if fErr != nil {
			log.Printf("Processing Tags Error reading file %s", fErr)
			return nil, fErr
		}

		// I could also make this smarter and do a single IN query and then a single insert
		tagType := "keywords"
		sc := bufio.NewScanner(f)
		for sc.Scan() {
			tagLine := strings.TrimSpace(sc.Text())
			if tagLine != "" {
				if strings.Contains(tagLine, "//") {
					if strings.Contains(tagLine, "//TagType:") {
						tagTypeArr := strings.Split(tagLine, ":")
						if len(tagTypeArr) > 1 {
							tagType = strings.TrimSpace(strings.Join(tagTypeArr[1:], " "))
						}
					}
				} else {
					name := strings.TrimSpace(tagLine)
					t := models.Tag{ID: name, TagType: tagType}
					tags = append(tags, t)
				}
			}

			// TODO: Make tags under their sections
		}
	} else {
		log.Printf("No tagfile found at %s", tagFile)
	}
	return &tags, nil
}

/**
 * Return the content with newly associated tags.
 */
func AssignTagsToContent(contentMap models.ContentMap, tagMap models.TagsMap) models.ContentMap {
	contents := maps.Values(contentMap)
	tags := maps.Values(tagMap)
	updatedContent := models.ContentMap{}
	for _, content := range contents {
		matchedTags := AssignTags(content, tags)
		if len(matchedTags) > 0 {
			content.Tags = maps.Values(matchedTags)
			updatedContent[content.ID] = content
		}
	}
	return updatedContent
}

/**
 * Try and assign some tag content
 */
func AssignTags(content models.Content, tags models.Tags) models.TagsMap {
	tagsMap := models.TagsMap{}
	if content.Src != "" {
		// Loop over the tags and check if the string contains the tag
		for _, tag := range tags {
			if TagMatches(content.Src, tag) {
				tagsMap[tag.ID] = tag
			}
		}
	}

	// len of description could do it from a tag optimization
	if content.Description != "" {
		for _, tag := range tags {
			if _, ok := tagsMap[tag.ID]; !ok {
				if TagMatches(content.Description, tag) {
					tagsMap[tag.ID] = tag
				}
			}
		}
	}
	// Do we want the keys or the values of these elements?
	return tagsMap
}

// Really just here to add possible complexity to the tag matching?
// To lower case?
func TagMatches(val string, tag models.Tag) bool {
	// Tag should maybe support case?
	// Case on tags should maybe be a config setting (ANOTHER ONE??!?!)
	return strings.Contains(strings.ToLower(val), strings.ToLower(tag.ID))
}
