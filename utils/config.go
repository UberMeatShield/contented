package utils
/**
* These are utilities for handling configuration settings for the app overall.  It sets up
* how content is discovered.  It will also centralize all the config lookup and managers loading
* in environment variables when running the full instance vs unit tests.
*/
import (
    "os"
    "errors"
    "strings"
    "strconv"
    "regexp"
    "log"
    "github.com/gobuffalo/envy"
)

// TODO: hard code vs Envy vs test stuff.  A pain in the butt
const sniffLen = 512  // How many bytes to read in a file when trying to determine mime type
const DefaultLimit int = 10000 // The max limit set by environment variable
const DefaultPreviewCount int = 8
const DefaultUseDatabase bool = false
const Default bool = false
const DefaultMaxSearchDepth int = 1
const DefaultMaxMediaPerContainer int = 90001    
const DefaultExcludeEmptyContainers bool = true

// Matchers that determine if you want to include specific filenames/content types
type MediaMatcher func(string, string) bool

func ExcludeNoFiles(filename string, content_type string) bool {
    return false
}

func IncludeAllFiles(filename string, content_type string) bool {
    return true
}

// Create a matcher that will check filename and content type and return true if it matches
// both in the case of AND matching (default) and true if either matches if matchType is OR
func CreateMatcher(filenameStrRE string, typesStrRE string, matchType string) MediaMatcher {
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

// TODO: this might be useful to add into the utils
type DirConfigEntry struct {
    Dir             string // The root of our loading (path to top level container directory)
    Limit           int    // The absolute max you can load in a single operation
    UseDatabase     bool   // Should it use the database or an in memory version
    CoreCount       int    // How many cores are likely available (used in creating multithread workers / previews)
    StaticResourcePath string  // The location where compiled js and css is hosted (container vs dev server)
    Initialized     bool   // Has the configuration actually be initialized properly

    // Config around creating preview images (used only by the task db:preview)
    PreviewCount    int    // How many files should be listed for a preview
    PreviewOverSize int64  // Over how many bytes should previews be created for the file
    PreviewVideoType string // This will be either gif or png based on config
    PreviewCreateFailIsFatal bool  // If set creating an image or movie preview will hard fail

    // Matchers that will determine which media elements to be included or excluded
    IncFiles MediaMatcher
    IncludeOperator string
    ExcFiles MediaMatcher
    ExcludeOperator string

    MaxSearchDepth  int    // When we search for data how far down the filesystem to search
    MaxMediaPerContainer  int    // When we search for data how far down the filesystem to search
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
       Initialized:  false,
       UseDatabase: true,
       Dir: "",
       CoreCount: 4,
       Limit: DefaultLimit,
       MaxSearchDepth: DefaultMaxSearchDepth,
       MaxMediaPerContainer: DefaultMaxMediaPerContainer,
       PreviewCount: DefaultPreviewCount,
       PreviewOverSize: 1024000,
       PreviewVideoType: "png",

       // Just grab all files by default
       IncFiles: IncludeAllFiles,
       IncludeOperator: "AND",
       ExcFiles: ExcludeNoFiles,
       ExcludeOperator: "AND",
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

    // There must be a cleaner way to do some of this default loading...
    excludeEmpty, emptyErr := strconv.ParseBool(envy.Get("EXCLUDE_EMPTY_CONTAINER", strconv.FormatBool(DefaultExcludeEmptyContainers)))
    maxSearchDepth, depthErr := strconv.Atoi(envy.Get("MAX_SEARCH_DEPTH", strconv.Itoa(DefaultMaxSearchDepth)))
    maxMediaPerContainer, medErr := strconv.Atoi(envy.Get("MAX_MEDIA_PER_CONTAINER", strconv.Itoa(DefaultMaxMediaPerContainer)))

    psize, perr := strconv.ParseInt(envy.Get("CREATE_PREVIEW_SIZE", "1024000"), 10, 64)
    previewType := envy.Get("PREVIEW_VIDEO_TYPE", "png")
    previewFailIsFatal, prevErr := strconv.ParseBool(envy.Get("PREVIEW_CREATE_FAIL_IS_FATAL", "false"))

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
    } else if medErr != nil {
        panic(medErr)
    } else if depthErr != nil {
        panic(depthErr)
    } else if emptyErr != nil {
        panic(emptyErr)
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
    cfg.PreviewCreateFailIsFatal = previewFailIsFatal
    cfg.IncludeOperator = envy.Get("INCLUDE_OPERATOR", "AND")
    cfg.ExcludeOperator = envy.Get("EXCLUDE_OPERATOR", "AND")

    cfg.ExcludeEmptyContainers = excludeEmpty
    cfg.MaxSearchDepth = maxSearchDepth
    cfg.MaxMediaPerContainer = maxMediaPerContainer
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
        cfg.IncFiles = CreateMatcher(y_fn, y_mime, cfg.IncludeOperator)
    } else {
        cfg.IncFiles = IncludeAllFiles
    }

    // If you do not specify exclusion regexes it will just include everything
    if n_fn != "" || n_mime != "" {
        cfg.ExcFiles = CreateMatcher(n_fn, n_mime, cfg.ExcludeOperator)
    } else {
        cfg.ExcFiles = ExcludeNoFiles
    }
}