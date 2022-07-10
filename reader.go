package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/nxadm/tail"
)

type Reader interface {
	Lines() int
	ResetLines()
	Readline() (string, error)
}

func NewReader(lcfg *LoonConfig, path string, stdin bool) (Reader, error) {
	size := lcfg.RingSize

	if stdin {
		reader := &PipeReader{
			lines:  0,
			reader: bufio.NewReader(os.Stdin),
		}

		return reader, nil
	}

	var cursor int64
	if _, err := os.Stat(path); err == nil {
		cursor, err = getPostionFromBottom(path, int64(size))
		if err != nil {
			return nil, fmt.Errorf("unable to seek file position: %w", err)
		}
	}

	tail, err := tailFile(cursor, path, stdin)
	if err != nil {
		return nil, fmt.Errorf("unable to tail file: %w", err)
	}

	return &FileReader{
		lines: 0,
		tail:  tail,
	}, nil
}

func tailFile(cursor int64, path string, stdin bool) (*tail.Tail, error) {
	config := tail.Config{Follow: true}

	if stdin {
		path = os.Stdin.Name()
		config.Pipe = stdin
	} else if path != "" {
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
	lines  int

	muLines sync.RWMutex
}

func (p *PipeReader) Readline() (string, error) {
	text, err := p.reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	p.muLines.Lock()
	p.lines++
	p.muLines.Unlock()

	return strings.TrimSuffix(text, "\n"), nil
}

func (p *PipeReader) Lines() (l int) {
	p.muLines.RLock()
	l = p.lines
	p.muLines.RUnlock()
	return
}

func (p *PipeReader) ResetLines() {
	p.muLines.Lock()
	p.lines = 0
	p.muLines.Unlock()
}

type FileReader struct {
	tail  *tail.Tail
	lines int

	muLines sync.RWMutex
}

func (f *FileReader) Lines() (l int) {
	f.muLines.RLock()
	l = f.lines
	f.muLines.RUnlock()
	return
}

func (f *FileReader) ResetLines() {
	f.muLines.Lock()
	f.lines = 0
	f.muLines.Unlock()
}

func (f *FileReader) Readline() (s string, err error) {
	line := <-f.tail.Lines
	if line != nil {
		f.muLines.Lock()
		f.lines++
		f.muLines.Unlock()

		s = line.Text
		return
	}

	if err = f.tail.Err(); err == nil {
		err = io.EOF
	}

	return
}
