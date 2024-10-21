package util

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/bmeg/jsonschema/v5"
)

func GetObjectID(data map[string]any, schema *jsonschema.Schema) (string, error) {
	if id, ok := data["id"]; ok {
		if idStr, ok := id.(string); ok {
			return idStr, nil
		}
	}
	return "", fmt.Errorf("object id not found")
}

func contains(elems []string, v string) bool {
	for _, s := range elems {
		if v == s {
			return true
		}
	}
	return false
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
	} else if strings.HasSuffix(filePath, ".ndjson") || strings.HasSuffix(filePath, ".json") {
		reader = bufio.NewReader(file)
	}

	// for some of these files the buffer wasn't large enough to get a line count
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

func ListFilesWithExtension(dir string, suffixes []string) ([]string, error) {
	var files []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Check if the file has the specified extension
		for _, suf := range suffixes {
			if !info.IsDir() && strings.HasSuffix(info.Name(), suf) {
				files = append(files, path)
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
