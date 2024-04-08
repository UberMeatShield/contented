package managers

/**
*  These are helpers for use in grifts, but we want them compiling in the dev service in case of breaks.
*
* Bad code in a grift is harder to notice and the compilation with tests also seems a little broken. ie
* you break the grift via main package changes and never notice.  If You break the test in a grift directory
* and then the compilation just failed with no error messages...
 */

import (
	"contented/models"
	"contented/utils"
	"fmt"
	"log"
	"path/filepath"
	"strings"

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
		if cfg.ExcludeEmptyContainers && len(ct.Content) == 0 {
			log.Printf("Excluding %s/%s as it had no content found\n", ct.Cnt.Path, ct.Cnt.Name)
			continue // SKIP empty container directories
		}

		// Prepare to at the container to the DB
		c := ct.Cnt
		c.Idx = idx
		content := ct.Content
		log.Printf("Adding Content to %s with total content %d \n", c.Name, len(content))

		// Use the database version of uuid generation (minimize the miniscule conflict)
		unset_uuid, _ := uuid.FromString("00000000-0000-0000-0000-000000000000")
		c.ID = unset_uuid

		// Assign a default preview (maybe move this into create Structure?)
		if len(ct.Content) > 0 {
			c.PreviewUrl = "/preview/" + ct.Content[0].ID.String()
		}

		// TODO: Port to using the manager somehow (note this is called from a grift)
		models.DB.Create(&c)
		log.Printf("Created %s with id %s\n", c.Name, c.ID)

		// There MUST be a way to do this as a single commit
		for _, mc := range content {
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

// For now this is fine but this could probably be better as something that
// just takes an array of strings and creates the tags that way in the manager.
func CreateTagsFromFile(cm ContentManager) (*models.Tags, error) {
	cfg := utils.GetCfg()
	tagFile := cfg.TagFile
	if tagFile == "" {
		log.Printf("No tag file so nothing to create")
		return &models.Tags{}, nil
	}
	tags, err := utils.ReadTagsFromFile(tagFile)
	if err != nil || tags == nil {
		log.Printf("Error reading tagfile %s", err)
		return nil, err
	}

	for _, tCheck := range *tags {
		tag, _ := cm.GetTag(tCheck.ID)
		/*
			if err != nil {
				log.Printf("Tag was not found trying to load tag %s", tCheck.ID)
			}
		*/
		// If we do not have the tag it should create it
		if tag == nil {
			tag = &tCheck
			cErr := cm.CreateTag(tag)
			if cErr != nil {
				log.Printf("Failed to create a tag bailing out %s err %s", tag.ID, cErr)
				return tags, cErr
			}
		}
	}
	return tags, nil
}

// Init a manager and pass it in or just do this via config value instead of a pass in
func CreateAllPreviews(cm ContentManager) error {
	log.Printf("Attempting to create all previews")
	cnts, _, c_err := cm.ListContainers(ContainerQuery{PerPage: 9001}) // Might need to make this smarter :(
	if c_err != nil {
		log.Printf("Failed to list all containers %s", c_err)
		return c_err
	}
	if cnts == nil {
		msg := "No Containers were found in the manager"
		log.Printf(msg)
		return errors.New(msg)
	}
	log.Printf("Found a number of containers %d", len(*cnts))

	err_msg := []string{}
	for _, cnt := range *cnts {
		err := CreateContainerPreviews(&cnt, cm)
		if err != nil {
			msg := fmt.Sprintf("Error creating previews in cnt %s - %s err: %s\n", cnt.ID.String(), cnt.Name, err)
			err_msg = append(err_msg, msg)
		}
	}
	// TODO: Cut down how much spam is getting kicked out by this summary
	if len(err_msg) > 0 {
		return errors.New(strings.Join(err_msg, "\n"))
	}
	return nil
}

// Attempts to look in a container for videos that were already encoded but where the original
// source video was not removed.
func FindDuplicateVideos(cm ContentManager) (models.Contents, error) {
	log.Printf("Attempting to remove Duplicate videos")
	cfg := utils.GetCfg()
	if cfg.EncodingFilenameModifier == "" {
		log.Fatalf("The encoding filename modifier is used to look for a dupe and it is not set.")
	}

	// Containers are cheap... maybe just grab them all initially?
	cq := ContainerQuery{PerPage: 9001}
	containers, totalCnt, err := cm.ListContainers(cq)
	if err != nil || totalCnt == 0 {
		log.Fatalf("No containers found in the system")
	}

	log.Printf("Looking in %d containers", len(*containers))
	duplicates := models.Contents{}
	for _, cnt := range *containers {
		dupes := FindDuplicateContents(cm, &cnt, "video")
		if len(dupes) > 0 {
			log.Printf("Found duplicates %d in cnt %s", len(dupes), cnt.Name)
			duplicates = append(duplicates, dupes...)
		}
	}
	return duplicates, nil
}

func FindDuplicateContents(cm ContentManager, cnt *models.Container, contentType string) models.Contents {
	cfg := utils.GetCfg()
	cs := ContentQuery{
		ContentType: contentType,
		PerPage:     90001, // TODO: just page it better
		ContainerID: cnt.ID.String(),
	}
	contents, total, err := cm.ListContent(cs)
	if total == 0 || err != nil {
		log.Printf("Could not find any content under %s", cnt.GetFqPath())
		return models.Contents{}
	}

	// We are only going to look for dupes in the same folder initially
	contentNames := models.ContentMapBySrc{}
	for _, content := range *contents {
		contentNames[content.Src] = content
	}

	// Initially we are only going to look for encoding dupes that are video
	cntPath := cnt.GetFqPath()
	duplicates := models.Contents{}

	// TODO: If I can trust the content.Encoding is always already set I could update to query on that
	// field in addition to the contentType but that is really only useful if I expend the video dupe
	// check to be a more complicated image hash & time lookup.
	log.Printf("Finding video already in %s so we can remove their dupes", cfg.EncodingFilenameModifier)
	for _, content := range *contents {
		if content.Encoding == cfg.CodecForConversionName {
			originalName := strings.Replace(content.Src, cfg.EncodingFilenameModifier, "", 1)
			if originalName == content.Src {
				continue
			}
			if mContent, ok := contentNames[originalName]; ok {
				log.Printf("Found a for a dupe called %s", originalName)
				encodedPath := filepath.Join(cntPath, content.Src)
				dupePath := filepath.Join(cntPath, mContent.Src)

				foundDupe, checkErr := utils.IsDuplicateVideo(encodedPath, dupePath)
				if checkErr != nil {
					log.Printf("Error attempting to determine if a video was a dupe %s", checkErr)
				} else if foundDupe {
					log.Printf("Found a duplicate at %s", dupePath)
					mContent.FqPath = dupePath
					duplicates = append(duplicates, mContent)
				}
			} else {
				log.Printf("No dupe with this name found %s", originalName)
			}
		}
	}
	return duplicates
}

// TODO: Should this return a total of previews created or something?
func CreateContainerPreviews(c *models.Container, cm ContentManager) error {
	log.Printf("About to try and create previews for %s:%s\n", c.Name, c.ID.String())
	// Reset the preview directory, then create it fresh (update tests if this changes)
	c_err := utils.ClearContainerPreviews(c)
	if c_err == nil {
		err := utils.MakePreviewPath(utils.GetContainerPreviewDst(c))
		if err != nil { // This is pretty fatal if we don't have dest permission
			log.Fatal(err)
		}
	}

	// TODO: It should fix up the total count there (-1 for unlimited?)
	cq := ContentQuery{ContainerID: c.ID.String(), PerPage: 90000, Direction: "asc", Order: "idx"}
	content, _, q_err := cm.ListContent(cq)
	if q_err != nil {
		log.Fatal(q_err) // Also fatal if we can no longer list content
	}

	// It would be nice to maybe abstract this into a better place?
	log.Printf("Found a set of content to make previews for %d", len(*content))
	if content != nil && len(*content) > 0 {
		mcs := *content
		c.PreviewUrl = "/preview/" + mcs[0].ID.String()
		cm.UpdateContainer(c)
	}

	update_list, err := CreateContentPreviews(c, *content)
	if err != nil {
		log.Printf("Summary of Errors while creating content previews %s", err)
	}
	log.Printf("Finished creating previews, now updating the database count(%d)", len(update_list))
	maybeScreens, _ := utils.GetPotentialScreens(c)
	for _, mc := range update_list {
		if mc.Preview != "" {
			log.Printf("Created a preview %s for mc %s", mc.Preview, mc.ID.String())
			screens := utils.AssignScreensFromSet(c, &mc, maybeScreens)
			if screens != nil {
				log.Printf("Found new screens we should create %d", len(*screens))
				for _, s := range *screens {
					cm.CreateScreen(&s)
				}
			}

			// Note that UpdateContent and create screen don't really work for in memory
			// Though it actually wouldn't be that bad to update the MemStorage...
			cm.UpdateContent(&mc)

			// TODO: Add in a search for getting screens based on the container and content
		} else if mc.Corrupt {
			cm.UpdateContent(&mc)
		}
	}
	return err
}

// TODO: This maybe could be ported to just a ContentTree Element or something
// This is complicated but a way to do many previews at once
func CreateContentPreviews(c *models.Container, content models.Contents) (models.Contents, error) {
	if len(content) == 0 {
		return models.Contents{}, nil
	}
	cfg := utils.GetCfg()
	processors := cfg.CoreCount
	if processors <= 0 {
		processors = 1 // Without at least one processor this will hang forever
	}
	log.Printf("Creating %d listeners for processing previews", processors)

	// We expect a result for every message so can create the channels in a way that they have a length
	expected_total := len(content)
	reply := make(chan utils.PreviewResult, expected_total)
	input := make(chan utils.PreviewRequest, expected_total)

	// Starts the workers
	for i := 0; i < processors; i++ {
		pw := utils.PreviewWorker{Id: i, In: input}
		go StartWorker(pw)
	}

	// Queue up a bunch of preview work
	contentMap := models.ContentMap{}
	for _, mc := range content {
		contentMap[mc.ID] = mc
		ref_mc := mc
		input <- utils.PreviewRequest{
			C:   c,
			Mc:  &ref_mc,
			Out: reply,
		}
	}

	// Exception handling should close the input and output probably
	total := 0
	previews := models.Contents{}

	error_list := ""
	for result := range reply {
		total++
		if total == expected_total {
			close(input) // Do I close this on error or on potential timeout (it seems like there should be a way)
			close(reply)
		}

		// Get a list of just the preview items?  Or just update by reference?
		log.Printf("Found a result for %s\n", result.MC_ID.String())
		if mc_update, ok := contentMap[result.MC_ID]; ok {
			if result.Preview != "" {
				log.Printf("We found a reply around this %s id was %s \n", result.Preview, result.MC_ID)
				mc_update.Preview = result.Preview
				previews = append(previews, mc_update)
			} else if result.Err != nil {
				log.Printf("ERROR: Failed to create a preview %s for %s \n", result.Err, mc_update.Src)
				error_list += "" + result.Err.Error()
				mc_update.Preview = ""
				mc_update.Corrupt = true
				previews = append(previews, mc_update)
			} else {
				log.Printf("No preview was needed for content %s", result.MC_ID)
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
		preview, err := utils.CreateContentPreview(c, mc)
		pr.Out <- utils.PreviewResult{
			C_ID:    c.ID,
			MC_ID:   mc.ID,
			Preview: preview,
			Err:     err,
		}
	}
}
