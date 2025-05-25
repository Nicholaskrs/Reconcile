package util

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
)

// ParseCSVRecordsAsync reads a CSV file asynchronously.
// It takes the CSV file path and a converter function that converts each CSV row (string slice)
// into a *T and an error.
// It returns two channels: one for the slice of parsed results and one for any error encountered.
// The parsing happens in a separate goroutine and results/errors are sent via channels.
func ParseCSVRecordsAsync[T any](filePath string, rowConverter func(record []string) (*T, error)) (<-chan []*T, <-chan error) {
	resultCh := make(chan []*T, 1)
	errCh := make(chan error, 1)

	go func() {
		res, err := ParseCSVRecords(filePath, rowConverter)
		if err != nil {
			errCh <- err
			return
		}
		resultCh <- res
	}()

	return resultCh, errCh
}

// ParseCSVRecords reads a CSV line-by-line and applies a converter function
// that returns a *T and an error. It collects and returns all parsed results.
func ParseCSVRecords[T any](filePath string, parseFn func(record []string) (*T, error)) ([]*T, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("could not open file %s: %w", filePath, err)
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Printf("failed to close file %s: %v", filePath, err)
		}
	}(file)

	reader := csv.NewReader(file)
	var result []*T
	rowIndex := 0

	for {
		record, err := reader.Read()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return nil, fmt.Errorf("error reading CSV at row %d: %w", rowIndex, err)
		}

		item, err := parseFn(record)
		if err != nil {
			return nil, fmt.Errorf("error parsing row %d: %w", rowIndex, err)
		}

		result = append(result, item)
		rowIndex++
	}

	return result, nil
}
