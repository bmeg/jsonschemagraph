package validate

import (
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/bmeg/golib"
	"github.com/bmeg/jsonschemagraph/graph"
	"github.com/bytedance/sonic"
	"github.com/hashicorp/go-multierror"
	"github.com/spf13/cobra"
)

// Cmd is the declaration of the command line
var Cmd = &cobra.Command{
	Use:   "data-validate [schemaFile] [inputFile]",
	Short: "Data Validate",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return ValidateData(args[0], args[1])
	},
}

func ValidateData(schemaFilePath, dataFilePath string) error {
	var overallErrors *multierror.Error
	var err error
	var sch *graph.GraphSchema

	type RowResult struct {
		Data map[string]any
		Err  error
	}

	sch, err = graph.Load(schemaFilePath)
	if err != nil {
		overallErrors = multierror.Append(overallErrors, fmt.Errorf("schema load error: %w", err))
		return overallErrors.ErrorOrNil()
	}

	var reader chan []byte
	if strings.HasSuffix(dataFilePath, ".gz") {
		reader, err = golib.ReadGzipLines(dataFilePath)
	} else {
		reader, err = golib.ReadFileLines(dataFilePath)
	}
	if err != nil {
		overallErrors = multierror.Append(overallErrors, fmt.Errorf("file reader error: %w", err))
		return overallErrors.ErrorOrNil()
	}

	procChan := make(chan RowResult, 100)
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		for line := range reader {
			var o map[string]any

			if len(line) == 0 {
				continue
			}

			if sonic.Unmarshal(line, &o) != nil {
				procChan <- RowResult{
					Err: fmt.Errorf("json unmarshal failed for line '%s': %w", line, err),
				}
				continue
			}
			procChan <- RowResult{Data: o}
		}
	}()
	go func() {
		wg.Wait()
		close(procChan)
	}()

	validCount := 0
	errorCount := 0
	for rowResult := range procChan {
		if rowResult.Err != nil {
			overallErrors = multierror.Append(overallErrors, rowResult.Err)
			errorCount++
			continue
		}

		row := rowResult.Data
		resourceType, ok := row["resourceType"].(string)
		if !ok {
			err = fmt.Errorf("data validation failed for line %v: required field 'resourceType' is missing or not a string", row["lineNumber"])
			overallErrors = multierror.Append(overallErrors, err)
			errorCount++
			continue
		}

		err = sch.Validate(resourceType, row)
		if err != nil {
			overallErrors = multierror.Append(overallErrors, err)
			errorCount++
		} else {
			validCount++
		}
	}

	log.Printf("%s results: %d valid records, %d invalid records\n", dataFilePath, validCount, errorCount)
	return overallErrors.ErrorOrNil()
}
