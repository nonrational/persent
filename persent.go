package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"regexp"
	"strings"

	"github.com/bradfitz/slice"
	"github.com/cdipaolo/sentiment"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

// SentComm is a comment with sentiment
type SentComm struct {
	author string
	score  uint8
	text   string
}

// SentScore is a sentiment score with comments by an author
type SentScore struct {
	percentPositive float32
	totalComments   int
	author          string
	comments        []SentComm
}

// NewSentScore makes a new SentScore
func NewSentScore(author string, percentPositive float32, comments []SentComm) *SentScore {
	return &SentScore{author: author, percentPositive: percentPositive, comments: comments, totalComments: len(comments)}
}

// NewSentComm makes a new SentComm
func NewSentComm(ghprc *github.PullRequestComment, model sentiment.Models) *SentComm {
	analysis := model.SentimentAnalysis(*ghprc.Body, sentiment.English)

	return &SentComm{author: *ghprc.User.Login, score: analysis.Score, text: *ghprc.Body}
}

func main() {
	orgName, repoName := parseArgs(os.Args[1:])

	scores := CalculatePersent(orgName, repoName)

	for _, v := range topCommenters(*scores) {
		fmt.Printf("%s: %.0f%% of %d\n", v.author, v.percentPositive, v.totalComments)
	}
}

func CalculatePersent(orgName string, repoName string) *[]SentScore {
	allComments := fetch(orgName, repoName)
	sentimentComments := analyze(allComments)

	commentsByAuthor := make(map[string][]SentComm)
	for _, c := range sentimentComments {
		commentsByAuthor[c.author] = append(commentsByAuthor[c.author], c)
	}

	var scores []SentScore

	for k := range commentsByAuthor {
		positiveScore := uint32(0)
		for _, c := range commentsByAuthor[k] {
			positiveScore = positiveScore + uint32(c.score)
		}

		totalComments := len(commentsByAuthor[k])
		percentPositive := (float32(positiveScore) / float32(totalComments)) * 100

		scores = append(scores, *NewSentScore(k, percentPositive, commentsByAuthor[k]))
	}

	return &scores
}

func topCommenters(scores []SentScore) (topCommenters []SentScore) {
	slice.Sort(scores[:], func(i, j int) bool {
		return scores[i].totalComments > scores[j].totalComments
	})
	// fmt.Printf("%# v\n", pretty.Formatter(commentsByAuthor))
	maxIdx := int(math.Min(10, float64(len(scores))))
	topCommenters = scores[0:maxIdx]

	return
}

func parseArgs(argv []string) (orgName, repoName string) {
	r, _ := regexp.Compile(`\w+[/.]\w+`)

	if len(argv) == 1 && r.MatchString(argv[0]) {
		dotOrSlash := func(c rune) bool { return c == '/' || c == '.' }
		orgAndRepo := strings.FieldsFunc(argv[0], dotOrSlash)

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

func analyze(comments []*github.PullRequestComment) []SentComm {
	analysisModel, _ := sentiment.Restore()

	var sentComms []SentComm

	for _, v := range comments {
		sentComms = append(sentComms, *NewSentComm(v, analysisModel))
	}

	return sentComms
}

func fetch(orgName string, repoName string) []*github.PullRequestComment {
	ghTokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: os.Getenv("GITHUB_API_TOKEN")})
	client := github.NewClient(oauth2.NewClient(oauth2.NoContext, ghTokenSource))

	nextPage := 1
	var comments []*github.PullRequestComment

	fileName := fmt.Sprintf("%s.%s.json", orgName, repoName)

	if _, err := os.Stat(fileName); err == nil {
		log.Println("Local")
		comments = readFromFile(fileName)
	} else {
		log.Println("GitHub")
		for nextPage > 0 {

			ctx := context.TODO()

			comms, resp, err := client.PullRequests.ListComments(ctx, orgName, repoName, 0,
				&github.PullRequestListCommentsOptions{ListOptions: github.ListOptions{PerPage: 100, Page: nextPage}},
			)

			if err != nil {
				log.Fatalf("%+v", err)
			}

			log.Printf("%v (%v)\n", len(comms), resp.NextPage-1)

			nextPage = resp.NextPage
			comments = append(comments, comms...)
		}
		writeToFile(comments, fileName)
	}

	log.Printf("Total: %v\n", len(comments))

	return comments
}

func writeToFile(comments []*github.PullRequestComment, fileName string) string {
	json, err := json.Marshal(comments)
	if err != nil {
		log.Fatalf("json.Marshall got err:%+v", err)
	}

	f, err := os.Create(fileName)
	if err != nil {
		log.Fatalf("os.Create got err:%+v", err)
	}
	defer f.Close()

	_, err = f.Write(json)
	if err != nil {
		log.Fatalf("f.Write got err:%+v", err)
	}

	return (f.Name())
}

func readFromFile(fileName string) []*github.PullRequestComment {
	b, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Fatalf("ioutil.ReadFile got err:%+v", err)
	}

	var rawComments []github.PullRequestComment
	json.Unmarshal(b, &rawComments)

	comments := make([]*github.PullRequestComment, len(rawComments))
	for i := range rawComments {
		comments[i] = &rawComments[i]
	}

	return comments
}
