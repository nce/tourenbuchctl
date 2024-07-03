package oauth

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"
)

const (
	timeoutDelay  = 1 * time.Minute
	serverTimeout = 1 * time.Minute
)

type callbackResult struct {
	Interface interface{}
	Error     error
}

func runCallbackServer(callback func(w http.ResponseWriter, r *http.Request) (interface{}, error)) func() (interface{}, error) {

	port := authCallbackPort
	redirectURI := authCallbackPath

	resultCh := make(chan callbackResult)

	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		fmt.Println(err)
	}

	ln, err := net.Listen("tcp", addr.String())
	if err != nil {
		fmt.Println(err)
	}
	callbackPort := ln.Addr().(*net.TCPAddr).Port

	m := http.NewServeMux()
	server := &http.Server{
		ReadTimeout: serverTimeout,
		Handler:     m,
		Addr:        fmt.Sprintf(":%d", callbackPort),
	}

	m.HandleFunc(redirectURI, func(w http.ResponseWriter, r *http.Request) {
		// Got a response, call the callback function.
		i, err := callback(w, r)
		resultCh <- callbackResult{Interface: i, Error: err}
	})

	// Start the server.
	go func() {
		err := server.Serve(ln)
		if err != http.ErrServerClosed {
			resultCh <- callbackResult{Error: err}
		}
	}()

	closeAndReturn := func() (interface{}, error) {
		var finalErr error

		// Block till the callback gives us a result.
		var r callbackResult
		select {
		case r = <-resultCh:
			finalErr = r.Error
		case <-time.After(timeoutDelay):
			finalErr = error(fmt.Errorf("timeout waiting for callback"))
		}

		// Shutdown the server.
		err := server.Shutdown(context.Background())
		if err != nil {
			return nil, err
		}

		// Return the result.
		return r.Interface, finalErr
	}

	return closeAndReturn

}
