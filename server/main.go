package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
)

type Repo struct {
	Name  string
	Stars float64
}

func assembleURL(repoName string) string {
	base := "https://api.github.com/"
	api := "repos/"
	return base + api + repoName
}

func getAuth() string {
	return os.Getenv("GIT_TOKEN")
}

func GetStars(repo Repo) float64 {
	//TODO  unmarshel directly into struct, handle errors, write tests

	client := &http.Client{}

	url := assembleURL(repo.Name)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "token "+getAuth())

	resp, _ := client.Do(req)

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		var obj map[string]interface{}
		json.Unmarshal([]byte(body), &obj)
		stars := obj["stargazers_count"].(float64)
		return stars
	}

	return 0 // handle errors
}

func GetStarsForRepos(repos []Repo) []Repo {
	// Do this concurently
	wg := sync.WaitGroup{}
	for i, repo := range repos {
		r := &repos[i]
		wg.Add(1)
		go func(repo Repo) {
			r.Stars = GetStars(repo)
			wg.Done()
		}(repo)
	}
	wg.Wait()
	return repos
}

func main() {
	// POC
	repos := []Repo{
		{Name: "rdelpret/cartographer"},
		{Name: "kubernetes/kubernetes"},
		{Name: "argoproj/argo-cd"},
		{Name: "hashicorp/terraform"},
		{Name: "hashicorp/terraform-provider-aws"},
		{Name: "ansible/ansible"},
		{Name: "envoyproxy/envoy"},
		{Name: "puppetlabs/puppet"},
		{Name: "jtblin/kube2iam"},
		{Name: "id-Software/DOOM"},
		{Name: "dty1er/kubecolor"},
		{Name: "jenkinsci/jenkins"},
		{Name: "istio/istio"},
		{Name: "nats-io/nats-operator"},
		{Name: "containerd/containerd"},
		{Name: "tektoncd/pipeline"},
	}

	repos = GetStarsForRepos(repos)

	for _, repo := range repos {
		fmt.Println(repo)
	}

}
