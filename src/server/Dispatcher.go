package main

 import "fmt"
// import "time"

type ConnectionHandler func(ctx *Context)

type Job struct {
	data         *Context
	fn           ConnectionHandler
}

type Worker struct {
	workerId   int
	workerPool chan chan Job
	jobChannel chan Job
	quit       chan bool
}

type Dispatcher struct {
	// A pool of workers channels that are registered with the dispatcher
	workerPool chan chan Job
	maxWorkers int

	// Recieve job and dispatch to registered worker.
	jobQueue chan Job
}

func newDispatcher(maxWorkers int) *Dispatcher {
	// For simplicity, the number of worker is equal to the number of job.
	// TODO: Check this is safe or make it different.
	return &Dispatcher{
		workerPool:  make(chan chan Job, maxWorkers),
		maxWorkers:  maxWorkers,
		jobQueue:    make(chan Job, maxWorkers),
	}
}

func (d *Dispatcher) run() {
	for i := 0; i < d.maxWorkers; i++ {
		worker := newWorker(d.workerPool, i)
		worker.start()
	}

	go d.dispatch()
}

func (d *Dispatcher) dispatch() {
	var jobChannel chan Job

	for {
		select {
		case job, ok := <-d.jobQueue:
			if ok != true {
				// TODO: chan is closed.
				return
			}

			// TODO:  non-blocking
			jobChannel = <-d.workerPool

			jobChannel <- job
		}
		
		
	}
}

func newWorker(workerPool chan chan Job, id int) *Worker {
	worker := new(Worker)

	(*worker).workerId = id
	(*worker).workerPool = workerPool
	(*worker).jobChannel = make(chan Job)
	(*worker).quit = make(chan bool)

	return worker

}

func (w Worker) start() {
	go func() {
		for {
			// Worker is ready. Register to workerPool
			w.workerPool <- w.jobChannel

			// This is blocking channel.
			select {
			case job := <-w.jobChannel:
				
				job.fn(job.data)
				
				jobPool.ReleaseResource(&job)
				fmt.Println("Release job to pool")
			case <-w.quit:
				// we have recieved a signal to stop
				// TODO: log something
				return
			}
		}
	}()
}