package gengraph

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	jsgraph "github.com/bmeg/jsonschemagraph/util"
	"github.com/bmeg/sifter/readers"
	"github.com/spf13/cobra"
)

// https://github.com/bmeg/sifter/blob/51a67b0de852e429d30b9371d9975dbefe3a8df9/transform/graph_build.go#L86
var Cmd = &cobra.Command{
	Use:   "gen-graph",
	Short: "Gen graph",
	Args:  cobra.MinimumNArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		if out, err := jsgraph.Load(args[0]); err == nil {
			fmt.Println("Loaded ", out.ListClasses())
			//fmt.Println("OUT ", out)
			entries, err := os.ReadDir(args[1])
			if err != nil {
				log.Fatal(err)
			}
			for _, e := range entries {
				if !strings.HasSuffix(e.Name(), ".gz") {
					fmt.Println(e.Name(), "Does not have suffix .gz -- continuing")
					continue
				}

				fmt.Println("Writing: ", args[2]+"/"+e.Name())
				if _, err := os.Stat(args[2]); os.IsNotExist(err) {
					fmt.Println("Path: ", args[2], " does not exist. Creating directory path")
					err := os.Mkdir(args[2], os.ModePerm)
					if err != nil {
						log.Fatal(err)
					}
				}

				if reader, err := readers.ReadGzipLines(args[1] + e.Name()); err == nil {
					procChan := make(chan map[string]any, 100)
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
					vertex_file, err := os.Create(args[2] + "/" + e.Name() + ".Vertex.json")
					if err != nil {
						fmt.Println("ERROR ON FILE CREATE")
					}
					defer vertex_file.Close()

					InEdge_file, err := os.Create(args[2] + "/" + e.Name() + ".InEdge.json")
					if err != nil {
						fmt.Println("ERROR ON FILE CREATE")
					}
					defer InEdge_file.Close()

					OutEdege_file, err := os.Create(args[2] + "/" + e.Name() + ".OutEdge.json")
					if err != nil {
						fmt.Println("ERROR ON FILE CREATE")
					}
					defer OutEdege_file.Close()

					//print("LENGTH OF PROCCHAN ", len(procChan))
					proc_len := len(procChan) - 1
					proc_count := 0
					for line := range procChan {
						// could either read off of the file name to choose the class or try every type and use the one that doesn't produce an error
						new_type := strings.Split(strings.Split(e.Name(), ".")[0], "_")[1]
						// toggle uncomment/comment of below line depending on SWAPI/BMEG data
						new_type = strings.ToUpper(new_type[:1]) + new_type[1:]
						//fmt.Println("NEW TYPE ", new_type)
						if result, err := out.Generate(new_type, line, false); err == nil {
							for _, lin := range result {
								//fmt.Println("THE VALUE OF LIN ", lin)
								// lin contains Vertex,OutEdge,InEdge refer to generate.go
								if b, err := json.Marshal(lin.Vertex); err == nil {
									if string(b) != "null" && (proc_count < proc_len) {
										vertex_file.WriteString(string(b) + "\n")
									} else if string(b) != "null" && !(proc_count < proc_len) {
										vertex_file.WriteString(string(b))
									}
								}

								if b, err := json.Marshal(lin.InEdge); err == nil {
									if string(b) != "null" && (proc_count < proc_len) {
										InEdge_file.WriteString(string(b) + "\n")
									} else if string(b) != "null" && !(proc_count < proc_len) {
										InEdge_file.WriteString(string(b))
									}
								}

								if b, err := json.Marshal(lin.OutEdge); err == nil {
									if string(b) != "null" && (proc_count < proc_len) {
										OutEdege_file.WriteString(string(b) + "\n")
									} else if string(b) != "null" && !(proc_count < proc_len) {
										OutEdege_file.WriteString(string(b))
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
						proc_count = proc_count + 1
					}
					jsgraph.Check_delete(args[2] + "/" + e.Name() + ".OutEdge.json")
					jsgraph.Check_delete(args[2] + "/" + e.Name() + ".InEdge.json")
					jsgraph.Check_delete(args[2] + "/" + e.Name() + ".Vertex.json")
				}
			}
		}
		return nil
	},
}

func init() {
}
