package main

import (
	"testing"

	"io"
	"net/http"
	"net/http/httptest"
	//"fmt"
	//"os"
	//"log"
)

// Things to test:
// - the empty app set
// - apps with bad specs
// - missing device_csv parameter
// - badly formatted csv

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
