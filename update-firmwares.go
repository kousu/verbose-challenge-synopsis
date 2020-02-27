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

	"encoding/csv"
	"encoding/json"
)

type ApplicationUpdateSpec struct {
	ApplicationId string `json:"applicationId"`
	Version       string `json:"version"`
}

type ApplicationUpdateProfile struct {
	Applications []ApplicationUpdateSpec `json:"applications"`
}

type ApplicationUpdate struct {
	Profile ApplicationUpdateProfile `json:"profile"`
}

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

	verboseLog.Println("TouchTunes Fleet Updater Starting Up")

	// Pre-compute the update-request JSON
	var appSpec []ApplicationUpdateSpec
	for app, version := range apps {
		appSpec = append(appSpec, ApplicationUpdateSpec{app, version})
	}
	verboseLog.Println(appSpec)

	update, err := json.MarshalIndent(
		ApplicationUpdate{
			Profile: ApplicationUpdateProfile{
				Applications: appSpec,
			},
		},
		"",
		"  ")
	if err != nil {
		log.Fatalf(err.Error())
	}
	verboseLog.Printf("Updating players with:\n%s\n", string(update))

	// TODO: add a worker pool here so that updates can happen in parallel.

	devices := csv.NewReader(input)

	header, err := devices.Read()
	if err != nil {
		log.Fatalf(err.Error())
	}
	headerM := make(map[string]int)
	for i, h := range header {
		headerM[h] = i
	} // go doesn't have an indexOf() method; this is the second-best way
	macAddresses_i, exists := headerM["mac_addresses"]
	if !exists {
		log.Fatalf("Batch input missing 'mac_addresses' column")
	}
	r := 0
	for {
		r += 1
		record, err := devices.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf(err.Error())
		}

		clientId := record[macAddresses_i]

		macAddresses := regexp.MustCompile(`^([0-9A-Fa-f]{2}:){5}([0-9A-Fa-f]{2})$`)
		if !macAddresses.MatchString(clientId) {
			log.Printf("Warning: invalid MAC address '%s' on line %d.\n", clientId, r)
			continue // ignore, don't break, the batch job on invalid lines
		}

		verboseLog.Printf("Updating client '%s'\n", clientId)
	}

	api = api

	verboseLog.Println("TouchTunes Fleet Updater Finished")
}
