package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"path/filepath"
	"regexp"
)

// TODO: for unit testing main(), you can use
// os.Args = []string{"something", "something"}

func main() {

	api, apiSet := os.LookupEnv("TOUCHTUNES_FLEET_API")
	if !apiSet {
		api = "http://localhost:6565"
	}

	apps := make(map[string]string)
	var input *os.File

	{
		// scope to limit the reach of the temporary argument parsing variables

		// Define and parse command line
		flag.Usage = func() {
			fmt.Fprintf(os.Stderr, "Usage: %s [options] [application_id=vM.m.p ...] device_csv\n", filepath.Base(os.Args[0]))
			fmt.Fprintf(os.Stderr, "Options:\n")
			flag.PrintDefaults()
		}
		verboseFlag := flag.Bool("v", false, "verbose mode")
		flag.Parse()
		rest := flag.Args()
		if len(rest) < 1 {
			log.Fatalf("Missing 'device_csv' argument.")
		}
		inputFname := rest[len(rest)-1]
		rest = rest[:len(rest)-1]

		// record target application versions to install
		re := regexp.MustCompile(`(?P<application_id>\w+)=(?P<version>v\d+\.\d+\.\d+)`)
		for _, app := range rest {
			pieces := re.FindStringSubmatch(app)
			if pieces == nil {
				log.Fatalf("Invalid app version specification '%v'", app)
			}
			apps[pieces[1]] = pieces[2]
		}

		verboseFlag = verboseFlag
		inputFname = inputFname
	}

	api = api
	input = input
}
