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
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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

//TODO - only print this once on startup
// helper function to pull a git token if it exists. Print warning if it doesnt
func getAuth() (string, error) {
	val, ok := os.LookupEnv("GITHUB_TOKEN")
	if !ok {
		return val, errors.New("WARNING: GITHUB_TOKEN environment variable not set. API requests to github will be rate limited")
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

func logger(targetMux http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		targetMux.ServeHTTP(w, r)

		// log request by who(IP address)
		requesterIP := r.RemoteAddr

		log.Printf(
			"%-20s%-20s%-20s%-20v",
			r.Method,
			r.RequestURI,
			requesterIP,
			time.Since(start),
		)
	})
}

// Function to call github api and get star count
// return error and -1 for bad requests
func GetStars(repo Repo) (int, error) {

	githubApiReqAll.Inc()
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

	// use github token if we have it
	if gitToken != "" {
		req.Header.Set("Authorization", "token "+gitToken)
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
		githubApiReq200.Inc()
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

	starsApiReqAll.Inc()

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

	starsApiReq200.Inc()
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

	json.NewEncoder(w).Encode(`{"status":"green"}`)
}

var (
	starsApiReqAll = promauto.NewCounter(prometheus.CounterOpts{
		Name: "api_requests_stars_ALL",
		Help: "The total number of processed requests from the stars api",
	})
)

var (
	starsApiReq200 = promauto.NewCounter(prometheus.CounterOpts{
		Name: "api_requests_stars_200",
		Help: "The total number of 200 requests from the stars api",
	})
)

var (
	githubApiReqAll = promauto.NewCounter(prometheus.CounterOpts{
		Name: "api_requests_github_all",
		Help: "The total number of outgoing requests to github",
	})
)

var (
	githubApiReq200 = promauto.NewCounter(prometheus.CounterOpts{
		Name: "api_requests_github_200",
		Help: "The total number of 200 requests to github",
	})
)

var gitToken, gitTokenErr = getAuth()

func main() {

	if gitTokenErr != nil {
		log.Println(gitTokenErr)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/stars", starsHandler)
	mux.HandleFunc("/health", healthCheckHandler)
	mux.Handle("/metrics", promhttp.Handler())
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

	err := http.ListenAndServe(":9999", logger(mux))

	if err != nil {
		log.Fatalf("Server exited with: %v", err)
	}
}

//TODO:
// comment tests better
// comment server better
// impliment 206, bubble github 400 if all requests are denied
// impliment faster hot relaod
