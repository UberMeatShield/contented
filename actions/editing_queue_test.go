package actions

import (
	"contented/models"
	"contented/test_common"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gobuffalo/nulls"
)

func CreateVideoContainer(as *ActionSuite) (*models.Container, *models.Content) {
	cntToCreate, contents := test_common.GetContentByDirName("dir2")

	cRes := as.JSON("/containers/").Post(&cntToCreate)
	as.Equal(http.StatusCreated, cRes.Code, fmt.Sprintf("It should create the container %s", cRes.Body.String()))

	cnt := models.Container{}
	json.NewDecoder(cRes.Body).Decode(&cnt)

	content := models.Content{}
	for _, contentToCreate := range contents {
		if strings.Contains(contentToCreate.Src, "donut") {
			contentToCreate.ContainerID = nulls.NewUUID(cnt.ID)
			contentRes := as.JSON("/content").Post(&contentToCreate)
			as.Equal(http.StatusCreated, contentRes.Code, fmt.Sprintf("Error %s", contentRes.Body.String()))
			json.NewDecoder(contentRes.Body).Decode(&content)
			break
		}
	}
	as.NotZero(cnt.ID)
	as.NotZero(content.ID)
	return &cnt, &content
}

// Do the screen grab in memory
func (as *ActionSuite) Test_EditingQueueScreenHandler() {

	as.Equal(models.TaskOperation.SCREENS.String(), "screen_capture")

	cfg := test_common.InitMemoryFakeAppEmpty()
	as.Equal(cfg.ReadOnly, false)
	_, content := CreateVideoContainer(as)
	timeSeconds := 3
	url := fmt.Sprintf("/editing_queue/%s/screens/%d/%d", content.ID.String(), 1, timeSeconds)
	res := as.JSON(url).Post(&content)
	as.Equal(http.StatusCreated, res.Code, fmt.Sprintf("Should be able to grab a screen %s", res.Body.String()))

	// Huh... ODD
	tr := models.TaskRequest{}
	json.NewDecoder(res.Body).Decode(&tr)
	as.NotZero(tr.ID, fmt.Sprintf("Ugh ogh %s", res.Body.String()))
	as.Equal(models.TaskStatus.PENDING, tr.Status, fmt.Sprintf("Task invalid %s", tr))
	as.Fail("No idea how to test background task")
}

// Validate it created some actual output.
