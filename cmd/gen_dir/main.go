package gen_dir

import (
	"compress/gzip"
	"encoding/json"
	"log"
	"os"
	"strings"

	"github.com/bmeg/jsonschemagraph/graph"
	"github.com/bmeg/jsonschemagraph/util"
	"github.com/spf13/cobra"
)

var extraArgs string
var gzip_files bool

// https://github.com/bmeg/sifter/blob/51a67b0de852e429d30b9371d9975dbefe3a8df9/transform/graph_build.go#L86
var Cmd = &cobra.Command{
	Use:   "gen-dir [schema dir] [data dir] [out dir]",
	Short: "Generates edges and vertices from source data files and schemas",
	Args:  cobra.MinimumNArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		var reader chan []byte
		var out graph.GraphSchema
		var err error

		files, err := util.ListFilesWithExtension(args[1], []string{".gz", ".ndjson", ".json"})
		if err != nil {
			log.Fatal("ListFilesWithExtension Error: ", err)
		}

		var mapstringArgs map[string]any
		err = json.Unmarshal([]byte(extraArgs), &mapstringArgs)
		if err != nil {
			log.Fatal("Error unmarshaling JSON:", err)
			return nil
		}

		if out, err = graph.Load(args[0]); err != nil {
			log.Fatal("graph.Load: ", err)
		}
		log.Println("Loaded ", out.ListClasses())
		for _, file := range files {
			if _, err := os.Stat(args[2]); os.IsNotExist(err) {
				log.Println("Path: ", args[2], " does not exist. Creating directory path")
				err := os.Mkdir(args[2], os.ModePerm)
				if err != nil {
					log.Fatal("os.Mkdir:", err)
				}
			}
			// current buffer size 1 MB
			if strings.HasSuffix(file, ".gz") {
				if reader, err = util.ReadGzipLines(file, 1024*1024); err != nil {
					log.Fatal("util.ReadGzipLines: ", err)
				}
			} else if strings.HasSuffix(file, ".json") || strings.HasSuffix(file, ".ndjson") {
				if reader, err = util.ReadFileLines(file, 1024*1024); err != nil {
					log.Fatal("util.ReadGzipLines: ", err)
				}
			}

			var line_count int = util.CountLines(file)
			log.Println("File Name:", file, "Line Count:", line_count)
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

			base_name := strings.Split(file, "/")
			ClassName := strings.Split(base_name[len(base_name)-1], ".")[0]
			log.Println("CLASSNAME: ", ClassName)
			log.Println("Writing:", args[2]+"/"+ClassName)

			vertex_file_path := args[2] + "/" + ClassName + ".vertex.json"
			inedge_file_path := args[2] + "/" + ClassName + ".out.edge.json"
			outedge_file_path := args[2] + "/" + ClassName + ".in.edge.json"

			if gzip_files {
				vertex_file_path += ".gz"
				inedge_file_path += ".gz"
				outedge_file_path += ".gz"
			}

			vertex_file, err := os.Create(vertex_file_path)
			if err != nil {
				log.Printf("os.Create(%s) Err: %s\n", vertex_file_path, err)
			}
			defer vertex_file.Close()

			InEdge_file, err := os.Create(inedge_file_path)
			if err != nil {
				log.Printf("os.Create(%s) Err: %s\n", inedge_file_path, err)
			}
			defer InEdge_file.Close()

			OutEdege_file, err := os.Create(outedge_file_path)
			if err != nil {
				log.Printf("os.Create(%s) Err: %s\n", outedge_file_path, err)
			}
			defer OutEdege_file.Close()

			var InEdge_gzWriter *gzip.Writer
			var OutEdge_gzWriter *gzip.Writer
			var Vertex_gzwriter *gzip.Writer
			if gzip_files {
				InEdge_gzWriter = gzip.NewWriter(InEdge_file)
				defer InEdge_gzWriter.Close()

				OutEdge_gzWriter = gzip.NewWriter(OutEdege_file)
				defer OutEdge_gzWriter.Close()

				Vertex_gzwriter = gzip.NewWriter(vertex_file)
				defer Vertex_gzwriter.Close()
			}

			var IedgeInit, VertexInit, OedegeInit = true, true, true
			for line := range procChan {
				if result, err := out.Generate(ClassName, line, false, mapstringArgs); err == nil {
					for _, lin := range result {
						if b, err := json.Marshal(lin.Edge); err == nil {
							IedgeInit = util.Write_line(IedgeInit, b, InEdge_file, InEdge_gzWriter)
						}
						if b, err := json.Marshal(lin.Edge); err == nil {
							OedegeInit = util.Write_line(OedegeInit, b, OutEdege_file, OutEdge_gzWriter)

						}
						if b, err := json.Marshal(lin.Vertex); err == nil {
							VertexInit = util.Write_line(VertexInit, b, vertex_file, Vertex_gzwriter)

						}
					}
				} else if err != nil {
					log.Fatal(err)
				}
			}
			util.Check_delete(vertex_file_path)
			util.Check_delete(inedge_file_path)
			util.Check_delete(outedge_file_path)

		}
		return nil
	},
}

func init() {
	Cmd.Flags().StringVar(&extraArgs, "extraArgs", "", "specify extra args in dict format. Args are applied to every vertex")
	Cmd.Flags().BoolVar(&gzip_files, "gzip_files", false, "specify output files to be gzipped")

}
