package fetrunner

import (
	"encoding/json"
	"fetrunner/internal/base"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

// Implement a framework for running a basic app using just the
// default logger.

var (
	stop_request bool = false
	run_finished bool = false
)

// The function passed as parameter calls `Do` and handles the result strings.
func RunLoop(do func(string, ...string) bool) {
	do("RUN_TT")
	for {
		if stop_request {
			do("_STOP_TT")
		}
		do("_POLL_TT")
		if run_finished {
			return
		}
	}
}

type DispatcherOp struct {
	Op   string
	Data []string
}

type DispatcherResult struct {
	Type string
	Text string
}

func Do(op string, data ...string) ([]string, bool) {
	jsonbytes, err := json.Marshal(DispatcherOp{op, data})
	if err != nil {
		panic(err)
	}
	res := Dispatch(string(jsonbytes))
	v := []DispatcherResult{}
	json.Unmarshal([]byte(res), &v)
	ok := true
	var resultlist []string
	for _, r := range v {
		if r.Type == base.ERROR.String() {
			ok = false
		}
		if r.Text == ".TICK=-1" {
			run_finished = true
		}

		if r.Type != base.OP_START.String() || r.Text[0] != '_' {
			//fmt.Println(r.Type, r.Text)
			resultlist = append(resultlist, r.Type+" "+r.Text)
		}
	}

	if run_finished {
		fmt.Printf("??? %s %v\n  %+v\n", op, data, v)
	}

	return resultlist, ok
}

// Catch "terminate" signal (goroutine)
func Termination() {
	// Catch termination signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan // wait for signal
	stop_request = true
}
