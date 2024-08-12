package managers

import (
	"contented/internals"
	"contented/models"
	"contented/test_common"
	"contented/utils"
	"log"
	"net/url"
	"os"
	"testing"

	"github.com/gobuffalo/envy"
	"github.com/gobuffalo/suite/v4"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

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

func TestMain(m *testing.M) {
	_, err := envy.MustGet("DIR")
	if err != nil {
		log.Println("DIR ENV REQUIRED ie: $export=DIR=`pwd`/mocks/content/ && buffalo test")
		panic(err)
	}
	code := m.Run()
	os.Exit(code)
}

func Test_ManagerSuite(t *testing.T) {
	app := internals.CreateBuffaloApp(true, "test")
	action := suite.NewAction(app)
	as := &ActionSuite{
		Action: action,
	}
	suite.Run(t, as)
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
	assert.Equal(t, count, 2, "And return all the contents")

	_, tCount, tErr := man.SearchContent(ContentQuery{Tags: []string{"A"}, Search: "File"})
	assert.NoError(t, tErr, "A tag join shouldn't explode")
	assert.Equal(t, tCount, 1, "And it should only get A Back")

	_, orCount, orErr := man.SearchContent(ContentQuery{Tags: []string{"OR", "A"}})
	assert.NoError(t, orErr, "A tag join shouldn't explode")
	assert.Equal(t, orCount, 2, "And it should get both objects Back")

	_, noCount, noErr := man.SearchContent(ContentQuery{Tags: []string{"A"}, Search: "YAAARG"})
	assert.NoError(t, noErr, "It should not error")
	assert.Equal(t, noCount, 0, "But it shouldn't match the text")
}
