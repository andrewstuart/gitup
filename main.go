package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"sync"
)

func main() {
	files, err := ioutil.ReadDir(".")
	if err != nil {
		log.Fatal(err)
	}

	var wg sync.WaitGroup

	for _, file := range files {
		if file.IsDir() {
			wg.Add(1)
		}
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
					fmt.Fprintln(os.Stderr, err)
					fmt.Fprintf(os.Stderr, "error with dir %q: %s\n", file.Name(), string(out))
				}
				fmt.Printf("Done with dir %q\n", file.Name())
			}
		}(file)
	}

	wg.Wait()
}
