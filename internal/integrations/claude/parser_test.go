package claude

import (
	"testing"
)

func Test_ExtractToolUses(t *testing.T) {
	t.Run("empty content array returns empty slice", func(t *testing.T) {
		// Arrange
		content := []ContentBlock{}

		// Act
		result := ExtractToolUses(content)

		// Assert
		if len(result) != 0 {
			t.Errorf("Expected empty slice, got %d items", len(result))
		}
	})

	t.Run("filters only tool_use blocks", func(t *testing.T) {
		// Arrange
		id := "tool-123"
		name := "test_tool"
		input := map[string]any{"param": "value"}
		text := "some text"

		content := []ContentBlock{
			{Type: "text", Text: &text},
			{Type: "tool_use", ID: &id, Name: &name, Input: input},
			{Type: "tool_result", ToolUseID: &id},
		}

		// Act
		result := ExtractToolUses(content)

		// Assert
		if len(result) != 1 {
			t.Errorf("Expected 1 result, got %d", len(result))
		}
		if result[0].ID != "tool-123" {
			t.Errorf("Expected ID 'tool-123', got '%s'", result[0].ID)
		}
		if result[0].Name != "test_tool" {
			t.Errorf("Expected Name 'test_tool', got '%s'", result[0].Name)
		}
		if result[0].Input == nil {
			t.Error("Expected Input to be non-nil")
		}
	})

	t.Run("generates ID if missing", func(t *testing.T) {
		// Arrange
		name := "test_tool"
		input := map[string]any{"param": "value"}

		content := []ContentBlock{
			{Type: "tool_use", Name: &name, Input: input},
		}

		// Act
		result := ExtractToolUses(content)

		// Assert
		if len(result) != 1 {
			t.Errorf("Expected 1 result, got %d", len(result))
		}
		if result[0].ID == "" {
			t.Error("Expected generated ID, got empty string")
		}
		if result[0].Name != "test_tool" {
			t.Errorf("Expected Name 'test_tool', got '%s'", result[0].Name)
		}
	})

	t.Run("generates ID if empty string", func(t *testing.T) {
		// Arrange
		id := ""
		name := "test_tool"
		input := map[string]any{"param": "value"}

		content := []ContentBlock{
			{Type: "tool_use", ID: &id, Name: &name, Input: input},
		}

		// Act
		result := ExtractToolUses(content)

		// Assert
		if len(result) != 1 {
			t.Errorf("Expected 1 result, got %d", len(result))
		}
		if result[0].ID == "" {
			t.Error("Expected generated ID, got empty string")
		}
	})

	t.Run("uses UNKNOWN_TOOL_NAME for missing name", func(t *testing.T) {
		// Arrange
		id := "tool-123"
		input := map[string]any{"param": "value"}

		content := []ContentBlock{
			{Type: "tool_use", ID: &id, Input: input},
		}

		// Act
		result := ExtractToolUses(content)

		// Assert
		if len(result) != 1 {
			t.Errorf("Expected 1 result, got %d", len(result))
		}
		if result[0].Name != UNKNOWN_TOOL_NAME {
			t.Errorf("Expected Name '%s', got '%s'", UNKNOWN_TOOL_NAME, result[0].Name)
		}
	})

	t.Run("uses UNKNOWN_TOOL_NAME for empty name", func(t *testing.T) {
		// Arrange
		id := "tool-123"
		name := ""
		input := map[string]any{"param": "value"}

		content := []ContentBlock{
			{Type: "tool_use", ID: &id, Name: &name, Input: input},
		}

		// Act
		result := ExtractToolUses(content)

		// Assert
		if len(result) != 1 {
			t.Errorf("Expected 1 result, got %d", len(result))
		}
		if result[0].Name != UNKNOWN_TOOL_NAME {
			t.Errorf("Expected Name '%s', got '%s'", UNKNOWN_TOOL_NAME, result[0].Name)
		}
	})

	t.Run("extracts parent_tool_use_id when present", func(t *testing.T) {
		// Arrange
		id := "tool-123"
		name := "test_tool"
		parentID := "parent-456"
		input := map[string]any{"param": "value"}

		content := []ContentBlock{
			{Type: "tool_use", ID: &id, Name: &name, Input: input, ParentToolUseID: &parentID},
		}

		// Act
		result := ExtractToolUses(content)

		// Assert
		if len(result) != 1 {
			t.Errorf("Expected 1 result, got %d", len(result))
		}
		if result[0].ParentToolUseID == nil {
			t.Error("Expected ParentToolUseID to be non-nil")
		} else if *result[0].ParentToolUseID != "parent-456" {
			t.Errorf("Expected ParentToolUseID 'parent-456', got '%s'", *result[0].ParentToolUseID)
		}
	})

	t.Run("parent_tool_use_id is nil when missing", func(t *testing.T) {
		// Arrange
		id := "tool-123"
		name := "test_tool"
		input := map[string]any{"param": "value"}

		content := []ContentBlock{
			{Type: "tool_use", ID: &id, Name: &name, Input: input},
		}

		// Act
		result := ExtractToolUses(content)

		// Assert
		if len(result) != 1 {
			t.Errorf("Expected 1 result, got %d", len(result))
		}
		if result[0].ParentToolUseID != nil {
			t.Errorf("Expected ParentToolUseID to be nil, got '%s'", *result[0].ParentToolUseID)
		}
	})

	t.Run("handles multiple tool_use blocks", func(t *testing.T) {
		// Arrange
		id1 := "tool-123"
		id2 := "tool-456"
		name1 := "tool_one"
		name2 := "tool_two"
		input1 := map[string]any{"a": 1}
		input2 := map[string]any{"b": 2}

		content := []ContentBlock{
			{Type: "tool_use", ID: &id1, Name: &name1, Input: input1},
			{Type: "tool_use", ID: &id2, Name: &name2, Input: input2},
		}

		// Act
		result := ExtractToolUses(content)

		// Assert
		if len(result) != 2 {
			t.Errorf("Expected 2 results, got %d", len(result))
		}
		if result[0].ID != "tool-123" {
			t.Errorf("Expected first ID 'tool-123', got '%s'", result[0].ID)
		}
		if result[0].Name != "tool_one" {
			t.Errorf("Expected first Name 'tool_one', got '%s'", result[0].Name)
		}
		if result[1].ID != "tool-456" {
			t.Errorf("Expected second ID 'tool-456', got '%s'", result[1].ID)
		}
		if result[1].Name != "tool_two" {
			t.Errorf("Expected second Name 'tool_two', got '%s'", result[1].Name)
		}
	})
}

func Test_ExtractToolResults(t *testing.T) {
	t.Run("empty content array returns empty slice", func(t *testing.T) {
		// Arrange
		content := []ContentBlock{}

		// Act
		result := ExtractToolResults(content)

		// Assert
		if len(result) != 0 {
			t.Errorf("Expected empty slice, got %d items", len(result))
		}
	})

	t.Run("filters only tool_result blocks", func(t *testing.T) {
		// Arrange
		toolUseID := "tool-123"
		resultContent := "success"
		text := "some text"

		content := []ContentBlock{
			{Type: "text", Text: &text},
			{Type: "tool_result", ToolUseID: &toolUseID, Content: resultContent},
			{Type: "tool_use", ID: &toolUseID},
		}

		// Act
		result := ExtractToolResults(content)

		// Assert
		if len(result) != 1 {
			t.Errorf("Expected 1 result, got %d", len(result))
		}
		if result[0].ID != "tool-123" {
			t.Errorf("Expected ID 'tool-123', got '%s'", result[0].ID)
		}
		if result[0].Result != "success" {
			t.Errorf("Expected Result 'success', got '%v'", result[0].Result)
		}
	})

	t.Run("generates ID if tool_use_id missing", func(t *testing.T) {
		// Arrange
		resultContent := "success"

		content := []ContentBlock{
			{Type: "tool_result", Content: resultContent},
		}

		// Act
		result := ExtractToolResults(content)

		// Assert
		if len(result) != 1 {
			t.Errorf("Expected 1 result, got %d", len(result))
		}
		if result[0].ID == "" {
			t.Error("Expected generated ID, got empty string")
		}
	})

	t.Run("generates ID if tool_use_id is empty", func(t *testing.T) {
		// Arrange
		toolUseID := ""
		resultContent := "success"

		content := []ContentBlock{
			{Type: "tool_result", ToolUseID: &toolUseID, Content: resultContent},
		}

		// Act
		result := ExtractToolResults(content)

		// Assert
		if len(result) != 1 {
			t.Errorf("Expected 1 result, got %d", len(result))
		}
		if result[0].ID == "" {
			t.Error("Expected generated ID, got empty string")
		}
	})

	t.Run("extracts name when present", func(t *testing.T) {
		// Arrange
		toolUseID := "tool-123"
		name := "test_tool"
		resultContent := "success"

		content := []ContentBlock{
			{Type: "tool_result", ToolUseID: &toolUseID, Name: &name, Content: resultContent},
		}

		// Act
		result := ExtractToolResults(content)

		// Assert
		if len(result) != 1 {
			t.Errorf("Expected 1 result, got %d", len(result))
		}
		if result[0].Name == nil {
			t.Error("Expected Name to be non-nil")
		} else if *result[0].Name != "test_tool" {
			t.Errorf("Expected Name 'test_tool', got '%s'", *result[0].Name)
		}
	})

	t.Run("name is nil when missing", func(t *testing.T) {
		// Arrange
		toolUseID := "tool-123"
		resultContent := "success"

		content := []ContentBlock{
			{Type: "tool_result", ToolUseID: &toolUseID, Content: resultContent},
		}

		// Act
		result := ExtractToolResults(content)

		// Assert
		if len(result) != 1 {
			t.Errorf("Expected 1 result, got %d", len(result))
		}
		if result[0].Name != nil {
			t.Errorf("Expected Name to be nil, got '%s'", *result[0].Name)
		}
	})

	t.Run("name is nil when empty", func(t *testing.T) {
		// Arrange
		toolUseID := "tool-123"
		name := ""
		resultContent := "success"

		content := []ContentBlock{
			{Type: "tool_result", ToolUseID: &toolUseID, Name: &name, Content: resultContent},
		}

		// Act
		result := ExtractToolResults(content)

		// Assert
		if len(result) != 1 {
			t.Errorf("Expected 1 result, got %d", len(result))
		}
		if result[0].Name != nil {
			t.Errorf("Expected Name to be nil, got '%s'", *result[0].Name)
		}
	})

	t.Run("is_error defaults to false when missing", func(t *testing.T) {
		// Arrange
		toolUseID := "tool-123"
		resultContent := "success"

		content := []ContentBlock{
			{Type: "tool_result", ToolUseID: &toolUseID, Content: resultContent},
		}

		// Act
		result := ExtractToolResults(content)

		// Assert
		if len(result) != 1 {
			t.Errorf("Expected 1 result, got %d", len(result))
		}
		if result[0].IsError != false {
			t.Errorf("Expected IsError to be false, got %v", result[0].IsError)
		}
	})

	t.Run("is_error is true when set", func(t *testing.T) {
		// Arrange
		toolUseID := "tool-123"
		isError := true
		resultContent := "error message"

		content := []ContentBlock{
			{Type: "tool_result", ToolUseID: &toolUseID, Content: resultContent, IsError: &isError},
		}

		// Act
		result := ExtractToolResults(content)

		// Assert
		if len(result) != 1 {
			t.Errorf("Expected 1 result, got %d", len(result))
		}
		if result[0].IsError != true {
			t.Errorf("Expected IsError to be true, got %v", result[0].IsError)
		}
	})

	t.Run("handles multiple tool_result blocks", func(t *testing.T) {
		// Arrange
		toolUseID1 := "tool-123"
		toolUseID2 := "tool-456"
		content1 := "result 1"
		content2 := "result 2"

		content := []ContentBlock{
			{Type: "tool_result", ToolUseID: &toolUseID1, Content: content1},
			{Type: "tool_result", ToolUseID: &toolUseID2, Content: content2},
		}

		// Act
		result := ExtractToolResults(content)

		// Assert
		if len(result) != 2 {
			t.Errorf("Expected 2 results, got %d", len(result))
		}
		if result[0].ID != "tool-123" {
			t.Errorf("Expected first ID 'tool-123', got '%s'", result[0].ID)
		}
		if result[0].Result != "result 1" {
			t.Errorf("Expected first Result 'result 1', got '%v'", result[0].Result)
		}
		if result[1].ID != "tool-456" {
			t.Errorf("Expected second ID 'tool-456', got '%s'", result[1].ID)
		}
		if result[1].Result != "result 2" {
			t.Errorf("Expected second Result 'result 2', got '%v'", result[1].Result)
		}
	})
}

func Test_ExtractToolErrors(t *testing.T) {
	t.Run("empty content array returns empty slice", func(t *testing.T) {
		// Arrange
		content := []ContentBlock{}

		// Act
		result := ExtractToolErrors(content)

		// Assert
		if len(result) != 0 {
			t.Errorf("Expected empty slice, got %d items", len(result))
		}
	})

	t.Run("filters only tool_error blocks", func(t *testing.T) {
		// Arrange
		toolUseID := "tool-123"
		errorMsg := "something went wrong"
		text := "some text"

		content := []ContentBlock{
			{Type: "text", Text: &text},
			{Type: "tool_error", ToolUseID: &toolUseID, Error: errorMsg},
			{Type: "tool_use", ID: &toolUseID},
		}

		// Act
		result := ExtractToolErrors(content)

		// Assert
		if len(result) != 1 {
			t.Errorf("Expected 1 result, got %d", len(result))
		}
		if result[0].ID != "tool-123" {
			t.Errorf("Expected ID 'tool-123', got '%s'", result[0].ID)
		}
		if result[0].Error != "something went wrong" {
			t.Errorf("Expected Error 'something went wrong', got '%v'", result[0].Error)
		}
	})

	t.Run("generates ID if tool_use_id missing", func(t *testing.T) {
		// Arrange
		errorMsg := "something went wrong"

		content := []ContentBlock{
			{Type: "tool_error", Error: errorMsg},
		}

		// Act
		result := ExtractToolErrors(content)

		// Assert
		if len(result) != 1 {
			t.Errorf("Expected 1 result, got %d", len(result))
		}
		if result[0].ID == "" {
			t.Error("Expected generated ID, got empty string")
		}
	})

	t.Run("generates ID if tool_use_id is empty", func(t *testing.T) {
		// Arrange
		toolUseID := ""
		errorMsg := "something went wrong"

		content := []ContentBlock{
			{Type: "tool_error", ToolUseID: &toolUseID, Error: errorMsg},
		}

		// Act
		result := ExtractToolErrors(content)

		// Assert
		if len(result) != 1 {
			t.Errorf("Expected 1 result, got %d", len(result))
		}
		if result[0].ID == "" {
			t.Error("Expected generated ID, got empty string")
		}
	})

	t.Run("extracts name when present", func(t *testing.T) {
		// Arrange
		toolUseID := "tool-123"
		name := "test_tool"
		errorMsg := "something went wrong"

		content := []ContentBlock{
			{Type: "tool_error", ToolUseID: &toolUseID, Name: &name, Error: errorMsg},
		}

		// Act
		result := ExtractToolErrors(content)

		// Assert
		if len(result) != 1 {
			t.Errorf("Expected 1 result, got %d", len(result))
		}
		if result[0].Name == nil {
			t.Error("Expected Name to be non-nil")
		} else if *result[0].Name != "test_tool" {
			t.Errorf("Expected Name 'test_tool', got '%s'", *result[0].Name)
		}
	})

	t.Run("name is nil when missing", func(t *testing.T) {
		// Arrange
		toolUseID := "tool-123"
		errorMsg := "something went wrong"

		content := []ContentBlock{
			{Type: "tool_error", ToolUseID: &toolUseID, Error: errorMsg},
		}

		// Act
		result := ExtractToolErrors(content)

		// Assert
		if len(result) != 1 {
			t.Errorf("Expected 1 result, got %d", len(result))
		}
		if result[0].Name != nil {
			t.Errorf("Expected Name to be nil, got '%s'", *result[0].Name)
		}
	})

	t.Run("name is nil when empty", func(t *testing.T) {
		// Arrange
		toolUseID := "tool-123"
		name := ""
		errorMsg := "something went wrong"

		content := []ContentBlock{
			{Type: "tool_error", ToolUseID: &toolUseID, Name: &name, Error: errorMsg},
		}

		// Act
		result := ExtractToolErrors(content)

		// Assert
		if len(result) != 1 {
			t.Errorf("Expected 1 result, got %d", len(result))
		}
		if result[0].Name != nil {
			t.Errorf("Expected Name to be nil, got '%s'", *result[0].Name)
		}
	})

	t.Run("handles multiple tool_error blocks", func(t *testing.T) {
		// Arrange
		toolUseID1 := "tool-123"
		toolUseID2 := "tool-456"
		error1 := "error 1"
		error2 := "error 2"

		content := []ContentBlock{
			{Type: "tool_error", ToolUseID: &toolUseID1, Error: error1},
			{Type: "tool_error", ToolUseID: &toolUseID2, Error: error2},
		}

		// Act
		result := ExtractToolErrors(content)

		// Assert
		if len(result) != 2 {
			t.Errorf("Expected 2 results, got %d", len(result))
		}
		if result[0].ID != "tool-123" {
			t.Errorf("Expected first ID 'tool-123', got '%s'", result[0].ID)
		}
		if result[0].Error != "error 1" {
			t.Errorf("Expected first Error 'error 1', got '%v'", result[0].Error)
		}
		if result[1].ID != "tool-456" {
			t.Errorf("Expected second ID 'tool-456', got '%s'", result[1].ID)
		}
		if result[1].Error != "error 2" {
			t.Errorf("Expected second Error 'error 2', got '%v'", result[1].Error)
		}
	})
}

func Test_ParseClaudeEvent(t *testing.T) {
	t.Run("parses valid JSON line", func(t *testing.T) {
		// Arrange
		line := `{"type":"assistant","message":{"content":[{"type":"text","text":"Hello"}]}}`

		// Act
		result, err := ParseClaudeEvent(line)
		// Assert
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if result == nil {
			t.Fatal("Expected non-nil result")
		}
		if result.Type != "assistant" {
			t.Errorf("Expected Type 'assistant', got '%s'", result.Type)
		}
		if result.Message == nil {
			t.Error("Expected Message to be non-nil")
		} else if len(result.Message.Content) != 1 {
			t.Errorf("Expected 1 content block, got %d", len(result.Message.Content))
		}
	})

	t.Run("returns error for invalid JSON", func(t *testing.T) {
		// Arrange
		line := `{invalid json`

		// Act
		result, err := ParseClaudeEvent(line)

		// Assert
		if err == nil {
			t.Error("Expected error for invalid JSON, got nil")
		}
		if result != nil {
			t.Errorf("Expected nil result, got %v", result)
		}
	})

	t.Run("returns error for empty string", func(t *testing.T) {
		// Arrange
		line := ""

		// Act
		result, err := ParseClaudeEvent(line)

		// Assert
		if err == nil {
			t.Error("Expected error for empty string, got nil")
		}
		if result != nil {
			t.Errorf("Expected nil result, got %v", result)
		}
	})

	t.Run("parses message with all fields", func(t *testing.T) {
		// Arrange
		line := `{
			"type": "stream_event",
			"subtype": "content_block_start",
			"parent_tool_use_id": "parent-123",
			"session_id": "session-456",
			"usage": {
				"input_tokens": 100,
				"output_tokens": 50
			},
			"total_cost_usd": 0.001,
			"duration_ms": 500,
			"is_error": true,
			"result": "error result"
		}`

		// Act
		result, err := ParseClaudeEvent(line)
		// Assert
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if result == nil {
			t.Fatal("Expected non-nil result")
		}
		if result.Type != "stream_event" {
			t.Errorf("Expected Type 'stream_event', got '%s'", result.Type)
		}
		if result.Subtype == nil {
			t.Error("Expected Subtype to be non-nil")
		} else if *result.Subtype != "content_block_start" {
			t.Errorf("Expected Subtype 'content_block_start', got '%s'", *result.Subtype)
		}
		if result.ParentToolUseID == nil {
			t.Error("Expected ParentToolUseID to be non-nil")
		} else if *result.ParentToolUseID != "parent-123" {
			t.Errorf("Expected ParentToolUseID 'parent-123', got '%s'", *result.ParentToolUseID)
		}
		if result.SessionID == nil {
			t.Error("Expected SessionID to be non-nil")
		} else if *result.SessionID != "session-456" {
			t.Errorf("Expected SessionID 'session-456', got '%s'", *result.SessionID)
		}
		if result.IsError != true {
			t.Errorf("Expected IsError to be true, got %v", result.IsError)
		}
		if result.Result == nil {
			t.Error("Expected Result to be non-nil")
		} else if *result.Result != "error result" {
			t.Errorf("Expected Result 'error result', got '%s'", *result.Result)
		}
	})

	t.Run("handles null optional fields", func(t *testing.T) {
		// Arrange
		line := `{"type":"assistant","message":null}`

		// Act
		result, err := ParseClaudeEvent(line)
		// Assert
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if result == nil {
			t.Fatal("Expected non-nil result")
		}
		if result.Type != "assistant" {
			t.Errorf("Expected Type 'assistant', got '%s'", result.Type)
		}
		if result.Message != nil {
			t.Errorf("Expected Message to be nil, got %v", result.Message)
		}
	})
}
