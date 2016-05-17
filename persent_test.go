package main

import (
	"os"
	"os/exec"
	"testing"
)

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
