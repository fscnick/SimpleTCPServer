package main

import "flag"
import "fmt"
import "os"
import "net/http"
import "runtime"

type helloHandler struct{}

var version bool 
var help bool
var maxConcurrency int
 
var listenAddr string
var servePath string

var concurrentCount chan struct{}


func parseArguments() {

	flag.BoolVar(&version, "build", false, "print GoLang build version.")
	flag.BoolVar(&help, "help", false, "print this message.")

	flag.IntVar(&maxConcurrency, "max", 30, "Max concurrent request.")

	flag.StringVar(&listenAddr, "listen", ":18080", "The address to listening.")
	flag.StringVar(&servePath, "path", "/", "The path wants to serve..")

	flag.Parse()

	return
}

func (h *helloHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if len(concurrentCount) > maxConcurrency {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte("Busy!"))
		return
	}
	
	// TODO: consider the chan might be full
	concurrentCount <- struct{}{}
	w.Write([]byte("Hello, world!"))
	<- concurrentCount
}

func main() {

	parseArguments()
	if flag.NArg() > 0 || help {
		flag.PrintDefaults()
		return
	}

	if version {
		fmt.Fprintf(os.Stderr, "%s\n", runtime.Version())
		return
	}

	concurrentCount = make(chan struct{}, 65536)


	http.Handle(servePath, &helloHandler{})
	err := http.ListenAndServe(listenAddr, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		return 
	}
}