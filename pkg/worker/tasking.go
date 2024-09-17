package worker

/*
 * Initial framework was built out using Cursor (color me impressed)
 */
import (
	"contented/pkg/models"
	"fmt"
	"log"
)

// Task represents a unit of work to be processed
type Task struct {
	ID        int64                    `json:"id"`
	Operation models.TaskOperationType `json:"operation" db:"operation"`
}

func (t Task) String() string {
	return fmt.Sprintf("Task %d: %s", t.ID, t.Operation)
}

// TaskHandler is a function type for processing tasks
type TaskHandler func(Task) error
type MaxConcurrentTasks int

// TaskQueue manages the distribution of tasks to different channels
type TaskQueue struct {
	inputQueue         chan Task
	typeHandlers       map[string]TaskHandler
	maxConcurrentTasks MaxConcurrentTasks
	runningTasks       chan struct{}
}

// NewTaskQueue creates a new TaskQueue
func NewTaskQueue(bufferSize int, maxConcurrent MaxConcurrentTasks) *TaskQueue {
	return &TaskQueue{
		inputQueue:         make(chan Task, bufferSize),
		typeHandlers:       make(map[string]TaskHandler),
		maxConcurrentTasks: maxConcurrent,
		runningTasks:       make(chan struct{}, maxConcurrent),
	}
}

// Start begins processing tasks from the input queue
func (tq *TaskQueue) Start() {
	go func() {
		for task := range tq.inputQueue {
			tq.runningTasks <- struct{}{} // Acquire a slot
			if handler, exists := tq.typeHandlers[task.Operation.String()]; exists {
				go func(t Task, h TaskHandler) {
					defer func() { <-tq.runningTasks }() // Release the slot when done
					h(t)
				}(task, handler)
			} else {
				<-tq.runningTasks // Release the slot immediately if no handler
				log.Printf("Warning: No handler for task type %s", task.Operation)
			}
		}
	}()
}

// RegisterTaskHandler adds a new task type and its corresponding handler function
func (tq *TaskQueue) RegisterTaskHandler(taskOperation string, handler TaskHandler) {
	tq.typeHandlers[taskOperation] = handler
}

// EnqueueTask adds a task to the input queue
func (tq *TaskQueue) EnqueueTask(task Task) {
	tq.inputQueue <- task
}

// Stop closes the input queue and stops processing new tasks
func (tq *TaskQueue) Stop() {
	close(tq.inputQueue)
}

// GetTaskHandler returns the handler function for a specific task type
func (tq *TaskQueue) GetTaskHandler(taskOperation string) (TaskHandler, error) {
	if handler, exists := tq.typeHandlers[taskOperation]; exists {
		return handler, nil
	}
	return nil, fmt.Errorf("no handler for task type %s", taskOperation)
}
