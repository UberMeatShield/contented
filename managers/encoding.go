package managers

import (
  "fmt"
  "log"
  "strings"
  "contented/models"
  "contented/utils"
  //"github.com/gobuffalo/nulls"
  //"github.com/gofrs/uuid"
  "github.com/pkg/errors" 
)

// Init a manager and pass it in or just do this via config value instead of a pass in
func EncodeVideos(cm ContentManager) error {
    cnts, c_err := cm.ListContainers(0, 9001) // Might need to make this smarter (obviously)
    if c_err != nil {
        return c_err
    }
    if len(*cnts) == 0 {
        return errors.New("No Containers were found in the database")
    }

    err_msg := []string{}
    for _, cnt := range *cnts {
        _, err := EncodeContainer(&cnt, cm)
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

func EncodeContainer(c* models.Container, cm ContentManager) (*utils.EncodingResults, error) {
    content, q_err := cm.ListContent(c.ID, 0, 90000)
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
            ref_mc := mc  // Ensure the pointers don't get messed up
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
    return nil, nil
}

// Do encoding in parallel but many fewer processors.  It will be an interesting mix of
// disk write and CPU use vs single process and heavy disk use.  Plus encoding video takes
// some pretty serious memory use.
func EncodeContainerContent(toEncode *utils.EncodingRequests, cm ContentManager) (*utils.EncodingResults, error) {
    expected := len(*toEncode)
    log.Printf("Attempting to encode N(%d) video files", expected)

    cfg := utils.GetCfg()
    processors := cfg.CoreCount // TODO: Another config... SO MANY
    if processors <= 0 {
        processors = 1 // Without at least one processor this will hang forever
    }

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
        log.Printf("Finished and found result: %s", res)
        //results = append(results, res)
    }
    return &results, nil
}

func StartEncoder(ew utils.EncodingWorker) {
    for req := range ew.In {
        c := req.C
        mc := req.Mc

        // Should check the on disk size and add a check to look at a post encode filesize
        log.Printf("Worker %d Doing encoding for %s - %s\n", ew.Id, mc.ID.String(), mc.Src)

        msg, err, converted := utils.ConvertVideoToH256(req.SrcFile, req.DstFile)

        if err == nil && converted == false {
            err = errors.New(fmt.Sprintf("A request was made to convert %s but it did not encode %s", req.SrcFile, msg))
        }

        // Check that the destination is ACTUALLY a valid file
        encodedSize := int64(0)
        if err == nil {
            _, encodedSize, err = utils.IsValidVideo(req.DstFile)
        }

        log.Printf("Size of the media %d and encoded %d", mc.SizeBytes, encodedSize)
        req.Out <- utils.EncodingResult{
            C_ID:    c.ID,
            MC_ID:   mc.ID,
            NewVideo: req.DstFile,
            InitialSize: mc.SizeBytes,
            EncodedSize: encodedSize,
            Err: err,
        }
    }
}
