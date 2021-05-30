package main

import (
	"fmt"
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

func TestGetServerUrl(t *testing.T) {

	// Get original server so we can put it back after
	originalURL, _ := os.LookupEnv("RAINBOW_ROAD_SERVER")

	// set server to test value
	err := os.Setenv("RAINBOW_ROAD_SERVER", "http://localhost:9999")
	assert.Nil(t, err)

	// test that we can get the server
	server, err := getServerURL()
	assert.Nil(t, err)
	assert.Equal(t, "http://localhost:9999", server)

	// unset the test server
	err = os.Unsetenv("RAINBOW_ROAD_SERVER")
	assert.Nil(t, err)

	// assert we get the correct error message
	_, err = getServerURL()

	assert.EqualError(t, err, "RAINBOW_ROAD_SERVER environment variable not set")

	// set server to invalid server name
	err = os.Setenv("RAINBOW_ROAD_SERVER", "localhost:9999")
	assert.Nil(t, err)

	// assert we get the correct error message
	val, err := getServerURL()

	assert.EqualError(t, err, "RAINBOW_ROAD_SERVER environment variable invalid: "+val)

	// put it back the way it was
	err = os.Setenv("RAINBOW_ROAD_SERVER", originalURL)
	assert.Nil(t, err)

}

func TestValidateRepoName(t *testing.T) {
	assert.False(t, validateRepo("foo"))
	assert.True(t, validateRepo("fizz/buzz"))
}

func TestValidateServerName(t *testing.T) {
	assert.False(t, validateServerName("localhost:9999"))
	assert.True(t, validateServerName("http://localhost:9999"))
}

func TestValidateRepos(t *testing.T) {
	repos := []string{"test", "foo", "baz"}
	err := validateRepos(repos)
	assert.EqualError(t, err, "Error: Invalid repo name test\nError: Invalid repo name foo\nError: Invalid repo name baz")

	repos = []string{"kubernetes/kubernetes", "foo", "baz"}
	err = validateRepos(repos)
	assert.EqualError(t, err, "Error: Invalid repo name foo\nError: Invalid repo name baz")

	repos = []string{"kubernetes/kubernetes", "istio/istio"}
	err = validateRepos(repos)
	assert.Nil(t, err)
}

func TestCreateRequestBody(t *testing.T) {
	repos := []string{"kubernetes/kubernetes", "istio/istio"}
	body := string(createRequestBody(repos))
	expected := "{\"repos\":[{\"name\":\"kubernetes/kubernetes\"},{\"name\":\"istio/istio\"}]}"
	assert.Equal(t, expected, body)
}

func TestCallServer(t *testing.T) {
	// test callServer happy path
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `{"repos":[{"name":"kubernetes/kubernetes","Stars":77634,"Error":"\u003cnil\u003e"}]}`)
	}))
	defer ts.Close()

	repos := []string{"kubernetes/kubernetes", "istio/istio"}

	res := callServer(repos, ts.URL)
	expected := map[string]interface{}(map[string]interface{}{"repos": []interface{}{map[string]interface{}{"Error": "<nil>", "Stars": float64(77634), "name": "kubernetes/kubernetes"}}})
	assert.Equal(t, expected, res)

}

func TestRun(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `{"repos":[{"name":"kubernetes/kubernetes","Stars":77634,"Error":"\u003cnil\u003e"}]}`)
	}))
	defer ts.Close()
	serverURLOverride = ts.URL
	repos := []string{"kubernetes/kubernetes"}
	str, err := run(repos)
	assert.Nil(t, err)
	expected := "REPO                                              STARS\nkubernetes/kubernetes                             77634"
	assert.Equal(t, expected, str)

	repos = []string{"kuberneteskubernetes", "istio/istio"}
	_, err = run(repos)
	assert.NotNil(t, err)
	assert.EqualError(t, err, "Error: Invalid repo name kuberneteskubernetes")

	repos = []string{}
	str, err = run(repos)
	assert.Nil(t, err)
	assert.Equal(t, "Usage: stars <git-repo-1> <git-repo-2> ...\n", str)

}
