package actions

import (
	"contented/pkg/config"
	"contented/pkg/test_common"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHomeHandler(t *testing.T) {
	cfg, _ := test_common.InitFakeApp(false)
	cfg.StaticResourcePath = fmt.Sprintf("../../%s", cfg.StaticResourcePath)
	config.SetCfg(*cfg)

	r := setupStatic()
	url := "/"
	req, _ := http.NewRequest("GET", url, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, fmt.Sprintf("Could not hit %s", url))
	assert.Contains(t, w.Body.String(), "Loading Up Contented")
}
