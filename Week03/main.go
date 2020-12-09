package main

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/sync/errgroup"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {

	ctx, cancel := context.WithCancel(context.Background())

	mux := http.NewServeMux()
	mux.HandleFunc("/", IndexHandler)
	server := &http.Server{
		Addr:    "0.0.0.0:8080",
		Handler: mux,
	}

	g, ctx1 := errgroup.WithContext(ctx)

	g.Go(func() error {
		return server.ListenAndServe()
	})

	g.Go(func() error {
		tick := time.NewTicker(10 * time.Second)
		select {
		case <-tick.C:
			cancel()
			server.Shutdown(ctx1)
			return errors.New("timeout")
		case <-ctx1.Done():
			return nil
		}

	})

	g.Go(func() error {
		stopSignal := make(chan os.Signal, 1)
		signal.Notify(stopSignal,
			os.Interrupt,
			syscall.SIGINT,
			os.Kill,
			syscall.SIGKILL,
			syscall.SIGTERM,
			syscall.SIGHUP,
			syscall.SIGQUIT,
		)
		select {
		case <-ctx1.Done():
			fmt.Println("http server shutdown")
			return nil
		case s := <-stopSignal:
			fmt.Println("catch stop signal" + s.String())
			cancel()
			_ = server.Shutdown(ctx1)
			return errors.New("stop by signal:" + s.String())
		}
	})

	err := g.Wait()
	fmt.Println("shutdown with :" + err.Error())
	os.Exit(0)

}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "hello world")
	fmt.Println("get request")
	fmt.Println(r.URL)

}
