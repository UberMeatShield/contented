/**
 * Functions that deal with handling something in the task queue.
 */
package managers

import (
	"contented/models"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gobuffalo/nulls"
	"github.com/gofrs/uuid"
)

/**
 * TODO: Move all the editing queue tasks into a new file.
 */
type DuplicateContent struct {
	KeepContentID uuid.UUID `json:"keep_id"`
	ContainerID   uuid.UUID `json:"container_id"`
	ContainerName string    `json:"container_name"`
	DuplicateID   uuid.UUID `json:"duplicate_id"`
	KeepSrc       string    `json:"keep_src"`
	DuplicateSrc  string    `json:"duplicate_src"`
}
type DuplicateContents []DuplicateContent

func (dupe DuplicateContent) String() string {
	s, _ := json.MarshalIndent(dupe, "", "  ")
	return string(s)
}

func (sr ContentQuery) String() string {
	s, _ := json.MarshalIndent(sr, "", "  ")
	return string(s)
}

// This is a little sketchy because the memory version already does a lookup on status
func ChangeTaskState(man ContentManager, task *models.TaskRequest, newStatus models.TaskStatusType, msg string) (*models.TaskRequest, error) {
	log.Printf("Changing Task State %s to %s", task, newStatus)
	status := task.Status.Copy()
	if status == newStatus {
		return nil, fmt.Errorf("task %s already in state %s", task, newStatus)
	}
	if newStatus == models.TaskStatus.IN_PROGRESS {
		task.StartedAt = time.Now().UTC()
	}
	task.Status = newStatus
	task.Message = strings.ReplaceAll(msg, man.GetCfg().Dir, "")
	return man.UpdateTask(task, status)
}

func FailTask(man ContentManager, task *models.TaskRequest, errMsg string) (*models.TaskRequest, error) {
	log.Print(errMsg)

	status := task.Status.Copy()
	if status == models.TaskStatus.ERROR {
		return nil, fmt.Errorf("task %s already in state %s", task, models.TaskStatus.ERROR)
	}
	task.Status = models.TaskStatus.ERROR
	task.ErrMsg = strings.ReplaceAll(errMsg, man.GetCfg().Dir, "")
	return man.UpdateTask(task, status)
}

/**
 * Grab a content related task
 */
func TakeContentTask(man ContentManager, id uuid.UUID, operation string) (*models.TaskRequest, *models.Content, error) {
	task, tErr := man.GetTask(id)
	if tErr != nil {
		log.Printf("%s Could not look up the task successfully %s", operation, tErr)
		return task, nil, tErr
	}
	task, pErr := ChangeTaskState(man, task, models.TaskStatus.PENDING, "Starting to execute task")
	if pErr != nil {
		msg := fmt.Sprintf("%s Couldn't move task into pending %s", operation, pErr)
		FailTask(man, task, msg)
		return task, nil, pErr
	}
	content, cErr := man.GetContent(task.ContentID.UUID)
	if cErr != nil {
		msg := fmt.Sprintf("%s Content not found %s %s", operation, task.ContentID.UUID, cErr)
		FailTask(man, task, msg)
		return task, content, cErr
	}
	task, upErr := ChangeTaskState(man, task, models.TaskStatus.IN_PROGRESS, fmt.Sprintf("Content was found %s", content.Src))
	if upErr != nil {
		msg := fmt.Sprintf("%s Failed to update task state to in progress %s", operation, upErr)
		FailTask(man, task, msg)
		return task, content, upErr
	}
	return task, content, nil
}

/**
 * Grab a container related task
 */
func TakeContainerTask(man ContentManager, id uuid.UUID, operation string) (*models.TaskRequest, *models.Container, error) {
	task, tErr := man.GetTask(id)
	if tErr != nil {
		log.Printf("%s Could not look up the task successfully %s", operation, tErr)
		return task, nil, tErr
	}
	task, pErr := ChangeTaskState(man, task, models.TaskStatus.PENDING, "Starting to execute task")
	if pErr != nil {
		msg := fmt.Sprintf("%s Couldn't move task into pending %s", operation, pErr)
		FailTask(man, task, msg)
		return task, nil, pErr
	}
	container, cErr := man.GetContainer(task.ContainerID.UUID)
	if cErr != nil {
		msg := fmt.Sprintf("%s Container not found %s %s", operation, task.ContainerID.UUID, cErr)
		FailTask(man, task, msg)
		return task, container, cErr
	}
	task, upErr := ChangeTaskState(man, task, models.TaskStatus.IN_PROGRESS, fmt.Sprintf("Container was found %s", container.Name))
	if upErr != nil {
		msg := fmt.Sprintf("%s Failed to update task state to in progress %s", operation, upErr)
		FailTask(man, task, msg)
		return task, container, upErr
	}
	return task, container, nil
}

/**
 * Capture a set of screens given a task
 */
func ScreenCaptureTask(man ContentManager, id uuid.UUID) error {
	log.Printf("Managers Screen Tasks taskID attempting to start %s", id)
	task, _, err := TakeContentTask(man, id, "Screenshots")
	if err != nil {
		return err
	}
	screens, sErr, pattern := CreateScreensForContent(man, task.ContentID.UUID, task.NumberOfScreens, task.StartTimeSeconds)
	if sErr != nil {
		failMsg := fmt.Sprintf("Failing to create screen %s", sErr)
		FailTask(man, task, failMsg)
		return sErr
	}
	// Should strip the path information out of the task state
	ChangeTaskState(man, task, models.TaskStatus.DONE, fmt.Sprintf("Successfully created screens %s", pattern))
	log.Printf("Screens %s and the pattern %s", screens, pattern)
	return sErr
}

/**
 * Capture a set of screens given a task
 */
func WebpFromScreensTask(man ContentManager, id uuid.UUID) error {
	log.Printf("Managers WebP taskID attempting to start %s", id)
	task, content, err := TakeContentTask(man, id, "WebpFromScreensTask")
	if err != nil {
		return err
	}

	webp, err := WebpFromContent(man, content)
	if err != nil {
		failMsg := fmt.Sprintf("Failing to create screen %s", err)
		FailTask(man, task, failMsg)
		return err
	}

	// Should strip the path information out of the task state
	ChangeTaskState(man, task, models.TaskStatus.DONE, fmt.Sprintf("Successfully created webp %s", webp))
	return err
}

/**
 * Capture a set of screens given a task
 */
func DetectDuplicatesTask(man ContentManager, id uuid.UUID) error {
	log.Printf("Managers duplicate content taskID attempting to start %s", id)
	task, container, err := TakeContainerTask(man, id, "DetectDuplicatesTask")
	if err != nil {
		return err
	}

	dupes := FindDuplicateContents(man, container, "video")
	if err != nil {
		failMsg := fmt.Sprintf("Failing to detect duplicates %s", err)
		FailTask(man, task, failMsg)
		return err
	}

	// JSON encode the duplicate information into the summary
	summary := fmt.Sprintf("%s", dupes)

	// TODO: This should actually generate something semi useful for the UI
	ChangeTaskState(man, task, models.TaskStatus.DONE, summary)
	return err
}

/**
 * Tag a piece of content, get this working on one item and then consider some other operation.
 */
func TaggingContentTask(man ContentManager, id uuid.UUID) error {
	log.Printf("Managers Tagging taskID attempting to start %s", id)
	task, content, err := TakeContentTask(man, id, "TaggingContentTask")
	if err != nil {
		return err
	}

	// Get all the tags (this is expensive if I am kicking off a lot of them)
	tq := TagQuery{PerPage: 90001}
	tags, total, tErr := man.ListAllTags(tq)
	if tErr != nil || total == 0 || tags == nil {
		failMsg := fmt.Sprintf("Failed to tag content %s", err)
		FailTask(man, task, failMsg)
		return err
	}

	// TODO: Make it so this can work on a single piece of content (refactor AssignTagsAndUpdate)
	contents := models.Contents{*content}
	assignmentError := AssignTagsToContents(man, &contents, tags)
	if assignmentError != nil {
		failMsg := fmt.Sprintf("Failed to tag content %s", err)
		FailTask(man, task, failMsg)
		return err
	}

	// Should strip the path information out of the task state
	ChangeTaskState(man, task, models.TaskStatus.DONE, fmt.Sprintf("successfully tagged content %s", content.ID))
	return err
}

/**
 * Could definitely make this a method assuming the next task uses the same logic.
 */
func EncodingVideoTask(man ContentManager, id uuid.UUID) error {
	log.Printf("Managers Video encoding taskID attempting to start %s", id)
	task, content, err := TakeContentTask(man, id, "VideoEncoding")
	if err != nil {
		return err
	}
	msg, encodeErr, shouldEncode, newFile := EncodeVideoContent(man, content, task.Codec)
	log.Printf("Video Encode video %s %s %t", msg, encodeErr, shouldEncode)
	if encodeErr != nil {
		failMsg := fmt.Sprintf("Failed to encode %s", encodeErr)
		FailTask(man, task, failMsg)
		return encodeErr
	}

	encodedContent, eErr := CreateContentAfterEncoding(man, content, newFile)
	if eErr != nil {
		failMsg := fmt.Sprintf("Failed to determine the newly encoded file %s", eErr)
		FailTask(man, task, failMsg)
		return eErr
	}

	task.CreatedID = nulls.NewUUID(encodedContent.ID) // Note that this could already have existed.
	taskMsg := fmt.Sprintf("Completed video encoding %s and had to encode %t", msg, shouldEncode)
	_, doneErr := ChangeTaskState(man, task, models.TaskStatus.DONE, taskMsg)
	return doneErr
}
