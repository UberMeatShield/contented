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

// Used in the case of async processing when encoding.
// Pretty much identical to previews but I might have to tweak this a lot
type EncodingResult struct {
    C_ID    uuid.UUID
    MC_ID   uuid.UUID
    NewVideo string
    Err     error

    // We should ensure that there are some stats gathered around this.
    InitialSize int64
    EncodedSize int64
}

type EncodingRequest struct {
    C    *models.Container
    Mc   *models.Content
    SrcFile string
    DstFile string
    Out  chan EncodingResult
}

type EncodingWorker struct {
    Id int
    In chan EncodingRequest
}
type EncodingRequests []EncodingRequest
type EncodingResults []EncodingResult


// Returns a message about the video codec, size of it, or if it is invalid.
func IsValidVideo(srcFile string) (string, int64, error) {
    srcStat, statErr := os.Stat(srcFile)
    fileSize := int64(0)
    codecName := ""
    if statErr != nil {
        return codecName, fileSize, statErr
    }
    fileSize = int64(srcStat.Size())
    if fileSize == 0 {
        return codecName, fileSize, errors.New(fmt.Sprintf("%s was empty or small %d", srcFile, fileSize))
    }

    filename := filepath.Base(srcFile)
    path := filepath.Dir(srcFile)
    contentType, tErr := GetMimeType(path, filename)
    if tErr != nil {
        return codecName, fileSize, tErr
    }
    if !strings.Contains(contentType, "video") {
        msg := fmt.Sprintf("Not a video file so not converting %s", contentType)
        return codecName, fileSize, errors.New(msg)
    }
    vidInfo, probeErr := ffmpeg.Probe(srcFile)
    if probeErr != nil {
        log.Printf("Couldn't probe %s", srcFile)
        return codecName, fileSize, probeErr
    }
    codecName = gjson.Get(vidInfo, "streams.0.codec_name").String()
    return codecName, fileSize, nil
}

/**
 *  This returns a message about what is happening with the encoding (should it do it etc).
 */
func ShouldEncodeVideo(srcFile string, dstFile string) (string, error, bool) {
    codecName, _, videoInvalid := IsValidVideo(srcFile)
    if videoInvalid != nil {
        return fmt.Sprintf("Not valid video %s", videoInvalid), nil, false
    }
    _, statErr := os.Stat(dstFile)
    if !os.IsNotExist(statErr) {
        existsMsg := fmt.Sprintf("Destination file already exists %s", dstFile)
        return "", errors.New(existsMsg), false
    }
    // TODO: Config setting where if the filesize is too small we should _NOT_ reencode

    // Check that the converted file doesn't exist
    cfg := GetCfg()
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

    msg := fmt.Sprintf("File will be converted from %s to %s\nOld File: %s\nNew File: %s", codecName, cfg.CodecForConversion, srcFile, dstFile)
    return msg, nil, true
}

// Just rename to something with the previous extension stripped 
func GetVideoConversionName(srcFile string) string {
    path := filepath.Dir(srcFile)
    filename := filepath.Base(srcFile)
    ext := filepath.Ext(filename)

    cfg := GetCfg()
    if cfg.EncodingDestination != "" {
        path = cfg.EncodingDestination
    }
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

    // If codec in list of conversion codecs, then do it
    cfg := GetCfg()
    log.Printf("About to convert %s to codec %s", reason, cfg.CodecForConversion)
    encode_err := ffmpeg.Input(srcFile).
        Output(dstFile, ffmpeg.KwArgs{"c:v": cfg.CodecForConversion, "tag:v": "hvc1"}).
		OverWriteOutput().ErrorToStdOut().Run()

    if encode_err != nil {
        log.Printf("Encoding error when actually running ffmpeg  %s", encode_err)
        return "", encode_err, false
    }
    return "Success: " + reason, nil, true
}
