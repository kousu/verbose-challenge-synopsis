package main

import (
	"bytes"
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
	"net/http"
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

type ApiError struct {
	StatusCode int    `json:"statusCode"`
	Error      string `json:"error"`
	Message    string `json:"message"`
}

var verboseLog *log.Logger

// TODO: for unit testing main(), you can use
// os.Args = []string{"something", "something"}

func main() {
	log.SetFlags(0) // disable timestamps ; XXX don't do this?

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

	// Load environment parameters
	api, apiSet := os.LookupEnv("TOUCHTUNES_FLEET_API")
	if !apiSet {
		api = "http://fleet.intra.touchtones.example.com:6565"
	}

	hostname, err := os.Hostname()
	if err != nil {
		log.Fatalf(err.Error())
	}

	authToken, authTokenSet := os.LookupEnv("TOUCHTUNES_AUTH_TOKEN")
	if !authTokenSet {
		log.Println("Warning: TOUCHTUNES_AUTH_TOKEN should be defined.")
		authToken = "<required>"
	}

	// Pre-compute the update-request JSON
	var appSpec []ApplicationUpdateSpec
	for app, version := range apps {
		appSpec = append(appSpec, ApplicationUpdateSpec{app, version})
	}

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

		url := api + "/profiles/clientId:" + clientId
		verboseLog.Printf("Updating '%s'\n", url)

		req, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(update))
		if err != nil {
			log.Println(err.Error())
			continue
		}
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("X-Client-ID", hostname) // Assumption: this is meant to identify the *user*; if it was meant to identify the software it would be "User-Agent"
		req.Header.Add("X-Authentication-Token", authToken)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Println(err.Error())
			continue
		}
		if resp.StatusCode == http.StatusOK {
			log.Printf("%s updated.\n", url)
		} else {
			if resp.Header.Get("Content-Type") == "application/json" {
				var respDetails ApiError
				err := json.NewDecoder(resp.Body).Decode(&respDetails)
				if err != nil {
					log.Printf("%s %s: %v\n", req.Method, url, err.Error())
				} else {
					log.Printf("%s %s: %v\n", req.Method, url, respDetails.Message)
				}
			} else {
				// Nonsense result; maybe we're not talking to the right API?
				log.Printf("%s %s: %v\n", req.Method, url, resp.Status)
			}
		}
	}

	verboseLog.Println("TouchTunes Fleet Updater Finished")
}
