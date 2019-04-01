package main

import (
	"log"
	"time"

	"github.com/danlock/go-rss-gen/feedgen"
)

const releasePollFrequency = 6 * time.Hour

func main() {
	// ctx, cancel := context.WithCancel(context.Background())
	// out := feedgen.PollMUForReleases(ctx, 2*time.Second)
	// timer := time.NewTimer(10 * time.Second)
	// for {
	// 	select {
	// 	case <-timer.C:
	// 		cancel()
	// 		return
	// 	case releases := <-out:
	// 		fmt.Printf("%+v", releases)
	// 	}
	// }
	infos, _ := feedgen.QueryAllMUSeries()
	log.Printf("%+v", infos)
}
