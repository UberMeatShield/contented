package actions

import (
	"contented/test_common"
	"contented/utils"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHomeHandler(t *testing.T) {
	cfg, _ := test_common.InitFakeApp(false)
	cfg.StaticResourcePath = fmt.Sprintf("../%s", cfg.StaticResourcePath)
	utils.SetCfg(*cfg)

	r := setupStatic()
	url := "/"
	req, _ := http.NewRequest("GET", url, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, fmt.Sprintf("Could not hit %s", url))
	assert.Contains(t, w.Body.String(), "Loading Up Contented")
}
