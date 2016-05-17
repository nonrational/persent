package main

import (
	"fmt"
	"os"
	"os/exec"
	"testing"
	// "github.com/nonrational/persent"
)

func TestTopCommenters_shortArray(t *testing.T) {
	var smallDummyScores []SentScore
	for i := 0; i < 2; i++ {
		smallDummyScores = append(smallDummyScores, *NewSentScore(fmt.Sprintf("me%d", i), 50, nil))
	}

	tc := topCommenters(smallDummyScores)

	if len(tc) != 2 {
		t.Fatalf("Expected %d, got %d", 2, len(tc))
	}
}

func TestPrintScores_largeArray(t *testing.T) {
	var largeDummyScores []SentScore
	for i := 0; i < 20; i++ {
		largeDummyScores = append(largeDummyScores, *NewSentScore(fmt.Sprintf("me%d", i), 50, nil))
	}

	tc := topCommenters(largeDummyScores)

	if len(tc) != 10 {
		t.Fatalf("Expected %d, got %d", 10, len(tc))
	}
}

func TestArgParsing_twoArguments(t *testing.T) {
	orgName, repoName := parseArgs([]string{"nonrational", "mySuperRepo"})
	if orgName != "nonrational" {
		t.Fatalf("Expected %s, got %s", "nonrational", orgName)
	}
	if repoName != "mySuperRepo" {
		t.Fatalf("Expected %s, got %s", "mySuperRepo", repoName)
	}
}

func TestArgParsing_singleArgument(t *testing.T) {
	orgName, repoName := parseArgs([]string{"nonrational/mySuperRepo"})
	if orgName != "nonrational" {
		t.Fatalf("Expected %s, got %s", "nonrational", orgName)
	}
	if repoName != "mySuperRepo" {
		t.Fatalf("Expected %s, got %s", "mySuperRepo", repoName)
	}
}

func TestArgParsing_singleDotArgument(t *testing.T) {
	orgName, repoName := parseArgs([]string{"nonrational.mySuperRepo"})
	if orgName != "nonrational" {
		t.Fatalf("Expected %s, got %s", "nonrational", orgName)
	}
	if repoName != "mySuperRepo" {
		t.Fatalf("Expected %s, got %s", "mySuperRepo", repoName)
	}
}

func TestArgParsing_tooManyArguments(t *testing.T) {
	if os.Getenv("BE_CRASHER") == "1" {
		parseArgs([]string{"too", "many", "arguments"})
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestArgParsing_tooManyArguments")
	cmd.Env = append(os.Environ(), "BE_CRASHER=1")
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	}
	t.Fatalf("process ran with err %v, want exit status 1", err)
}
