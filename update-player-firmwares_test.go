package main

import (
	"testing"

	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"sort"
	"strings"
	"sync"
)

func TestUpdate200(t *testing.T) {
	loaded := false

	var s http.ServeMux
	s.HandleFunc("/profiles/clientId:aa:bb:cc:dd:ee:ff", func(w http.ResponseWriter, r *http.Request) {
		// echo input
		w.Header().Add("Content-Type", r.Header.Get("Content-Type"))
		io.Copy(w, r.Body)

		loaded = true
	})
	srv := httptest.NewServer(&s)
	defer srv.Close()

	api = srv.URL

	err := update("aa:bb:cc:dd:ee:ff", []byte(`{
  "profile": {    
    "applications": [
      {
        "applicationId": "music_app"
        "version": "v1.4.10"
      },
      {
        "applicationId": "diagnostic_app",
        "version": "v1.2.6"
      },
      {
        "applicationId": "settings_app",
        "version": "v1.1.5"
      }
    ]
  }
}
`))
	if err != nil {
		t.Error(err)
	}

	if !loaded {
		t.Errorf("Expected URL was not accessed correctly")
	}

}

func TestUpdate401(t *testing.T) {
	loaded := false

	var s http.ServeMux
	s.HandleFunc("/profiles/clientId:aa:bb:cc:dd:ee:ff", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")

		http.Error(w, `{
  "statusCode": 401,
  "error": "Unauthorized",
  "message": "invalid clientId or token supplied"
}
`, http.StatusUnauthorized)

		loaded = true
	})
	srv := httptest.NewServer(&s)
	defer srv.Close()

	api = srv.URL

	err := update("aa:bb:cc:dd:ee:ff", []byte(`{
  "profile": {
    "applications": [
      {
        "applicationId": "music_app"
        "version": "v1.4.10"
      },
      {
        "applicationId": "diagnostic_app",
        "version": "v1.2.6"
      },
      {
        "applicationId": "settings_app",
        "version": "v1.1.5"
      }
    ]
  }
}
`))
	if err == nil {
		t.Errorf("update() incorrectly succeeded on %v", http.StatusUnauthorized)
	} else if !strings.Contains(err.Error(), "401 Unauthorized") { // this is a spot check, it's not an exhaustive verification
		t.Error("update() did not report expected HTTP error code")
	}

	if !loaded {
		t.Errorf("Expected URL was not accessed correctly")
	}

}

func TestUpdate404(t *testing.T) {
	loaded := false

	var s http.ServeMux
	s.HandleFunc("/profiles/clientId:aa:bb:cc:dd:ee:ff", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")

		http.Error(w, `{
  "statusCode": 404,
  "error": "Not Found",
  "message": "profile of client aa:bb:cc:dd:ee:ff does not exist"
}
`, http.StatusNotFound)

		loaded = true
	})
	srv := httptest.NewServer(&s)
	defer srv.Close()

	api = srv.URL

	err := update("aa:bb:cc:dd:ee:ff", []byte(`{
  "profile": {
    "applications": [
      {
        "applicationId": "music_app"
        "version": "v1.4.10"
      },
      {
        "applicationId": "diagnostic_app",
        "version": "v1.2.6"
      },
      {
        "applicationId": "settings_app",
        "version": "v1.1.5"
      }
    ]
  }
}
`))
	if err == nil {
		t.Errorf("update() incorrectly succeeded on %v", http.StatusUnauthorized)
	} else if !strings.Contains(err.Error(), "404 Not Found") { // this is a spot check, it's not an exhaustive verification
		t.Error("update() did not report expected HTTP error code")
	}

	if !loaded {
		t.Errorf("Expected URL was not accessed correctly")
	}

}

func TestUpdate409(t *testing.T) {
	loaded := false

	var s http.ServeMux
	s.HandleFunc("/profiles/clientId:aa:bb:cc:dd:ee:ff", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")

		http.Error(w, `{
  "statusCode": 409,
  "error": "Conflict",
  "message": "child \"profile\" fails because [child \"applications\" fails because [\"applications\" is required]]"
}
`, http.StatusConflict)

		loaded = true
	})
	srv := httptest.NewServer(&s)
	defer srv.Close()

	api = srv.URL

	err := update("aa:bb:cc:dd:ee:ff", []byte(`{"profile": {}`))
	if err == nil {
		t.Errorf("update() incorrectly succeeded on %v", http.StatusUnauthorized)
	} else if !strings.Contains(err.Error(), "409 Conflict") { // this is a spot check, it's not an exhaustive verification
		t.Error("update() did not report expected HTTP error code")
	}

	if !loaded {
		t.Errorf("Expected URL was not accessed correctly")
	}

}

func TestUpdate500(t *testing.T) {
	loaded := false

	var s http.ServeMux
	s.HandleFunc("/profiles/clientId:aa:bb:cc:dd:ee:ff", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")

		http.Error(w, `{
  "statusCode": 500,
  "error": "Internal Server Error",
  "message": "An internal server error occurred"
}`, http.StatusInternalServerError)

		loaded = true
	})
	srv := httptest.NewServer(&s)
	defer srv.Close()

	api = srv.URL

	err := update("aa:bb:cc:dd:ee:ff", []byte(`{"profile": {}`))
	if err == nil {
		t.Errorf("update() incorrectly succeeded on %v", http.StatusUnauthorized)
	} else if !strings.Contains(err.Error(), "500 Internal Server Error") { // this is a spot check, it's not an exhaustive verification
		t.Error("update() did not report expected HTTP error code")
	}

	if !loaded {
		t.Errorf("Expected URL was not accessed correctly")
	}
}

func TestUpdateNoApps(t *testing.T) {
	log.SetOutput(ioutil.Discard)

	passed := false

	var s http.ServeMux
	s.HandleFunc("/profiles/clientId:aa:bb:cc:dd:ee:ff", func(w http.ResponseWriter, r *http.Request) {
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			fmt.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		expected := []byte(`{
  "profile": {
    "applications": null
  }
}`)
		if bytes.Compare(expected, data) != 0 {
			err := fmt.Errorf("Expected %s, recieved %s", expected, data)
			t.Error(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			w.Header().Add("Content-Type", "application/json")
			w.Write([]byte("{}"))
			passed = true
		}
	})
	srv := httptest.NewServer(&s)
	defer srv.Close()

	api = srv.URL

	err := batchUpdate(bytes.NewBufferString("mac_addresses\naa:bb:cc:dd:ee:ff"), map[string]string{}, 1)
	if err != nil {
		t.Error(err)
	}
	if !passed {
		t.Errorf("Did not pass")
	}
}

func TestUpdateMultipleDevices(t *testing.T) {

	log.SetOutput(ioutil.Discard)

	var lock sync.Mutex
	var loaded []string = nil
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lock.Lock()
		loaded = append(loaded, r.URL.String())
		lock.Unlock()

		w.Header().Add("Content-Type", "application/json")
		w.Write([]byte("{}"))
	}))
	defer srv.Close()

	api = srv.URL

	err := batchUpdate(bytes.NewBufferString("mac_addresses\naa:bb:cc:dd:ee:ff\nee:ee:ee:ee:ee:ee\n12:34:56:78:9a:bc"), map[string]string{}, 2)
	if err != nil {
		t.Error(err)
	}

	sort.Strings(loaded)
	expected := []string{"/profiles/clientId:12:34:56:78:9a:bc", "/profiles/clientId:aa:bb:cc:dd:ee:ff", "/profiles/clientId:ee:ee:ee:ee:ee:ee"}
	if !reflect.DeepEqual(expected, loaded) {
		t.Errorf("Expected to see %v but actually %v were loaded", expected, loaded)
	}
}

func TestUpdateMultipleDevicesMalformed1(t *testing.T) {

	log.SetOutput(ioutil.Discard)

	var lock sync.Mutex
	var loaded []string = nil
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lock.Lock()
		loaded = append(loaded, r.URL.String())
		lock.Unlock()

		w.Header().Add("Content-Type", "application/json")
		w.Write([]byte("{}"))
	}))
	defer srv.Close()

	api = srv.URL

	err := batchUpdate(bytes.NewBufferString(
		`mac_addresses
aa:bb:cc:dd:ee:ff
192.168.0.1
12:34:56:78:9a:bc
`), map[string]string{}, 1)
	if err != nil {
		t.Error(err)
	}

	sort.Strings(loaded)
	expected := []string{"/profiles/clientId:12:34:56:78:9a:bc", "/profiles/clientId:aa:bb:cc:dd:ee:ff"}
	if !reflect.DeepEqual(expected, loaded) {
		t.Errorf("Expected to see %v but actually %v were loaded", expected, loaded)
	}
}

func TestUpdateMultipleDevicesMalformed2(t *testing.T) {

	log.SetOutput(ioutil.Discard)

	var lock sync.Mutex
	var loaded []string = nil
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lock.Lock()
		loaded = append(loaded, r.URL.String())
		lock.Unlock()

		w.Header().Add("Content-Type", "application/json")
		w.Write([]byte("{}"))
	}))
	defer srv.Close()

	api = srv.URL

	err := batchUpdate(bytes.NewBufferString(`size,ip_addresses
9,10.0.0.1
3,10.0.0.2
9,10.0.0.3
`), map[string]string{}, 3)
	if err == nil {
		t.Errorf("Expected an error but got none")
	} else if !strings.Contains(err.Error(), "Batch input missing 'mac_addresses' column") {
		t.Errorf("Expected 'Batch input missing 'mac_addresses' column' but received: %s", err.Error())
	}
}

func TestUpdateParseNoApps(t *testing.T) {
	os.Args = append(os.Args[:1], "-")
	csv, _, apps, err := parseCommand()
	if err != nil {
		t.Fatalf("Error should not have happened: %v", err)
	}

	if csv != os.Stdin {
		t.Errorf("Input file should be stdin")
	}

	expectedApps := map[string]string{}
	if !reflect.DeepEqual(expectedApps, apps) {
		t.Errorf("App specification was mis-parsed. Expected %v, got %v", expectedApps, apps)
	}
}

func TestUpdateAppParsing(t *testing.T) {
	os.Args = append(os.Args[:1], "app1=v3.4.0", "app2=v5.4.3", "-")
	csv, _, apps, err := parseCommand()
	if err != nil {
		t.Fatalf("Error should not have happened: %v", err)
	}

	if csv != os.Stdin {
		t.Errorf("Input file should be stdin")
	}

	expectedApps := map[string]string{
		"app2": "v5.4.3",
		"app1": "v3.4.0",
	}
	if !reflect.DeepEqual(expectedApps, apps) {
		t.Errorf("App specification was mis-parsed. Expected %v, got %v", expectedApps, apps)
	}
}

func TestUpdateAppParsingMalformed(t *testing.T) {
	os.Args = append(os.Args[:1], "app1=v3.4.0", "-", "app2=v5.4.3")
	_, _, _, err := parseCommand()
	if err == nil {
		t.Fatalf("Error should have happened")
	} else if !strings.Contains(err.Error(), "Invalid app version specification") {
		t.Fatalf("Unexpected error: %v", err.Error())
	}
}

func TestUpdateMalformedArgs(t *testing.T) {
	os.Args = append(os.Args[:1])
	_, _, _, err := parseCommand()
	if err == nil {
		t.Fatalf("Error should have happened")
	} else if !strings.Contains(err.Error(), "Missing 'device_csv' argument.") {
		t.Fatalf("Unexpected error: %v", err.Error())
	}
}
