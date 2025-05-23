/**
 * Functions that deal with handling something in the task queue.
 */
package managers

import (
	"contented/pkg/models"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

/**
 * TODO: Move all the editing queue tasks into a new file.
 */
type DuplicateContent struct {
	KeepContentID int64  `json:"keep_id"`
	ContainerID   *int64 `json:"container_id"`
	ContainerName string `json:"container_name"`
	DuplicateID   int64  `json:"duplicate_id"`
	KeepSrc       string `json:"keep_src"`
	DuplicateSrc  string `json:"duplicate_src"`
	FqPath        string `json:"-"`
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
func TakeContentTask(man ContentManager, id int64, operation string) (*models.TaskRequest, *models.Content, error) {
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
	content, cErr := man.GetContent(*task.ContentID)
	if cErr != nil {
		msg := fmt.Sprintf("%s Content not found %d %s", operation, task.ContentID, cErr)
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
 * Grab a container related task that requires a container
 */
func TakeContainerTask(man ContentManager, id int64, operation string) (*models.TaskRequest, *models.Container, *models.Content, error) {
	task, container, content, err := TakeTask(man, id, operation)
	if err != nil {
		return task, container, content, err
	}

	log.Printf("Took the task at least? %s container %s", task, container)

	task, upErr := ChangeTaskState(man, task, models.TaskStatus.IN_PROGRESS, fmt.Sprintf("Container was found %s", container.Name))
	if upErr != nil {
		msg := fmt.Sprintf("%s Failed to update task state to in progress %s", operation, upErr)
		FailTask(man, task, msg)
		return task, container, content, upErr
	}
	return task, container, content, nil
}

func TakeTask(man ContentManager, id int64, operation string) (*models.TaskRequest, *models.Container, *models.Content, error) {
	task, tErr := man.GetTask(id)
	if tErr != nil {
		log.Printf("%s Could not look up the task successfully %s", operation, tErr)
		return task, nil, nil, tErr
	}
	task, pErr := ChangeTaskState(man, task, models.TaskStatus.PENDING, "Starting to execute task")
	if pErr != nil {
		msg := fmt.Sprintf("%s Couldn't move task into pending %s", operation, pErr)
		FailTask(man, task, msg)
		return task, nil, nil, pErr
	}

	//TODO: Make this less ugly....
	var container *models.Container = nil
	if task.ContainerID != nil && *task.ContainerID > 0 {
		cnt, cErr := man.GetContainer(*task.ContainerID)
		container = cnt
		if cErr != nil {
			msg := fmt.Sprintf("%s Container not found %d %s", operation, task.ContainerID, cErr)
			FailTask(man, task, msg)
			return task, container, nil, cErr
		}
	}

	var content *models.Content = nil
	if task.ContentID != nil && *task.ContentID > 0 {
		mc, cErr := man.GetContent(*task.ContentID)
		content = mc
		if cErr != nil {
			msg := fmt.Sprintf("%s Content not found %d %s", operation, task.ContentID, cErr)
			FailTask(man, task, msg)
			return task, container, content, cErr
		}
		// Fallback to containerID lookup based on content (nulls.uint is ANNOYING)
		if content != nil && content.ContainerID != nil && *content.ContainerID > 0 && container == nil {
			container, _ = man.GetContainer(*content.ContainerID)
		}
	}
	return task, container, content, nil
}

/**
 * Capture a set of screens given a task
 */
func ScreenCaptureTask(man ContentManager, id int64) error {
	log.Printf("Managers Screen Tasks taskID attempting to start %d", id)
	task, _, err := TakeContentTask(man, id, "Screenshots")
	if err != nil {
		return err
	}
	screens, pattern, sErr := CreateScreensForContent(man, *task.ContentID, task.NumberOfScreens, task.StartTimeSeconds)
	if sErr != nil {
		failMsg := fmt.Sprintf("Failing to create screen %s", sErr)
		FailTask(man, task, failMsg)
		return sErr
	}
	// Should strip the path information out of the task state
	ChangeTaskState(man, task, models.TaskStatus.DONE, fmt.Sprintf("Successfully created screens %s", pattern))
	log.Printf("Screens %s and the pattern %s", screens, pattern)

	// TODO: Come up with a way to submit a follow-up task to create a webp from the screens
	return sErr
}

/**
 * Capture a set of screens given a task
 */
func WebpFromScreensTask(man ContentManager, id int64) error {
	log.Printf("Managers WebP taskID attempting to start %d", id)
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
 * Remove a duplicate content
 */
func RemoveDuplicateContentTask(man ContentManager, id int64) error {
	log.Printf("Managers duplicate content taskID attempting to start %d", id)
	task, cnt, _, err := TakeContainerTask(man, id, "RemoveDuplicateContentTask")
	if err != nil {
		return err
	}
	removed, err := RemoveDuplicateContents(man, cnt)
	if err != nil {
		failMsg := fmt.Sprintf("Failed to remove duplicate contents %s", err)
		FailTask(man, task, failMsg)
		return err
	}
	// Remove the content from the database
	successMsg := fmt.Sprintf("Successfully removed %d duplicate contents from container %d", removed, cnt.ID)
	ChangeTaskState(man, task, models.TaskStatus.DONE, successMsg)
	return err
}

/**
 * Capture a set of screens given a task
 */
func DetectDuplicatesTask(man ContentManager, id int64) error {
	log.Printf("Managers duplicate content taskID attempting to start %d", id)
	task, container, content, err := TakeContainerTask(man, id, "DetectDuplicatesTask")

	// Make it so the task has a contentId
	// Maybe a take task should return the container AND the content if specified
	// It should be smart enough to get a container or content
	// TODO: Figure out how to get container vs content(if there is an error her)
	if err != nil {
		log.Printf("Error attempting to start a dupe task %s", err)
		return err
	}

	// ContainerID is currently required
	cs := ContentQuery{
		ContentType: "video",
		PerPage:     9001, // TODO: Seriously come up with a better paging method...
	}
	if content != nil {
		cs.ContentID = strconv.FormatInt(content.ID, 10)
	}
	if container != nil {
		cs.ContainerID = strconv.FormatInt(container.ID, 10)
	}
	log.Printf("DetectDuplicates Output starting with query container %s and contentID %s", cs.ContentID, cs.ContainerID)

	// Is this actually a full failure?
	dupes, dupeErrors := FindDuplicateContents(man, container, cs)
	if len(dupeErrors) > 0 {
		failMsg := fmt.Sprintf("Failing to detect duplicates %s", dupeErrors)
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
func TaggingContentTask(man ContentManager, id int64) error {
	log.Printf("Managers Tagging taskID attempting to start %d", id)
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
	ChangeTaskState(man, task, models.TaskStatus.DONE, fmt.Sprintf("successfully tagged content %d", content.ID))
	return err
}

/**
 * Could definitely make this a method assuming the next task uses the same logic.
 */
func EncodingVideoTask(man ContentManager, id int64) error {
	log.Printf("Managers Video encoding taskID attempting to start %d", id)
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

	task.CreatedID = &encodedContent.ID // Note that this could already have existed.
	taskMsg := fmt.Sprintf("Completed video encoding %s and had to encode %t", msg, shouldEncode)
	_, doneErr := ChangeTaskState(man, task, models.TaskStatus.DONE, taskMsg)
	return doneErr
}
