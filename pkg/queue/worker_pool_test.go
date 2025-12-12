package queue

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWorkerPool_Submit(t *testing.T) {
	wp := NewWorkerPool(2, 10)
	defer wp.Shutdown()

	var counter int
	var mu sync.Mutex

	task := func(ctx context.Context) error {
		mu.Lock()
		counter++
		mu.Unlock()
		return nil
	}

	// Submit 5 tasks
	for i := 0; i < 5; i++ {
		err := wp.Submit(task)
		assert.NoError(t, err)
	}

	// Wait for tasks to complete
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	assert.Equal(t, 5, counter)
	mu.Unlock()
}

func TestWorkerPool_SubmitError(t *testing.T) {
	wp := NewWorkerPool(1, 5)
	defer wp.Shutdown()

	taskErr := errors.New("task failed")

	task := func(ctx context.Context) error {
		return taskErr
	}

	err := wp.Submit(task)
	assert.NoError(t, err) // Submission should succeed

	// Wait for task to execute
	time.Sleep(50 * time.Millisecond)

	metrics := wp.GetMetrics()
	assert.Equal(t, int64(1), metrics.tasksFailed)
}

func TestWorkerPool_QueueFull(t *testing.T) {
	// Create a small pool
	wp := NewWorkerPool(1, 2)
	defer wp.Shutdown()

	blockingTask := func(ctx context.Context) error {
		time.Sleep(100 * time.Millisecond)
		return nil
	}

	// Submit tasks until queue is full
	for i := 0; i < 3; i++ {
		wp.Submit(blockingTask)
	}

	// This should fail because queue is full
	err := wp.Submit(blockingTask)
	
	// Either ErrQueueFull or context error
	assert.Error(t, err)
}

func TestWorkerPool_Metrics(t *testing.T) {
	wp := NewWorkerPool(2, 10)
	defer wp.Shutdown()

	successTask := func(ctx context.Context) error {
		time.Sleep(10 * time.Millisecond)
		return nil
	}

	failTask := func(ctx context.Context) error {
		return errors.New("fail")
	}

	// Submit successful tasks
	for i := 0; i < 3; i++ {
		wp.Submit(successTask)
	}

	// Submit failing tasks
	for i := 0; i < 2; i++ {
		wp.Submit(failTask)
	}

	// Wait for completion
	time.Sleep(200 * time.Millisecond)

	metrics := wp.GetMetrics()
	assert.Equal(t, int64(5), metrics.tasksSubmitted)
	assert.Equal(t, int64(3), metrics.tasksCompleted)
	assert.Equal(t, int64(2), metrics.tasksFailed)
}

func TestWorkerPool_Shutdown(t *testing.T) {
	wp := NewWorkerPool(2, 10)

	task := func(ctx context.Context) error {
		time.Sleep(50 * time.Millisecond)
		return nil
	}

	wp.Submit(task)
	wp.Submit(task)

	// Shutdown should wait for tasks to complete
	wp.Shutdown()

	// After shutdown, submissions should fail
	err := wp.Submit(task)
	assert.Error(t, err)
}

