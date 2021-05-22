package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Disable log output durring tests (nice for negetive tests)
func TestMain(m *testing.M) {
	log.SetOutput(ioutil.Discard)
	os.Exit(m.Run())
}
func TestAssembleURL(t *testing.T) {
	// Test Valid Repo
	url, err := assembleURL("rdelprete/cartographer")
	assert.Nil(t, err)
	assert.Equal(t, "https://api.github.com/repos/rdelprete/cartographer", url)

	// Test Invalid Repo
	_, err = assembleURL("invalidrepo")
	assert.NotNil(t, err)
}

func TestGetAuth(t *testing.T) {

	//Set Env Var for test
	originalToken, _ := os.LookupEnv("GIT_TOKEN")

	err := os.Setenv("GIT_TOKEN", "test")
	assert.Nil(t, err)

	token, err := getAuth()
	assert.Nil(t, err)
	assert.Equal(t, "test", token)

	err = os.Unsetenv("GIT_TOKEN")
	assert.Nil(t, err)

	_, err = getAuth()

	assert.EqualError(t, err, "WARNING: GIT_TOKEN environment variable not set. API requests to github will be rate limited")

	err = os.Setenv("GIT_TOKEN", originalToken)
	assert.Nil(t, err)

}

func TestGetStars(t *testing.T) {
	var r Repo
	r.Name = "rdelpret/cartographer"
	stars, err := GetStars(r)
	assert.Nil(t, err)
	assert.Equal(t, 1, stars)

	r.Name = "invalid"
	_, err = GetStars(r)
	assert.EqualError(t, err, "Recieved invalid repo name: invalid")

	r.Name = "lkajef023093i2sdfaj09cff/9dieadf09ejd92d23"
	_, err = GetStars(r)
	assert.EqualError(t, err, "Repo Not found: lkajef023093i2sdfaj09cff/9dieadf09ejd92d23")

}

func TestGetStarsForRepos(t *testing.T) {

	repos := Repos{Repos: []Repo{
		{Name: "rdelpret/kfx"},
		{Name: "rdelpret/cartographer-infra-test-repo"}}}

	repos = GetStarsForRepos(repos)

	expected := Repos{Repos: []Repo{
		{Name: "rdelpret/kfx", Stars: 0, Error: "<nil>"},
		{Name: "rdelpret/cartographer-infra-test-repo", Stars: 1, Error: "<nil>"}}}

	assert.Equal(t, expected, repos)

	repos = Repos{Repos: []Repo{
		{Name: "invalid"},
		{Name: "lkajef023093i2sdfaj09cff/9dieadf09ejd92d23"}}}

	repos = GetStarsForRepos(repos)

	expected = Repos{Repos: []Repo{
		{Name: "invalid", Stars: -1, Error: "Recieved invalid repo name: invalid"},
		{Name: "lkajef023093i2sdfaj09cff/9dieadf09ejd92d23", Stars: -1, Error: "Repo Not found: lkajef023093i2sdfaj09cff/9dieadf09ejd92d23"}}}

	assert.Equal(t, expected, repos)

	repos = Repos{Repos: []Repo{
		{},
		{}}}

	repos = GetStarsForRepos(repos)

	expected = Repos{Repos: []Repo{
		{Name: "", Stars: -1, Error: "Recieved invalid repo name: "},
		{Name: "", Stars: -1, Error: "Recieved invalid repo name: "}}}

	assert.Equal(t, expected, repos)

}

func TestStarsHandler(t *testing.T) {

	// test happy path
	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(starsHandler)

	json := []byte(`{"repos": [{"name": "rdelpret/cartographer"}]}`)

	req, err := http.NewRequest("POST", "/stars", bytes.NewBuffer(json))
	req.Header.Set("Content-Type", "application/json")

	if err != nil {
		t.Fatal(err)
	}

	handler.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	expected := "{\"repos\":[{\"name\":\"rdelpret/cartographer\",\"Stars\":1,\"Error\":\"\\u003cnil\\u003e\"}]}\n"
	assert.Equal(t, expected, recorder.Body.String())

	// test malformed json

	recorder = httptest.NewRecorder()
	handler = http.HandlerFunc(starsHandler)

	json = []byte(`{"repos": [{"name" "rdelpret/cartographer"}]}`)

	req, err = http.NewRequest("POST", "/stars", bytes.NewBuffer(json))
	req.Header.Set("Content-Type", "application/json")

	if err != nil {
		t.Fatal(err)
	}

	handler.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	assert.Equal(t, "Malformed Request.\n", recorder.Body.String())

	// Test using wrong HTTP Method

	recorder = httptest.NewRecorder()
	handler = http.HandlerFunc(starsHandler)

	req, err = http.NewRequest("GET", "/stars", nil)

	if err != nil {
		t.Fatal(err)
	}

	handler.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusNotFound, recorder.Code)

	assert.Equal(t, "Method is not supported.\n", recorder.Body.String())

	// Test bad route with this handler

	recorder = httptest.NewRecorder()
	handler = http.HandlerFunc(starsHandler)

	json = []byte(`{"repos": [{"name" "rdelpret/cartographer"}]}`)

	req, err = http.NewRequest("POST", "/wrongroute", bytes.NewBuffer(json))
	req.Header.Set("Content-Type", "application/json")

	if err != nil {
		t.Fatal(err)
	}

	handler.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusNotFound, recorder.Code)

	assert.Equal(t, "404 not found.\n", recorder.Body.String())

}

func TestHealthCheckHandler(t *testing.T) {
	// test happy path
	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(healthCheckHandler)

	req, err := http.NewRequest("GET", "/health", nil)

	if err != nil {
		t.Fatal(err)
	}

	handler.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	expected := "\"{\\\"status\\\" : \\\"green\\\"}\"\n"
	assert.Equal(t, expected, recorder.Body.String())

	// Test using wrong HTTP Method

	recorder = httptest.NewRecorder()
	handler = http.HandlerFunc(healthCheckHandler)

	req, err = http.NewRequest("POST", "/health", nil)

	if err != nil {
		t.Fatal(err)
	}

	handler.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusNotFound, recorder.Code)

	assert.Equal(t, "Method is not supported.\n", recorder.Body.String())

	// Test bad route with this handler

	recorder = httptest.NewRecorder()
	handler = http.HandlerFunc(healthCheckHandler)

	req, err = http.NewRequest("GET", "/wrongroute", nil)
	req.Header.Set("Content-Type", "application/json")

	if err != nil {
		t.Fatal(err)
	}

	handler.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusNotFound, recorder.Code)

	assert.Equal(t, "404 not found.\n", recorder.Body.String())

}
