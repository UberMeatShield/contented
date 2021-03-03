package grifts

import (
	"contented/actions"
	"github.com/gobuffalo/buffalo"
)

// We do need a DB if you are going to run the grifts
func init() {
    UseDatabase := true
	buffalo.Grifts(actions.App(UseDatabase))
}
