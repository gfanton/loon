package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/nxadm/tail"
	"github.com/teacat/noire"
)

type Reader interface {
	Lines() int
	ResetLines()
	Sources() []File
	Readline() (line string, src SourceID, err error)
}

type SourceID uint32

var huecolor noire.Color = noire.NewRGB(235, 0, 0)

func (sid SourceID) Color(shade float64) tcell.Color {
	hue := huecolor.AdjustHue(float64(sid%72) * 5).Shade(shade)
	return tcell.NewRGBColor(int32(hue.Red), int32(hue.Green), int32(hue.Blue))
}

type File struct {
	ID    SourceID
	Path  string
	Stdin bool
}

func NewFile(path string, stdin bool) (f File) {
	f.Path = path
	f.Stdin = stdin
	f.ID = SourceID(hashString(path))
	return
}

func NewReader(lcfg *LoonConfig, f File) (Reader, error) {
	size := lcfg.RingSize

	var cursor int64
	if !f.Stdin {
		if _, err := os.Stat(f.Path); err == nil {
			cursor, err = getPostionFromBottom(f.Path, int64(size))
			if err != nil {
				return nil, fmt.Errorf("unable to seek file position: %w", err)
			}
		}
	}

	tail, err := tailFile(cursor, f)
	if err != nil {
		return nil, fmt.Errorf("unable to tail file: %w", err)
	}

	return &TailReader{
		lines: 0,
		file:  f,
		tail:  tail,
	}, nil
}

func tailFile(cursor int64, f File) (*tail.Tail, error) {
	config := tail.Config{
		ReOpen: true,
		Follow: true,
		Logger: tail.DiscardingLogger,
		Pipe:   f.Stdin,
	}

	if f.Stdin {
		f.Path = os.Stdin.Name()
		config.Pipe = f.Stdin
	} else if f.Path != "" {
		config.Location = &tail.SeekInfo{Offset: cursor}
	} else {
		return nil, fmt.Errorf("no valid path given")
	}

	return tail.TailFile(f.Path, config)
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

type TailReader struct {
	tail  *tail.Tail
	lines int
	file  File

	muLines sync.RWMutex
}

func (f *TailReader) Lines() (l int) {
	f.muLines.RLock()
	l = f.lines
	f.muLines.RUnlock()
	return
}

func (f *TailReader) Sources() (src []File) {
	f.muLines.RLock()
	src = []File{f.file}
	f.muLines.RUnlock()
	return
}

func (f *TailReader) ResetLines() {
	f.muLines.Lock()
	f.lines = 0
	f.muLines.Unlock()
}

func (f *TailReader) Readline() (s string, sid SourceID, err error) {
	sid = f.file.ID
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

type multiReaderSource struct {
	sid  SourceID
	line string
	err  error
}

type MultiReader struct {
	rootCtx context.Context

	readers  []Reader
	sources  []File
	cline    chan *multiReaderSource
	muReader sync.RWMutex
}

func NewMultiReader(readers ...Reader) *MultiReader {
	wg := sync.WaitGroup{}
	cline := make(chan *multiReaderSource)
	for _, reader := range readers {
		wg.Add(1)
		go func(reader Reader) {
			defer wg.Done()
			for {
				line, sid, err := reader.Readline()
				cline <- &multiReaderSource{
					sid:  sid,
					line: line,
					err:  err,
				}

				if err != nil {
					return
				}
			}
		}(reader)
	}

	go func() {
		wg.Wait()
		close(cline)
	}()

	sources := []File{}
	for _, reader := range readers {
		sources = append(sources, reader.Sources()...)
	}

	return &MultiReader{
		readers: readers,
		cline:   cline,
		sources: sources,
	}
}

func (m *MultiReader) Lines() (l int) {
	m.muReader.RLock()
	for _, r := range m.readers {
		l += r.Lines()
	}

	m.muReader.RUnlock()
	return
}

func (m *MultiReader) Sources() (src []File) {
	m.muReader.RLock()
	src = m.sources
	m.muReader.RUnlock()
	return
}

func (m *MultiReader) ResetLines() {
	m.muReader.Lock()
	for _, r := range m.readers {
		r.ResetLines()
	}
	m.muReader.Unlock()
}

func (m *MultiReader) Readline() (s string, sid SourceID, err error) {
	if source := <-m.cline; source != nil {
		s, sid, err = source.line, source.sid, source.err
		return
	}

	err = fmt.Errorf("no more reader to read")
	return
}
