package main

import (
	"flag"
	"os"

	"github.com/danlock/go-rss-gen/lib/logger"

	"github.com/joho/godotenv"
)

func helpAndQuit() {
	flag.Usage()
	os.Exit(0)
}

func main() {
	logger.SetupLogger()
	var (
		dotenvLocation string
		help           bool
	)
	flag.StringVar(&dotenvLocation, "e", "./ops/.env", "Location of .env file with environment variables in KEY=VALUE format")
	flag.BoolVar(&help, "h", false, "Show help")
	flag.Parse()
	if help {
		helpAndQuit()
	}
	godotenv.Overload(dotenvLocation)
	switch flag.Arg(0) {
	case "poll":

	default:
		helpAndQuit()
	}
}
