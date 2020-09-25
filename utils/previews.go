package utils

import (
    "os"
    "log"
    "image"
    "image/jpeg"
    "image/png"
    "path/filepath"
    "github.com/nfnt/resize"
)

func MakePreviewPath(dstPath string) error {
    if _, err := os.Stat(dstPath); os.IsNotExist(err) {
        return os.MkdirAll(dstPath, 0740) 
    }
    return nil
}


// Make sure dstPath already exists before you call this
func GetImagePreview(path string, filename string, dstPath string) (string, error) {

    // Does the preview already exist, return that
    dstFile := filepath.Join(dstPath, "preview_" + filename)
    if _, e_err := os.Stat(dstFile); os.IsNotExist(e_err) {
        return dstFile, nil
    }

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
        return fqFile, nil
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
    defer previewImg.Close()

    // All previews should then be jpeg (change file extensioni?)
    jpeg.Encode(previewImg, dstImg, nil)
    return dstFile, nil
}
