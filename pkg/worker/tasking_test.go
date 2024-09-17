package worker

import (
	"contented/models"
	"testing"
)

// TestNewTaskQueueWithMaxConcurrent tests the creation of a new TaskQueue with max concurrent tasks
// Also mostly generated with Cursor.  The update was pretty close to working but didn't get all the
// constructor elements.
func TestNewTaskQueueWithMaxConcurrent(t *testing.T) {
	bufferSize := 10
	maxConcurrent := MaxConcurrentTasks(5)
	tq := NewTaskQueue(bufferSize, maxConcurrent)

	if tq == nil {
		t.Error("NewTaskQueue returned nil")
		return
	}

	if cap(tq.inputQueue) != bufferSize {
		t.Errorf("Expected input queue capacity of %d, got %d", bufferSize, cap(tq.inputQueue))
	}

	if len(tq.typeHandlers) != 0 {
		t.Errorf("Expected empty typeHandlers, got %d handlers", len(tq.typeHandlers))
	}

	if tq.maxConcurrentTasks != maxConcurrent {
		t.Errorf("Expected maxConcurrentTasks to be %d, got %d", maxConcurrent, tq.maxConcurrentTasks)
	}

	if cap(tq.runningTasks) != int(maxConcurrent) {
		t.Errorf("Expected runningTasks capacity to be %d, got %d", maxConcurrent, cap(tq.runningTasks))
	}
}

// TestRegisterTaskHandler tests registering a task handler
func TestRegisterTaskHandler(t *testing.T) {
	tq := NewTaskQueue(1, 5)
	handler := func(Task) error { return nil }
	tq.RegisterTaskHandler("test", handler)
	if len(tq.typeHandlers) != 1 {
		t.Errorf("Expected 1 handler, got %d", len(tq.typeHandlers))
	}
	if _, exists := tq.typeHandlers["test"]; !exists {
		t.Error("Handler not registered for 'test' type")
	}
}

// TestEnqueueTask tests enqueueing a task
func TestEnqueueTask(t *testing.T) {
	tq := NewTaskQueue(1, 5)
	task := Task{ID: int64(123), Operation: models.TaskOperation.ENCODING}
	tq.EnqueueTask(task)
	select {
	case receivedTask := <-tq.inputQueue:
		if receivedTask != task {
			t.Errorf("Expected task %v, got %v", task, receivedTask)
		}
	default:
		t.Error("Task was not enqueued")
	}
}
