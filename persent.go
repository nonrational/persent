package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"
	s "strings"

	"github.com/cdipaolo/sentiment"
	// "github.com/codegangsta/cli"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

// SentComm is a comment with sentiment
type SentComm struct {
	author string
	score  uint8
	text   string
}

// NewSentComm makes a new SentComm
func NewSentComm(ghprc github.PullRequestComment, model sentiment.Models) *SentComm {
	analysis := model.SentimentAnalysis(*ghprc.Body, sentiment.English)

	return &SentComm{author: *ghprc.User.Login, score: analysis.Score, text: *ghprc.Body}
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	orgName, repoName := parseArgs(os.Args[1:])

	commentsByAuthor := make(map[string][]SentComm)
	for _, c := range analyze(orgName, repoName) {
		commentsByAuthor[c.author] = append(commentsByAuthor[c.author], c)
	}

	var authors []string
	for k := range commentsByAuthor {
		authors = append(authors, k)
	}
	sort.Strings(authors)

	for _, k := range authors {
		positive := uint32(0)
		for _, c := range commentsByAuthor[k] {
			positive = positive + uint32(c.score)
		}
		fmt.Printf("%s: %.2f%% (%d/%d)\n", k, ((float32(positive) / float32(len(commentsByAuthor[k]))) * 100), positive, len(commentsByAuthor[k]))
	}

	// fmt.Printf("%# v\n", pretty.Formatter(commentsByAuthor))
}

func parseArgs(argv []string) (orgName, repoName string) {
	if len(argv) == 1 && s.Contains(argv[0], "/") {
		orgAndRepo := s.Split(argv[0], "/")
		orgName = orgAndRepo[0]
		repoName = orgAndRepo[1]
	} else if len(argv) == 2 {
		orgName = argv[0]
		repoName = argv[1]
	} else {
		log.Fatal("usage: persent owner repo")
		os.Exit(1)
	}
	return
}

func analyze(orgName string, repoName string) (sentComms []SentComm) {
	ghTokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: os.Getenv("GITHUB_API_TOKEN")})
	client := github.NewClient(oauth2.NewClient(oauth2.NoContext, ghTokenSource))

	nextPage := 1
	var comments []github.PullRequestComment

	fileName := fmt.Sprintf("%s.%s.json", orgName, repoName)

	if _, err := os.Stat(fileName); err == nil {
		log.Println("Local")
		comments = readFromFile(fileName)
	} else {
		log.Println("GitHub")
		for nextPage > 0 {
			comms, resp, err := client.PullRequests.ListComments(orgName, repoName, 0,
				&github.PullRequestListCommentsOptions{ListOptions: github.ListOptions{PerPage: 100, Page: nextPage}},
			)

			check(err)

			log.Printf("%v (%v)\n", len(comms), resp.NextPage-1)

			nextPage = resp.NextPage
			comments = append(comments, comms...)
		}
		writeToFile(comments, fileName)
	}

	model, _ := sentiment.Restore()

	for _, v := range comments {
		sentComms = append(sentComms, *NewSentComm(v, model))
	}

	log.Printf("Total: %v\n", len(comments))

	return
}

func writeToFile(comments []github.PullRequestComment, fileName string) string {
	json, err := json.Marshal(comments)
	check(err)

	f, err := os.Create(fileName)
	check(err)
	defer f.Close()

	_, err = f.Write(json)
	check(err)

	return (f.Name())
}

func readFromFile(fileName string) []github.PullRequestComment {
	b, err := ioutil.ReadFile(fileName)
	check(err)

	var comments []github.PullRequestComment
	json.Unmarshal(b, &comments)

	return comments
}
