package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/Sirupsen/logrus"
)

const (
	GIT_ADD    = 0
	GIT_EDIT   = 1
	GIT_DELETE = 2
)

func gitClone(url string) {
	files, err := ioutil.ReadDir(vaultDir)
	if err != nil {
		logrus.Fatalf("%s does not exist", vaultDir)
	}
	if len(files) > 0 {
		logrus.Fatalf("%s directory already exists and contains files", vaultDir)
	}
	runGitCommand(false, "clone", url, vaultDir)
}

func gitInit() {
	runGitCommand(false, "init")
}

func gitRemote(url string) {
	runGitCommand(true, "remote", "rm", "origin")
	runGitCommand(false, "remote", "add", "-f", "origin", url)
}

func gitCommit(file string, op int) {
	message := ""
	switch op {
	case GIT_ADD:
		message = fmt.Sprintf("Added secret '%s'", file)
	case GIT_EDIT:
		message = fmt.Sprintf("Edited secret '%s'", file)
	case GIT_DELETE:
		message = fmt.Sprintf("Deleted secret '%s'", file)
	}

	runGitCommand(true, "add", file)
	runGitCommand(true, "commit", "-m", message)
}

func gitPush() {
	runGitCommand(false, "add", "-A")
	runGitCommand(false, "commit", "-m", "Vault store update.")
	runGitCommand(false, "push", "-u", "origin", "master")
}

func gitPull() {
	runGitCommand(false, "pull", "origin", "master")
}

func runGitCommand(suppress bool, args ...string) error {
	err := os.Chdir(vaultDir)
	if err != nil {
		logrus.Fatalf("could not run git command: %s", err)
	}

	cmd := exec.Command("git", args...)
	if !suppress {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	err = cmd.Run()

	return err
}
