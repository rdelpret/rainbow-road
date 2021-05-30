package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func mockGithubAPI(output string, code int) func() {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(code)
		fmt.Fprintln(w, output)
	}))

	serverURLOverride = ts.URL

	return func() {
		serverURLOverride = ""
	}
}

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

	// Get original token so we can put it back after
	originalToken, _ := os.LookupEnv("GITHUB_TOKEN")

	// set token to test value
	err := os.Setenv("GITHUB_TOKEN", "test")
	assert.Nil(t, err)

	// test that we can get the token
	token, err := getAuth()
	assert.Nil(t, err)
	assert.Equal(t, "test", token)

	// unset the test token
	err = os.Unsetenv("GITHUB_TOKEN")
	assert.Nil(t, err)

	// assert we get the correct error message
	_, err = getAuth()

	assert.EqualError(t, err, "WARNING: GITHUB_TOKEN environment variable not set. API requests to github will be rate limited")

	// put it back the way it was
	err = os.Setenv("GITHUB_TOKEN", originalToken)
	assert.Nil(t, err)

}

func TestGetStars(t *testing.T) {
	// Test error handling for our calls to github.

	close := mockGithubAPI(`{"stargazers_count" : 1}`, 200)
	defer close()

	var r Repo
	r.Name = "rdelpret/cartographer"
	stars, err := GetStars(r)
	assert.Nil(t, err)
	assert.Equal(t, 1, stars)

	r.Name = "invalid"
	_, err = GetStars(r)
	assert.EqualError(t, err, "Recieved invalid repo name: invalid")

	mockGithubAPI(`{}`, 404)

	r.Name = "lkajef023093i2sdfaj09cff/9dieadf09ejd92d23"
	_, err = GetStars(r)
	assert.EqualError(t, err, "Repo Not found: lkajef023093i2sdfaj09cff/9dieadf09ejd92d23")

	serverURLOverride = ""

}

func TestGetStarsForRepos(t *testing.T) {

	// test getting multiple repos

	close := mockGithubAPI(`{"stargazers_count" : 1}`, 200)
	defer close()

	repos := Repos{Repos: []Repo{
		{Name: "rdelpret/kfx"},
		{Name: "rdelpret/cartographer-infra-test-repo"}}}

	repos = GetStarsForRepos(repos)

	expected := Repos{Repos: []Repo{
		{Name: "rdelpret/kfx", Stars: 1, Error: "<nil>"},
		{Name: "rdelpret/cartographer-infra-test-repo", Stars: 1, Error: "<nil>"}}}

	assert.Equal(t, expected, repos)

	mockGithubAPI(`{}`, 404)
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

	close := mockGithubAPI(`{"stargazers_count" : 1}`, 200)
	defer close()

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
	expected := "\"{\\\"status\\\":\\\"green\\\"}\"\n"
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
