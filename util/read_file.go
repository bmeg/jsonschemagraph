package util

import (
	"bufio"
	"compress/gzip"
	"io"
	"os"
)

func ReadGzipLines(path string, size int) (chan []byte, error) {
	var err error
	var gfile *os.File
	if gfile, err = os.Open(path); err == nil {
		file, err := gzip.NewReader(gfile)
		if err != nil {
			return nil, err
		}
		return ReadLines(file, size)
	}
	return nil, err
}

func ReadFileLines(path string, size int) (chan []byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return ReadLines(file, size)
}

func ReadLines(r io.Reader, size int) (chan []byte, error) {
	out := make(chan []byte, 100)
	go func() {
		reader := bufio.NewReaderSize(r, size)
		var isPrefix bool = true
		var err error = nil
		var line, ln []byte
		for err == nil {
			line, isPrefix, err = reader.ReadLine()
			ln = append(ln, line...)
			if !isPrefix {
				out <- ln
				ln = []byte{}
			}
		}
		close(out)
	}()
	return out, nil
}
