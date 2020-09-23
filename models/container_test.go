package models

func (ms *ModelSuite) Test_Container() {
	// ms.LoadFixture("Container")
	// ms.LoadFixture("MediaContainer")
	count, err := ms.DB.Count("containers")
	ms.NoError(err)
	if count > 0 {
		ms.Fail("The DB was not reset")
	}

	c := Container{
		Total: 2,
		Path:  "root_path",
		Name:  "test_dir",
	}
	ms.DB.Create(&c)
	ms.NotZero(c.ID)
}


func (ms *ModelSuite) Test_Container_Query() {
	c := Container{
		Total: 2,
		Path:  "query_test",
		Name:  "Shit",
	}
	ms.DB.Create(&c)
    ms.NotZero(c.ID)

    containers := []Container{}
    err := ms.DB.All(&containers)

    if err != nil {
      ms.Fail("Could not query the DB %s", err)
    }

}
