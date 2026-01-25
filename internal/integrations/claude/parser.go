package claude

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

// UNKNOWN_TOOL_NAME is used when a tool name is missing or empty
const UNKNOWN_TOOL_NAME = "unknown_tool"

// ExtractToolUses extracts tool_use blocks from content array
// Filters for type="tool_use" and generates ID if missing
func ExtractToolUses(content []ContentBlock) []ClaudeToolUse {
	var results []ClaudeToolUse

	for _, block := range content {
		if block.Type != "tool_use" {
			continue
		}

		// Extract ID, generate if missing or empty
		var id string
		if block.ID != nil && *block.ID != "" {
			id = *block.ID
		} else {
			id = generateID()
		}

		// Extract name, use UNKNOWN_TOOL_NAME if missing or empty
		name := UNKNOWN_TOOL_NAME
		if block.Name != nil && *block.Name != "" {
			name = *block.Name
		}

		// Extract parent_tool_use_id if present
		var parentID *string
		if block.ParentToolUseID != nil && *block.ParentToolUseID != "" {
			parentID = block.ParentToolUseID
		}

		results = append(results, ClaudeToolUse{
			ID:              id,
			Name:            name,
			Input:           block.Input,
			ParentToolUseID: parentID,
		})
	}

	return results
}

// ExtractToolResults extracts tool_result blocks from content array
// Filters for type="tool_result" and generates ID if tool_use_id is missing
func ExtractToolResults(content []ContentBlock) []ClaudeToolResult {
	var results []ClaudeToolResult

	for _, block := range content {
		if block.Type != "tool_result" {
			continue
		}

		// Extract tool_use_id, generate if missing or empty
		var id string
		if block.ToolUseID != nil && *block.ToolUseID != "" {
			id = *block.ToolUseID
		} else {
			id = generateID()
		}

		// Extract name if present and non-empty
		var name *string
		if block.Name != nil && *block.Name != "" {
			name = block.Name
		}

		// Extract is_error, default to false
		isError := false
		if block.IsError != nil {
			isError = *block.IsError
		}

		results = append(results, ClaudeToolResult{
			ID:      id,
			Name:    name,
			Result:  block.Content,
			IsError: isError,
		})
	}

	return results
}

// ExtractToolErrors extracts tool_error blocks from content array
// Filters for type="tool_error" and generates ID if tool_use_id is missing
func ExtractToolErrors(content []ContentBlock) []ClaudeToolError {
	var results []ClaudeToolError

	for _, block := range content {
		if block.Type != "tool_error" {
			continue
		}

		// Extract tool_use_id, generate if missing or empty
		var id string
		if block.ToolUseID != nil && *block.ToolUseID != "" {
			id = *block.ToolUseID
		} else {
			id = generateID()
		}

		// Extract name if present and non-empty
		var name *string
		if block.Name != nil && *block.Name != "" {
			name = block.Name
		}

		results = append(results, ClaudeToolError{
			ID:    id,
			Name:  name,
			Error: block.Error,
		})
	}

	return results
}

// ParseClaudeEvent parses a JSON line into a ClaudeMessage
// Returns error if the line is not valid JSON or is empty
func ParseClaudeEvent(line string) (*ClaudeMessage, error) {
	if line == "" {
		return nil, errors.New("empty line")
	}

	var message ClaudeMessage
	if err := json.Unmarshal([]byte(line), &message); err != nil {
		return nil, errors.Wrapf(err, "failed to parse Claude event")
	}

	return &message, nil
}

// generateID generates a unique ID using UUID v4
func generateID() string {
	return uuid.New().String()
}
