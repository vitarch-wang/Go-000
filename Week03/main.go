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

	g, ctx1 := errgroup.WithContext(ctx)

	g.Go(func() error {
		go HTTPServer(cancel)

		go func() {
			time.Sleep(10 * time.Second)
			cancel()
			return
		}()

		<-ctx1.Done()
		return nil
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
			return errors.New("catch stop signal")
		}
	})

	if err := g.Wait(); err == nil {
		fmt.Println("success shutdown")
		os.Exit(0)
	} else {
		fmt.Println("error shutdown")
		os.Exit(1)
	}

}

func HTTPServer(cancel context.CancelFunc) error {
	http.HandleFunc("/", IndexHandler)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		cancel()
		return err
	}
	return nil
}
func IndexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "hello world")
	fmt.Println("get request")
	fmt.Println(r.URL)

}
