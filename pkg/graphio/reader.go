package graphio

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
)

func ReadCSVWithHeader(filename string, skipHeader bool, expectedHeaderFirst string) ([][]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}

	if skipHeader && len(records) > 0 && records[0][0] == expectedHeaderFirst {
		records = records[1:]
	}

	return records, nil
}

func ParseIntRecord(record []string, lineNum int) ([]int, error) {
	result := make([]int, len(record))
	for i, field := range record {
		val, err := strconv.Atoi(field)
		if err != nil {
			return nil, fmt.Errorf("line %d, column %d: invalid integer: %w", lineNum, i+1, err)
		}
		result[i] = val
	}
	return result, nil
}

func ValidateRecordLength(record []string, expected int, lineNum int) error {
	if len(record) != expected {
		return fmt.Errorf("line %d: expected %d columns, got %d", lineNum, expected, len(record))
	}
	return nil
}
