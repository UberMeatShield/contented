package actions

import (
  "os"
  "net/http"
)


func (as *ActionSuite) Test_HomeHandler() {
    os.Chdir("../") // The Index file expects to be under the serve director/public/build
	res := as.JSON("/").Get()

	as.Equal(http.StatusOK, res.Code)
	as.Contains(res.Body.String(), "Loading Up Contented")
}
