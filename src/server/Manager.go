package main

import "fmt"
import "net/http"
import "encoding/json"

type Status struct {
	Free	int 	`json:"Free"`
	Used 	int 	`json:"Used"`
	Rps 	int 	`json:"Rps"`
}

var isManagerInit bool

var httpServerAddr string
var statusPath string

func initManager(opts Options) error {

	if len(opts.httpListenAddr) == 0 {
		return fmt.Errorf("Http listening address is empty.")
	}
	httpServerAddr = opts.httpListenAddr

	if len(opts.statusPath) == 0 {
		return fmt.Errorf("Status path is emtpy.")
	}
	statusPath = opts.statusPath

	isManagerInit = true
	return nil

}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	freeCount := jobPool.getFreeRescourceCount()
	useCount := jobPool.getUsedRescourceCount()

	// TODO: Make this more precisely.
	rps := useCount

	jj := Status{
		Free: freeCount,
		Used: useCount,
		Rps: rps,
	}

	json.NewEncoder(w).Encode(jj)

}

func startManagerServer() error {
	if isManagerInit == false {
		return fmt.Errorf("Manager has not initialized yet.")
	}

	go func() {
		http.HandleFunc("/"+statusPath, statusHandler)
		err := http.ListenAndServe(httpServerAddr, nil)
		if err != nil {
			fmt.Println("Start manager server error.", err )
		}
	}()
	
	return nil
}