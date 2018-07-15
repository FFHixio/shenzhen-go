// The http_server command was automatically generated by Shenzhen Go.
package main

import (
	"context"
	"fmt"
	"github.com/google/shenzhen-go/dev/parts"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"time"
)

var _ = runtime.Compiler

func HTTPServeMux(metrics chan<- *parts.HTTPRequest, requests <-chan *parts.HTTPRequest, root chan<- *parts.HTTPRequest) {
	multiplicity := runtime.NumCPU()
	mux := http.NewServeMux()
	mux.Handle("/", parts.HTTPHandler(root))
	mux.Handle("/metrics", parts.HTTPHandler(metrics))

	defer func() {
		close(root)
		close(metrics)

	}()
	var multWG sync.WaitGroup
	multWG.Add(multiplicity)
	defer multWG.Wait()
	for n := 0; n < multiplicity; n++ {
		go func(instanceNumber int) {
			defer multWG.Done()
			for req := range requests {
				// Borrow fix for Go issues #3692 and #5955.
				if req.Request.RequestURI == "*" {
					if req.Request.ProtoAtLeast(1, 1) {
						req.ResponseWriter.Header().Set("Connection", "close")
					}
					req.ResponseWriter.WriteHeader(http.StatusBadRequest)
					req.Close()
					continue
				}
				h, _ := mux.Handler(req.Request)
				hh, ok := h.(parts.HTTPHandler)
				if !ok {
					// ServeMux may return handlers that weren't added above.
					h.ServeHTTP(req.ResponseWriter, req.Request)
					req.Close()
					continue
				}
				hh <- req
			}
		}(n)
	}
}

func HTTPServer(errors chan<- error, manager <-chan parts.HTTPServerManager, requests chan<- *parts.HTTPRequest) {
	multiplicity := 1

	defer func() {
		close(requests)
		if errors != nil {
			close(errors)
		}

	}()
	const instanceNumber = 0

	for mgr := range manager {
		svr := &http.Server{
			Handler: parts.HTTPHandler(requests),
			Addr:    mgr.Addr(),
		}
		done := make(chan struct{})
		go func() {
			err := svr.ListenAndServe()
			if err != nil && errors != nil {
				errors <- err
			}
			close(done)
		}()
		err := svr.Shutdown(mgr.Wait())
		if err != nil && errors != nil {
			errors <- err
		}
		<-done
	}
}

func Hello_World(requests <-chan *parts.HTTPRequest) {
	multiplicity := runtime.NumCPU()

	var multWG sync.WaitGroup
	multWG.Add(multiplicity)
	defer multWG.Wait()
	for n := 0; n < multiplicity; n++ {
		go func(instanceNumber int) {
			defer multWG.Done()
			for rw := range requests {
				rw.Write([]byte("Hello, HTTP!\n"))
				rw.Close()
			}
		}(n)
	}
}

func Log_errors(errors <-chan error) {
	multiplicity := 1

	const instanceNumber = 0
	for err := range errors {
		log.Printf("HTTP server: %v", err)
	}
}

func Metrics(requests <-chan *parts.HTTPRequest) {
	multiplicity := runtime.NumCPU()

	var multWG sync.WaitGroup
	multWG.Add(multiplicity)
	defer multWG.Wait()
	for n := 0; n < multiplicity; n++ {
		go func(instanceNumber int) {
			defer multWG.Done()

			h := promhttp.Handler()
			for r := range requests {
				h.ServeHTTP(r.ResponseWriter, r.Request)
				r.Close()
			}

		}(n)
	}
}

func Send_a_manager(manager chan<- parts.HTTPServerManager) {
	multiplicity := 1

	defer func() {
		close(manager)
	}()
	const instanceNumber = 0
	mgr := parts.NewHTTPServerManager(":8765")
	manager <- mgr

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	fmt.Println("Press Ctrl-C (or SIGINT) to shut down.")
	<-sig

	timeout := 5 * time.Second
	fmt.Printf("Shutting down within %v...\n", timeout)
	ctx, _ := context.WithTimeout(context.Background(), timeout)
	mgr.Shutdown(ctx)
}

func main() {

	channel0 := make(chan *parts.HTTPRequest, 0)
	channel3 := make(chan error, 0)
	channel5 := make(chan *parts.HTTPRequest, 0)
	channel6 := make(chan *parts.HTTPRequest, 0)
	channel9 := make(chan parts.HTTPServerManager, 0)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		HTTPServeMux(channel5, channel0, channel6)
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		HTTPServer(channel3, channel9, channel0)
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		Hello_World(channel6)
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		Log_errors(channel3)
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		Metrics(channel5)
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		Send_a_manager(channel9)
		wg.Done()
	}()

	// Wait for the various goroutines to finish.
	wg.Wait()
}
