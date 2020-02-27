package main

import (
	"testing"

	"io"
	"net/http"
	"net/http/httptest"
	"strings"
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
