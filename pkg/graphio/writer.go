package graphio

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
)

func WriteCSV(filePath string, headers []string, data [][]string) error {
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if err := writer.Write(headers); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	for i, row := range data {
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write row %d: %w", i+1, err)
		}
	}

	return nil
}

func WriteIntMapToCSV(filePath string, headers []string, data map[int]int) error {
	var keys []int
	for key := range data {
		keys = append(keys, key)
	}
	sort.Ints(keys)

	rows := make([][]string, len(keys))
	for i, key := range keys {
		value := data[key]
		rows[i] = []string{strconv.Itoa(key), strconv.Itoa(value)}
	}

	return WriteCSV(filePath, headers, rows)
}
