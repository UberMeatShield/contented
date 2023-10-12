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
	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/suite/v4"
)

func GetManagerActionSuite(cfg *utils.DirConfigEntry, as *ActionSuite) ContentManager {
	ctx := test_common.GetContext(as.App)
	get_params := func() *url.Values {
		vals := ctx.Params().(url.Values)
		return &vals
	}
	get_conn := func() *pop.Connection {
		// as.DB should work, but it is of a type pop.v5.Connection instead of pop.Connection
		return models.DB
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

func ManagersTagSearchValidation(as *ActionSuite, man ContentManager) {
	as.NoError(man.CreateTag(&models.Tag{ID: "A"}))
	as.NoError(man.CreateTag(&models.Tag{ID: "B"}))
	as.NoError(man.CreateTag(&models.Tag{ID: "OR"}))

	a := &models.Content{NoFile: true, Src: "AFile"}
	b := &models.Content{NoFile: true, Src: "BFile"}
	as.NoError(man.CreateContent(a))
	as.NoError(man.CreateContent(b))

	as.NoError(man.AssociateTagByID("A", a.ID))
	as.NoError(man.AssociateTagByID("B", b.ID))
	as.NoError(man.AssociateTagByID("OR", b.ID))

	_, count, err := man.SearchContent(SearchQuery{})
	as.NoError(err, "It should search empty content")
	as.Equal(count, 2, "And return all the contents")

	_, tCount, tErr := man.SearchContent(SearchQuery{Tags: []string{"A"}, Text: "File"})
	as.NoError(tErr, "A tag join shouldn't explode")
	as.Equal(tCount, 1, "And it should only get A Back")

	_, orCount, orErr := man.SearchContent(SearchQuery{Tags: []string{"OR", "A"}})
	as.NoError(orErr, "A tag join shouldn't explode")
	as.Equal(orCount, 2, "And it should get both objects Back")

	_, noCount, noErr := man.SearchContent(SearchQuery{Tags: []string{"A"}, Text: "YAAARG"})
	as.NoError(noErr, "It should not error")
	as.Equal(noCount, 0, "But it shouldn't match the text")
}
