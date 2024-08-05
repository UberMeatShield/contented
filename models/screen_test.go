package models

import (
	"fmt"
	"path/filepath"
)

func (ms *ModelSuite) Test_ContentScreens() {
	count, err := ms.DB.Count("contents")
	ms.NoError(err)
	if count > 0 {
		ms.Fail("The DB was not reset")
	}

	mc := Content{
		Src:         "We should be able to create a set of screens",
		Preview:     "preview_location",
		ContentType: "image/png",
	}
	ms.DB.Create(&mc)
	ms.NotZero(mc.ID)

	p1 := Screen{
		Src:       "fake1",
		Idx:       0,
		ContentID: mc.ID,
	}
	p2 := Screen{
		Src:       "fake2.png",
		Idx:       1,
		Path:      "Derp/Monkey",
		ContentID: mc.ID,
	}

	p2Loc := p2.GetFqPath()
	if p2Loc != filepath.Join(p2.Path, p2.Src) {
		ms.Fail("Didn't create the right fq path")
	}
	perr1 := ms.DB.Create(&p1)
	if perr1 != nil {
		ms.Fail("Couldn't create preview screen 1 %s", perr1)
	}
	perr2 := ms.DB.Create(&p2)
	if perr2 != nil {
		ms.Fail(fmt.Sprintf("Couldn't create preview screen 2 %s", perr2))
	}

	check := Content{}
	q_err := ms.DB.Eager().Find(&check, mc.ID)
	if q_err != nil {
		ms.Fail(fmt.Sprintf("Could not query for this id %d %s", mc.ID, q_err))
	}
	if check.Screens == nil {
		ms.Fail(fmt.Sprintf("Failed to load screens %d", mc.ID))
	}
	if len(check.Screens) != 2 {
		ms.Fail(fmt.Sprintf("The screens did not load back: %d", mc.ID))
	}
}
