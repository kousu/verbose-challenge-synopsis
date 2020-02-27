package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

	"path/filepath"
	"regexp"
)

var verboseLog *log.Logger

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

		// Configure logging
		{
			log.SetFlags(0) // disable timestamps ; XXX don't do this?

			var target io.Writer
			if *verboseFlag {
				target = os.Stderr
			} else {
				target = ioutil.Discard
			}
			verboseLog = log.New(target, log.Prefix(), log.Flags())
		}

		// Prepare batch input file
		if inputFname == "-" {
			input = os.Stdin
		} else {
			var err error
			input, err = os.Open(inputFname)
			if err != nil {
				log.Fatalf(err.Error())
			}
		}
	}

	api = api
	input = input
	verboseLog.Println("TouchTunes Fleet Updater Starting Up")
}
