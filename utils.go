package main

import (
	"os"
	"os/user"
	"strings"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

var (
	homeDir  string
	homeOnce sync.Once
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

func emitStr(s tcell.Screen, x, y int, style tcell.Style, str string) {
	for _, c := range str {
		var comb []rune
		w := runewidth.RuneWidth(c)
		if w == 0 {
			comb = []rune{c}
			c = ' '
			w = 1
		}
		s.SetContent(x, y, c, comb, style)
		x += w
	}
}
