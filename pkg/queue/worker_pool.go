package queue

import (
	"context"
	"fmt"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// Task represents a unit of work
type Task func(ctx context.Context) error

// WorkerPool manages a pool of workers to process tasks
type WorkerPool struct {
	tasks   chan Task
	workers int
	wg      sync.WaitGroup
	ctx     context.Context
	cancel  context.CancelFunc
	metrics *PoolMetrics
}

// PoolMetrics tracks worker pool statistics
type PoolMetrics struct {
	mu             sync.RWMutex
	tasksSubmitted int64
	tasksCompleted int64
	tasksFailed    int64
	tasksRejected  int64
}

// NewWorkerPool creates a new worker pool
func NewWorkerPool(workers int, queueSize int) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())
	wp := &WorkerPool{
		tasks:   make(chan Task, queueSize),
		workers: workers,
		ctx:     ctx,
		cancel:  cancel,
		metrics: &PoolMetrics{},
	}
	wp.start()
	return wp
}

func (wp *WorkerPool) start() {
	for i := 0; i < wp.workers; i++ {
		wp.wg.Add(1)
		go wp.worker(i)
	}
	log.Infof("Started worker pool with %d workers", wp.workers)
}

func (wp *WorkerPool) worker(id int) {
	defer wp.wg.Done()

	for {
		select {
		case <-wp.ctx.Done():
			log.Infof("Worker %d shutting down", id)
			return
		case task := <-wp.tasks:
			startTime := time.Now()
			if err := task(wp.ctx); err != nil {
				wp.metrics.incrementFailed()
				log.WithFields(log.Fields{
					"worker_id": id,
					"error":     err.Error(),
					"duration":  time.Since(startTime),
				}).Error("Worker task failed")
			} else {
				wp.metrics.incrementCompleted()
				log.WithFields(log.Fields{
					"worker_id": id,
					"duration":  time.Since(startTime),
				}).Debug("Worker task completed")
			}
		}
	}
}

// Submit adds a task to the queue
func (wp *WorkerPool) Submit(task Task) error {
	wp.metrics.incrementSubmitted()

	select {
	case wp.tasks <- task:
		return nil
	case <-wp.ctx.Done():
		wp.metrics.incrementRejected()
		return wp.ctx.Err()
	default:
		wp.metrics.incrementRejected()
		return ErrQueueFull
	}
}

// Shutdown gracefully stops the worker pool
func (wp *WorkerPool) Shutdown() {
	log.Info("Shutting down worker pool...")
	wp.cancel()

	// Wait for all workers to finish with timeout
	done := make(chan struct{})
	go func() {
		wp.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Info("All workers stopped gracefully")
	case <-time.After(30 * time.Second):
		log.Warn("Worker pool shutdown timeout, some tasks may not have completed")
	}

	close(wp.tasks)
	wp.logMetrics()
}

// GetMetrics returns current pool metrics
func (wp *WorkerPool) GetMetrics() PoolMetrics {
	wp.metrics.mu.RLock()
	defer wp.metrics.mu.RUnlock()
	return PoolMetrics{
		tasksSubmitted: wp.metrics.tasksSubmitted,
		tasksCompleted: wp.metrics.tasksCompleted,
		tasksFailed:    wp.metrics.tasksFailed,
		tasksRejected:  wp.metrics.tasksRejected,
	}
}

func (wp *WorkerPool) logMetrics() {
	metrics := wp.GetMetrics()
	log.WithFields(log.Fields{
		"submitted": metrics.tasksSubmitted,
		"completed": metrics.tasksCompleted,
		"failed":    metrics.tasksFailed,
		"rejected":  metrics.tasksRejected,
	}).Info("Worker pool final metrics")
}

func (m *PoolMetrics) incrementSubmitted() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tasksSubmitted++
}

func (m *PoolMetrics) incrementCompleted() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tasksCompleted++
}

func (m *PoolMetrics) incrementFailed() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tasksFailed++
}

func (m *PoolMetrics) incrementRejected() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tasksRejected++
}

var ErrQueueFull = fmt.Errorf("task queue is full")
