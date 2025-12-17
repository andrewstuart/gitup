package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
)

// isDirty checks if the git repository at the given directory has uncommitted changes.
func isDirty(dir string) bool {
	c := exec.Command("git", "status", "--porcelain")
	c.Dir = dir
	out, err := c.Output()
	if err != nil {
		logrus.WithField("dir", dir).WithError(err).Error("failed to check git status")
		return false
	}
	return len(strings.TrimSpace(string(out))) > 0
}

// getCommitMsgAndPush prompts the user for a commit message, commits all changes, and pushes.
func getCommitMsgAndPush(dir string) error {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("[%s] Enter commit message (or press Enter to skip): ", dir)
	msg, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	msg = strings.TrimSpace(msg)
	if msg == "" {
		logrus.WithField("dir", dir).Info("skipping commit (no message provided)")
		return nil
	}

	// Commit
	commit := exec.Command("git", "commit", "-a", "-m", msg)
	commit.Dir = dir
	if out, err := commit.CombinedOutput(); err != nil {
		return fmt.Errorf("git commit failed: %s: %w", string(out), err)
	}

	// Push
	push := exec.Command("git", "push")
	push.Dir = dir
	if out, err := push.CombinedOutput(); err != nil {
		return fmt.Errorf("git push failed: %s: %w", string(out), err)
	}

	logrus.WithField("dir", dir).Info("committed and pushed")
	return nil
}

func main() {
	// Collect all git directories
	var gitDirs []string
	err := filepath.Walk(".", filepath.WalkFunc(func(p string, file os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if path.Base(p) == ".git" && file.IsDir() {
			gitDirs = append(gitDirs, path.Dir(p))
		}
		return nil
	}))

	if err != nil {
		logrus.WithError(err).Error("could not walk filesystem")
		return
	}

	// Handle dirty repos first (requires user input, must be serial)
	for _, dir := range gitDirs {
		if isDirty(dir) {
			logrus.WithField("dir", dir).Warn("repository has uncommitted changes")
			if err := getCommitMsgAndPush(dir); err != nil {
				logrus.WithField("dir", dir).WithError(err).Error("failed to commit and push")
			}
		}
	}

	// Now do concurrent pulls
	var wg sync.WaitGroup
	for _, dir := range gitDirs {
		wg.Add(1)
		go func(dir string) {
			defer wg.Done()
			c := exec.Command("git", "pull")
			c.Dir = dir
			out, err := c.CombinedOutput()
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"directory": dir,
					"output":    string(out),
				}).WithError(err).Error("error with directory")
			}
			logrus.WithField("dir", dir).Info("done")
		}(dir)
	}

	wg.Wait()
}
