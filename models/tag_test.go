package models
import ("fmt")

func (ms *ModelSuite) Test_Tag() {
    t := Tag{
        ID: "Test",
    }
    ms.DB.Create(&t)

    t_check := Tag{}
    q_err := ms.DB.Find(&t_check, "Test")
    if q_err != nil {
        ms.Fail(fmt.Sprintf("Failed to load tag %s", q_err))
    }
}
