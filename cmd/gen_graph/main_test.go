package gen_graph

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

type exon struct {
	Label string `json:"label"`
	From  string `json:"from"`
	To    string `json:"to"`
}

type exon_vertex struct {
	Gid   string                 `json:"gid"`
	Label string                 `json:"label"`
	Data  map[string]interface{} `json:"data"`
}

type ExonDataInfo struct {
	Chromosome  string `json:"chromosome"`
	End         int    `json:"end"`
	ExonID      string `json:"exon_id"`
	Genome      string `json:"genome"`
	ID          string `json:"id"`
	ProjectID   string `json:"project_id"`
	Start       int    `json:"start"`
	Strand      string `json:"strand"`
	SubmitterID string `json:"submitter_id"`
}

var Exons_In_Edge = []exon{
	{
		Label: "exons",
		From:  "ENST00000673477",
		To:    "ENSE00003889014",
	},
	{
		Label: "exons",
		From:  "ENST00000673477",
		To:    "ENSE00003467707",
	},
	{
		Label: "exons",
		From:  "ENST00000673477",
		To:    "ENSE00003569130",
	},
	{
		Label: "exons",
		From:  "ENST00000673477",
		To:    "ENSE00003608502",
	},
	{
		Label: "exons",
		From:  "ENST00000673477",
		To:    "ENSE00003474888",
	},
	{
		Label: "exons",
		From:  "ENST00000673477",
		To:    "ENSE00003654064",
	},
}

var Exons_Out_Edge = []exon{
	{
		Label: "transcripts",
		To:    "ENST00000673477",
		From:  "ENSE00003889014",
	},
	{
		Label: "transcripts",
		To:    "ENST00000673477",
		From:  "ENSE00003467707",
	},
	{
		Label: "transcripts",
		To:    "ENST00000673477",
		From:  "ENSE00003569130",
	},
	{
		Label: "transcripts",
		To:    "ENST00000673477",
		From:  "ENSE00003608502",
	},
	{
		Label: "transcripts",
		To:    "ENST00000673477",
		From:  "ENSE00003474888",
	},
	{
		Label: "transcripts",
		To:    "ENST00000673477",
		From:  "ENSE00003654064",
	},
}

func checkNullFields(data map[string]interface{}) bool {
	for key, value := range data {
		if value == "" {
			fmt.Printf("Warning: key %s has no value\n", key)
			return false
		} else if subData, ok := value.(map[string]interface{}); ok {
			if !checkNullFields(subData) {
				return false
			}
		}
	}
	return true
}

func Test_main(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "Match Data"},
	}
	//read file in and compare to the one in memory. How hard could it be ?
	//https: //stackoverflow.com/questions/34388083/read-entire-file-of-newline-delimited-json-blobs-to-memory-and-unmarshal-each-bl
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var arrayName = make([][]map[string]interface{}, 0)
			arrayName = append(arrayName, []map[string]interface{}{})

			for _, exon := range Exons_In_Edge {
				map_interface := make(map[string]interface{})
				map_interface["label"] = exon.Label
				map_interface["from"] = exon.From
				map_interface["to"] = exon.To
				arrayName[0] = append(arrayName[0], map_interface)
			}

			arrayName = append(arrayName, []map[string]interface{}{})
			for _, exon := range Exons_Out_Edge {
				map_interface := make(map[string]interface{})
				map_interface["label"] = exon.Label
				map_interface["from"] = exon.From
				map_interface["to"] = exon.To
				arrayName[1] = append(arrayName[1], map_interface)
			}

			if tt.name == "Match Data" {
				files, err := filepath.Glob("../../output/*")
				if err != nil {
					t.Errorf("FATAL ERROR %s, filepath.Glob failed for directory:  ../../output/*", err)
					return
				}
				for q, file := range files {
					fmt.Println(file)
					lines, err := ioutil.ReadFile(file)
					if err != nil {
						t.Errorf("FATAL ERROR %s, ioutil.Readfile failed for file %s", err, string(file))
						continue
					}
					// check number of lines in every file
					//if len(bytes.Split(lines, []byte{'\n'})) != 6 {
					///	t.Errorf("ERROR number of lines in every file does not equal 6")
					//}

					// Iterate over the NDJSON data read from file
					for i, line := range bytes.Split(lines, []byte{'\n'}) {
						var v map[string]interface{}
						if err := json.Unmarshal(line, &v); err != nil {
							t.Errorf("FATAL ERROR %s, json unmarshal failed for file %s", err, string(file))
							return
						}
						parity := checkNullFields(v)
						if parity == false {
							t.Error()
						}

						// Check equivalence of the hardcoded data to the json file data
						if strings.Contains(string(file), "../../output/ensembl_gtf.exons.exon.json.gz.InEdge.json") ||
							strings.Contains(string(file), "ensembl_gtf.exons.exon.json.gz.OutEdge.json") {
							if reflect.DeepEqual(v, arrayName[q][i]) == false {
								t.Errorf("ERROR: generated data does not match data structures")
							}

						}
					}
				}
			}

		})
	}
}
