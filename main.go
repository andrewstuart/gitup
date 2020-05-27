package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sync"

	"github.com/sirupsen/logrus"
)

func main() {
	files, err := filepath.Glob("**/*.git/")
	if err != nil {
		log.Fatal(err)
	}

	var wg sync.WaitGroup

	for _, file := range files {
		file, _ := os.Stat(file)
		go func(file os.FileInfo) {
			defer wg.Done()
			if file.IsDir() {
				if fi, err := os.Stat(path.Join(file.Name(), ".git")); err != nil || !fi.IsDir() {
					return
				}

				c := exec.Command("git", "pull")
				c.Dir = file.Name()
				out, err := c.CombinedOutput()
				if err != nil {
					logrus.WithFields(logrus.Fields{
						"directory": file.Name(),
						"output":    string(out),
					}).Error("error with directory")
					fmt.Fprintln(os.Stderr, err)
					fmt.Fprintf(os.Stderr, "error with dir %q: %s\n", file.Name(), string(out))
				}
				fmt.Printf("Done with dir %q\n", file.Name())
			}
		}(file)
	}

	wg.Wait()
}
