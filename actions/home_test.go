package actions

import (
    "contented/test_common"
    "net/http"
    "os"
)

func (as *ActionSuite) Test_HomeHandler() {
    test_common.InitFakeApp(false)
    os.Chdir("../") // The Index file expects to be under the serve director/public/build
    res := as.HTML("/").Get()

    as.Equal(http.StatusOK, res.Code)
    as.Contains(res.Body.String(), "Loading Up Contented")
}
