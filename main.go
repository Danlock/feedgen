package main

import (
	"context"
	"fmt"
	"time"

	"github.com/danlock/go-rss-gen/feedgen"
)

const mangaUpdatePollFrequency = 6 * time.Hour

func throwIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	out := feedgen.PollMangaUpdatesForReleases(ctx, 2*time.Second)
	timer := time.NewTimer(10 * time.Second)
	for {
		select {
		case <-timer.C:
			cancel()
			return
		case releases := <-out:
			fmt.Printf("%+v", releases)
		}
	}
}
