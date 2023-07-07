package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	jsgraph "github.com/bmeg/jsonschemagraph"
	"github.com/bmeg/sifter/readers"
)

func check_delete(filePath string) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		fmt.Println("Error getting file information:", err)
		return
	}
	if fileInfo.Size() == 0 {
		defer func() {
			err = os.Remove(filePath)
		}()
	}
}

// https://github.com/bmeg/sifter/blob/51a67b0de852e429d30b9371d9975dbefe3a8df9/transform/graph_build.go#L86
func main() {
	if out, err := jsgraph.Load("old_schemas"); err == nil {
		fmt.Println(out, out.ListClasses())
		entries, err := os.ReadDir("data")
		if err != nil {
			log.Fatal(err)
		}
		for _, e := range entries {
			if !strings.HasSuffix(e.Name(), ".gz") {
				fmt.Println(e.Name(), "Does not have suffix .gz -- continuing")
				continue
			}
			fmt.Println("Writing: ", "data/"+e.Name())
			if reader, err := readers.ReadGzipLines("swapi/" + e.Name()); err == nil {
				procChan := make(chan map[string]interface{}, 100)
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
				vertex_file, err := os.Create("output/" + e.Name() + ".Vertex.json")
				if err != nil {
					fmt.Println("ERROR ON FILE CREATE")
				}
				defer vertex_file.Close()

				InEdge_file, err := os.Create("output/" + e.Name() + ".InEdge.json")
				if err != nil {
					fmt.Println("ERROR ON FILE CREATE")
				}
				defer InEdge_file.Close()

				OutEdege_file, err := os.Create("output/" + e.Name() + ".OutEdge.json")
				if err != nil {
					fmt.Println("ERROR ON FILE CREATE")
				}
				defer OutEdege_file.Close()

				print("LENGTH OF PROCCHAN ", len(procChan))
				proc_len := len(procChan) - 1
				proc_count := 0
				for line := range procChan {
					// could either read off of the file name to choose the class or try every type and use the one that doesn't produce an error
					//new_type := strings.ToUpper(class_type[(len(class_type) - 3)][:1]) + class_type[(len(class_type) - 3)][1:]
					new_type := strings.Split(strings.Split(e.Name(), ".")[0], "_")[1]
					//fmt.Println("HERE", new_type)
					if result, err := out.Generate(new_type, line, false); err == nil {
						for _, lin := range result {
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
							//if (lin  != jsgraph.GraphElement{} && lin.Vertex != nil && lin.Vertex.Label != ""){
							//    fmt.Println(lin.Vertex.Label)
							//}
						}
					}
					proc_count = proc_count + 1
				}
				check_delete("output/" + e.Name() + ".OutEdge.json")
				check_delete("output/" + e.Name() + ".InEdge.json")
				check_delete("output/" + e.Name() + ".Vertex.json")
			}
		}
	}
}
