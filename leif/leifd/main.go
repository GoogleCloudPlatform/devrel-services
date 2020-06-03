package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/GoogleCloudPlatform/devrel-services/leif"

	// "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

func main() {
	fmt.Println("Hello World!")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)
	go func() {
		sig := <-signalCh
		fmt.Printf("termination signal received: %s", sig)
		cancel()
	}()

	corpus := leif.Corpus{}
	err := corpus.Init(ctx)

	if err != nil {
		fmt.Println(err)
		return
	}

	group, ctx := errgroup.WithContext(context.Background())

	group.Go(func() error {
		return corpus.Sync(ctx)
	})

	fmt.Println(group.Wait())
}