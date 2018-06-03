package main

import "os"
import "fmt"
import "net"
import "time"
import "bytes"
import "unsafe"
import "regexp"
import "strings"
import "net/http"

var isInitDone bool
// var contextPool *ResourcePoolWrapper
var jobPool *ResourcePoolWrapper
var listenAddr string
var dstUrl string
var numWorker int
var incomingTimeout time.Duration
var outgoingTimeout time.Duration
var bufferSize int

var regexValidator *regexp.Regexp

var dispatcher *Dispatcher

var MSG_TEMP_UNAVAILABLE = []byte("Temporary unavailable.")
var MSG_SERVER_ERROR = []byte("Server Error.")
var MSG_INVALID_INPUT = []byte("Invalid input.")
var MSG_BACKEND_BUSY = []byte("Backend busy.")
var MSG_OK = []byte("OK.")
var MSG_QUIT = []byte("Bye bye.")

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
	incomingTimeout =  time.Second * time.Duration(opts.incomingTimeout)

	if opts.outgoingTimeout <= 0 || opts.outgoingTimeout >= 120 {
		return fmt.Errorf("Option forwardTimeout must be within the range of 1 to 120 seconds.", opts.worker)
	}
	outgoingTimeout =  time.Second * time.Duration(opts.outgoingTimeout)

	if opts.bufferSize <= 128 || opts.bufferSize >= 65536 {
		return fmt.Errorf("Option bufferSize must be within the range of 128 to 65536 seconds.", opts.worker)
	}
	bufferSize = opts.bufferSize

	// contextPool = &ResourcePoolWrapper{}
	// err := contextPool.InitPool(numWorker, initContext)
	// if err != nil {
	// 	fmt.Errorf("Init context pool fails.")
	// 	return err
	// }

	jobPool = &ResourcePoolWrapper{}
	err := jobPool.InitPool(numWorker, initJob)
	if err != nil {
		fmt.Errorf("Init context pool fails.")
		return err
	}

	dispatcher = newDispatcher(numWorker)

	isInitDone = true
	return nil
}

func initContext() (interface{}, error) {
	return &Context{
		buff: make([]byte, bufferSize),
		httpClient: &http.Client{Timeout: outgoingTimeout},
	}, nil
}

func initJob() (interface{}, error) {
	return &Job{
		data: &Context{
			buff: make([]byte, bufferSize),
			httpClient: &http.Client{Timeout: time.Duration(opts.outgoingTimeout) * time.Second},
		},
		fn: nil,
	}, nil
}

func serve() error {

	if isInitDone != true {
		return fmt.Errorf("Please call initServer at first.")
	}

	dispatcher.run()

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

		// get resource if fail return busy
		resource := jobPool.GetResource()
		if resource == nil {
			// TODO: log something
			conn.Write(MSG_TEMP_UNAVAILABLE)
			conn.Close()
			continue
		}


		job, ok := resource.(*Job)
		if ok != true {
			// This should not happen.
			// TODO: log something
			conn.Write(MSG_SERVER_ERROR)
			conn.Close()
			continue
		}

		job.data.conn = conn
		job.data.dstUrl = dstUrl
		job.fn = sendToBackend


		select {
		case dispatcher.jobQueue <- *job:
			fmt.Println("Enqueue job success.")
		default:
			fmt.Println("Job queue is full.")
			jobPool.ReleaseResource(job)
		}

		// bb := make([]byte, bufferSize)
		// nn, err := conn.Read(bb)
		// if err != nil {
		// 	// TODO: log something
		// 	fmt.Println("Read fails")
		// 	continue
		// }

		// fmt.Println("Success read ", nn, "bytes.")
		// fmt.Println(string(bb[:bytes.IndexByte(bb, 0)]))
		// conn.Close()
	}

}

func sendToBackend(context *Context) {
	conn := context.conn
	buff := context.buff
	client := context.httpClient
	url := context.dstUrl

	for {
		now := time.Now()
		err := conn.SetReadDeadline(now.Add(incomingTimeout))
		if err != nil {
			fmt.Println("Set timeout fail: ", err)
			// TODO: notify client if possible
			return
		}

		nn, err := conn.Read(buff)
		if err != nil {
			fmt.Println("Read fail: ", err)
			// TODO: notify client if possible
			return
		}

		isValid := isValidText(buff[:nn])
		if isValid != true {
			fmt.Println("Recv invalid content")
			_, err := conn.Write(MSG_INVALID_INPUT)
			if err != nil {
				fmt.Println("Write fail:", err)
				return
			}
			// TODO: drain unread
			continue
		}

		shouldQuit := isQuit(buff[:nn])
		if shouldQuit == true {
			fmt.Println("Client wants to quit.")
			_, err := conn.Write(MSG_QUIT)
			if err != nil {
				fmt.Println("Write fail:", err)
				return
			}

			return
		}

		// TODO: escape quit

		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(buff[:nn]))
		if err != nil {
			fmt.Println("Create POST fail:", err)
			_, err := conn.Write(MSG_SERVER_ERROR)
			if err != nil {
				fmt.Println("Write fail:", err)
				return
			}

			return
		}

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Send POST fail:", err)

			// TODO: maybe use another message instead of MSG_BACKEND_BUSY
			_, err := conn.Write(MSG_BACKEND_BUSY)
			if err != nil {
				fmt.Println("Write fail:", err)
				return
			}

			// TODO: drain unread
			continue
		}
		
		// For simplicity.
		if resp.StatusCode != http.StatusOK {
			fmt.Println("Response with status:", resp.StatusCode)
			resp.Body.Close()
			_, err := conn.Write(MSG_BACKEND_BUSY)
			if err != nil {
				fmt.Println("Write fail:", err)
				return
			}

			// TODO: drain unread
			continue
		}
		
		// TODO: drain unread buffer in conn

		_, err = conn.Write(MSG_OK)
		if err != nil {
			fmt.Println("Write fail:", err)
			resp.Body.Close()
			return
		}

		// TODO: if return length is not equal the MSG_OK.


		resp.Body.Close()
	}

}

func isValidText(input []byte) bool {
	ss := UnsafeBytesToString(input[:]) 

	if strings.IndexFunc(ss, charChecker) != -1 {
        return false
	}
	
	return true

}

func isQuit(input []byte) bool {
	ss := UnsafeBytesToString(input[:]) 
	if ss == "quit" {
		return true
	}

	return false
}

func charChecker(r rune) bool {
	return 	r < ' ' || r > '~'
}

func UnsafeBytesToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func unsafeStringToBytes(ss string) []byte {
	return *(*[]byte)(unsafe.Pointer(&ss))
}

