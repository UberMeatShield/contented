package config

/**
* These are utilities for handling configuration settings for the app overall.  It sets up
* how content is discovered.  It will also centralize all the config lookup and managers loading
* in environment variables when running the full instance vs unit tests.
 */
import (
	"log"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// TODO: hard code vs Envy vs test stuff.  A pain in the butt
const SniffLen = 512           // How many bytes to read in a file when trying to determine mime type
const DefaultLimit int = 10000 // The max limit set by environment variable
const DefaultPreviewCount int = 8
const DefaultUseDatabase bool = false
const DefaultMaxSearchDepth int = 1
const DefaultMaxContentPerContainer = int64(90001)
const DefaultExcludeEmptyContainers bool = true
const DefeaultTotalScreens = 12
const DefaultPreviewFirstScreenOffset = 5
const DefaultCodecsToConvert = ".*"          // regex match
const DefaultCodecsToIgnore = "hevc"         // regex match (do not try and double encode)
const DefaultCodecForConversionName = "hevc" // The name of the encoding (sometimes not the lib)
const DefaultEncodingDestination = ""
const DefaultCodecForConversion = "libx265"
const DefaultEncodingFilenameModifier = "_h265" // This is used when encoding a new video file name <name>_h265.mp4

var ValidPreviewTypes = []string{"png", "gif", "screens"}

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
	StaticLibraryPath  string // Library includes (monaco just doesn't want to build in)
	ReadOnly           bool   // Can you edit content on this site
	TagFile            string // A file to use in populating tags
	Initialized        bool   // Has the configuration actually be initialized properly

	// Config around creating preview images (used only by the task db:preview)
	PreviewCount             int    // How many files should be listed for a preview
	PreviewOverSize          int64  // Over how many bytes should previews be created for the file
	ScreensOverSize          int64  // Over a certain size video select filters are slow
	PreviewVideoType         string // gif|screens|png are the video preview output type
	PreviewCreateFailIsFatal bool   // If set creating an image or movie preview will hard fail
	PreviewNumberOfScreens   int    // How many screens should be created to make the preview?
	PreviewFirstScreenOffset int    // Seconds to skip before taking a screen (black screen / titles)

	// Convertion script configuration
	CodecsToConvert        string // A matching regex for codecs to convert
	CodecsToIgnore         string // Which codecs should not be converted (hevc is libx265 so default ignore)
	CodecForConversion     string // libx265 is the default and that makes hevc files
	CodecForConversionName string // ie libx265 becomes hevc
	EncodingDestination    string // defaults to same directory but can override as well

	// TODO: Handle it being mp4?
	EncodingFilenameModifier string // After re-encoding filename<EncodingFilenameModifier>.mp4
	RemoveDuplicateFiles     bool   // Removing old video files after re-encoding

	StartQueueWorkers bool // Should we process requested tasks on this server

	// Matchers that will determine which content elements to be included or excluded
	IncContent      ContentMatcher
	IncludeOperator string
	ExcContent      ContentMatcher
	ExcludeOperator string

	IncContainer ContainerMatcher
	ExcContainer ContainerMatcher

	MaxSearchDepth         int   // When we search for data how far down the filesystem to search
	MaxContentPerContainer int64 // When we search for data how far down the filesystem to search
	ExcludeEmptyContainers bool  // If there is no content, should we list the container default true

	// Splash Endpoint configuration for the 'home' page
	SplashContainerName string // A Container you would like to load and send back
	SplashContentID     string // A Content ID you would like to load and send back
	SplashRendererType  string // video view | container | content | editor ?

	SplashTitle       string // The title string for the html file
	SplashHtmlFile    string // A fq path to a splash file to render into the page
	SplashContentHTML string // Raw HTML you would like to render in

	// TODO: Implement
	SplashContentFile string // A full file path on the host (raw HTML?)
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
		MaxContentPerContainer:   DefaultMaxContentPerContainer,
		PreviewCount:             DefaultPreviewCount,
		PreviewOverSize:          1024000,
		PreviewVideoType:         "png",
		ScreensOverSize:          50 * 1024000,
		PreviewNumberOfScreens:   DefeaultTotalScreens,
		PreviewFirstScreenOffset: DefaultPreviewFirstScreenOffset,

		// Conversion codecs
		CodecsToConvert:        DefaultCodecsToConvert,
		CodecsToIgnore:         DefaultCodecsToIgnore,
		CodecForConversion:     DefaultCodecForConversion,
		CodecForConversionName: DefaultCodecForConversionName,
		EncodingDestination:    DefaultEncodingDestination,

		EncodingFilenameModifier: DefaultEncodingFilenameModifier,
		RemoveDuplicateFiles:     false,

		// Should this server start up processing tasks for tasking screens, encoding etc.
		StartQueueWorkers: true,

		// Just grab all files by default
		IncContent:             IncludeAllFiles,
		IncludeOperator:        "AND",
		ExcContent:             ExcludeNoFiles,
		ExcludeOperator:        "AND",
		IncContainer:           IncludeAllContainers,
		ExcContainer:           ExcludeContainerDefault,
		ExcludeEmptyContainers: DefaultExcludeEmptyContainers,

		SplashContainerName: "",
		SplashContentID:     "",
		SplashContentHTML:   "",
		SplashRendererType:  "",
		SplashHtmlFile:      "",
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

func MustGetEnvString(key string) string {
	val := GetEnvString(key, "")
	if val == "" {
		log.Fatalf("failed to find key %s", key)
	}
	return val
}

// These need to be valid or we will bail the app
func GetEnvString(key string, defaultVal string) string {
	valStr := os.Getenv(key)
	if valStr != "" {
		return valStr
	}
	return defaultVal
}

func GetEnvBool(key string, defaultBool bool) bool {
	valStr := os.Getenv(key)
	if valStr != "" {
		val, err := strconv.ParseBool(valStr)
		if err != nil {
			log.Fatalf("Failed to parse boolean key(%s) val (%s) err %s", key, valStr, err)
		}
		return val
	}
	return defaultBool
}

func GetEnvInt64(key string, defaultInt int64) int64 {
	valStr := os.Getenv(key)
	if valStr != "" {
		val, err := strconv.ParseInt(valStr, 10, 64)
		if err != nil {
			log.Fatalf("Failed to parse Int64 key(%s) val (%s) err %s", key, valStr, err)
		}
		return val
	}
	return defaultInt
}

func GetEnvInt(key string, defaultInt int) int {
	valStr := os.Getenv(key)
	if valStr != "" {
		val, err := strconv.Atoi(valStr)
		if err != nil {
			log.Fatalf("Failed to parse Int key(%s) val (%s) err %s", key, valStr, err)
		}
		return val
	}
	return defaultInt
}

// Should I move this into the config itself?
func InitConfigEnvy(cfg *DirConfigEntry) *DirConfigEntry {

	envErr := godotenv.Load()
	if envErr != nil {
		log.Printf("No .env file found if this is not running a service this is probably fatal")
	}

	dir := GetEnvString("DIR", "")
	if dir == "" {
		dir = GetEnvString("CONTENT_DIR", "") // From the .env file
		if dir == "" {
			log.Fatalf("No DIR environment variable or CONTENT_DIR in the .env found")
		}
	}
	if !strings.HasSuffix(dir, "/") {
		dir = dir + "/"
	}

	if _, noDirErr := os.Stat(dir); os.IsNotExist(noDirErr) {
		log.Fatalf("Failed to find directory %s error %s", dir, noDirErr)
	}

	cfg.Dir = dir
	cfg.UseDatabase = GetEnvBool("USE_DATABASE", DefaultUseDatabase)
	cfg.StaticResourcePath = GetEnvString("STATIC_RESOURCE_PATH", "./public/build")
	cfg.StaticLibraryPath = GetEnvString("STATIC_LIBRARY_PATH", "./public/static")
	cfg.TagFile = GetEnvString("TAG_FILE", "")

	cfg.Limit = GetEnvInt("LIMIT", DefaultLimit)
	cfg.CoreCount = GetEnvInt("CORE_COUNT", 4)
	cfg.StartQueueWorkers = GetEnvBool("START_QUEUE_WORKERS", true)
	cfg.PreviewCount = GetEnvInt("PREVIEW", DefaultPreviewCount)

	// Could make this an enum?
	cfg.PreviewVideoType = GetEnvString("PREVIEW_VIDEO_TYPE", "png")
	if !(slices.Contains(ValidPreviewTypes, cfg.PreviewVideoType)) {
		log.Fatalf("the video preview type is not png or gif %s", cfg.PreviewVideoType)
	}

	cfg.PreviewOverSize = GetEnvInt64("CREATE_PREVIEW_SIZE", int64(1024000))
	cfg.ScreensOverSize = GetEnvInt64("SEEK_SCREEN_OVER_SIZE", int64(7168000))
	cfg.PreviewCreateFailIsFatal = GetEnvBool("PREVIEW_CREATE_FAIL_IS_FATAL", false)
	cfg.PreviewNumberOfScreens = GetEnvInt("TOTAL_SCREENS", DefeaultTotalScreens)
	cfg.PreviewFirstScreenOffset = GetEnvInt("FIRST_SCREEN_OFFSET", DefaultPreviewFirstScreenOffset)

	cfg.ReadOnly = GetEnvBool("READ_ONLY", false)
	cfg.IncludeOperator = GetEnvString("INCLUDE_OPERATOR", "AND")
	cfg.ExcludeOperator = GetEnvString("EXCLUDE_OPERATOR", "AND")

	// For video conversion options (eventually)
	cfg.CodecsToConvert = GetEnvString("CODECS_TO_CONVERT", DefaultCodecsToConvert)
	cfg.CodecsToIgnore = GetEnvString("CODECS_TO_IGNORE", DefaultCodecsToIgnore)
	cfg.CodecForConversion = GetEnvString("CODEC_FOR_CONVERSION", DefaultCodecForConversion)
	cfg.CodecForConversionName = GetEnvString("CODEC_FOR_CONVERSION_NAME", DefaultCodecForConversionName)
	cfg.EncodingDestination = GetEnvString("ENCODING_DESTINATION", DefaultEncodingDestination)

	// TODO: Make this a little saner on the name side
	cfg.EncodingFilenameModifier = GetEnvString("ENCODING_FILENAME_MODIFIER", DefaultEncodingFilenameModifier)
	cfg.RemoveDuplicateFiles = GetEnvBool("REMOVE_DUPLICATE_FILES", false)

	// There must be a cleaner way to do some of this default loading...

	cfg.ExcludeEmptyContainers = GetEnvBool("EXCLUDE_EMPTY_CONTAINER", DefaultExcludeEmptyContainers)
	cfg.MaxSearchDepth = GetEnvInt("MAX_SEARCH_DEPTH", DefaultMaxSearchDepth)
	cfg.MaxContentPerContainer = GetEnvInt64("MAX_MEDIA_PER_CONTAINER", DefaultMaxContentPerContainer)

	cfg.SplashContainerName = GetEnvString("SPLASH_CONTAINER_NAME", "")
	cfg.SplashContentID = GetEnvString("SPLASH_CONTENT_ID", "")
	cfg.SplashRendererType = GetEnvString("SPLASH_RENDERER_TYPE", "")
	cfg.SplashHtmlFile = GetEnvString("SPLASH_HTML_FILE", "")
	cfg.SplashTitle = GetEnvString("SPLASH_TITLE", "")

	SetupContentMatchers(
		cfg,
		GetEnvString("INCLUDE_MEDIA_MATCH", ""),
		GetEnvString("INCLUDE_TYPES_MATCH", ""),
		GetEnvString("EXCLUDE_MEDIA_MATCH", ""),
		GetEnvString("EXCLUDE_TYPES_MATCH", ""),
	)
	SetupContainerMatchers(
		cfg,
		GetEnvString("INCLUDE_CONTAINER_MATCH", ""),
		GetEnvString("EXCLUDE_CONTAINER_MATCH", ""),
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
