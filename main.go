package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	//"time"
)

const EnvName = "TP_DIRECTORY"

func main() {

	//startTime := time.Now()
	root := os.Getenv(EnvName)

	if root == "" {
		fmt.Println("TP_DIRECTORY is not set")
		os.Exit(1)
	}

	stat, err := os.Stat(root)

	if err != nil {
		fmt.Printf("%s does not exist (%s)\n", EnvName, root)
		os.Exit(1)
	}

	if !stat.IsDir() {
		fmt.Printf("%s is not a directory (%s)\n", EnvName, root)
		os.Exit(1)
	}

	err = filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {

		if err != nil {
			return err
		}

		relativePath, err := filepath.Rel(root, path)

		if err != nil {
			return err
		}

		depth := len(strings.Split(relativePath, string(os.PathSeparator)))

		if depth > 3 {
			if info.IsDir() {
				return filepath.SkipDir // skip current dir in walk
			}
			return nil
		}

		if info.Name() == ".tmux" && !info.IsDir() {
			fmt.Println(path)
		}
		return nil
	})

	//fmt.Printf("time taken %s\n", time.Since(startTime))
}
