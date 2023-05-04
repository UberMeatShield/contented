package utils

/**
* These are utilities for handling configuration settings for the app overall.  It sets up
* how content is discovered.  It will also centralize all the config lookup and managers loading
* in environment variables when running the full instance vs unit tests.
 */
import (
        "errors"
        "github.com/gobuffalo/envy"
        "log"
        "os"
        "regexp"
        "strconv"
        "strings"
)

// TODO: hard code vs Envy vs test stuff.  A pain in the butt
const sniffLen = 512           // How many bytes to read in a file when trying to determine mime type
const DefaultLimit int = 10000 // The max limit set by environment variable
const DefaultPreviewCount int = 8
const DefaultUseDatabase bool = false
const Default bool = false
const DefaultMaxSearchDepth int = 1
const DefaultMaxContentPerContainer int = 90001
const DefaultExcludeEmptyContainers bool = true
const DefeaultTotalScreens = 12
const DefaultPreviewFirstScreenOffset = 5
const DefaultCodecsToConvert = ".*"  // regex match
const DefaultCodecsToIgnore = "hevc" // regex match
const DefaultEncodingDestination = ""
const DefaultCodecForConversion = "libx265"

// Matchers that determine if you want to include specific filenames/content types
type ContentMatcher func(string, string) bool
type ContainerMatcher func(string) bool

func ExcludeNoFiles(filename string, content_type string) bool {
        return false
}

func ExcludeNoContainers(name string) bool {
        return false
}

func ExcludeContainerDefault(name string) bool {
        defaultCntExclude := regexp.MustCompile("DS_Store|container_previews")
        return defaultCntExclude.MatchString(name)
}

func IncludeAllFiles(filename string, content_type string) bool {
        return true
}

func IncludeAllContainers(name string) bool {
        return true
}

// Create a matcher that will check filename and content type and return true if it matches
// both in the case of AND matching (default) and true if either matches if matchType is OR
func CreateContentMatcher(filenameStrRE string, typesStrRE string, matchType string) ContentMatcher {
    filenameRE := regexp.MustCompile(filenameStrRE)
    typeRE := regexp.MustCompile(typesStrRE)

    if matchType == "OR" {
        return func(filename string, content_type string) bool {
                return filenameRE.MatchString(filename) || typeRE.MatchString(content_type)
        }
    }
    return func(filename string, content_type string) bool {
        return filenameRE.MatchString(filename) && typeRE.MatchString(content_type)
    }
}

func CreateContainerMatcher(filenameStrRE string) ContainerMatcher {
    cntRE := regexp.MustCompile(filenameStrRE)
    return func(cntName string) bool {
        return cntRE.MatchString(cntName)
    }
}

// TODO: this might be useful to add into the utils
type DirConfigEntry struct {
        Dir                string // The root of our loading (path to top level container directory)
        Limit              int    // The absolute max you can load in a single operation
        UseDatabase        bool   // Should it use the database or an in memory version
        CoreCount          int    // How many cores are likely available (used in creating multithread workers / previews)
        StaticResourcePath string // The location where compiled js and css is hosted (container vs dev server)
        Initialized        bool   // Has the configuration actually be initialized properly

        // Config around creating preview images (used only by the task db:preview)
        PreviewCount             int    // How many files should be listed for a preview
        PreviewOverSize          int64  // Over how many bytes should previews be created for the file
        ScreensOverSize   int64  // Over a certain size video select filters are slow
        PreviewVideoType         string // gif|screens|png are the video preview output type
        PreviewCreateFailIsFatal bool   // If set creating an image or movie preview will hard fail
        PreviewNumberOfScreens   int    // How many screens should be created to make the preview?
        PreviewFirstScreenOffset int    // Seconds to skip before taking a screen (black screen / titles)

        // Convertion script configuration
        CodecsToConvert string  // A matching regex for codecs to convert 
        CodecsToIgnore string   // Which codecs should not be converted (hevc is libx265 so default ignore)
        CodecForConversion string  // libx265 is the default and that makes hevc files
        EncodingDestination string // defaults to same directory but can override as well

        // Matchers that will determine which content elements to be included or excluded
        IncContent        ContentMatcher
        IncludeOperator string
        ExcContent        ContentMatcher
        ExcludeOperator string

        IncContainer ContainerMatcher
        ExcContainer ContainerMatcher

        MaxSearchDepth         int  // When we search for data how far down the filesystem to search
        MaxContentPerContainer   int  // When we search for data how far down the filesystem to search
        ExcludeEmptyContainers bool // If there is no content, should we list the container default true
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
                Initialized:              false,
                UseDatabase:              true,
                Dir:                      "",
                CoreCount:                4,
                Limit:                    DefaultLimit,
                MaxSearchDepth:           DefaultMaxSearchDepth,
                MaxContentPerContainer:     DefaultMaxContentPerContainer,
                PreviewCount:             DefaultPreviewCount,
                PreviewOverSize:          1024000,
                PreviewVideoType:         "png",
                ScreensOverSize:   50 * 1024000,
                PreviewNumberOfScreens:   DefeaultTotalScreens,
                PreviewFirstScreenOffset: DefaultPreviewFirstScreenOffset,

                // Conversion codecs 
                CodecsToConvert: DefaultCodecsToConvert,
                CodecsToIgnore: DefaultCodecsToIgnore,
                CodecForConversion: DefaultCodecForConversion,
                EncodingDestination: DefaultEncodingDestination,

                // Just grab all files by default
                IncContent:               IncludeAllFiles,
                IncludeOperator:        "AND",
                ExcContent:               ExcludeNoFiles,
                ExcludeOperator:        "AND",
                IncContainer:           IncludeAllContainers,
                ExcContainer:           ExcludeContainerDefault,
                ExcludeEmptyContainers: DefaultExcludeEmptyContainers,
        }
}

/*
 * Build out a valid configuration given the directory etc.
 *
 * Note we do not create a new instance, we are updating the overall app config.
 * TODO: Figure out how to do this "right" for a Buffalo app.
 */
func InitConfig(dir_root string, cfg *DirConfigEntry) *DirConfigEntry {
        cfg.Dir = dir_root // Always Common
        cfg.Initialized = true
        SetupContentMatchers(cfg, "", "", "", "")
        SetupContainerMatchers(cfg, "", "")
        return cfg
}

// Should I move this into the config itself?
func InitConfigEnvy(cfg *DirConfigEntry) *DirConfigEntry {
        var err error
        dir := envy.Get("DIR", "")
        if dir == "" {
                dir, err = envy.MustGet("CONTENT_DIR") // From the .env file
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

        // There must be a cleaner way to do some of this default loading...
        excludeEmpty, emptyErr := strconv.ParseBool(envy.Get("EXCLUDE_EMPTY_CONTAINER", strconv.FormatBool(DefaultExcludeEmptyContainers)))
        maxSearchDepth, depthErr := strconv.Atoi(envy.Get("MAX_SEARCH_DEPTH", strconv.Itoa(DefaultMaxSearchDepth)))
        maxContentPerContainer, medErr := strconv.Atoi(envy.Get("MAX_MEDIA_PER_CONTAINER", strconv.Itoa(DefaultMaxContentPerContainer)))

        psize, perr := strconv.ParseInt(envy.Get("CREATE_PREVIEW_SIZE", "1024000"), 10, 64)
        useSeekScreenSize, seekErr := strconv.ParseInt(envy.Get("SEEK_SCREEN_OVER_SIZE", "7168000"), 10, 64)

        previewType := envy.Get("PREVIEW_VIDEO_TYPE", "png")
        previewFailIsFatal, prevErr := strconv.ParseBool(envy.Get("PREVIEW_CREATE_FAIL_IS_FATAL", "false"))

        previewNumberOfScreens, totalScreenErr := strconv.Atoi(envy.Get("TOTAL_SCREENS", strconv.Itoa(DefeaultTotalScreens)))
        previewFirstScreenOffset, offsetErr := strconv.Atoi(envy.Get("FIRST_SCREEN_OFFSET", strconv.Itoa(DefaultPreviewFirstScreenOffset)))


        if err != nil {
                panic(err)
        } else if limErr != nil {
                panic(limErr)
        } else if previewErr != nil {
                panic(previewErr)
        } else if prevErr != nil {
                panic(prevErr)
        } else if _, noDirErr := os.Stat(dir); os.IsNotExist(noDirErr) {
                panic(noDirErr)
        } else if connErr != nil {
                panic(connErr)
        } else if coreErr != nil {
                panic(coreErr)
        } else if perr != nil {
                panic(perr)
        } else if seekErr != nil {
                panic(seekErr)
        } else if medErr != nil {
                panic(medErr)
        } else if depthErr != nil {
                panic(depthErr)
        } else if emptyErr != nil {
                panic(emptyErr)
        } else if (totalScreenErr) != nil {
                panic(totalScreenErr)
        } else if (offsetErr) != nil {
                panic(offsetErr)
        }

        if !(previewType == "png" || previewType == "gif" || previewType == "screens") {
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
        cfg.ScreensOverSize = useSeekScreenSize
        cfg.PreviewCreateFailIsFatal = previewFailIsFatal
        cfg.PreviewNumberOfScreens = previewNumberOfScreens
        cfg.PreviewFirstScreenOffset = previewFirstScreenOffset

        cfg.IncludeOperator = envy.Get("INCLUDE_OPERATOR", "AND")
        cfg.ExcludeOperator = envy.Get("EXCLUDE_OPERATOR", "AND")

        // For video conversion options (eventually)
        cfg.CodecsToConvert = envy.Get("CODECS_TO_CONVERT", DefaultCodecsToConvert)
        cfg.CodecsToIgnore = envy.Get("CODECS_TO_IGNORE", DefaultCodecsToIgnore)
        cfg.CodecForConversion = envy.Get("CODEC_FOR_CONVERSION", DefaultCodecForConversion)
        cfg.EncodingDestination = envy.Get("ENCODING_DESTINATION", DefaultEncodingDestination)

        cfg.ExcludeEmptyContainers = excludeEmpty
        cfg.MaxSearchDepth = maxSearchDepth
        cfg.MaxContentPerContainer = maxContentPerContainer

        SetupContentMatchers(
                cfg,
                envy.Get("INCLUDE_MEDIA_MATCH", ""),
                envy.Get("INCLUDE_TYPES_MATCH", ""),
                envy.Get("EXCLUDE_MEDIA_MATCH", ""),
                envy.Get("EXCLUDE_TYPES_MATCH", ""),
        )
        SetupContainerMatchers(
                cfg,
                envy.Get("INCLUDE_CONTAINER_MATCH", ""),
                envy.Get("EXCLUDE_CONTAINER_MATCH", ""),
        )
        cfg.Initialized = true
        return cfg
}

// Setup the matchers on the configuration, these are used to determine which content elments should match
// yes filename matches, yes mime matches, no if the filename matches, no if the mime matches.
func SetupContentMatchers(cfg *DirConfigEntry, y_fn string, y_mime string, n_fn string, n_mime string) {

    //To include content only if it matches the filename or mime type
    if y_fn != "" || y_mime != "" {
            cfg.IncContent = CreateContentMatcher(y_fn, y_mime, cfg.IncludeOperator)
    } else {
            cfg.IncContent = IncludeAllFiles
    }

    // If you do not specify exclusion regexes it will just include everything
    if n_fn != "" || n_mime != "" {
            cfg.ExcContent = CreateContentMatcher(n_fn, n_mime, cfg.ExcludeOperator)
    } else {
            cfg.ExcContent = ExcludeNoFiles
    }
}

func SetupContainerMatchers(cfg *DirConfigEntry, y_cnt string, n_cnt string) {
    if y_cnt != "" {
            cfg.IncContainer = CreateContainerMatcher(y_cnt)
    } else {
            cfg.IncContainer = IncludeAllContainers
    }
    if n_cnt != "" {
            cfg.ExcContainer = CreateContainerMatcher(n_cnt)
    } else {
            // ExcludeNoContainers is not used because we really don't want .DS_Store and previews
            cfg.ExcContainer = ExcludeContainerDefault
    }
}
