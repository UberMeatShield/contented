package models

import (
    "path/filepath"
)

func (ms *ModelSuite) Test_MediaContainerScreens() {
	count, err := ms.DB.Count("media_containers")
	ms.NoError(err)
	if count > 0 {
		ms.Fail("The DB was not reset")
	}

	mc := MediaContainer{
		Src:         "We should be able to create a set of screens",
		Preview:     "preview_location",
		ContentType: "image/png",
	}
	ms.DB.Create(&mc)
	ms.NotZero(mc.ID)

	p1 := PreviewScreen{
		Src:     "fake1",
		Idx:     0,
		MediaID: mc.ID,
	}
	p2 := PreviewScreen{
		Src:     "fake2.png",
		Idx:     1,
        Path: "Derp/Monkey",
		MediaID: mc.ID,
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
		ms.Fail("Couldn't create preview screen 2 %s", perr2)
	}

	check := MediaContainer{}
	q_err := ms.DB.Eager().Find(&check, mc.ID)
	if q_err != nil {
		ms.Fail("Could not query for this id %s" + mc.ID.String(), q_err)
	}
	if check.Screens == nil {
		ms.Fail("Failed to load screens" + mc.ID.String())
	}
	if len(check.Screens) != 2 {
		ms.Fail("The screens did not load back: " + mc.ID.String())
	}
}
