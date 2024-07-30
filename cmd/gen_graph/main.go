package gen_graph

import (
	"encoding/json"
	"log"
	"os"
	"strings"

	"github.com/bmeg/jsonschemagraph/util"
	"github.com/spf13/cobra"
)

var project_id string

// https://github.com/bmeg/sifter/blob/51a67b0de852e429d30b9371d9975dbefe3a8df9/transform/graph_build.go#L86
var Cmd = &cobra.Command{
	Use:   "gen-graph [schema dir] [data dir] [out dir] [class name]",
	Short: "Generates edges and vertices from source data files and schemas",
	Args:  cobra.MinimumNArgs(4),
	RunE: func(cmd *cobra.Command, args []string) error {
		var reader chan []byte
		var out util.GraphSchema
		var err error

		if out, err = util.Load(args[0]); err != nil {
			log.Fatal("ERROR: ", err)
		}

		log.Println("Loaded ", out.ListClasses())
		if _, err := os.Stat(args[2]); os.IsNotExist(err) {
			log.Println("Path: ", args[2], " does not exist. Creating directory path")
			err := os.Mkdir(args[2], os.ModePerm)
			if err != nil {
				log.Fatal(err)
			}
		}
		// current buffer size 1 MB
		if strings.HasSuffix(args[1], ".gz") {
			if reader, err = util.ReadGzipLines(args[1], 1024*1024); err != nil {
				log.Fatal("ERROR: ", err)
			}
		} else if strings.HasSuffix(args[1], ".json") || strings.HasSuffix(args[1], ".ndjson") {
			if reader, err = util.ReadFileLines(args[1], 1024*1024); err != nil {
				log.Fatal("ERROR: ", err)
			}
		}

		var line_count int = util.CountLines(args[1])
		log.Println("File Name:", args[1], "Line Count:", line_count)
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

		base_name := strings.Split(args[1], "/")
		ClassName := strings.Split(base_name[len(base_name)-1], ".")[0]
		log.Println("Writing:", args[2]+"/"+ClassName)

		vertex_file, err := os.Create(args[2] + "/" + ClassName + ".Vertex.json")
		if err != nil {
			log.Println("ERROR ON FILE CREATE", err)
		}
		defer vertex_file.Close()

		InEdge_file, err := os.Create(args[2] + "/" + ClassName + ".InEdge.json")
		if err != nil {
			log.Println("ERROR ON FILE CREATE", err)
		}
		defer InEdge_file.Close()

		OutEdege_file, err := os.Create(args[2] + "/" + ClassName + ".OutEdge.json")
		if err != nil {
			log.Println("ERROR ON FILE CREATE", err)
		}
		defer OutEdege_file.Close()

		var IedgeInit, VertexInit, OedegeInit = true, true, true
		for line := range procChan {
			if result, err := out.Generate(args[3], line, false, project_id); err == nil {
				for _, lin := range result {
					if b, err := json.Marshal(lin.InEdge); err == nil {
						if string(b) != "null" {
							if IedgeInit {
								_, err := InEdge_file.WriteString(string(b))
								IedgeInit = !IedgeInit
								if err != nil {
									log.Fatal("Write File error")
								}
							} else {
								_, err := InEdge_file.WriteString("\n" + string(b))
								if err != nil {
									log.Fatal("Write File error")
								}
							}
						}
					}
					if b, err := json.Marshal(lin.OutEdge); err == nil {
						if string(b) != "null" {
							if OedegeInit {
								_, err := OutEdege_file.WriteString(string(b))
								OedegeInit = !OedegeInit
								if err != nil {
									log.Fatal("Write File error")
								}
							} else {
								_, err := OutEdege_file.WriteString("\n" + string(b))
								if err != nil {
									log.Fatal("Write File error")
								}
							}
						}
					}
					if b, err := json.Marshal(lin.Vertex); err == nil {
						if string(b) != "null" {
							if VertexInit {
								_, err := vertex_file.WriteString(string(b))
								VertexInit = !VertexInit
								if err != nil {
									log.Fatal("Write File error")
								}
							} else {
								_, err := vertex_file.WriteString("\n" + string(b))
								if err != nil {
									log.Fatal("Write File error")
								}
							}
						}
					}
				}
			} else if err != nil {
				log.Fatal(err)
			}
		}
		util.Check_delete(args[2] + "/" + ClassName + ".OutEdge.json")
		util.Check_delete(args[2] + "/" + ClassName + ".InEdge.json")
		util.Check_delete(args[2] + "/" + ClassName + ".Vertex.json")

		return nil
	},
}

func init() {
	Cmd.Flags().StringVar(&project_id, "project_id", "", "specify project_id if loading into gen3")
}
