package gengraph

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"log"
	"os"

	jsgraph "github.com/bmeg/jsonschemagraph/util"
	"github.com/bmeg/sifter/readers"
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

	scanner := bufio.NewScanner(gzReader)
	count := 0
	for scanner.Scan() {
		count++
	}

	return count
}

// https://github.com/bmeg/sifter/blob/51a67b0de852e429d30b9371d9975dbefe3a8df9/transform/graph_build.go#L86
var Cmd = &cobra.Command{
	Use:   "gen-graph",
	Short: "Gen graph",
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

			if reader, err := readers.ReadGzipLines(args[1]); err == nil {
				var lines int = countLines(args[1])
				fmt.Println("HELLO?", args[1], lines)
				procChan := make(chan map[string]any, lines)
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
				vertex_file, err := os.Create(args[2] + "/" + args[3] + ".Vertex.json")
				if err != nil {
					fmt.Println("ERROR ON FILE CREATE")
				}
				defer vertex_file.Close()

				InEdge_file, err := os.Create(args[2] + "/" + args[3] + ".InEdge.json")
				if err != nil {
					fmt.Println("ERROR ON FILE CREATE")
				}
				defer InEdge_file.Close()

				OutEdege_file, err := os.Create(args[2] + "/" + args[3] + ".OutEdge.json")
				if err != nil {
					fmt.Println("ERROR ON FILE CREATE")
				}
				defer OutEdege_file.Close()
				for line := range procChan {
					if result, err := out.Generate(args[3], line, false); err == nil {
						for _, lin := range result {
							//fmt.Println("THE VALUE OF LIN ", lin)
							// lin contains Vertex,OutEdge,InEdge refer to generate.go
							if b, err := json.Marshal(lin.Vertex); err == nil {
								if string(b) != "null" {
									vertex_file.WriteString(string(b) + "\n")
								}
							}
							if b, err := json.Marshal(lin.InEdge); err == nil {
								if string(b) != "null" {
									InEdge_file.WriteString(string(b) + "\n")
								}
							}
							if b, err := json.Marshal(lin.OutEdge); err == nil {
								if string(b) != "null" {
									OutEdege_file.WriteString(string(b) + "\n")
								}
							}
							if err != nil {
								fmt.Println("Error during write")
							}
							//if (lin != jsgraph.GraphElement{} && lin.OutEdge != nil) {
							//	fmt.Println("LIN INEDGE ", lin.OutEdge)
							//}
						}
					}
				}
				jsgraph.Check_delete(args[2] + "/" + args[3] + ".OutEdge.json")
				jsgraph.Check_delete(args[2] + "/" + args[3] + ".InEdge.json")
				jsgraph.Check_delete(args[2] + "/" + args[3] + ".Vertex.json")
			}
			//}
		}
		return nil
	},
}

func init() {
}
