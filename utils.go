package main

import (
	"os"
	"os/user"
	"strings"
	"sync"

	ansi "github.com/leaanthony/go-ansi-parser"
)

var (
	homeDir  string
	homeOnce sync.Once

	defaultParseOptions []ansi.ParseOption
)

func getHomedirectory() string {
	homeOnce.Do(func() {
		u, err := user.Current()
		if err != nil {
			panic(err)
		}

		homeDir = u.HomeDir
	})

	return homeDir
}

func expandPath(path string) string {
	home := getHomedirectory()
	path = os.ExpandEnv(path)
	if strings.HasPrefix(path, "~") {
		return strings.Replace(path, "~", home, 1)
	}

	return path
}
