package models

import (
    "testing"
    "github.com/gobuffalo/suite/v3"
)

type ModelSuite struct {
    *suite.Model
}

func Test_ModelSuite(t *testing.T) {
    model := suite.NewModel()
    as := &ModelSuite{
        Model: model,
    }
    suite.Run(t, as)
}
