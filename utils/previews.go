package utils

import (
	"errors"
	"image"
	"image/jpeg"
	"image/png"
	"log"
	"os"
	"path/filepath"
	"github.com/nfnt/resize"
)

func MakePreviewPath(dstPath string) error {
	if _, err := os.Stat(dstPath); os.IsNotExist(err) {
		return os.MkdirAll(dstPath, 0740)
	}
	return nil
}

func ShouldCreatePreview(f *os.File, fsize int64) bool {
	finfo, err := f.Stat()
	if err != nil {
		log.Printf("Error determining file stat for %s", err)
		return false
	}

	if finfo.Size() > fsize {
		// log.Printf("How big was the size %d", finfo.Size())
		return true
	}
	return false
}

//  No need to check
func PreviewExists(filename string, dstPath string) (string, error) {
	// Does the preview already exist, return that
	dstFile := filepath.Join(dstPath, filename)
	if _, e_err := os.Stat(dstFile); os.IsNotExist(e_err) {
		return dstFile, nil
	}
	return "", errors.New("Preview Already Exists: " + dstFile)
}

// Break this down better using just a file object?
func CreateImagePreview(srcImg *os.File, dstFile string, contentType string) (string, error) {
	// Attempt to create previews for different media types
	var img image.Image
	var dErr error

	// HOW TO DO THIS in a sane extensible fashion?
	if contentType == "image/png" {
		img, dErr = png.Decode(srcImg)
	} else if contentType == "image/jpeg" {
		img, dErr = jpeg.Decode(srcImg)
	} else {
		log.Printf("No provided method for this file type %s", contentType)
		fname, _ := srcImg.Stat()
		return "", errors.New("Cannot handle type for file: " + fname.Name())
	}
	if dErr != nil {
		log.Fatal(dErr)
	}

	// Now creat the preview image
	dstImg := resize.Resize(640, 0, img, resize.Lanczos3)
	previewImg, errCreate := os.Create(dstFile)
	if errCreate != nil {
		log.Fatal(errCreate)
		return "Error" + dstFile, errCreate
	}

	// All previews should then be jpeg (change file extensioni?)
	jpeg.Encode(previewImg, dstImg, nil)
	return dstFile, previewImg.Close()
}

// Make sure dstPath already exists before you call this (MakePreviewPath)
func GetImagePreview(path string, filename string, dstPath string, pIfSize int64) (string, error) {

	// HMMM (if exists do not do anything)
	dstFile, dErr := PreviewExists(filename, dstPath)
	if dErr != nil {
		return dstFile, dErr
	}

	// Try and determine the content type (required for doing encoding and decoding)
	contentType, tErr := GetMimeType(path, filename)
	if tErr != nil {
		log.Fatal(tErr)
		return "Could not determine img type", tErr
	}

	// The file we are going to check about making a preview of
	fqFile := filepath.Join(path, filename)
	srcImg, fErr := os.Open(fqFile)
	if fErr != nil {
		log.Fatal(fErr)
		return "Error Generating Preview", fErr
	}
	defer srcImg.Close()

	// Check to see if the image is ACTUALLY over a certain size to be worth previewing
	if ShouldCreatePreview(srcImg, pIfSize) == true {
		return CreateImagePreview(srcImg, dstFile, contentType)
	}
    // No Preview is required
	return "", nil
}
