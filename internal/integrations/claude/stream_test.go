package claude

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/pkg/errors"
)

// mockResponseWriter wraps httptest.ResponseRecorder for SSE testing
type mockResponseWriter struct {
	*httptest.ResponseRecorder
	events []map[string]any
}

func newMockResponseWriter() *mockResponseWriter {
	return &mockResponseWriter{
		ResponseRecorder: httptest.NewRecorder(),
		events:           []map[string]any{},
	}
}

func (m *mockResponseWriter) Write(p []byte) (n int, err error) {
	n, err = m.ResponseRecorder.Write(p)
	if err != nil {
		return n, errors.Wrapf(err, "failed to write to recorder")
	}

	// Parse SSE events from the written data
	body := string(p)
	lines := strings.Split(body, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "data: ") {
			jsonData := strings.TrimPrefix(line, "data: ")
			var event map[string]any
			if err := json.Unmarshal([]byte(jsonData), &event); err == nil {
				m.events = append(m.events, event)
			}
		}
	}

	return n, nil
}

// getEventsByType returns all events of a specific type
func (m *mockResponseWriter) getEventsByType(eventType string) []map[string]any {
	var results []map[string]any
	for _, event := range m.events {
		if t, ok := event["type"].(string); ok && t == eventType {
			results = append(results, event)
		}
	}
	return results
}

// Test batched tool calls without stream_event messages
// This reproduces the real Claude CLI behavior from the user's sample
func Test_ProcessStream_batchedToolCalls(t *testing.T) {
	t.Run("emits tool-input-end before tool-call for batched messages", func(t *testing.T) {
		// Arrange - Simulate Claude CLI batched output (no stream_event messages)
		toolID := "toolu_bdrk_01HFQFe3dQH6MWbJUF4unRpk"
		input := `{"command":"pwd","description":"Show current working directory"}`

		claudeOutput := strings.Join([]string{
			// Assistant message with tool_use
			`{"type":"assistant","message":{"content":[{"type":"tool_use","id":"` + toolID + `","name":"Bash","input":` + input + `}]}}`,
			// User message with tool_result
			`{"type":"user","message":{"content":[{"tool_use_id":"` + toolID + `","type":"tool_result","content":"/mnt/workspace","is_error":false}]}}`,
			// Result message
			`{"type":"result","subtype":"success","usage":{"input_tokens":10,"output_tokens":5}}`,
		}, "\n")

		reader := bytes.NewBufferString(claudeOutput)
		writer := newMockResponseWriter()

		// Act
		err := ProcessStream(context.Background(), reader, writer, nil)
		// Assert
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		// Verify event sequence
		var eventTypes []string
		for _, event := range writer.events {
			if eventType, ok := event["type"].(string); ok {
				eventTypes = append(eventTypes, eventType)
			}
		}

		// Expected: stream-start, tool-input-start, tool-input-delta, tool-input-end, tool-call, tool-result, finish
		expectedSequence := []string{
			"stream-start",
			"tool-input-start",
			"tool-input-delta",
			"tool-input-end", // THIS IS THE KEY FIX
			"tool-call",
			"tool-result",
			"finish",
		}

		if len(eventTypes) != len(expectedSequence) {
			t.Errorf("Expected %d events, got %d\nGot: %v\nExpected: %v",
				len(expectedSequence), len(eventTypes), eventTypes, expectedSequence)
		}

		for i, expected := range expectedSequence {
			if i >= len(eventTypes) {
				t.Errorf("Missing event at index %d: expected '%s'", i, expected)
				continue
			}
			if eventTypes[i] != expected {
				t.Errorf("Event at index %d: expected '%s', got '%s'", i, expected, eventTypes[i])
			}
		}

		// Verify tool-input-end comes before tool-call
		toolInputEndIndex := -1
		toolCallIndex := -1
		for i, eventType := range eventTypes {
			if eventType == "tool-input-end" {
				toolInputEndIndex = i
			}
			if eventType == "tool-call" {
				toolCallIndex = i
			}
		}

		if toolInputEndIndex == -1 {
			t.Error("tool-input-end event not found")
		}
		if toolCallIndex == -1 {
			t.Error("tool-call event not found")
		}
		if toolInputEndIndex > -1 && toolCallIndex > -1 && toolInputEndIndex >= toolCallIndex {
			t.Errorf("tool-input-end (index %d) must come before tool-call (index %d)",
				toolInputEndIndex, toolCallIndex)
		}
	})

	t.Run("handles multiple parallel tool calls with correct event sequence", func(t *testing.T) {
		// Arrange - Multiple tool calls in parallel (like the user's sample)
		toolID1 := "toolu_bdrk_01HFQFe3dQH6MWbJUF4unRpk"
		toolID2 := "toolu_bdrk_01GDJupxgyY8czwAdEyChdTL"
		toolID3 := "toolu_bdrk_01WLES7mD5DGQFM65Ht6PoSX"

		claudeOutput := strings.Join([]string{
			// Assistant messages with tool_use (one per line, like real Claude CLI)
			`{"type":"assistant","message":{"content":[{"type":"tool_use","id":"` + toolID1 + `","name":"Bash","input":{"command":"pwd"}}]}}`,
			`{"type":"assistant","message":{"content":[{"type":"tool_use","id":"` + toolID2 + `","name":"Bash","input":{"command":"ls -la"}}]}}`,
			`{"type":"assistant","message":{"content":[{"type":"tool_use","id":"` + toolID3 + `","name":"Bash","input":{"command":"ls -la /mnt/workspace"}}]}}`,
			// User messages with tool_result (one per line)
			`{"type":"user","message":{"content":[{"tool_use_id":"` + toolID1 + `","type":"tool_result","content":"/mnt/workspace","is_error":false}]}}`,
			`{"type":"user","message":{"content":[{"tool_use_id":"` + toolID2 + `","type":"tool_result","content":"total 1\ndrwxr-xr-x...","is_error":false}]}}`,
			`{"type":"user","message":{"content":[{"tool_use_id":"` + toolID3 + `","type":"tool_result","content":"lrwxrwxrwx...","is_error":false}]}}`,
			// Result message
			`{"type":"result","subtype":"success","usage":{"input_tokens":10,"output_tokens":5}}`,
		}, "\n")

		reader := bytes.NewBufferString(claudeOutput)
		writer := newMockResponseWriter()

		// Act
		err := ProcessStream(context.Background(), reader, writer, nil)
		// Assert
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		// Verify we have tool-input-end for each tool
		toolInputEndEvents := writer.getEventsByType("tool-input-end")
		if len(toolInputEndEvents) != 3 {
			t.Errorf("Expected 3 tool-input-end events, got %d", len(toolInputEndEvents))
		}

		// Verify we have tool-call for each tool
		toolCallEvents := writer.getEventsByType("tool-call")
		if len(toolCallEvents) != 3 {
			t.Errorf("Expected 3 tool-call events, got %d", len(toolCallEvents))
		}

		// Verify tool-result for each tool
		toolResultEvents := writer.getEventsByType("tool-result")
		if len(toolResultEvents) != 3 {
			t.Errorf("Expected 3 tool-result events, got %d", len(toolResultEvents))
		}

		// Verify each tool has complete lifecycle: input-start, input-delta, input-end, call, result
		for _, toolID := range []string{toolID1, toolID2, toolID3} {
			var foundInputStart, foundInputDelta, foundInputEnd, foundCall, foundResult bool

			for _, event := range writer.events {
				eventType, _ := event["type"].(string)
				eventID, hasID := event["id"].(string)
				eventToolCallID, hasToolCallID := event["toolCallId"].(string)

				if hasID && eventID == toolID {
					switch eventType {
					case "tool-input-start":
						foundInputStart = true
					case "tool-input-delta":
						foundInputDelta = true
					case "tool-input-end":
						foundInputEnd = true
					}
				}

				if hasToolCallID && eventToolCallID == toolID {
					switch eventType {
					case "tool-call":
						foundCall = true
					case "tool-result":
						foundResult = true
					}
				}
			}

			if !foundInputStart {
				t.Errorf("Missing tool-input-start for tool %s", toolID)
			}
			if !foundInputDelta {
				t.Errorf("Missing tool-input-delta for tool %s", toolID)
			}
			if !foundInputEnd {
				t.Errorf("Missing tool-input-end for tool %s", toolID)
			}
			if !foundCall {
				t.Errorf("Missing tool-call for tool %s", toolID)
			}
			if !foundResult {
				t.Errorf("Missing tool-result for tool %s", toolID)
			}
		}
	})
}

// Test text content emission in assistant messages
func Test_ProcessStream_textContent(t *testing.T) {
	t.Run("emits text events for text content in assistant messages", func(t *testing.T) {
		// Arrange
		textContent := "I'll search the current directory to see what's available."

		claudeOutput := strings.Join([]string{
			// Assistant message with text content (before tool calls)
			`{"type":"assistant","message":{"content":[{"type":"text","text":"` + textContent + `"}]}}`,
			// Result message
			`{"type":"result","subtype":"success","usage":{"input_tokens":10,"output_tokens":5}}`,
		}, "\n")

		reader := bytes.NewBufferString(claudeOutput)
		writer := newMockResponseWriter()

		// Act
		err := ProcessStream(context.Background(), reader, writer, nil)
		// Assert
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		// Verify text events were emitted
		textStartEvents := writer.getEventsByType("text-start")
		if len(textStartEvents) != 1 {
			t.Errorf("Expected 1 text-start event, got %d", len(textStartEvents))
		}

		textDeltaEvents := writer.getEventsByType("text-delta")
		if len(textDeltaEvents) != 1 {
			t.Errorf("Expected 1 text-delta event, got %d", len(textDeltaEvents))
		}

		textEndEvents := writer.getEventsByType("text-end")
		if len(textEndEvents) != 1 {
			t.Errorf("Expected 1 text-end event, got %d", len(textEndEvents))
		}

		// Verify text content matches
		if len(textDeltaEvents) > 0 {
			delta, ok := textDeltaEvents[0]["delta"].(string)
			if !ok {
				t.Error("text-delta event missing 'delta' field")
			} else if delta != textContent {
				t.Errorf("Expected delta '%s', got '%s'", textContent, delta)
			}
		}

		// Verify event sequence: stream-start, text-start, text-delta, text-end, finish
		var eventTypes []string
		for _, event := range writer.events {
			if eventType, ok := event["type"].(string); ok {
				eventTypes = append(eventTypes, eventType)
			}
		}

		expectedSequence := []string{
			"stream-start",
			"text-start",
			"text-delta",
			"text-end",
			"finish",
		}

		if len(eventTypes) != len(expectedSequence) {
			t.Errorf("Expected %d events, got %d\nGot: %v\nExpected: %v",
				len(expectedSequence), len(eventTypes), eventTypes, expectedSequence)
		}
	})

	t.Run("emits text before tool calls in same message", func(t *testing.T) {
		// Arrange - Text followed by tool in same assistant message
		textContent := "Let me check that for you."
		toolID := "toolu_123"

		claudeOutput := strings.Join([]string{
			// Assistant message with both text and tool_use
			`{"type":"assistant","message":{"content":[` +
				`{"type":"text","text":"` + textContent + `"},` +
				`{"type":"tool_use","id":"` + toolID + `","name":"Bash","input":{"command":"pwd"}}` +
				`]}}`,
			// User message with tool_result
			`{"type":"user","message":{"content":[{"tool_use_id":"` + toolID + `","type":"tool_result","content":"/home/user","is_error":false}]}}`,
			// Result message
			`{"type":"result","subtype":"success","usage":{"input_tokens":10,"output_tokens":5}}`,
		}, "\n")

		reader := bytes.NewBufferString(claudeOutput)
		writer := newMockResponseWriter()

		// Act
		err := ProcessStream(context.Background(), reader, writer, nil)
		// Assert
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		// Verify event sequence
		var eventTypes []string
		for _, event := range writer.events {
			if eventType, ok := event["type"].(string); ok {
				eventTypes = append(eventTypes, eventType)
			}
		}

		// Expected: stream-start, text-start, text-delta, text-end, tool-input-start, tool-input-delta, tool-input-end, tool-call, tool-result, finish
		expectedSequence := []string{
			"stream-start",
			"text-start",
			"text-delta",
			"text-end",
			"tool-input-start",
			"tool-input-delta",
			"tool-input-end",
			"tool-call",
			"tool-result",
			"finish",
		}

		if len(eventTypes) != len(expectedSequence) {
			t.Errorf("Expected %d events, got %d\nGot: %v\nExpected: %v",
				len(expectedSequence), len(eventTypes), eventTypes, expectedSequence)
		}

		// Verify text events come before tool events
		textEndIndex := -1
		toolInputStartIndex := -1
		for i, eventType := range eventTypes {
			if eventType == "text-end" {
				textEndIndex = i
			}
			if eventType == "tool-input-start" {
				toolInputStartIndex = i
			}
		}

		if textEndIndex > -1 && toolInputStartIndex > -1 && textEndIndex >= toolInputStartIndex {
			t.Errorf("text-end (index %d) must come before tool-input-start (index %d)",
				textEndIndex, toolInputStartIndex)
		}
	})
}

// Test edge cases
func Test_ProcessStream_edgeCases(t *testing.T) {
	t.Run("handles empty input gracefully", func(t *testing.T) {
		// Arrange
		reader := bytes.NewBufferString("")
		writer := newMockResponseWriter()

		// Act
		err := ProcessStream(context.Background(), reader, writer, nil)
		// Assert
		if err != nil {
			t.Errorf("Expected no error for empty input, got %v", err)
		}

		// Should emit at least stream-start
		events := writer.getEventsByType("stream-start")
		if len(events) != 1 {
			t.Errorf("Expected 1 stream-start event, got %d", len(events))
		}
	})

	t.Run("skips unparseable lines", func(t *testing.T) {
		// Arrange
		claudeOutput := strings.Join([]string{
			"Some random output",
			`{"type":"assistant","message":{"content":[{"type":"text","text":"Hello"}]}}`,
			"Another random line",
			`{"type":"result","subtype":"success","usage":{"input_tokens":5,"output_tokens":3}}`,
		}, "\n")

		reader := bytes.NewBufferString(claudeOutput)
		writer := newMockResponseWriter()

		// Act
		err := ProcessStream(context.Background(), reader, writer, nil)
		// Assert - Should not error, should skip non-JSON lines
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		// Should still process valid JSON lines
		finishEvents := writer.getEventsByType("finish")
		if len(finishEvents) != 1 {
			t.Errorf("Expected 1 finish event, got %d", len(finishEvents))
		}
	})

	t.Run("handles context cancellation", func(t *testing.T) {
		// Arrange
		claudeOutput := strings.Join([]string{
			`{"type":"assistant","message":{"content":[{"type":"text","text":"Hello"}]}}`,
			`{"type":"result","subtype":"success","usage":{"input_tokens":5,"output_tokens":3}}`,
		}, "\n")

		reader := bytes.NewBufferString(claudeOutput)
		writer := newMockResponseWriter()
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		// Act
		err := ProcessStream(ctx, reader, writer, nil)

		// Assert - Should return context error
		if err == nil {
			t.Error("Expected error for cancelled context, got nil")
		}
	})
}

// Test token callback
func Test_ProcessStream_tokenCallback(t *testing.T) {
	t.Run("calls token callback with usage stats", func(t *testing.T) {
		// Arrange
		claudeOutput := strings.Join([]string{
			`{"type":"result","subtype":"success","usage":{"input_tokens":100,"output_tokens":50,"cache_read_input_tokens":25}}`,
		}, "\n")

		reader := bytes.NewBufferString(claudeOutput)
		writer := newMockResponseWriter()

		var callbackCalled bool
		var receivedInput, receivedOutput, receivedCache int64

		callback := func(inputTokens, outputTokens, cacheTokens int64) {
			callbackCalled = true
			receivedInput = inputTokens
			receivedOutput = outputTokens
			receivedCache = cacheTokens
		}

		// Act
		err := ProcessStream(context.Background(), reader, writer, callback)
		// Assert
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if !callbackCalled {
			t.Error("Expected token callback to be called")
		}

		if receivedInput != 100 {
			t.Errorf("Expected input tokens 100, got %d", receivedInput)
		}

		if receivedOutput != 50 {
			t.Errorf("Expected output tokens 50, got %d", receivedOutput)
		}

		if receivedCache != 25 {
			t.Errorf("Expected cache tokens 25, got %d", receivedCache)
		}
	})
}
