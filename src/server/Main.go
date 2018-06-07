package main

import "flag"
import "fmt"
import "os"
//import "net/http"
import "runtime"


var opts Options

var resoucePool *ResourcePoolWrapper

func parseArguments() {

	flag.BoolVar(&opts.version, "build", false, "print GoLang build version.")
	flag.BoolVar(&opts.debug, "debug", false, "print debug message.")
	flag.BoolVar(&opts.help, "help", false, "print this message.")

	// TODO: Is this need?
	flag.IntVar(&opts.maxgo, "maxgo", 4, "maximum threads for go-routine.")
	flag.IntVar(&opts.worker, "worker", 30, "The number of work to process incoming text and forwarding message.")
	flag.IntVar(&opts.incomingTimeout, "incomingTimeout", 10, "The seconds to disconnect if no input message from client.")
	flag.IntVar(&opts.outgoingTimeout, "outgoingTimeout", 10, "The seconds to disconnect if no response of external api.")
	flag.IntVar(&opts.bufferSize, "bufferSize", 1024, "The buffer of processing incoming text in bytes.")

	flag.StringVar(&opts.listen, "listen", ":8080", "The address to listening.")
	flag.StringVar(&opts.to, "to", "http://127.0.0.1:18080/", "The url forward to external api.")

	flag.Parse()

	return
}

func initGlobal(opts Options) error {
	if opts.maxgo <= 0 {
		return fmt.Errorf("Option maxgo %d is invalid.", opts.maxgo)
	}

	return nil
}

func main() {

	parseArguments()
	if flag.NArg() > 0 || opts.help {
		flag.PrintDefaults()
		return
	}

	if opts.version {
		fmt.Fprintf(os.Stderr, "%s\n", runtime.Version())
		return
	}

	err := initGlobal(opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Init fail: %s\n", err.Error())
		os.Exit(1)
	}

	err = initServer(opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Init fail: %s\n", err.Error())
		os.Exit(1)
	}

	err = serve()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Start server fail: %s\n", err.Error())
		os.Exit(1)
	}

	fmt.Println("Service Terminated.")
}