package actions

import (
	"contented/models"
	"contented/test_common"
	"encoding/json"
	"fmt"
	"net/http"
)

func (as *ActionSuite) Test_TagsResource_List_DB() {
	test_common.ResetConfig()
	test_common.InitFakeApp(true)

	a := models.Tag{ID: "A", Description: "Create a tag, show it"}
	b := models.Tag{ID: "B", Description: "Create a tag, show it"}
	as.DB.Create(&a)
	as.DB.Create(&b)

	res := as.JSON("/tags/").Get()
	as.Equal(http.StatusOK, res.Code, fmt.Sprintf("Failed to load tags %s", res.Body.String()))

	tags := models.Tags{}
	json.NewDecoder(res.Body).Decode(&tags)
	as.Equal(len(tags), 2, fmt.Sprintf("There should be two tags %s", tags))
}

func (as *ActionSuite) Test_TagsResource_Show_DB() {
	test_common.ResetConfig()
	test_common.InitFakeApp(true)
	t := models.Tag{ID: "A", Description: "Create a tag, show it"}
	err := as.DB.Create(&t)
	as.NoError(err)

	res := as.JSON(fmt.Sprintf("/tags/%s/", t.ID)).Get()
	as.Equal(http.StatusOK, res.Code, fmt.Sprintf("It should find the tag %s", res.Body.String()))

	check := models.Tag{}
	cErr := as.DB.Find(&check, t.ID)
	as.NoError(cErr)
	as.Equal(check.Description, t.Description)
}

func (as *ActionSuite) Test_TagsResource_Create_DB() {
	test_common.ResetConfig()
	test_common.InitFakeApp(true)
	t := models.Tag{ID: "Monkey", Description: "What?"}
	res := as.JSON("/tags/").Post(t)
	as.Equal(http.StatusCreated, res.Code, fmt.Sprintf("Error %s", res.Body.String()))

	check := as.JSON("/tags").Get()
	as.Equal(http.StatusOK, check.Code, fmt.Sprintf("It should find the tag %s", res.Body.String()))

	checkTags := models.Tags{}
	json.NewDecoder(check.Body).Decode(&checkTags)
	as.Equal(1, len(checkTags), fmt.Sprintf("We should have a tag %s", check.Body.String()))
}

func (as *ActionSuite) Test_TagsResource_Update_DB() {
	test_common.ResetConfig()
	test_common.InitFakeApp(true)
	t := models.Tag{ID: "Tag", Description: "Original"}
	as.DB.Create(&t)
	t.Description = "Updated"
	res := as.JSON(fmt.Sprintf("/tags/%s", t.ID)).Put(t)
	as.Equal(http.StatusOK, res.Code, fmt.Sprintf("It should update %s", res.Body.String()))

	check := models.Tag{}
	err := as.DB.Find(&check, t.ID)
	as.NoError(err)
	as.Equal("Updated", check.Description)
}

func (as *ActionSuite) Test_TagsResource_Destroy_DB() {
	test_common.ResetConfig()
	test_common.InitFakeApp(true)

	t := models.Tag{ID: "A"}
	err := as.DB.Create(&t)
	as.NoError(err)

	tags := models.Tags{}
	tErr := as.DB.All(&tags)
	as.NoError(tErr)
	as.Equal(len(tags), 1, "There should be a tag")

	res := as.JSON(fmt.Sprintf("/tags/%s", t.ID)).Delete()
	as.Equal(http.StatusOK, res.Code)

	checkTags := models.Tags{}
	cErr := as.DB.All(&checkTags)
	as.NoError(cErr)
	as.Equal(len(checkTags), 0, "It should be deleted")
}
