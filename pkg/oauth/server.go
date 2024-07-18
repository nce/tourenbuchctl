package oauth

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"
)

const (
	timeoutDelay  = 1 * time.Minute
	serverTimeout = 1 * time.Minute
)

var ErrFailedTimeout = errors.New("timeout waiting for callback")

type callbackResult struct {
	Interface interface{}
	Error     error
}

func runCallbackServer(callback func(w http.ResponseWriter,
	r *http.Request) (interface{}, error),
) func() (interface{}, error) {
	port := authCallbackPort
	redirectURI := authCallbackPath

	resultCh := make(chan callbackResult)

	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		return func() (interface{}, error) {
			return nil, fmt.Errorf("TCP connection initialization: %w", err)
		}
	}

	listen, err := net.Listen("tcp", addr.String())
	if err != nil {
		return func() (interface{}, error) {
			return nil, fmt.Errorf("TCP server listen initialization: %w", err)
		}
	}

	//nolint: forcetypeassert
	callbackPort := listen.Addr().(*net.TCPAddr).Port

	mux := http.NewServeMux()
	server := &http.Server{
		ReadTimeout: serverTimeout,
		Handler:     mux,
		Addr:        fmt.Sprintf(":%d", callbackPort),
	}

	mux.HandleFunc(redirectURI, func(w http.ResponseWriter, r *http.Request) {
		// Got a response, call the callback function.
		i, err := callback(w, r)
		resultCh <- callbackResult{Interface: i, Error: err}
	})

	// Start the server.
	go func() {
		err := server.Serve(listen)
		if errors.Is(err, http.ErrServerClosed) {
			resultCh <- callbackResult{Error: err}
		}
	}()

	closeAndReturn := func() (interface{}, error) {
		var finalErr error

		// Block till the callback gives us a result.
		var result callbackResult
		select {
		case result = <-resultCh:
			finalErr = result.Error
		case <-time.After(timeoutDelay):
			finalErr = fmt.Errorf("failed waiting %w", ErrFailedTimeout)
		}

		// Shutdown the server.
		err := server.Shutdown(context.Background())
		if err != nil {
			return nil, fmt.Errorf("server shutdown: %w", err)
		}

		// Return the result.
		return result.Interface, finalErr
	}

	return closeAndReturn
}
