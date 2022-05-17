package managers

import (
	"contented/internals"
	"contented/models"
	"contented/utils"
	"github.com/gobuffalo/envy"
	"github.com/gobuffalo/packr/v2"
	"github.com/gobuffalo/pop/v5"
	"github.com/gobuffalo/suite/v3"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"testing"
)

var expect_len = map[string]int{
	"dir1":            12,
	"dir2":            3,
	"dir3":            8,
	"screens":         4,
	"screens_sub_dir": 2,
}

func GetManagerActionSuite(cfg *utils.DirConfigEntry, as *ActionSuite) ContentManager {
	ctx := internals.GetContext(as.App)
	get_params := func() *url.Values {
		vals := ctx.Params().(url.Values)
		return &vals
	}
	get_conn := func() *pop.Connection {
		// as.DB should work, but it is of a type pop.v5.Connection instead of pop.Connection
		return models.DB
	}
	return CreateManager(cfg, get_conn, get_params)
}

// Why are no tests working?
func TestMain(m *testing.M) {
	_, err := envy.MustGet("DIR")
	if err != nil {
		log.Println("DIR ENV REQUIRED$ export=DIR=`pwd`/mocks/content/ && buffalo test")
		panic(err)
	}
	code := m.Run()
	os.Exit(code)
}

func Test_ManagerSuite(t *testing.T) {
	app := internals.CreateBuffaloApp(true, "test")
	action, err := suite.NewActionWithFixtures(app, packr.New("Test_ManagerSuite", "../fixtures"))
	if err != nil {
		t.Fatal(err)
	}
	as := &ActionSuite{
		Action: action,
	}
	suite.Run(t, as)
}

func (as *ActionSuite) Test_ManagerContainers() {
	internals.InitFakeApp(false)
	ctx := internals.GetContext(as.App)
	man := GetManager(&ctx)
	containers, err := man.ListContainersContext()
	as.NoError(err)

	for _, c := range *containers {
		c_mem, err := man.GetContainer(c.ID)
		if err != nil {
			as.Fail("It should not have an issue finding valid containers")
		}
		as.Equal(c_mem.ID, c.ID)
	}
}

func (as *ActionSuite) Test_ManagerMediaContainer() {
	internals.InitFakeApp(false)
	ctx := internals.GetContext(as.App)
	man := GetManager(&ctx)
	mcs, err := man.ListAllMedia(1, 9001)
	as.NoError(err)

	for _, mc := range *mcs {
		cm, err := man.FindFileRef(mc.ID)
		if err != nil {
			as.Fail("It should not have an issue finding valid containers")
		}
		as.Equal(cm.ID, mc.ID)
	}
}

func (as *ActionSuite) Test_AssignManager() {
	cfg := internals.ResetConfig()
	cfg.UseDatabase = false
	utils.InitConfig(cfg.Dir, cfg)

	mem := ContentManagerMemory{}
	mem.validate = "Memory"
	mem.SetCfg(cfg)
	mem.Initialize()

	memCfg := mem.GetCfg()
	as.NotNil(memCfg, "It should be defined")
	mcs, err := mem.ListAllMedia(1, 9001)
	as.NoError(err)
	as.Greater(len(*mcs), 0, "It should have valid files in the manager")

	cfg.UseDatabase = false
	ctx := internals.GetContext(as.App)
	man := GetManager(&ctx) // New Reference but should have the same count of media
	mcs_2, _ := man.ListAllMedia(1, 9000)

	as.Equal(len(*mcs), len(*mcs_2), "A new instance should use the same storage")
}

func (as *ActionSuite) Test_MemoryManagerPaginate() {
	cfg := internals.InitFakeApp(false)
	cfg.UseDatabase = false

	ctx := internals.GetContextParams(as.App, "/containers", "1", "2")
	man := GetManager(&ctx)
	as.Equal(man.CanEdit(), false, "Memory manager should not allow editing")

	containers, err := man.ListContainers(1, 1)
	as.NoError(err, "It should list with pagination")
	as.Equal(1, len(*containers), "It should respect paging")

	cnt := (*containers)[0]
	as.NotNil(cnt, "There should be a container with 12 entries")
	as.Equal(cnt.Total, 12, "There should be 12 test images in the first ORDERED containers")
	as.NoError(err)
	as.NotEqual("", cnt.PreviewUrl, "The previewUrl should be set")
	media_page_1, _ := man.ListMedia(cnt.ID, 1, 4)
	as.Equal(len(*media_page_1), 4, "It should respect page size")

	media_page_3, _ := man.ListMedia(cnt.ID, 3, 4)
	as.Equal(len(*media_page_3), 4, "It should respect page size and get the last page")
	as.NotEqual((*media_page_3)[3].ID, (*media_page_1)[3].ID, "Ensure it actually paged")

	// Last container pagination check
	l_cnts, _ := man.ListContainers(4, 1)
	as.Equal(1, len(*l_cnts), "It should still return only as we are on the last page")
	l_cnt := (*l_cnts)[0]
	as.Equal(expect_len[l_cnt.Name], l_cnt.Total, "There are 3 entries in the ordered test data last container")
}

func (as *ActionSuite) Test_ManagerInitialize() {
	internals.InitFakeApp(false)

	ctx := internals.GetContext(as.App)
	man := GetManager(&ctx)
	as.NotNil(man, "It should have a manager defined after init")

	containers, err := man.ListContainersContext()
	as.NoError(err, "It should list all containers")
	as.NotNil(containers, "It should have containers")
	as.Equal(len(*containers), internals.TOTAL_CONTAINERS, "Unexpected container count")

	// Memory test working
	for _, c := range *containers {
		// fmt.Printf("Searching for this container %s with name %s\n", c.ID, c.Name)
		media, err := man.ListMediaContext(c.ID)
		as.NoError(err)
		as.NotNil(media)

		media_len := len(*media)
		// fmt.Printf("Media length was %d\n", media_len)
		as.Greater(media_len, 0, "There should be a number of media")
		as.Equal(expect_len[c.Name], media_len, "It should have this many instances: "+c.Name)
		as.Greater(c.Total, 0, "All of them should have a total assigned")
	}
}

func (as *ActionSuite) Test_MemoryManagerSearch() {
	internals.InitFakeApp(false)

	ctx := internals.GetContext(as.App)
	man := GetManager(&ctx)
	as.NotNil(man, "It should have a manager defined after init")

	containers, err := man.ListContainersContext()
	as.NoError(err, "It should list all containers")
	as.NotNil(containers, "It should have containers")
	as.Equal(len(*containers), internals.TOTAL_CONTAINERS, "Wrong number of containers found")

	mcs, total, err := man.SearchMedia("donut", 1, 20, "")
	as.NoError(err, "Can we search in the memory manager")
	as.Equal(len(*mcs), 1, "One donut should be found")
	as.Equal(total, len(*mcs), "It should get the total right")

	mcs_1, _, err_1 := man.SearchMedia("Large", 1, 6, "")
	as.NoError(err_1, "Can we search in the memory manager")
	as.Equal(3, len(*mcs_1), "One donut should be found")

	all_mc, _, err_all := man.SearchMedia("", 0, 9000, "")
	as.NoError(err_all, "Can in search everything")
	as.Equal(len(*all_mc), internals.TOTAL_MEDIA, "The Kitchen sink")
}

func (as *ActionSuite) Test_MemoryManagerSearchMulti() {
	// Test that a search restricting containerID works
	// Test that search restricting container and text works
	internals.InitFakeApp(false)
	//man := GetManagerActionSuite(cfg, as)
	ctx := internals.GetContext(as.App)
	man := GetManager(&ctx)

	// Ensure we initialized with a known search
	as.Equal(man.CanEdit(), false)
	mcs, total, err := man.SearchMedia("donut", 1, 20, "")
	as.NoError(err, "Can we search in the memory manager")
	as.Equal(len(*mcs), 1, "One donut should be found")
	as.Equal(total, len(*mcs), "It should get the total right")

	cnts, eep := man.ListContainers(0, 10)
	as.NoError(eep, "It should have 4 containers")
	as.Greater(len(*cnts), 1, "We should have containers")

	allMedia, errAll := man.ListAllMedia(0, 50)
	as.Greater(len(*allMedia), 0, "We should have media")
	as.NoError(errAll)

	all_media, wild_total, _ := man.SearchMedia("", 0, 40, "")
	as.Greater(wild_total, 0)
	as.Equal(len(*all_media), wild_total)

	for _, cnt := range *cnts {
		if cnt.Name == "dir1" {
			_, no_total, n_err := man.SearchMedia("donut", 1, 20, cnt.ID.String())
			as.NoError(n_err)
			as.Equal(no_total, 0, "It should not be in this directory")
		}
		if cnt.Name == "dir2" {
			yes_match, y_total, r_err := man.SearchMedia("donut", 1, 20, cnt.ID.String())
			as.NoError(r_err)
			as.Equal(y_total, 1, "We did not find the expected media")

			movie := (*yes_match)[0]
			as.Equal(movie.Src, "donut.mp4")
		}
		if cnt.Name == "dir3" {
			has_media, _, err := man.SearchMedia("", 0, 1, cnt.ID.String())
			as.NoError(err, "We should have media")
			as.Greater(len(*has_media), 0)
		}
	}
}

// A basic DB search (ilike matching)
func (as *ActionSuite) Test_DbManagerSearch() {
	models.DB.TruncateAll()
	cfg := internals.InitFakeApp(true)

	man := GetManagerActionSuite(cfg, as)
	as.Equal(man.CanEdit(), true, "It should be a DB manager")

	cnt, media := internals.GetMediaByDirName("dir1")
	c_err := models.DB.Create(cnt)
	as.NoError(c_err)
	for _, mc := range media {
		models.DB.Create(&mc)
	}
	mcs, _, err := man.SearchMedia("Large", 1, 20, "")
	as.NoError(err, "It should be able to search")
	as.NotNil(mcs, "It should be")
	as.Equal(3, len(*mcs), "We should have 3 large images with an ilike compare")

	all_mcs, total, err := man.SearchMedia("", 1, 10, "")
	as.NoError(err, "It should be able to empty search")
	as.Equal(12, total, "The total count for this dir is 12")
	as.Equal(10, len(*all_mcs), "But we limited the pagination")
}

func (as *ActionSuite) Test_DbManagerSearchMulti() {
	// Test that a search restricting containerID works
	// Test that search restricting container and text works
	models.DB.TruncateAll()
	cfg := internals.InitFakeApp(true)

	man := GetManagerActionSuite(cfg, as)
	as.Equal(man.CanEdit(), true)

	cnt1, media1, err1 := internals.CreateMediaByDirName("dir1")
	as.NoError(err1)
	as.Greater(len(media1), 1)

	cnt2, media2, err2 := internals.CreateMediaByDirName("dir2")
	as.NoError(err2)
	as.Greater(len(media2), 1)

	found, count, err := man.SearchMedia(media1[1].Src, 0, 10, cnt1.ID.String())
	as.Equal(len(*found), 1, "We should have found our item")
	as.Equal(count, 1)
	as.NoError(err)

	_, n_count, n_err := man.SearchMedia("blah", 0, 10, cnt1.ID.String())
	as.Equal(n_count, 0, "It should not find this the media name is invalid")
	as.NoError(n_err)

	_, not_in_cnt_count, not_err := man.SearchMedia(media1[1].Src, 0, 10, cnt2.ID.String())
	as.Equal(not_in_cnt_count, 0, "It should not find this valid media as it is not in the container")
	as.NoError(not_err)
}

func (as *ActionSuite) Test_MemoryPreviewInitialization() {
	cfg := internals.ResetConfig()
	cfg.MaxSearchDepth = 1
	utils.SetupMediaMatchers(cfg, "", "video", "DS_Store", "")
	utils.SetCfg(*cfg)

	// Create a fake file that would sub in by name for a preview
	var testDir, _ = envy.MustGet("DIR")
	srcDir := filepath.Join(testDir, "dir2")
	dstDir := utils.GetPreviewDst(srcDir)
	testFile := "donut.mp4"

	// Create a fake preview
	utils.ResetPreviewDir(dstDir)

	fqPath := utils.GetPreviewPathDestination(testFile, dstDir, "video/mp4")
	f, err := os.Create(fqPath)
	if err != nil {
		as.T().Errorf("Could not create the file at %s", fqPath)
	}
	_, wErr := f.WriteString("Now something exists in the file")
	if wErr != nil {
		as.T().Errorf("Could not write to the file at %s", fqPath)
	}
	as.Contains(fqPath, "donut.mp4.png")
	f.Sync()

	// Checks that if a preview exists
	cnts, media := utils.PopulateMemoryView(cfg.Dir)
	as.Equal(1, len(cnts), "We should only pull in containers that have media")
	as.Equal(len(media), 1, "But there is only one video by mime type")
	for _, mc := range media {
		as.Equal("/container_previews/donut.mp4.png", mc.Preview)
	}

	cfg.ExcludeEmptyContainers = false
	all_cnts, one_media := utils.PopulateMemoryView(cfg.Dir)
	as.Equal(1, len(one_media), "But there is only one video by mime type")

	as.Equal(internals.TOTAL_CONTAINERS, len(all_cnts), "Allow it to pull in all containers")
}

func (as *ActionSuite) Test_ManagerDB() {
	models.DB.TruncateAll()
	cfg := internals.InitFakeApp(true)

	cnt, media := internals.GetMediaByDirName("dir1")
	as.Equal("dir1", cnt.Name, "It should be the right dir")
	as.Equal(12, cnt.Total, "The container total should be this for dir1")
	as.Equal(12, len(media))

	c_err := models.DB.Create(cnt)
	as.NoError(c_err)
	for _, mc := range media {
		models.DB.Create(&mc)
	}

	man := GetManagerActionSuite(cfg, as)
	q_media, err := man.ListAllMedia(0, 14)
	as.NoError(err, "We should be able to list")
	as.Equal(len(*q_media), 12, "there should be 12 results")

	lim_media, _ := man.ListAllMedia(0, 3)
	as.Equal(3, len(*lim_media), "The DB should be setup with 10 items")
}


func (as *ActionSuite) Test_ManagerDBPreviews() {
	models.DB.TruncateAll()
	cfg := internals.InitFakeApp(true)

    mc := models.MediaContainer{Src: "A", Preview: "p", ContentType: "i",}
    mc2 := models.MediaContainer{Src: "A", Preview: "p", ContentType: "i",}
    as.DB.Create(&mc)
    as.DB.Create(&mc2)
    as.NotZero(mc.ID)

    p1 := models.PreviewScreen{Src: "fake1", Idx: 0, MediaID: mc.ID,}
    p2 := models.PreviewScreen{Src: "fake2.png", Idx: 1, MediaID: mc.ID,}
    p3 := models.PreviewScreen{Src: "fake2.png", Idx: 1, MediaID: mc2.ID,}
    as.DB.Create(&p1)
    as.DB.Create(&p2)
    as.DB.Create(&p3)

    man := GetManagerActionSuite(cfg, as)
    previewList, err := man.ListScreens(mc.ID, 1, 10)
    as.NoError(err)
    as.Equal(len(*previewList), 2, "We should have two previews")

    previewOne, p_err := man.ListScreens(mc2.ID, 1, 10)
    as.NoError(p_err)
    as.Equal(len(*previewOne), 1, "Now there should be 1")
}

func (as *ActionSuite) Test_ManagerMemoryPreviews() {
	cfg := internals.InitFakeApp(false)

    man := GetManagerActionSuite(cfg, as)
    media := man.ListAllMedia(1, 100)
    as.Greater(len(media), 0, "It should have media setup")

    // Generate some fake screens
    as.Fail("Implement Memory Manager Preview Test")
}
