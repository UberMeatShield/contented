package models

func (ms *ModelSuite) Test_MediaContainer() {
	count, err := ms.DB.Count("media_containers")
	ms.NoError(err)
	if count > 0 {
		ms.Fail("The DB was not reset")
	}

	mc := MediaContainer{
		Src:     "This is the name of the item",
		Type:    "image/png",
		Preview: "preview_location",
	}
	ms.DB.Create(&mc)

	ms.NotZero(mc.ID)
}
