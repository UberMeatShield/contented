package managers

/**
*  These are helpers for use in grifts, but we want them compiling in the dev service in case of breaks.
*
* Bad code in a grift is harder to notice and the compilation with tests also seems a little broken. ie
* you break the grift via main package changes and never notice.  If You break the test in a grift directory
* and then the compilation just failed with no error messages...
 */

import (
	//    "os"
	"fmt"
	"log"
	// "time"
	// "strings"
	//"path/filepath"
	"contented/models"
	"contented/utils"
	"github.com/gobuffalo/nulls"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
)

// Process all the directories and get a valid setup into the DB
// Probably should return a count of everything
func CreateInitialStructure(cfg *utils.DirConfigEntry) error {

	contentTree, err := utils.CreateStructure(cfg.Dir, cfg, &utils.ContentTree{}, 0)
	content := *contentTree
	if len(content) == 0 {
		return errors.New("No subdirectories found under path: " + cfg.Dir)
	}
	log.Printf("Found %d sub-directories.\n", len(content))

	// TODO: Optional?  Some sort of crazy merge for later?
	db_err := models.DB.TruncateAll()
	if db_err != nil {
		return errors.WithStack(err)
	}

	// TODO: Need to do this in a single transaction vs partial
	for idx, ct := range content {
		if cfg.ExcludeEmptyContainers && len(ct.Media) == 0 {
			log.Printf("Excluding %s/%s as it had no content found\n", ct.Cnt.Path, ct.Cnt.Name)
			continue // SKIP empty container directories
		}

		// Prepare to at the container to the DB
		c := ct.Cnt
		c.Idx = idx
		media := ct.Media
		log.Printf("Adding Media to %s with total media %d \n", c.Name, len(media))

		// Use the database version of uuid generation (minimize the miniscule conflict)
		unset_uuid, _ := uuid.FromString("00000000-0000-0000-0000-000000000000")
		c.ID = unset_uuid

		// Assign a default preview (maybe move this into create Structure?)
		if len(ct.Media) > 0 {
			c.PreviewUrl = "/preview/" + ct.Media[0].ID.String()
		}

        // TODO: Port to using the manager somehow (note this is called from a grift)
		models.DB.Create(&c)
		log.Printf("Created %s with id %s\n", c.Name, c.ID)

		// There MUST be a way to do this as a single commit
		for _, mc := range media {
			mc.ContainerID = nulls.NewUUID(c.ID)
			c_err := models.DB.Create(&mc)

			// This is pretty damn fatal so we want it to die if the DB bails.
			if c_err != nil {
				log.Fatal(c_err)
			}
		}
	}
	return nil
}

// Init a manager and pass it in or just do this via config value instead of a pass in
func CreateAllPreviews(cm ContentManager) error {
	cnts, c_err := cm.ListContainers(0, 9001)
	if c_err != nil {
		return c_err
	}
	if len(*cnts) == 0 {
		return errors.New("No Containers were found in the database")
	}

	err_msg := ""
	for _, cnt := range *cnts {
		err := CreateContainerPreviews(&cnt, cm)
		if err != nil {
			err_msg += fmt.Sprintf("Error creating previews in cnt %s err: %s\n", cnt.ID.String(), err)
		}
	}
	if err_msg != "" {
		return errors.New(err_msg)
	}
	return nil
}

// TODO: Should this return a total of previews created or something?
func CreateContainerPreviews(c *models.Container, cm ContentManager) error {
	log.Printf("About to try and create previews for %s:%s\n", c.Name, c.ID.String())
	// Reset the preview directory, then create it fresh (update tests if this changes)
	c_err := utils.ClearContainerPreviews(c)
	if c_err == nil {
		err := utils.MakePreviewPath(utils.GetContainerPreviewDst(c))
		if err != nil { // This is pretty fatal if we don't have dist permission
			log.Fatal(err)
		}
	}

	// TODO: It should fix up the total count there (-1 for unlimited?)
	media, q_err := cm.ListMedia(c.ID, 0, 90000)
	if q_err != nil {
		log.Fatal(q_err) // Also fatal if we can no longer list media
	}

	// It would be nice to maybe abstract this into a better place?
	log.Printf("Found a set of media to make previews for %d", len(*media))
	if media != nil && len(*media) > 0 {
		mcs := *media
		c.PreviewUrl = "/preview/" + mcs[0].ID.String()
		// log.Printf("What was the container preview src %s", c.PreviewUrl)
		cm.UpdateContainer(c)
	}

	update_list, err := CreateMediaPreviews(c, *media)
    if err != nil {
        log.Printf("Errors while creating media previews %s", err)
    }
	log.Printf("Finished creating previews, now updating the database count(%d)", len(update_list))
    maybeScreens, _ := utils.GetPotentialScreens(c)
	for _, mc := range update_list {
		if mc.Preview != "" {
			log.Printf("Created a preview %s for mc %s", mc.Preview, mc.ID.String())
            screens := utils.AssignScreensFromSet(c, &mc, maybeScreens)
            if screens != nil {
                log.Printf("Found new screens we should create %d", len(*screens))
                for _, s := range(*screens) {
                    cm.CreateScreen(&s)
                }
            }

            // Note that UpdateMedia and create screen don't really work for in memory
            // Though it actually wouldn't be that bad to update the MemStorage...
			cm.UpdateMedia(&mc)

            // TODO: Add in a search for getting screens based on the container and media
		} else if mc.Corrupt {
			cm.UpdateMedia(&mc)
		}
	}
	return err
}

// TODO: This maybe could be ported to just a ContentTree Element or something
// This is complicated but a way to do many previews at once
func CreateMediaPreviews(c *models.Container, media models.MediaContainers) (models.MediaContainers, error) {
	if len(media) == 0 {
		return models.MediaContainers{}, nil
	}
	cfg := utils.GetCfg()
	processors := cfg.CoreCount
	if processors <= 0 {
		processors = 1 // Without at least one processor this will hang forever
	}
	log.Printf("Creating %d listeners for processing previews", processors)

	// We expect a result for every message so can create the channels in a way that they have a length
	expected_total := len(media)
	reply := make(chan utils.PreviewResult, expected_total)
	input := make(chan utils.PreviewRequest, expected_total)

	// Starts the workers
	for i := 0; i < processors; i++ {
		pw := utils.PreviewWorker{Id: i, In: input}
		go StartWorker(pw)
	}

	// Queue up a bunch of preview work
	mediaMap := models.MediaMap{}
	for _, mc := range media {
		mediaMap[mc.ID] = mc
		ref_mc := mc
		input <- utils.PreviewRequest{
			C:   c,
			Mc:  &ref_mc,
			Out: reply,
		}
	}

	// Exception handling should close the input and output probably
	total := 0
	previews := models.MediaContainers{}

	error_list := ""
	for result := range reply {
		total++
		if total == expected_total {
			close(input) // Do I close this immediately
			close(reply)
		}

		// Get a list of just the preview items?  Or just update by reference?
		log.Printf("Found a result for %s\n", result.MC_ID.String())
		if mc_update, ok := mediaMap[result.MC_ID]; ok {
			if result.Preview != "" {
				log.Printf("We found a reply around this %s id was %s \n", result.Preview, result.MC_ID)
				mc_update.Preview = result.Preview
				previews = append(previews, mc_update)
			} else if result.Err != nil {
				log.Printf("ERROR: Failed to create a preview %s\n", result.Err)
				error_list += "" + result.Err.Error()

				mc_update.Preview = ""
				mc_update.Corrupt = true
				previews = append(previews, mc_update)
			} else {
				log.Printf("No preview was needed for media %s", result.MC_ID)
			}
		} else {
			log.Printf("Missing Response ID, something went really wrong %s\n", result.MC_ID)
		}
	}
	if error_list != "" {
		return previews, errors.New(error_list)
	}
	return previews, nil
}

func StartWorker(w utils.PreviewWorker) {
	// sleepTime := time.Duration(w.Id) * time.Millisecond
	// log.Printf("Worker %d with sleep %d\n", w.Id, sleepTime)
	// Sleep before kicking off?  Kinda don't need to
	for pr := range w.In {
		c := pr.C
		mc := pr.Mc
		log.Printf("Worker %d Doing a preview for %s\n", w.Id, mc.ID.String())
		preview, err := utils.CreateMediaPreview(c, mc)
		pr.Out <- utils.PreviewResult{
			C_ID:    c.ID,
			MC_ID:   mc.ID,
			Preview: preview,
			Err:     err,
		}
	}
}
