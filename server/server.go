package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"sync"
)

// Struct that represents a /stars request
type Repos struct {
	Repos []Repo `json:"repos"`
}

// Struct that represents a repo, used in both request and response
type Repo struct {
	Name  string `json:"name"`
	Stars int
	Error string
}

// helper function to pull a git token if it exists. Print warning if it doesnt
func getAuth() (string, error) {
	val, ok := os.LookupEnv("GIT_TOKEN")
	if !ok {
		return val, errors.New("WARNING: GIT_TOKEN environment variable not set. API requests to github will be rate limited")
	}
	return val, nil
}

// helper function to validate repo name and return api url to make request
func assembleURL(repoName string) (string, error) {
	base := "https://api.github.com/"
	api := "repos/"

	match, _ := regexp.MatchString("(.*)/(.*)", repoName)

	if !match {
		return base + api, errors.New("Recieved invalid repo name: " + repoName)
	}

	return base + api + repoName, nil
}

// Function to call github api and get star count
// return error and -1 for bad requests
func GetStars(repo Repo) (int, error) {

	client := &http.Client{}

	// validate repo name and generate github api url
	url, err := assembleURL(repo.Name)

	if err != nil {
		log.Println(err)
		return -1, err
	}

	// setup request
	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		log.Println(err)
		return -1, err
	}

	// get github token from env if we can
	token, err := getAuth()

	if err != nil {
		log.Println(err)
	}

	// use github token if we have it
	if token != "" {
		req.Header.Set("Authorization", "token "+token)
	}

	resp, err := client.Do(req)

	if err != nil {
		log.Println(err)
		return -1, err
	}

	// close body stream when we are done with it
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)

		if err != nil {
			log.Println(err)
		}

		// load into interface, pull out stargazers and convert to int
		var obj map[string]interface{}
		json.Unmarshal([]byte(body), &obj)
		stars := int(obj["stargazers_count"].(float64))
		return stars, nil
	}

	return -1, errors.New("Repo Not found: " + repo.Name)
}

// Handle Bulk requests for stars concurrently
func GetStarsForRepos(repos Repos) Repos {

	wg := sync.WaitGroup{}
	for i, repo := range repos.Repos {
		r := &repos.Repos[i]
		wg.Add(1)
		go func(repo Repo) {
			var err error
			r.Stars, err = GetStars(repo)

			// Convert the error to a string so we can
			// json encode and pass to the client
			// not every request will have an error so
			// we don't want to block good data
			r.Error = fmt.Sprint(err)

			wg.Done()
		}(repo)
	}
	wg.Wait()
	// return repos object with stars and errors after wg has finished
	return repos
}

// HTTP route to handle stars requests
func starsHandler(w http.ResponseWriter, r *http.Request) {

	// Ensure this handler can only be called from /stars route
	if r.URL.Path != "/stars" {
		http.Error(w, "404 not found.", http.StatusNotFound)
		return
	}

	// Ensure POST is the only method used
	if r.Method != "POST" {
		http.Error(w, "Method is not supported.", http.StatusNotFound)
		return
	}

	// Read body
	body, err := ioutil.ReadAll(r.Body)

	// not sure how to get code coverage here
	if err != nil {
		http.Error(w, "Malformed Request.", http.StatusBadRequest)
		log.Println(err)
		return

	}

	// unmarshal request into repos obj
	var repos Repos

	err = json.Unmarshal([]byte(body), &repos)

	if err != nil {
		http.Error(w, "Malformed Request.", http.StatusBadRequest)
		log.Println(err)
		return
	}

	// get stars and errors
	repos = GetStarsForRepos(repos)

	w.Header().Set("Content-Type", "application/json")

	// send em on back
	json.NewEncoder(w).Encode(repos)
}

// HTTP route to handle health checks
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {

	// Ensure this handler can only be called from /stars route
	if r.URL.Path != "/health" {
		http.Error(w, "404 not found.", http.StatusNotFound)
		return
	}

	// Ensure POST is the only method used
	if r.Method != "GET" {
		http.Error(w, "Method is not supported.", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(`{"status" : "green"}`)
}

func main() {
	http.HandleFunc("/stars", starsHandler)
	http.HandleFunc("/health", healthCheckHandler)
	fmt.Println(`

Starting Rainbow Road Server!
            .
           ,O,
          ,OOO,
    'oooooOOOOOooooo'
      'OOOOOOOOOOO'
        'OOOOOOO'
        OOOO'OOOO
       OOO'   'OOO
      O'         'O
  
 Listening on port: 9999
 
 `)

	http.ListenAndServe(":9999", nil)
}

//TODO log requests maybe?
