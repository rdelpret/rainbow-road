package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
)

// Get server env var and do some validation
func getServerURL() (string, error) {

	val, ok := os.LookupEnv("RAINBOW_ROAD_SERVER")
	if !ok {
		return val, errors.New("RAINBOW_ROAD_SERVER environment variable not set")
	}
	if !validateServerName(val) {
		return val, errors.New("RAINBOW_ROAD_SERVER environment variable invalid: " + val)

	}
	return val, nil
}

// validate repo names
func validateRepo(repo string) bool {
	match, _ := regexp.MatchString("(.*)/(.*)", repo)
	return match
}

// validate server name (just make sure it has http or https :/ for now)
func validateServerName(server string) bool {
	match, _ := regexp.MatchString("https?:(.*)", server)
	return match
}

// loop through and validate all the repos recieved by client input
func validateRepos(repos []string) error {
	var invalidRepos string = ""
	for _, repo := range repos {
		if !validateRepo(repo) {
			invalidRepos += "Error: Invalid repo name " + repo + "\n"
		}
	}
	if invalidRepos != "" {
		invalidRepos = strings.TrimSuffix(invalidRepos, "\n")
		return errors.New(invalidRepos)
	}
	return nil
}

// formulate the request body for the POST
func createRequestBody(repos []string) []byte {
	body := make(map[string][]map[string]string)
	body["repos"] = make([]map[string]string, 0)
	for _, repo := range repos {
		repoMap := make(map[string]string)
		repoMap["name"] = repo
		body["repos"] = append(body["repos"], repoMap)
	}
	bodyByte, err := json.Marshal(body)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return bodyByte
}

// make the request to the /stars api
func callServer(repos []string, url string) map[string]interface{} {
	body := createRequestBody(repos)
	resp, err := http.Post(url+"/stars", "application/json", bytes.NewBuffer(body))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	defer resp.Body.Close()

	// would have used a struct but wanted to demonstrate use of maps and type assertions
	var res map[string]interface{}

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if resp.StatusCode == http.StatusOK {
		json.NewDecoder(resp.Body).Decode(&res)
	}

	return res

}

var serverURLOverride string = ""

func run(repos []string) (string, error) {

	// get and validate server url
	url, err := getServerURL()

	// this is for testing so we can setup a mock http server
	if serverURLOverride != "" {
		url = serverURLOverride
	}

	if err != nil {
		return "", err
	}

	// if there is no repos return a help string
	if len(repos) == 0 {
		str := "Usage: stars <git-repo-1> <git-repo-2> ...\n"
		return str, nil
	}

	// make sure they are valid, don't make a request if they are not.
	// I think forcing the user to fix is a better UX for errors,
	// especially with Larger requests
	err = validateRepos(repos)

	if err != nil {
		return "", err
	}

	// call the stars api
	res := callServer(repos, url)

	str := "REPO                                              STARS\n"

	for _, repo := range res["repos"].([]interface{}) {

		name := fmt.Sprint(repo.(map[string]interface{})["name"])
		stars := fmt.Sprint(repo.(map[string]interface{})["Stars"])
		resErr := fmt.Sprint(repo.(map[string]interface{})["Error"])

		// print the error under stars column if one exists
		if stars == "-1" {
			stars = resErr
		}
		str += fmt.Sprintf("%-50s%s\n", name, stars)

	}

	str = strings.TrimSuffix(str, "\n")

	return str, nil
}

func main() {

	// get all the repos
	repos := os.Args[1:]

	// run command and handle errors
	str, err := run(repos)

	if err != nil {
		// print errors
		fmt.Println(err)
	}
	// print output
	fmt.Println(str)

}
