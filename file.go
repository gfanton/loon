package main

import (
	"bufio"
	"container/ring"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/nxadm/tail"
)

type Reader interface {
	Readline() (string, error)
}

type File struct {
	Stdin bool
	Path  string
}

func (f *File) NewRing(size int) (*Ring, error) {
	var lines int64
	var reader Reader

	ring := ring.New(size)
	if !f.Stdin {
		var fsize, cursor int64

		fi, err := os.Stat(f.Path)
		if err == nil {
			fsize = fi.Size()
			cursor, err = getPostionFromBottom(f.Path, int64(size))
			if err != nil {
				return nil, fmt.Errorf("unable to seek file position: %w", err)
			}
		}

		tail, err := f.tailFile(cursor)
		if err != nil {
			return nil, fmt.Errorf("unable to tail file: %w", err)
		}

		// fill ring until the end of the file
		for line := range tail.Lines {
			ring.Value = line.Text
			ring = ring.Next()
			lines++

			if line.SeekInfo.Offset >= fsize {
				break
			}

		}

		reader = &FileReader{tail}
	} else {
		file := bufio.NewReader(os.Stdin)
		reader = &PipeReader{file}
	}

	return &Ring{
		updateRing: make(chan struct{}),
		ringSize:   int(size),
		ring:       ring,
		reader:     reader,
		lines:      lines,
	}, nil
}

func (f *File) tailFile(cursor int64) (*tail.Tail, error) {
	var path string
	config := tail.Config{Follow: true}

	if f.Stdin {
		path = os.Stdin.Name()
		config.Pipe = f.Stdin
	} else if f.Path != "" {
		path = f.Path
		config.Location = &tail.SeekInfo{Offset: cursor}
	} else {
		return nil, fmt.Errorf("no valid path given")
	}

	return tail.TailFile(path, config)
}

func getPostionFromBottom(path string, lines int64) (int64, error) {
	const buffersize = 2048

	file, err := os.Open(path)
	if err != nil {
		return 0, fmt.Errorf("cannot open file: %w", err)
	}
	defer file.Close()

	var cursor, counter int64
	stat, _ := file.Stat()
	filesize := stat.Size()
	for {
		cursor += buffersize
		if cursor >= filesize {
			return 0, nil
		}
		file.Seek(-cursor, io.SeekEnd)

		slice := make([]byte, buffersize)
		file.Read(slice)

		var i int
		for ; i < len(slice); i++ {
			if slice[i] != '\n' && slice[i] != '\r' {
				continue
			}

			if counter = counter + 1; counter >= lines { // stop if we are at the begining
				return filesize - (cursor + int64(i)), nil
			}
		}
	}

}

type PipeReader struct {
	reader *bufio.Reader
}

func (f *PipeReader) Readline() (string, error) {
	text, err := f.reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(text, "\n"), nil
}

type FileReader struct {
	tail *tail.Tail
}

func (f *FileReader) Readline() (s string, err error) {
	line := <-f.tail.Lines
	if line != nil {
		s = line.Text
		return
	}

	if err = f.tail.Err(); err == nil {
		err = io.EOF
	}

	return
}
