package main

import "os"
import "fmt"
import "net"
import "bytes"
import "time"
import "net/http"

// type Worker struct {
// 	workerId   int
// 	workerPool chan chan Job
// 	jobChannel chan Job
// 	quit       chan bool
// 	//crc32Table  *crc32.Table
// 	//crc32Engine hash.Hash32
// }

// type Dispatcher struct {
// 	// A pool of workers channels that are registered with the dispatcher
// 	workerPool chan chan Job
// 	maxWorkers int

// 	// Recieve job and dispatch to registered worker.
// 	jobQueue chan Job

// 	// Write out queue
// 	outputQueue chan *Job
// }

var isInitDone bool
var contextPool *ResourcePoolWrapper
var listenAddr string
var dstUrl string
var numWorker int
var incomingTimeout int
var outgoingTimeout int
var bufferSize int

// var dispatcher *Dispatcher

func initServer(opts Options) error {
	listenAddr = opts.listen
	dstUrl = opts.to

	if opts.worker <= 0 {
		return fmt.Errorf("Option worker %d is invalid.", opts.worker)
	}
	numWorker = opts.worker

	if opts.incomingTimeout <= 0 || opts.incomingTimeout >= 120 {
		return fmt.Errorf("Option incomingTimeout must be within the range of 1 and 120 seconds.", opts.worker)
	}
	incomingTimeout = opts.incomingTimeout

	if opts.outgoingTimeout <= 0 || opts.outgoingTimeout >= 120 {
		return fmt.Errorf("Option forwardTimeout must be within the range of 1 to 120 seconds.", opts.worker)
	}
	outgoingTimeout = opts.outgoingTimeout

	if opts.bufferSize <= 128 || opts.bufferSize >= 65536 {
		return fmt.Errorf("Option bufferSize must be within the range of 128 to 65536 seconds.", opts.worker)
	}
	bufferSize = opts.bufferSize

	contextPool = &ResourcePoolWrapper{}
	err := contextPool.InitPool(numWorker, initContext)
	if err != nil {
		fmt.Errorf("Init context pool fails.")
		return err
	}

	isInitDone = true
	return nil
}

func initContext() (interface{}, error) {
	return &Context{
		buff: make([]byte, bufferSize),
		httpClient: &http.Client{Timeout: time.Duration(opts.outgoingTimeout) * time.Second},
	}, nil
}

func serve() error {
	// TODO: server start

	if isInitDone != true {
		return fmt.Errorf("Please call initServer at first.")
	}

	server, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stdout, "Listening on %s\n", listenAddr)

	for {
		conn, err := server.Accept()
		if err != nil {
			// TODO: log something
			fmt.Println("Accept fail.")
			continue
		}

		bb := make([]byte, bufferSize)
		nn, err := conn.Read(bb)
		if err != nil {
			// TODO: log something
			fmt.Println("Read fails")
			continue
		}

		fmt.Println("Success read ", nn, "bytes.")
		fmt.Println(string(bb[:bytes.IndexByte(bb, 0)]))
		conn.Close()
	}

}