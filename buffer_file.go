package main

import (
	"bufio"
	"container/ring"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/nxadm/tail"
)

type Reader interface {
	Lines() int64
	ResetLines()
	Readline() (string, error)
}

type File struct {
	Stdin bool
	Path  string
}

func (f *File) NewRing(lcfg *LoonConfig) (*Buffer, error) {
	size := lcfg.RingSize

	var parser Parser
	if lcfg.NoAnsi {
		parser = &RawParser{}
	} else {
		parser = &ANSIParser{NoColor: lcfg.NoColor}
	}

	if f.Stdin {
		reader := &PipeReader{
			lines:  0,
			reader: bufio.NewReader(os.Stdin),
		}

		return NewBuffer(size, parser, reader), nil
	}

	var fsize, cursor, lines int64
	ring := ring.New(size + 1)

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

	reader := &FileReader{
		lines: lines,
		tail:  tail,
	}

	return NewBufferFromRing(ring, parser, reader), nil
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
	lines  int64

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

func (p *PipeReader) Lines() (l int64) {
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
	lines int64

	muLines sync.RWMutex
}

func (f *FileReader) Lines() (l int64) {
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
