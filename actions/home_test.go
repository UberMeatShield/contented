package actions

import (
	"net/http"
	"os"
)

func (as *ActionSuite) Test_HomeHandler() {
	os.Chdir("../") // The Index file expects to be under the serve director/public/build
	res := as.HTML("/").Get()

	as.Equal(http.StatusOK, res.Code)
	as.Contains(res.Body.String(), "Loading Up Contented")
}
