package managers

/*
 * These functions help with the grift encoding of content.  Some of this
 * could all be done via task and maybe should be.
 */

import (
	"contented/pkg/config"
	"contented/pkg/models"
	"contented/pkg/utils"
	"fmt"
	"log"
	"strconv"

	"github.com/pkg/errors"
)

// Init a manager and pass it in or just do this via config value instead of a pass in
func EncodeVideos(cm ContentManager) error {
	cnts, _, c_err := cm.ListContainers(ContainerQuery{PerPage: 9001})
	if c_err != nil {
		return c_err
	}
	if cnts == nil {
		return errors.New("No Containers were found in the database")
	}

	all_results := utils.EncodingResults{}
	for _, cnt := range *cnts {
		results, err := EncodeContainer(&cnt, cm)
		if err == nil && results != nil && len(*results) > 0 {
			all_results = append(all_results, *results...)
		}
	}

	if len(all_results) == 0 {
		log.Printf("Found nothing that should be encoded (or everything is already encoded)")
		return nil
	}

	lineBreak := "===================================================="
	log.Printf("Encoding complete\n%s\n", lineBreak)
	for _, res := range all_results {
		if res.Err == nil {
			msg := fmt.Sprintf("Encoding Success %s media ID %d\n", res.NewVideo, res.MC_ID)
			log.Print(msg)
		}
	}
	log.Print(lineBreak)
	log.Printf("Failures after this line %s", lineBreak)
	err_cnt := 0
	for _, res := range all_results {
		if res.Err != nil {
			// Might want get the full link to the original video but it should be in the error
			content, miss_err := cm.GetContent(res.MC_ID)
			if miss_err == nil {
				msg := fmt.Sprintf("Failure encoding %s failure was %s id %d", res.Err, content.Src, content.ID)
				log.Print(msg)
			} else {
				msg := fmt.Sprintf("Failure encoding %s failure was %d (deleted?)", res.Err, res.MC_ID)
				log.Print(msg)
			}
			err_cnt++
		}
	}
	if err_cnt > 0 {
		return errors.New(fmt.Sprintf("Encoding had errors count(%d)", err_cnt))
	}
	return nil
}

func EncodeContainer(c *models.Container, cm ContentManager) (*utils.EncodingResults, error) {
	content, _, q_err := cm.ListContent(ContentQuery{ContainerID: strconv.FormatInt(c.ID, 10), PerPage: 90000})
	if q_err != nil {
		log.Fatal(q_err) // Also fatal if we can no longer list content (empty is just [])
	}

	// Remember that references in a range loop CHANGE the pointer on each loop so you MUST
	// re-assign a variable if you want to build a new object with pointers.
	toEncode := utils.EncodingRequests{}
	for _, mc := range *content {
		srcFile, _ := utils.GetFilePathInContainer(mc.Src, c.GetFqPath())

		// DstFile should split off the final extension \.xyz and replace it
		dstFile := utils.GetVideoConversionName(srcFile)
		msg, err, encode := utils.ShouldEncodeVideo(srcFile, dstFile)

		if encode {
			log.Printf("Will attempt to convert %s", msg)
			ref_mc := mc // Ensure the pointers don't get messed up
			req := utils.EncodingRequest{C: c, Mc: &ref_mc, DstFile: dstFile, SrcFile: srcFile}
			toEncode = append(toEncode, req)
		} else if err != nil {
			log.Printf("Error attempting to determine encoding err: %s msg: %s", err, msg)
		} else {
			// log.Printf("Ignoring this file msg: %s", msg)
		}
	}
	if len(toEncode) > 0 {
		return EncodeContainerContent(&toEncode, cm)
	}
	log.Printf("Did not find anything to encode under %s", c.Name)
	return nil, nil
}

// Do encoding in parallel but many fewer processors.  It will be an interesting mix of
// disk write and CPU use vs single process and heavy disk use.  Plus encoding video takes
// some pretty serious memory use. ffmpeg also uses multiple cores already.
func EncodeContainerContent(toEncode *utils.EncodingRequests, cm ContentManager) (*utils.EncodingResults, error) {
	expected := len(*toEncode)
	log.Printf("Attempting to encode N(%d) video files", expected)
	cfg := config.GetCfg()
	processors := cfg.CoreCount / 2 // TODO: Another config... SO MANY
	if processors <= 0 {
		processors = 1 // Without at least one processor this will hang forever
	}
	log.Printf("Starting %d processors for video encoding", processors)

	reply := make(chan utils.EncodingResult, expected)
	input := make(chan utils.EncodingRequest, expected)
	// Starts the workers
	for i := 0; i < processors; i++ {
		pw := utils.EncodingWorker{Id: i, In: input}
		go StartEncoder(pw)
	}

	for _, req := range *toEncode {
		req.Out = reply
		input <- req
	}

	total := 0
	results := utils.EncodingResults{}
	for res := range reply {
		total++
		if total == expected {
			close(input)
			close(reply)
		}
		r_cp := res
		results = append(results, r_cp)
	}
	// We don't really have an error case here.
	return &results, nil
}

func StartEncoder(ew utils.EncodingWorker) {
	for req := range ew.In {
		c := req.C
		mc := req.Mc

		// Should check the on disk size and add a check to look at a post encode filesize
		log.Printf("Worker %d doing encoding for %d - %s\n", ew.Id, mc.ID, mc.Src)
		msg, err, converted := utils.ConvertVideoToH265(req.SrcFile, req.DstFile)
		if err == nil && !converted {
			err = fmt.Errorf("a request was made to convert %s but it did not encode %s", req.SrcFile, msg)
		}

		// ConvertVideoToH265 does a lot of the 'same size', is encoded tests.
		encodedSize := int64(0)
		if err == nil {
			_, encodedSize, err, _ = utils.IsValidVideo(req.DstFile)
		}

		log.Printf("Size of the media %d and encoded size %d", mc.SizeBytes, encodedSize)
		req.Out <- utils.EncodingResult{
			C_ID:        c.ID,
			MC_ID:       mc.ID,
			NewVideo:    req.DstFile,
			InitialSize: mc.SizeBytes,
			EncodedSize: encodedSize,
			Err:         err,
		}
	}
}
