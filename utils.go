package main

import (
	"os"
	"os/user"
	"strings"
	"sync"

	"github.com/gdamore/tcell/v2"
	ansi "github.com/leaanthony/go-ansi-parser"
)

var (
	homeDir  string
	homeOnce sync.Once

	defaultParseOptions []ansi.ParseOption
)

func init() {
	defaultParseOptions = append(defaultParseOptions,
		ansi.WithDefaultBackgroundColor("black"),
	)
	defaultParseOptions = append(defaultParseOptions,
		ansi.WithDefaultForegroundColor("white"),
	)
}

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

func fillUpLine(printer Printer, startx, y, width int) {
	// fillup screen
	if startx < width {
		printer.Print(startx, y, tcell.StyleDefault, strings.Repeat(" ", width-startx))
	}
}

func styledcell(as *ansi.StyledText) (ts tcell.Style) {
	ts = tcell.StyleDefault
	if as == nil {
		return
	}

	if as.BgCol != nil {
		name := strings.TrimRight(strings.ToLower(as.BgCol.Name), "123456789")
		if c, ok := tcell.ColorNames[name]; ok {
			ts = ts.Background(c)
		} else {
			ts = ts.Background(tcell.NewRGBColor(
				int32(as.BgCol.Rgb.R), int32(as.BgCol.Rgb.G), int32(as.BgCol.Rgb.B),
			))
		}
	}

	if as.FgCol != nil {
		name := strings.TrimRight(strings.ToLower(as.FgCol.Name), "123456789")
		if c, ok := tcell.ColorNames[name]; ok {
			ts = ts.Foreground(c)
		} else {
			ts = ts.Foreground(tcell.NewRGBColor(
				int32(as.FgCol.Rgb.R), int32(as.FgCol.Rgb.G), int32(as.FgCol.Rgb.B),
			))
		}
	}

	return ts.
		Italic(as.Italic()).
		Bold(as.Bold()).
		Underline(as.Underlined()).
		Reverse(as.Inversed()).
		Blink(as.Blinking()).
		StrikeThrough(as.Strikethrough())
}
