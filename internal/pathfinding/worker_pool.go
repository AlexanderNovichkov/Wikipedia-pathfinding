package pathfinding

import (
	"log"
	"sync/atomic"
	"time"
)

type workerPool struct {
	tasksQueue   chan func()
	workersCount int
	stopWorkers  chan struct{}

	tasksStarted int32
	tasksDone    int32
	startTime    time.Time
}

func newWorkersPool(queueSize int, workersCount int) *workerPool {
	pool := &workerPool{
		tasksQueue:   make(chan func(), queueSize),
		workersCount: workersCount,
		stopWorkers:  make(chan struct{}),
		startTime:    time.Now(),
	}
	go pool.logInfo()
	return pool
}

func (pool *workerPool) logInfo() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			log.Println("Tasks started:", pool.tasksStarted, "Tasks done:", pool.tasksDone,
				"efficiency:", float64(pool.tasksDone)/time.Now().Sub(pool.startTime).Seconds())
		case <-pool.stopWorkers:
			return
		}
	}
}

func (pool *workerPool) start() {
	for i := 0; i < pool.workersCount; i++ {
		go func() {
			for {
				select {
				case task := <-pool.tasksQueue:
					task()
					atomic.AddInt32(&pool.tasksDone, 1)
				case <-pool.stopWorkers:
					return
				}
			}
		}()
	}
}

func (pool *workerPool) addTask(task func()) {
	pool.tasksQueue <- task
	atomic.AddInt32(&pool.tasksStarted, 1)
}

func (pool *workerPool) stop() {
	close(pool.stopWorkers)
}
