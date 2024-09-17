package managers

import (
	"contented/pkg/models"
	"contented/pkg/test_common"
	"contented/pkg/utils"
	"net/url"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestMain(m *testing.M) {
	models.RebuildDatabase("test")
	code := m.Run()
	os.Exit(code)
}

func GetManagerTestSuite(cfg *utils.DirConfigEntry) ContentManager {
	ctx := test_common.GetContext()
	get_params := func() *url.Values {
		return GinParamsToUrlValues(ctx.Params, url.Values{})
	}

	// Likely need to get the url values into a better places as well.
	get_conn := func() *gorm.DB {
		// as.DB should work, but it is of a type pop.v5.Connection instead of pop.Connection
		return models.InitGorm(false)
	}
	return CreateManager(cfg, get_conn, get_params)
}

// Called by the various manager tests
func ManagersTagSearchValidation(t *testing.T, man ContentManager) {
	assert.NoError(t, man.CreateTag(&models.Tag{ID: "A"}))
	assert.NoError(t, man.CreateTag(&models.Tag{ID: "B"}))
	assert.NoError(t, man.CreateTag(&models.Tag{ID: "OR"}))

	a := &models.Content{NoFile: true, Src: "AFile"}
	b := &models.Content{NoFile: true, Src: "BFile"}
	assert.NoError(t, man.CreateContent(a))
	assert.NoError(t, man.CreateContent(b))

	assert.NoError(t, man.AssociateTagByID("A", a.ID))
	assert.NoError(t, man.AssociateTagByID("B", b.ID))
	assert.NoError(t, man.AssociateTagByID("OR", b.ID))

	_, count, err := man.SearchContent(ContentQuery{})
	assert.NoError(t, err, "It should search empty content")
	assert.Equal(t, int64(2), count, "And return all the contents")

	_, tCount, tErr := man.SearchContent(ContentQuery{Tags: []string{"A"}, Search: "File"})
	assert.NoError(t, tErr, "A tag join shouldn't explode")
	assert.Equal(t, int64(1), tCount, "And it should only get A Back")

	_, orCount, orErr := man.SearchContent(ContentQuery{Tags: []string{"OR", "A"}})
	assert.NoError(t, orErr, "A tag join shouldn't explode")
	assert.Equal(t, int64(2), orCount, "And it should get both objects Back")

	_, noCount, noErr := man.SearchContent(ContentQuery{Tags: []string{"A"}, Search: "YAAARG"})
	assert.NoError(t, noErr, "It should not error")
	assert.Equal(t, int64(0), noCount, "But it shouldn't match the text")
}
