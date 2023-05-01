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

    cnts, c_err := cm.ListContainers(0, 9001) // Might need to make this smarter :(
    if c_err != nil {
        return c_err
    }
    if len(*cnts) == 0 {
        return errors.New("No Containers were found in the database")
    }

    err_msg := []string{}
    for _, cnt := range *cnts {
        err := EncodeContainer(&cnt, cm)
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

func EncodeContainer(c* models.Container, cm ContentManager) (error){
    content, q_err := cm.ListContent(c.ID, 0, 90000)
    if q_err != nil {
      log.Fatal(q_err) // Also fatal if we can no longer list content (empty is just [])
    }
    for _, mc := range *content {
        srcFile, _ := utils.GetFilePathInContainer(mc.Src, c.GetFqPath())

        // DstFile should split off the final extension \.xyz and replace it
        dstFile := fmt.Sprintf("%s.%s", srcFile, "[h256].mp4")
        msg, err, encode := utils.ShouldEncodeVideo(srcFile, dstFile)

        if encode {
            log.Printf("Will attempt to convert %s", msg)
        } else if err != nil {
            log.Printf("Error attempting to encode %s msg: %s", err, msg)
        } else {
            log.Printf("Ignoring this file msg: %s", msg)
        }
    }
    return nil
}
