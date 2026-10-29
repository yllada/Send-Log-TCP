// Package logimport parses log files (JSON, CSV, plain text) into message
// lists, reusing the exact parsing rules of the GUI batch import feature.
package logimport

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Result represents the result of a log import operation
type Result struct {
	Messages   []string `json:"messages"`
	TotalLines int      `json:"totalLines"`
	Errors     []string `json:"errors"`
}

// Import reads and parses a log file, returning the messages
func Import(filePath string) (Result, error) {
	result := Result{
		Messages: []string{},
		Errors:   []string{},
	}

	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return result, fmt.Errorf("failed to read file: %w", err)
	}

	// Detect format based on extension
	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case ".json":
		return parseJSONLogs(content)
	case ".csv":
		return parseCSVLogs(content)
	default:
		// Try JSON first, then fallback to plain text
		if len(content) > 0 && (content[0] == '[' || content[0] == '{') {
			jsonResult, err := parseJSONLogs(content)
			if err == nil && len(jsonResult.Messages) > 0 {
				return jsonResult, nil
			}
		}
		return parseTextLogs(content)
	}
}

// parseJSONLogs parses JSON format logs
func parseJSONLogs(content []byte) (Result, error) {
	result := Result{
		Messages: []string{},
		Errors:   []string{},
	}

	// Try parsing as array of strings
	var stringArray []string
	if err := json.Unmarshal(content, &stringArray); err == nil {
		for _, msg := range stringArray {
			if trimmed := strings.TrimSpace(msg); trimmed != "" {
				result.Messages = append(result.Messages, trimmed)
			}
		}
		result.TotalLines = len(stringArray)
		return result, nil
	}

	// Try parsing as array of objects with "message" field
	var objArray []map[string]interface{}
	if err := json.Unmarshal(content, &objArray); err == nil {
		result.TotalLines = len(objArray)
		for i, obj := range objArray {
			if msg, ok := obj["message"].(string); ok && strings.TrimSpace(msg) != "" {
				result.Messages = append(result.Messages, strings.TrimSpace(msg))
			} else if msg, ok := obj["msg"].(string); ok && strings.TrimSpace(msg) != "" {
				result.Messages = append(result.Messages, strings.TrimSpace(msg))
			} else if msg, ok := obj["log"].(string); ok && strings.TrimSpace(msg) != "" {
				result.Messages = append(result.Messages, strings.TrimSpace(msg))
			} else {
				result.Errors = append(result.Errors, fmt.Sprintf("Line %d: no 'message', 'msg', or 'log' field found", i+1))
			}
		}
		return result, nil
	}

	// Try parsing as newline-delimited JSON (NDJSON)
	lines := strings.Split(string(content), "\n")
	result.TotalLines = len(lines)
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var obj map[string]interface{}
		if err := json.Unmarshal([]byte(line), &obj); err == nil {
			if msg, ok := obj["message"].(string); ok && strings.TrimSpace(msg) != "" {
				result.Messages = append(result.Messages, strings.TrimSpace(msg))
			} else if msg, ok := obj["msg"].(string); ok && strings.TrimSpace(msg) != "" {
				result.Messages = append(result.Messages, strings.TrimSpace(msg))
			} else {
				result.Errors = append(result.Errors, fmt.Sprintf("Line %d: no message field found", i+1))
			}
		} else {
			result.Errors = append(result.Errors, fmt.Sprintf("Line %d: invalid JSON", i+1))
		}
	}

	return result, nil
}

// parseCSVLogs parses CSV format logs
func parseCSVLogs(content []byte) (Result, error) {
	result := Result{
		Messages: []string{},
		Errors:   []string{},
	}

	reader := csv.NewReader(strings.NewReader(string(content)))
	reader.TrimLeadingSpace = true
	reader.LazyQuotes = true

	records, err := reader.ReadAll()
	if err != nil {
		return result, fmt.Errorf("failed to parse CSV: %w", err)
	}

	if len(records) == 0 {
		return result, nil
	}

	result.TotalLines = len(records)

	// Check for header row
	messageColIdx := 0
	hasHeader := false
	header := records[0]

	for i, col := range header {
		colLower := strings.ToLower(strings.TrimSpace(col))
		if colLower == "message" || colLower == "msg" || colLower == "log" {
			messageColIdx = i
			hasHeader = true
			break
		}
	}

	startRow := 0
	if hasHeader {
		startRow = 1
	}

	for i := startRow; i < len(records); i++ {
		record := records[i]
		if len(record) > messageColIdx {
			msg := strings.TrimSpace(record[messageColIdx])
			if msg != "" {
				result.Messages = append(result.Messages, msg)
			}
		} else {
			result.Errors = append(result.Errors, fmt.Sprintf("Line %d: insufficient columns", i+1))
		}
	}

	return result, nil
}

// parseTextLogs parses plain text logs (one message per line)
func parseTextLogs(content []byte) (Result, error) {
	result := Result{
		Messages: []string{},
		Errors:   []string{},
	}

	lines := strings.Split(string(content), "\n")
	result.TotalLines = len(lines)

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			result.Messages = append(result.Messages, trimmed)
		}
	}

	return result, nil
}
