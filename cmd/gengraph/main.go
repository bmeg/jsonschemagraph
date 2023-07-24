package gengraph

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	jsgraph "github.com/bmeg/jsonschemagraph/util"
	"github.com/spf13/cobra"
)

func countLines(filePath string) int {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println(err)
		return 0
	}
	defer file.Close()

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		fmt.Println(err)
		return 0
	}
	defer gzReader.Close()
	// for some of these files the buffer wasn't large enough to get a line count
	const maxCapacity = 1024 * 1024
	buf := make([]byte, maxCapacity)
	scanner := bufio.NewScanner(gzReader)
	scanner.Buffer(buf, maxCapacity)
	count := 0
	for scanner.Scan() {
		count++
	}

	return count
}

func write_to_file(lin jsgraph.GraphElement, file *os.File, visited bool) (bool, error) {
	if b, err := json.Marshal(lin.InEdge); err == nil {
		if string(b) != "null" {
			if visited {
				_, err := file.WriteString(string(b))
				visited = !visited
				return visited, err
			} else {
				_, err := file.WriteString("\n" + string(b))
				return visited, err
			}
		}
	}
}

// https://github.com/bmeg/sifter/blob/51a67b0de852e429d30b9371d9975dbefe3a8df9/transform/graph_build.go#L86
var Cmd = &cobra.Command{
	Use:   "gen-graph [schema dir] [data dir] [out dir] [class name]",
	Short: "Generates edges and vertices from source data files and schemas",
	Args:  cobra.MinimumNArgs(4),
	RunE: func(cmd *cobra.Command, args []string) error {
		if out, err := jsgraph.Load(args[0]); err == nil {
			fmt.Println("Loaded ", out.ListClasses())
			fmt.Println("Writing: ", args[2]+"/"+args[3])
			if _, err := os.Stat(args[2]); os.IsNotExist(err) {
				fmt.Println("Path: ", args[2], " does not exist. Creating directory path")
				err := os.Mkdir(args[2], os.ModePerm)
				if err != nil {
					log.Fatal(err)
				}
			}
			// this fuction "ReadGzipLines" used to come from sifter but needed to add an argument to specify buffer size
			// for a single line because some of the file lines needed a larger buffer to load
			// current buffer size 1 MB
			if reader, err := jsgraph.ReadGzipLines(args[1], 1024*1024); err == nil {
				var line_count int = countLines(args[1])
				fmt.Println("INIT VALS ", args[1], line_count)
				procChan := make(chan map[string]any, line_count)
				go func() {
					for line := range reader {

						o := map[string]any{}
						if len(line) > 0 {
							json.Unmarshal(line, &o)
							procChan <- o
						}
					}
					close(procChan)
				}()

				// creates output string that is unique so that files of the same class do not overwrite eachother
				base_name := strings.Split(args[1], "/")
				unqiue_str_tmp := strings.Split(base_name[len(base_name)-1], ".")
				unique_str := strings.Join(unqiue_str_tmp[0:len(unqiue_str_tmp)-2], ".")

				vertex_file, err := os.Create(args[2] + "/" + unique_str + ".Vertex.json")
				if err != nil {
					fmt.Println("ERROR ON FILE CREATE", err)
				}
				defer vertex_file.Close()

				InEdge_file, err := os.Create(args[2] + "/" + unique_str + ".InEdge.json")
				if err != nil {
					fmt.Println("ERROR ON FILE CREATE", err)
				}
				defer InEdge_file.Close()

				OutEdege_file, err := os.Create(args[2] + "/" + unique_str + ".OutEdge.json")
				if err != nil {
					fmt.Println("ERROR ON FILE CREATE", err)
				}
				defer OutEdege_file.Close()

				var IedgeInit, VertexInit, OedegeInit = true, true, true
				for line := range procChan {
					if result, err := out.Generate(args[3], line, false); err == nil {
						for _, lin := range result {
							IedgeInit, err = write_to_file(lin, InEdge_file, IedgeInit)
							OedegeInit, err = write_to_file(lin, OutEdege_file, OedegeInit)
							VertexInit, err = write_to_file(lin, vertex_file, VertexInit)

							if err != nil {
								fmt.Println("Error during write")
							}
						}
					}
				}
				jsgraph.Check_delete(args[2] + "/" + unique_str + ".OutEdge.json")
				jsgraph.Check_delete(args[2] + "/" + unique_str + ".InEdge.json")
				jsgraph.Check_delete(args[2] + "/" + unique_str + ".Vertex.json")
			}
		}
		return nil
	},
}

func init() {
}
