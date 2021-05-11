package actions

/**
*  These are helpers for use in grifts, but we want them compiling in the dev service in case of breaks.
*
* Bad code in a grift is harder to notice and the compilation with tests also seems a little broken. ie
* you break the grift via main package changes and never notice.  You break the test in a grift directory
* and then the compilation just failed with no error messages.
*/

import (
    "os"
    "log"
    "fmt"
    "strings"
    "path/filepath"
    "contented/models"
    "contented/utils"
    "github.com/pkg/errors"
    "github.com/gofrs/uuid"
    "github.com/gobuffalo/nulls"
)


// Process all the directories and get a valid setup into the DB
func CreateInitialStructure(dir_name string) error {
    dirs := utils.FindContainers(dir_name)
    fmt.Printf("Found %d sub-directories.\n", len(dirs))
    if len(dirs) == 0 {
        return errors.New("No subdirectories found under path: " + dir_name)
    }

    err := models.DB.TruncateAll()
    if err != nil {
        return errors.WithStack(err)
    }

    // TODO: Need to do this in a single transaction vs partial
    // TODO: Print vs log... might need to setup logging in the Grift I guess
    for _, dir := range dirs {
        fmt.Printf("Adding directory %s with id %s\n", dir.Name, dir.ID)

        media := utils.FindMedia(dir, 90001, 0) // A more sensible limit?
        fmt.Printf("Adding Media to %s with total media %d \n", dir.Name, len(media))

        // Use the database version of uuid generation (minimize the miniscule conflict)
        unset_uuid, _ := uuid.FromString("00000000-0000-0000-0000-000000000000")
        dir.ID = unset_uuid
        dir.Total = len(media)
        models.DB.Create(&dir)
        fmt.Printf("Created %s with id %s\n", dir.Name, dir.ID)

        for _, mc := range media {
            mc.ContainerID = nulls.NewUUID(dir.ID) 
            c_err := models.DB.Create(&mc)
            if c_err != nil {
                log.Fatal(c_err)
            }
        }
    }
    return nil
}

// TODO: Move this code into manager (likely?)
func ClearContainerPreviews(c *models.Container) error {
    dst := GetContainerPreviewDst(c) 
    if _, err := os.Stat(dst); os.IsNotExist(err) {
        return nil
    }
    r_err := os.RemoveAll(dst)
    if r_err != nil {
        log.Fatal(r_err)
        return r_err
    }
    return nil
}

// TODO: Move to utils or make it wrapped for some reason?
func GetContainerPreviewDst(c *models.Container) string {
    return filepath.Join(appCfg.Dir, c.Name, "container_previews")
}

func CreateAllPreviews(preview_above_size int64) error {
    cnts := models.Containers{}
    models.DB.All(&cnts)

    if len(cnts) == 0 {
        return errors.New("No Containers were found in the database")
    }
    for _, cnt := range cnts {
        err := CreateContainerPreviews(&cnt, preview_above_size)    
        if err != nil {
            return err
        }
    }
    return nil
}

// TODO: Should this return a total of previews created or something?
func CreateContainerPreviews(c *models.Container, preview_above_size int64) error {
    // Reset the preview directory, then create it fresh
    c_err := ClearContainerPreviews(c)
    if c_err == nil {
        err := utils.MakePreviewPath(GetContainerPreviewDst(c))
        if err != nil {
            log.Fatal(err)
        }
    }
    media := models.MediaContainers{}
    q_err := models.DB.Where("container_id = ?", c.ID).All(&media)
    if q_err != nil {
        log.Fatal(q_err)
        return q_err
    }

    for _, mc := range media {
        prev_path, mc_err := CreateMediaPreview(c, &mc, preview_above_size)
        if mc_err != nil {
            log.Fatal(mc_err)
            return mc_err
        } else {
            if prev_path != "" {
                log.Printf("Created a preview %s for mc %s", prev_path, mc.ID.String())
                mc.Preview = prev_path
                models.DB.Update(&mc)
            } 
        }
    }
    return nil
}

func CreateMediaPreview(c *models.Container, mc *models.MediaContainer, fsize int64) (string, error) {
    cntPath := filepath.Join(appCfg.Dir, c.Name)
    dstPath := GetContainerPreviewDst(c)

    _, exist_err :=  utils.PreviewExists(mc.Src, dstPath)
    if exist_err != nil {
        return "", exist_err
    }
    dstFqPath, err := utils.GetImagePreview(cntPath, mc.Src, dstPath, fsize)
    if err != nil {
        log.Fatal(err)
    }
    return strings.ReplaceAll(dstFqPath, cntPath, ""), err
}

// In the case you want to do this in an more async manner (maybe important if it wraps video content)
func AsyncMediaPreview(c *models.Container, mc *models.MediaContainer, fsize int64, reply chan<- utils.ProcessingResult) {
    preview, err :=  CreateMediaPreview(c, mc, fsize)
    pr := utils.ProcessingResult{
        C_ID: c.ID,
        MC_ID: mc.ID,
        Preview: preview,
        Err: err,
    }
    reply <- pr
}