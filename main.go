package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/oklog/run"
	"github.com/peterbourgon/ff/v3"
	"github.com/peterbourgon/ff/v3/ffcli"
	"github.com/peterbourgon/ff/v3/fftoml"
)

type LoonConfig struct {
	RingSize   int
	LineSize   int
	ConfigFile string
	Json       bool

	// color
	NoColor       bool
	NoAnsi        bool
	BgSourceColor bool
	FgSourceColor bool

	// Debug      bool
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
	rootFlagSet.BoolVar(&cfg.BgSourceColor, "bgcolor", false, "enable background color on multiple sources")
	rootFlagSet.BoolVar(&cfg.FgSourceColor, "fgcolor", true, "enable forground color on multiple sources")
	rootFlagSet.IntVar(&cfg.RingSize, "ringsize", 100000, "ring line capacity")
	rootFlagSet.IntVar(&cfg.LineSize, "linesize", 10000, "max line size")
	// rootFlagSet.BoolVar(&cfg.Debug, "debug", false, "debug mode") // @TODO

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
	log.SetOutput(os.Stderr)

	args := os.Args[1:]

	lcfg, err := parseRootConfig(args)
	if err != nil {
		panic(err)
	}

	root := &ffcli.Command{
		Name:    "loon [flags] <files...>",
		FlagSet: rootFlagSet,
		Exec: func(ctx context.Context, args []string) error {
			readers := []Reader{}

			// check stdin
			{
				fi, _ := os.Stdin.Stat()
				if (fi.Mode() & os.ModeCharDevice) == 0 {
					path, stdin := os.Stdin.Name(), true
					file := NewFile(path, stdin)
					reader, err := NewReader(lcfg, file)
					if err != nil {
						return fmt.Errorf("unable to create reader from stdin: %w", err)
					}

					readers = append(readers, reader)
				}
			}

			// check for files
			{
				for _, arg := range args {
					file := NewFile(arg, false)
					reader, err := NewReader(lcfg, file)
					if err != nil {
						return fmt.Errorf("unable to create reader from `%s`: %w", arg, err)
					}

					readers = append(readers, reader)
				}
			}

			var reader Reader

			switch len(readers) {
			case 0:
				return flag.ErrHelp
			case 1:
				reader = readers[0]
			default:
				reader = NewMultiReader(readers...)
			}

			s, err := NewScreen(lcfg, reader)
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
