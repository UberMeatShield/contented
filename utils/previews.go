package utils

import (
    "os"
    "log"
    //"image"
    "image/jpeg"
    //"image/png"
    "path/filepath"
    "github.com/nfnt/resize"
)

func MakePreviewPath(dstPath string) error {
    if _, err := os.Stat(dstPath); os.IsNotExist(err) {
        return os.MkdirAll(dstPath, 0740) 
    }
    return nil
}

// TODO: Only run if image is over a certain size (wrapper method)

// Make sure dstPath already exists before you call this
func GetImagePreview(path string, filename string, dstPath string) (string, error) {

    // Does the preview already exist, return that
    /*
    dstFile := filepath.Join(dstPath, "preview_" + filename)
    if _, e_err := os.Stat(dstFile); os.IsNotExist(e_err) {
        return dstFile, nil
    }
    */

    // The file we are going to check about making a preview of
    fqFile := filepath.Join(path, filename)    
    srcImg, err := os.Open(fqFile)
    if err != nil {
        log.Fatal(err)
        return "Error Generating Preview", err
    }
    defer srcImg.Close()

    img, dErr := jpeg.Decode(srcImg)
    if dErr != nil {
        log.Fatal(dErr)
    }
    dstImg := resize.Resize(640, 0, img, resize.Lanczos3)

    dstFile := filepath.Join(dstPath, "preview_" + filename)
    previewImg, errCreate := os.Create(dstFile)

    if errCreate != nil {
        log.Fatal(errCreate)
        return "Error" + dstFile, errCreate
    }
    defer previewImg.Close()

    jpeg.Encode(previewImg, dstImg, nil)
    return dstFile, nil
}

/*

func main() {
    imagePath, _ := os.Open("jellyfish.jpg")
    defer imagePath.Close()
    srcImage, _, _ := image.Decode(imagePath)

    // Dimension of new thumbnail 80 X 80
    dstImage := image.NewRGBA(image.Rect(0, 0, 80, 80))
    // Thumbnail function of Graphics
    graphics.Thumbnail(dstImage, srcImage)

    newImage, _ := os.Create("thumbnail.jpg")
    defer newImage.Close()
    jpeg.Encode(newImage, dstImage, &jpeg.Options{jpeg.DefaultQuality})
}
*/
