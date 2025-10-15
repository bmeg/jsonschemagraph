package util

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/bmeg/jsonschema/v6"
)

func GetObjectID(data map[string]any, schema *jsonschema.Schema) (string, error) {
	id, ok := data["id"].(string)
	if !ok {
		return "", fmt.Errorf("object 'id' not found in data: %s", data)
	}
	return id, nil
}

func CountLines(filePath string) int {
	file, err := os.Open(filePath)
	if err != nil {
		log.Println(err)
		return 0
	}
	defer file.Close()

	var reader *bufio.Reader
	if strings.HasSuffix(filePath, ".gz") {
		gzReader, err := gzip.NewReader(file)
		if err != nil {
			log.Println(err)
			return 0
		}
		defer gzReader.Close()
		reader = bufio.NewReader(gzReader)
	} else if strings.HasSuffix(filePath, ".ndjson") {
		reader = bufio.NewReader(file)
	}

	const maxCapacity = 1024 * 1024
	buf := make([]byte, maxCapacity)
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(buf, maxCapacity)
	count := 0
	for scanner.Scan() {
		count++
	}
	return count
}

func ListFilesWithExtension(path string, suffixes []string) ([]string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		for _, suf := range suffixes {
			if strings.HasSuffix(info.Name(), suf) {
				return []string{path}, nil
			}
		}
		return []string{}, nil
	}
	var files []string
	err = filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		for _, suf := range suffixes {
			if strings.HasSuffix(info.Name(), suf) {
				files = append(files, p)
			}
		}
		return nil
	})
	return files, err
}

func Write_line(init bool, b []byte, file_writer *os.File, gz_writer *gzip.Writer) bool {
	var err error
	if string(b) != "null" {
		if init {
			if gz_writer != nil {
				_, err = gz_writer.Write(b)
			} else {
				_, err = file_writer.WriteString(string(b))
			}
			init = !init
		} else {
			if gz_writer != nil {
				_, err = gz_writer.Write([]byte("\n" + string(b)))
			} else {
				_, err = file_writer.WriteString("\n" + string(b))
			}
		}
	}
	if err != nil {
		log.Fatal("Write File error", err)
	}
	return init
}
