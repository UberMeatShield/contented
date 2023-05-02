package utils

/**
 * These functions deal with using ffmpeg to encode video to new formats (h265)
 */
import (
    "errors"
    "log"
    "os"
    "regexp"
    "strings"
    "fmt"
    "path/filepath"
    "github.com/tidwall/gjson"
    "github.com/gofrs/uuid"
    ffmpeg "github.com/u2takey/ffmpeg-go"
    "contented/models"
)

// Used in the case of async processing when creating Preview results
type EncodingResult struct {
    C_ID    uuid.UUID
    MC_ID   uuid.UUID
    NewVideo string
    Err     error
}

type EncodingRequest struct {
    C    *models.Container
    Mc   *models.Content
    Out  chan EncodingResult
    Size int64
}

type EncodingWorker struct {
    Id int
    In chan EncodingRequest
}


func ShouldEncodeVideo(srcFile string, dstFile string) (string, error, bool) {
    filename := filepath.Base(srcFile)
    path := filepath.Dir(srcFile)
    contentType, tErr := GetMimeType(path, filename)
    if tErr != nil {
        return "", tErr, false
    }
    if !strings.Contains(contentType, "video") {
        msg := fmt.Sprintf("Not a video file so not converting %s", contentType)
        return msg, nil, false
    }
    vidInfo, err := ffmpeg.Probe(srcFile)
    if err != nil {
        log.Printf("Couldn't probe %s", srcFile)
        return "", err, false
    }
    // TODO: Optional config to overwrite?
    if _, err := os.Stat(dstFile); !os.IsNotExist(err) {
        existsMsg := fmt.Sprintf("Destination file already exists %s", dstFile)
        return "", errors.New(existsMsg), false
    }

    // Check that the converted file doesn't exist
    cfg := GetCfg()
    codecName := gjson.Get(vidInfo, "streams.0.codec_name").String()
    log.Printf("The vidinfo codec %s converting to %s", codecName, cfg.CodecForConversion)

    // The CodecForConversion is NOT the name of the thing you convert to...
    if cfg.CodecForConversion == codecName {
        okMsg := fmt.Sprintf("%s Already in the desired codec %s", srcFile, cfg.CodecForConversion)
        return okMsg, nil, false
    }

    matcher := regexp.MustCompile(cfg.CodecsToConvert)
    if !matcher.MatchString(codecName) {
        ignoreMsg := fmt.Sprintf("%s Not on the conversion list %s", srcFile, cfg.CodecsToConvert)
        return ignoreMsg, nil, false
    }
    ignore := regexp.MustCompile(cfg.CodecsToIgnore)
    if ignore.MatchString(codecName) {
        ignoreMsg := fmt.Sprintf("%s ignored because it matched %s", srcFile, cfg.CodecsToIgnore)
        return ignoreMsg, nil, false
    }
    msg := fmt.Sprintf("%s will be converted from %s to %s named %s", srcFile, codecName, cfg.CodecForConversion, filepath.Base(dstFile))
    // log.Printf(msg)
    return msg, nil, true
}

// Just rename to something with the previous extension stripped 
func GetVideoConversionName(srcFile string) string {
    path := filepath.Dir(srcFile)
    filename := filepath.Base(srcFile)
    ext := filepath.Ext(filename)
    stripExtension := regexp.MustCompile(fmt.Sprintf("%s$", ext))
    newFilename := stripExtension.ReplaceAllString(filename, "_h256.mp4")
    return filepath.Join(path, newFilename) 
}

// This will check if we should convert the source file then run the ffmpeg converter
// Returns 
//  - (msg: string) : What happened in human readable form
//  - (err: error) : did we hit a full error state
//  - (encoded: bool) : Did actual encoding take place vs just 'should not do it (ie: already encoded)'
func ConvertVideoToH256(srcFile string, dstFile string) (string, error, bool) {
    reason, err, shouldConvert := ShouldEncodeVideo(srcFile, dstFile)
    if shouldConvert == false {
        log.Printf("Not converting %s", reason)
        return reason, err, shouldConvert
    }
    log.Printf("About to convert %s", reason)

    // If codec in list of conversion codecs, then do it
    cfg := GetCfg()
    encode_err := ffmpeg.Input(srcFile).
        Output(dstFile, ffmpeg.KwArgs{"c:v": cfg.CodecForConversion, "vtag": "hvc1"}).
		OverWriteOutput().ErrorToStdOut().Run()
    if encode_err != nil {
        return "", err, false
    }
    return "Success: " + reason, nil, true
}
