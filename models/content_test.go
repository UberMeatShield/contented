package models

import (
	"fmt"
)

func (ms *ModelSuite) Test_Content() {
	count, err := ms.DB.Count("contents")
	ms.NoError(err)
	if count > 0 {
		ms.Fail("The DB was not reset")
	}

	tags := Tags{
		Tag{ID: "A"},
		Tag{ID: "B"},
	}
	mc := Content{
		Src:         "This is the name of the item",
		Preview:     "preview_location",
		ContentType: "image/png",
		Tags:        tags,
	}
	for _, t := range tags {
		t_err := ms.DB.Create(&t)
		if t_err != nil {
			ms.Fail(fmt.Sprintf("Not creating tag %s\n", t_err))
		}
	}
	ms.DB.Eager().Create(&mc)
	ms.NotZero(mc.ID)

	check := Content{}
	q_err := ms.DB.Eager("Tags").Find(&check, mc.ID)

	tags_check := Tags{}
	t_err := ms.DB.All(&tags_check)
	if t_err != nil {
		ms.Fail(fmt.Sprintf("Could not find any tags %s", t_err))
	}
	fmt.Printf("Loaded tags %s \n", tags_check)

	if q_err != nil {
		ms.Fail("Could not query for this id" + mc.ID.String())
	}
	if len(check.Tags) == 0 {
		ms.Fail("None of the tags have been loaded")
	}
}
