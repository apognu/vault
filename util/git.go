package util

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

func GitClone(url string) {
	os.MkdirAll(GetVaultPath(), 0700)

	files, err := ioutil.ReadDir(GetVaultPath())
	if err != nil {
		logrus.Fatalf("%s does not exist", GetVaultPath())
	}
	if len(files) > 0 {
		logrus.Fatalf("%s directory already exists and contains files", GetVaultPath())
	}

	logrus.Info("cloning from remote repository")

	RunGitCommand(false, "clone", url, GetVaultPath())
}

func GitInit() {
	RunGitCommand(false, "init")
}

func GitRemote(url string) {
	RunGitCommand(true, "remote", "rm", "origin")
	RunGitCommand(false, "remote", "add", "-f", "origin", url)
}

func GitCommit(file string, op int, message string) {
	if message == "" {
		switch op {
		case GIT_ADD:
			message = fmt.Sprintf("Added secret '%s'", file)
		case GIT_EDIT:
			message = fmt.Sprintf("Edited secret '%s'", file)
		case GIT_DELETE:
			message = fmt.Sprintf("Deleted secret '%s'", file)
		}
	}

	RunGitCommand(true, "add", file)
	RunGitCommand(true, "commit", "-m", message)
}

func GitCommitRename(oldFile, newFile string) {
	RunGitCommand(true, "add", oldFile)
	RunGitCommand(true, "add", newFile)
	RunGitCommand(true, "commit", "-m", fmt.Sprintf("Renamed '%s' to '%s'", oldFile, newFile))
}

func GitPush() {
	logrus.Info("pushing to remote repository")

	RunGitCommand(false, "add", "-A")
	RunGitCommand(false, "commit", "-m", "Vault store update.")
	RunGitCommand(false, "push", "-u", "origin", "master")
}

func GitPull() {
	logrus.Info("pulling from remote repository")

	RunGitCommand(false, "pull", "origin", "master")
}

func RunGitCommand(suppress bool, args ...string) error {
	err := os.Chdir(GetVaultPath())
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
