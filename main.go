package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/oklog/run"
	"github.com/peterbourgon/ff/v3"
	"github.com/peterbourgon/ff/v3/ffcli"
	"github.com/peterbourgon/ff/v3/fftoml"
)

type LoonConfig struct {
	RingSize   int
	ConfigFile string
	NoColor    bool
	NoAnsi     bool
	Json       bool
}

// flags
var (
	rootFlagSet = flag.NewFlagSet("loon", flag.ExitOnError)
)

func parseRootConfig(args []string) (*LoonConfig, error) {
	var cfg LoonConfig

	defaultLoonConfig := expandPath("~/.loonrc")

	rootFlagSet.StringVar(&cfg.ConfigFile, "config", defaultLoonConfig, "root config project")
	rootFlagSet.BoolVar(&cfg.Json, "json", false, "parsed is a json line file")
	rootFlagSet.BoolVar(&cfg.NoColor, "nocolor", false, "disable color")
	rootFlagSet.BoolVar(&cfg.NoAnsi, "noansi", false, "do not parse ansi sequence")
	rootFlagSet.IntVar(&cfg.RingSize, "ringsize", 100000, "ring line size")

	err := ff.Parse(rootFlagSet, args,
		ff.WithEnvVarPrefix("LOON"),
		ff.WithConfigFileFlag("config"),
		ff.WithAllowMissingConfigFile(true),
		ff.WithConfigFileParser(fftoml.Parser),
	)

	// expand path
	if err != nil {
		return nil, fmt.Errorf("unable to parse flags: %w", err)
	}

	return &cfg, nil
}

func main() {
	// disable logger
	log.SetOutput(ioutil.Discard)

	args := os.Args[1:]

	lcfg, err := parseRootConfig(args)
	if err != nil {
		panic(err)
	}

	root := &ffcli.Command{
		Name:    "loon [flags] <file>",
		FlagSet: rootFlagSet,
		Exec: func(ctx context.Context, args []string) error {
			var stdin bool
			var path string

			fi, _ := os.Stdin.Stat()
			if (fi.Mode() & os.ModeCharDevice) == 0 {
				stdin = true
				path = os.Stdin.Name()
			} else if len(args) == 1 {
				path = args[0]
			} else {
				return flag.ErrHelp
			}

			file := File{
				Stdin: stdin,
				Path:  path,
			}

			ring, err := file.NewRing(lcfg)
			if err != nil {
				return fmt.Errorf("unable to init ring: %w", err)
			}

			s, err := NewScreen(lcfg, ring)
			if err != nil {
				return err
			}

			return s.Run()
		},
	}

	// create process context
	processCtx, processCancel := context.WithCancel(context.Background())
	var process run.Group
	{

		// add root command to process
		process.Add(func() error {
			return root.ParseAndRun(processCtx, args)
		}, func(error) {
			processCancel()
		})
	}

	// start process
	switch err := process.Run(); err {
	case flag.ErrHelp, nil: // ok
	case context.Canceled, context.DeadlineExceeded:
		fmt.Fprintf(os.Stderr, "interrupted: %s\n", err.Error())
	default:
		fmt.Fprintf(os.Stderr, "process error: %s\n", err.Error())
	}
}
