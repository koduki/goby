package vm

import (
	"testing"
	//"net/http/httptest"
	//"net/http"
	"fmt"
	"io/ioutil"
	"net/http"
)

//chan parameter for blocking until server is prepared
func startTestServer(c chan bool) {
	m := http.NewServeMux()

	m.HandleFunc("/index", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)

		if r.Method == http.MethodPost {
			b, err := ioutil.ReadAll(r.Body)
			if err != nil {
				panic(err)
			}
			fmt.Fprintf(w, "POST %s", b)
		} else {
			fmt.Fprint(w, "GET Hello World")
		}

	})

	m.HandleFunc("/error", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		fmt.Fprint(w, "oops")
	})

	m.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.URL.Path)
	})

	c <- true

	http.ListenAndServe(":3000", m)
}

func TestHTTPObject(t *testing.T) {

	//blocking channel
	c := make(chan bool, 1)

	//server to test off of
	go startTestServer(c)

	tests := []struct {
		input    string
		expected interface{}
	}{
		//test get request
		{`
		require "net/http"

		Net::HTTP.get("http://127.0.0.1:3000/index")
		`, "GET Hello World"},
		{`
		require "net/http"

		Net::HTTP.post("http://127.0.0.1:3000/index", "text/plain", "Hi Again")
		`, "POST Hi Again"},
		{`
		require "net/http"

		res = Net::HTTP.head("http://127.0.0.1:3000/index")
		res["Content-Length"]
		`, "15"},
	}

	//block until server is ready
	<-c

	for i, tt := range tests {
		v := initTestVM()
		evaluated := v.testEval(t, tt.input, getFilename())
		checkExpected(t, i, evaluated, tt.expected)
		v.checkCFP(t, i, 0)
		v.checkSP(t, i, 1)
	}
}

func TestHTTPObjectFail(t *testing.T) {
	//blocking channel
	c := make(chan bool, 1)

	//server to test off of
	go startTestServer(c)

	testsFail := []errorTestCase{
		//HTTPErrors for get()
		{`
		require "net/http"

		Net::HTTP.get("http://127.0.0.1:3000/error")
		`, "HTTPError: Non-200 response, 404 Not Found (404)", 4},
		{`
		require "net/http"

		Net::HTTP.get("http://127.0.0.1:3001")
		`, "HTTPError: Could not complete request, Get http://127.0.0.1:3001: dial tcp 127.0.0.1:3001: getsockopt: connection refused", 4},
		//Argument errors for get()
		{`
		require "net/http"

		Net::HTTP.get(42)
		`, "ArgumentError: Expect argument 0 to be string, got: Integer", 4},
		{`
		require "net/http"

		Net::HTTP.get("http://127.0.0.1:3000/error", 40, 2)
		`, "ArgumentError: Splat arguments must be a string, got: Integer for argument 0", 4},
		//HTTPErrors for post()
		{`
		require "net/http"

		Net::HTTP.post("http://127.0.0.1:3000/error", "text/plain", "Let me down")
		`, "HTTPError: Non-200 response, 404 Not Found (404)", 4},
		{`
		require "net/http"

		Net::HTTP.post("http://127.0.0.1:3001", "text/plain", "Let me down")
		`, "HTTPError: Could not complete request, Post http://127.0.0.1:3001: dial tcp 127.0.0.1:3001: getsockopt: connection refused", 4},
		//Argument errors for post()
		{`
		require "net/http"

		Net::HTTP.post("http://127.0.0.1:3001", "text/plain", "Let me down", "again")
		`, "ArgumentError: Expect 3 arguments. got: 4", 4},
		{`
		require "net/http"

		Net::HTTP.post(42, "text/plain", "Let me down")
		`, "ArgumentError: Expect argument 0 to be string, got: Integer", 4},
		//HTTPErrors for head()
		{`
		require "net/http"

		Net::HTTP.head("http://127.0.0.1:3000/error")
		`, "HTTPError: Non-200 response, 404 Not Found (404)", 4},
		{`
		require "net/http"

		Net::HTTP.head("http://127.0.0.1:3001")
		`, "HTTPError: Could not complete request, Head http://127.0.0.1:3001: dial tcp 127.0.0.1:3001: getsockopt: connection refused", 4},
		//Argument errors for head()
		{`
		require "net/http"

		Net::HTTP.head(42)
		`, "ArgumentError: Expect argument 0 to be string, got: Integer", 4},
		{`
		require "net/http"

		Net::HTTP.head("http://127.0.0.1:3000/error", 40, 2)
		`, "ArgumentError: Splat arguments must be a string, got: Integer for argument 0", 4},
	}

	//block until server is ready
	<-c

	for i, tt := range testsFail {
		v := initTestVM()
		evaluated := v.testEval(t, tt.input, getFilename())
		checkError(t, i, evaluated, tt.expected, getFilename(), tt.errorLine)
		v.checkCFP(t, i, 1)
		v.checkSP(t, i, 1)
	}
}
