package actions

import (
    "fmt"
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
            models.DB.Create(&mc)
        }
    }
    return nil
}
