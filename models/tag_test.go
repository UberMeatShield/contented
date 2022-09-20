package models

func (ms *ModelSuite) Test_Tag() {
    t := Tag{
        Name: "Test",
    }
    ms.DB.Create(&t)
}
