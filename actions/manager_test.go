package actions

import (
	//"fmt"
    "log"
	"contented/utils"
	"contented/models"
    "net/http"
    "net/url"
    /*
	"encoding/json"
	"net/http"
	"os"
	"testing"
	"github.com/gobuffalo/packr/v2"
	"github.com/gobuffalo/suite"
	"github.com/gofrs/uuid"
	"github.com/gobuffalo/envy"
    "github.com/gobuffalo/buffalo"
    */
    //"context"
    //"sync"
    //"github.com/gobuffalo/logger"
    "github.com/gobuffalo/envy"
    "github.com/gobuffalo/pop/v5"
    //"github.com/gobuffalo/buffalo"
)

var expect_len = map[string]int{
    "dir1": 12,
    "dir2": 3,
    "dir3": 6,
    "screens": 4,
}


func Get_Manager_ActionSuite(cfg *utils.DirConfigEntry, as *ActionSuite) ContentManager{
    ctx := getContext(as.App)
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


func GetMediaByDirName(test_dir_name string) (*models.Container, models.MediaContainers) {
    dir, _ := envy.MustGet("DIR")
    cfg := utils.GetCfg()
    cfg.Dir = dir
    cnts := utils.FindContainers(dir)

    var cnt *models.Container = nil
    for _, c := range cnts {
        if c.Name == test_dir_name {
            cnt = &c
            break
        }
    }
    if cnt == nil {
        log.Panic("Could not find the directory: " +  test_dir_name)
    }
    media := utils.FindMedia(*cnt, 42, 0)
    cnt.Total = len(media)
    return cnt, media
}


func (as *ActionSuite) Test_ManagerContainers() {
    init_fake_app(false)
    ctx := getContext(as.App)
    man := GetManager(&ctx)
    containers, err := man.ListContainersContext()
    as.NoError(err)

    for _, c := range *containers {
        c_mem, err := man.FindDirRef(c.ID)
        if err != nil {
            as.Fail("It should not have an issue finding valid containers")
        }
        as.Equal(c_mem.ID, c.ID)
    }
}


func (as *ActionSuite) Test_ManagerMediaContainer() {
    init_fake_app(false)
    ctx := getContext(as.App)
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
    dir, _ := envy.MustGet("DIR")
    cfg := utils.GetCfg()
    cfg.UseDatabase = false
    utils.InitConfig(dir, cfg)

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
    ctx := getContext(as.App)
    man := GetManager(&ctx)  // New Reference but should have the same count of media
    mcs_2, _ := man.ListAllMedia(1, 9000)

    as.Equal(len(*mcs), len(*mcs_2), "A new instance should use the same storage")
}


func (as *ActionSuite) Test_MemoryDenyEdit() {
    cfg := init_fake_app(false)
    cfg.UseDatabase = false
    ctx := getContext(as.App)
    man := GetManager(&ctx)

    containers, err := man.ListContainersContext()
    as.NoError(err, "It should list containers")

    as.Greater(len(*containers), 0, "There should be containers")

    for _, c := range *containers {
        c.Name = "Update Should fail"
        res := as.JSON("/containers/" + c.ID.String()).Put(&c)
        as.Equal(http.StatusNotImplemented, res.Code)
    }
}

func (as *ActionSuite) Test_MemoryManagerPaginate() {
    cfg := init_fake_app(false)
    cfg.UseDatabase = false

    ctx := getContextParams(as.App, "/containers", "1", "2")
    man := GetManager(&ctx)
    as.Equal(man.CanEdit(), false, "Memory manager should not be editing")

    // Hate
    containers, err := man.ListContainers(1, 1)
    as.NoError(err, "It should list with pagination")
    as.Equal(1, len(*containers), "It should respect paging")

    cnt := (*containers)[0]
    as.NotNil(cnt, "There should be a container with 12 entries")
    as.Equal(cnt.Total, 12, "There should be 12 test images in the first ORDERED containers")
    as.NoError(err)
    media_page_1, _ := man.ListMedia(cnt.ID, 1, 4)
    as.Equal(len(*media_page_1), 4, "It should respect page size")

    media_page_3, _ := man.ListMedia(cnt.ID, 3, 4)
    as.Equal(len(*media_page_3), 4, "It should respect page size and get the last page")

    as.NotEqual((*media_page_3)[3].ID, (*media_page_1)[3].ID, "Ensure it actually paged")

    // Last container pagination check
    l_cnts, _ := man.ListContainers(4, 1)
    as.Equal(1, len(*l_cnts), "It should still return only as we are on the last page")
    l_cnt := (*l_cnts)[0]
    as.Equal(l_cnt.Total, expect_len[l_cnt.Name], "There are 3 entries in the ordered test data last container")
}

func (as *ActionSuite) Test_ManagerInitialize() {
    init_fake_app(false)

    ctx := getContext(as.App)
    man := GetManager(&ctx)
    as.NotNil(man, "It should have a manager defined after init")

    containers, err := man.ListContainersContext()
    as.NoError(err, "It should list all containers")
    as.NotNil(containers, "It should have containers")
    as.Equal(len(*containers), 4, "It should have 4 of them")

    // Memory test working
    for _, c := range *containers {
        // fmt.Printf("Searching for this container %s with name %s\n", c.ID, c.Name)
        media, err := man.ListMediaContext(c.ID)
        as.NoError(err)
        as.NotNil(media)

        media_len := len(*media)
        // fmt.Printf("Media length was %d\n", media_len)
        as.Greater(media_len, 0, "There should be a number of media")
        as.Equal(expect_len[c.Name], media_len, "It should have this many instances: " + c.Name )
        as.Greater(c.Total, 0, "All of them should have a total assigned")
    }
}

func (as *ActionSuite) Test_ManagerDB() {
    models.DB.TruncateAll()

    cfg := utils.GetCfg()
    cfg.UseDatabase = true
    init_fake_app(true)

    cnt, media := GetMediaByDirName("dir1")
    as.Equal("dir1", cnt.Name, "It should be the right dir")
    as.Equal(12, cnt.Total, "The container total should be this for dir1")
    as.Equal(12, len(media))

    c_err := models.DB.Create(cnt)
    as.NoError(c_err)
    for _, mc := range media {
        models.DB.Create(&mc)
    }

    man := Get_Manager_ActionSuite(cfg, as)
    q_media, err := man.ListAllMedia(0, 14)
    as.NoError(err, "We should be able to list")
    as.Equal(len(*q_media), 12, "there should be 12 results")

    lim_media, _ := man.ListAllMedia(0, 3)
    as.Equal(3, len(*lim_media),"The DB should be setup with 10 items")
}
