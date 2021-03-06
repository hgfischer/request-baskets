package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequestData_Forward(t *testing.T) {
	basket := "demo"

	// Test request
	data := new(RequestData)
	data.Header = make(http.Header)
	data.Header.Add("Content-Type", "application/json")
	data.Header.Add("User-Agent", "Unit-Test")
	data.Header.Add("Accept", "plain/text")
	data.Method = "POST"
	data.Body = "{ \"name\" : \"test\", \"action\" : \"add\" }"
	data.ContentLength = int64(len(data.Body))
	// path contains basket name
	data.Path = "/" + basket + "/service/actions"
	data.Query = "id=15&max=10"

	// Test HTTP server
	var forwardedData *RequestData
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		forwardedData = ToRequestData(r)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	// Config to forward requests to test HTTP server
	config := BasketConfig{ForwardUrl: ts.URL, ExpandPath: false, Capacity: 20}
	data.Forward(config, basket)

	// Validate forwarded request
	assert.Equal(t, data.Method, forwardedData.Method, "wrong request method")
	// path is not expanded during forward
	assert.Equal(t, "/", forwardedData.Path, "wrong request path")
	assert.Equal(t, data.Query, forwardedData.Query, "wrong request query")
	assert.Equal(t, data.ContentLength, forwardedData.ContentLength, "wrong content length")
	assert.Equal(t, data.Body, forwardedData.Body, "wrong request body")

	// expect all original headers to present in forwarded request (additional headers might be added)
	for k, v := range data.Header {
		fv := forwardedData.Header[k]
		if assert.NotNil(t, fv, "missing expected header: %v = %v", k, v) {
			assert.Equal(t, v, fv, "wrong value of request header: %v", k)
		}
	}
}

func TestRequestData_Forward_ComplexForwardURL(t *testing.T) {
	basket := "zooapi"
	pathSuffix := "/rooms/1/pets/12"

	// Test request
	data := new(RequestData)
	data.Header = make(http.Header)
	data.Header.Add("Content-Type", "application/json")
	data.Header.Add("User-Agent", "Unit-Test")
	data.Method = "PUT"
	data.Body = "{ \"id\" : \"12\", \"kind\" : \"elephant\", \"name\" : \"Bibi\" }"
	data.ContentLength = int64(len(data.Body))
	// path contains basket name
	data.Path = "/" + basket + pathSuffix
	data.Query = "expose=true&pattern=*"

	// Test HTTP server
	var forwardedData *RequestData
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		forwardedData = ToRequestData(r)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	// Config to forward requests to test HTTP server (also enable expanding URL)
	forwardUrl := ts.URL + "/captures?from=" + basket
	config := BasketConfig{ForwardUrl: forwardUrl, ExpandPath: true, Capacity: 20}
	data.Forward(config, basket)

	// Validate forwarded path
	assert.Equal(t, "/captures"+pathSuffix, forwardedData.Path, "wrong request path")
	assert.Equal(t, "from="+basket+"&"+data.Query, forwardedData.Query, "wrong request query")
}

func TestRequestData_Forward_BrokenURL(t *testing.T) {
	basket := "test"

	// Test request
	data := new(RequestData)
	data.Header = make(http.Header)
	data.Header.Add("Content-Type", "application/json")
	data.Method = "GET"
	data.Body = "{ \"name\" : \"test\", \"action\" : \"add\" }"
	data.ContentLength = int64(len(data.Body))
	// path contains basket name
	data.Path = "/" + basket

	// Config to forward requests to broken URL
	config := BasketConfig{ForwardUrl: "-.'", ExpandPath: false, Capacity: 20}

	// Should not fail, warning in log is expected
	data.Forward(config, basket)
}
