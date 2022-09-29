package models

import (
    "github.com/gobuffalo/nulls"
)

func (ms *ModelSuite) Test_Container() {
    // ms.LoadFixture("Container")
    // ms.LoadFixture("Content")
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
        Name:  "KindaCrap",
    }
    ms.DB.Create(&c)
    ms.NotZero(c.ID)

    mc1 := Content{
        Src:         "first",
        ContainerID: nulls.NewUUID(c.ID),
    }
    mc2 := Content{
        Src:         "second",
        ContainerID: nulls.NewUUID(c.ID),
    }
    ms.DB.Create(&mc1)
    ms.DB.Create(&mc2)

    load_back := Container{}
    err := ms.DB.Eager().Find(&load_back, c.ID)

    if err != nil {
        ms.Fail("Could not query the DB %s", err)
    }
    if len(load_back.Contents) != 2 {
        ms.Fail("Could not load up the contents media containers %s", load_back)
    }

}
