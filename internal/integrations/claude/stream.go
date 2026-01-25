package claude

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

// MapFinishReason maps Claude finish reason to AI SDK format
func MapFinishReason(subtype *string) string {
	if subtype == nil {
		return "stop"
	}

	switch *subtype {
	case "success":
		return "stop"
	case "max_tokens":
		return "length"
	case "error":
		return "error"
	case "cancelled":
		return "stop"
	default:
		return "stop"
	}
}

// TokenUpdateCallback is called when token usage is parsed from the result message
type TokenUpdateCallback func(inputTokens, outputTokens, cacheTokens int64)

// ProcessStream processes Claude stdout and emits structured SSE events
func ProcessStream(
	ctx context.Context,
	stdout io.Reader,
	writer http.ResponseWriter,
	tokenCallback TokenUpdateCallback,
) error {
	parser := NewStreamParser()
	scanner := bufio.NewScanner(stdout)

	// Track state
	var textPartID *string
	var currentReasoningPartID *string
	reasoningBlocksByIndex := make(map[int]string)
	var accumulatedText strings.Builder
	streamedTextLength := 0

	// Emit stream-start
	err := EmitStreamStart(writer, []Warning{})
	if err != nil {
		return errors.Wrapf(err, "failed to emit stream-start")
	}

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return errors.Wrapf(ctx.Err(), "stream cancelled")
		default:
		}

		line := scanner.Text()
		if line == "" {
			continue
		}

		fmt.Println("[Claude Stream Line] ", line)

		// Parse Claude event
		message, err := ParseClaudeEvent(line)
		if err != nil {
			// Skip unparseable events (may be informational output)
			continue
		}

		// Route by message type
		switch message.Type {
		case "stream_event":
			err = handleStreamEvent(ctx, message, parser, writer, &textPartID, &currentReasoningPartID, reasoningBlocksByIndex)
			if err != nil {
				return err
			}

		case "assistant":
			err = handleAssistantMessage(ctx, message, parser, writer, &textPartID, &accumulatedText, &streamedTextLength)
			if err != nil {
				return err
			}

		case "user":
			err = handleUserMessage(ctx, message, parser, writer)
			if err != nil {
				return err
			}

		case "result":
			// Close any open text part
			if textPartID != nil {
				err = EmitTextEnd(writer, *textPartID)
				if err != nil {
					return errors.Wrapf(err, "failed to emit text-end")
				}
				textPartID = nil
			}

			// Finalize pending tool calls
			parser.FinalizeToolCalls()

			// Extract usage and finish reason
			usage := UsageStats{}
			if message.Usage != nil {
				usage = ConvertClaudeCodeUsage(message.Usage)
			}

			finishReason := MapFinishReason(message.Subtype)

			// Build metadata
			metadata := make(map[string]any)
			if message.SessionID != nil {
				metadata["sessionId"] = *message.SessionID
			}
			if message.TotalCostUSD != nil {
				metadata["costUsd"] = *message.TotalCostUSD
			}
			if message.DurationMS != nil {
				metadata["durationMs"] = *message.DurationMS
			}

			// Emit finish event
			err = EmitFinish(writer, finishReason, usage, metadata)
			if err != nil {
				return errors.Wrapf(err, "failed to emit finish")
			}

			// Update token counts via callback
			if tokenCallback != nil && message.Usage != nil {
				inputTokens := int64(0)
				outputTokens := int64(0)
				cacheTokens := int64(0)

				if message.Usage.InputTokens != nil {
					inputTokens = *message.Usage.InputTokens
				}
				if message.Usage.OutputTokens != nil {
					outputTokens = *message.Usage.OutputTokens
				}
				if message.Usage.CacheReadTokens != nil {
					cacheTokens = *message.Usage.CacheReadTokens
				}

				tokenCallback(inputTokens, outputTokens, cacheTokens)
			}

			return nil
		}
	}

	if err := scanner.Err(); err != nil {
		return errors.Wrapf(err, "error reading Claude output")
	}

	return nil
}

func handleStreamEvent(
	_ context.Context,
	message *ClaudeMessage,
	parser *StreamParser,
	writer http.ResponseWriter,
	textPartID **string,
	currentReasoningPartID **string,
	reasoningBlocksByIndex map[int]string,
) error {
	if message.Event == nil {
		return nil
	}

	event := message.Event

	switch event.Type {
	case "content_block_start":
		return handleContentBlockStart(event, parser, writer, textPartID, currentReasoningPartID, reasoningBlocksByIndex)

	case "content_block_delta":
		return handleContentBlockDelta(event, parser, writer, *textPartID, *currentReasoningPartID, reasoningBlocksByIndex)

	case "content_block_stop":
		return handleContentBlockStop(event, parser, writer, textPartID, currentReasoningPartID, reasoningBlocksByIndex)
	}

	return nil
}

func handleContentBlockStart(
	event *StreamEventDetails,
	parser *StreamParser,
	writer http.ResponseWriter,
	textPartID **string,
	currentReasoningPartID **string,
	reasoningBlocksByIndex map[int]string,
) error {
	if event.ContentBlock == nil {
		return nil
	}

	block := event.ContentBlock
	index := 0
	if event.Index != nil {
		index = *event.Index
	}

	switch block.Type {
	case "tool_use":
		// Close any active text part before tool starts
		if *textPartID != nil {
			err := EmitTextEnd(writer, **textPartID)
			if err != nil {
				return errors.Wrapf(err, "failed to emit text-end")
			}
			*textPartID = nil
		}

		// Extract tool ID and name
		toolID := uuid.New().String()
		if block.ID != nil && *block.ID != "" {
			toolID = *block.ID
		}

		toolName := UNKNOWN_TOOL_NAME
		if block.Name != nil && *block.Name != "" {
			toolName = *block.Name
		}

		// Determine parent (Task tools never have parent)
		var parentToolCallID *string
		if toolName != "Task" {
			parentToolCallID = parser.GetFallbackParentID()
		}

		// Create tool state
		parser.toolStates[toolID] = &ToolStreamState{
			Name:             toolName,
			InputStarted:     true,
			InputClosed:      false,
			CallEmitted:      false,
			ParentToolCallID: parentToolCallID,
		}

		// Track in maps
		parser.toolBlocksByIndex[index] = toolID
		parser.toolInputAccumulators[toolID] = ""

		// Track Task tools
		if toolName == "Task" {
			parser.TrackTaskTool(toolID, true)
		}

		// Emit tool-input-start
		err := EmitToolInputStart(writer, toolID, toolName, parentToolCallID)
		if err != nil {
			return errors.Wrapf(err, "failed to emit tool-input-start")
		}

	case "text":
		// Generate text part ID
		partID := uuid.New().String()
		parser.textBlocksByIndex[index] = partID
		*textPartID = &partID

		// Emit text-start
		err := EmitTextStart(writer, partID)
		if err != nil {
			return errors.Wrapf(err, "failed to emit text-start")
		}

	case "thinking":
		// Close any active text part before reasoning starts
		if *textPartID != nil {
			err := EmitTextEnd(writer, **textPartID)
			if err != nil {
				return errors.Wrapf(err, "failed to emit text-end")
			}
			*textPartID = nil
		}

		// Generate reasoning part ID
		reasoningPartID := uuid.New().String()
		reasoningBlocksByIndex[index] = reasoningPartID
		*currentReasoningPartID = &reasoningPartID

		// Emit reasoning-start
		err := EmitReasoningStart(writer, reasoningPartID)
		if err != nil {
			return errors.Wrapf(err, "failed to emit reasoning-start")
		}
	}

	return nil
}

func handleContentBlockDelta(
	event *StreamEventDetails,
	parser *StreamParser,
	writer http.ResponseWriter,
	textPartID *string,
	currentReasoningPartID *string,
	reasoningBlocksByIndex map[int]string,
) error {
	if event.Delta == nil {
		return nil
	}

	delta := event.Delta
	index := 0
	if event.Index != nil {
		index = *event.Index
	}

	switch delta.Type {
	case "text_delta":
		if delta.Text != nil && textPartID != nil {
			err := EmitTextDelta(writer, *textPartID, *delta.Text)
			if err != nil {
				return errors.Wrapf(err, "failed to emit text-delta")
			}
		}

	case "input_json_delta":
		if delta.PartialJSON != nil {
			// Route to tool-input-delta if we have a tracked tool
			if toolID, ok := parser.toolBlocksByIndex[index]; ok {
				// Accumulate input
				current := parser.toolInputAccumulators[toolID]
				parser.toolInputAccumulators[toolID] = current + *delta.PartialJSON

				// Emit delta
				err := EmitToolInputDelta(writer, toolID, *delta.PartialJSON)
				if err != nil {
					return errors.Wrapf(err, "failed to emit tool-input-delta")
				}
			}
		}

	case "thinking_delta":
		if delta.Thinking != nil {
			// Find reasoning part ID
			var reasoningPartID *string
			if id, ok := reasoningBlocksByIndex[index]; ok {
				reasoningPartID = &id
			} else if currentReasoningPartID != nil {
				reasoningPartID = currentReasoningPartID
			}

			if reasoningPartID != nil {
				err := EmitReasoningDelta(writer, *reasoningPartID, *delta.Thinking)
				if err != nil {
					return errors.Wrapf(err, "failed to emit reasoning-delta")
				}
			}
		}
	}

	return nil
}

func handleContentBlockStop(
	event *StreamEventDetails,
	parser *StreamParser,
	writer http.ResponseWriter,
	textPartID **string,
	currentReasoningPartID **string,
	reasoningBlocksByIndex map[int]string,
) error {
	index := 0
	if event.Index != nil {
		index = *event.Index
	}

	// Check if this is a tool block
	if toolID, ok := parser.toolBlocksByIndex[index]; ok {
		state := parser.toolStates[toolID]
		if state != nil {
			// Get accumulated input
			accumulatedInput := parser.toolInputAccumulators[toolID]
			state.LastSerializedInput = &accumulatedInput

			// Emit tool-input-end
			err := EmitToolInputEnd(writer, toolID)
			if err != nil {
				return errors.Wrapf(err, "failed to emit tool-input-end")
			}
			state.InputClosed = true

			// Emit tool-call immediately
			err = EmitToolCall(writer, toolID, state.Name, accumulatedInput, state.ParentToolCallID)
			if err != nil {
				return errors.Wrapf(err, "failed to emit tool-call")
			}
			state.CallEmitted = true
		}

		// Clean up
		delete(parser.toolBlocksByIndex, index)
		delete(parser.toolInputAccumulators, toolID)
		return nil
	}

	// Check if this is a text block
	if textID, ok := parser.textBlocksByIndex[index]; ok {
		err := EmitTextEnd(writer, textID)
		if err != nil {
			return errors.Wrapf(err, "failed to emit text-end")
		}

		delete(parser.textBlocksByIndex, index)
		if *textPartID != nil && **textPartID == textID {
			*textPartID = nil
		}
		return nil
	}

	// Check if this is a reasoning block
	if reasoningPartID, ok := reasoningBlocksByIndex[index]; ok {
		err := EmitReasoningEnd(writer, reasoningPartID)
		if err != nil {
			return errors.Wrapf(err, "failed to emit reasoning-end")
		}

		delete(reasoningBlocksByIndex, index)
		if *currentReasoningPartID != nil && **currentReasoningPartID == reasoningPartID {
			*currentReasoningPartID = nil
		}
		return nil
	}

	return nil
}

func handleAssistantMessage(
	_ context.Context,
	message *ClaudeMessage,
	parser *StreamParser,
	writer http.ResponseWriter,
	textPartID **string,
	accumulatedText *strings.Builder,
	_ *int,
) error {
	if message.Message == nil || len(message.Message.Content) == 0 {
		return nil
	}

	content := message.Message.Content

	// Extract SDK parent tool ID
	sdkParentToolUseID := message.ParentToolUseID

	// Extract tools
	tools := ExtractToolUses(content)

	// Process text blocks first and emit text events
	for _, block := range content {
		if block.Type == "text" && block.Text != nil {
			// Start text part if not already started
			if *textPartID == nil {
				partID := uuid.New().String()
				*textPartID = &partID
				err := EmitTextStart(writer, partID)
				if err != nil {
					return errors.Wrapf(err, "failed to emit text-start")
				}
			}

			// Emit text delta
			err := EmitTextDelta(writer, **textPartID, *block.Text)
			if err != nil {
				return errors.Wrapf(err, "failed to emit text-delta")
			}

			// Accumulate text (WriteString on strings.Builder never errors in practice)
			_, _ = accumulatedText.WriteString(*block.Text)
		}
	}

	// Close any active text part before tool calls start
	if *textPartID != nil && len(tools) > 0 {
		err := EmitTextEnd(writer, **textPartID)
		if err != nil {
			return errors.Wrapf(err, "failed to emit text-end")
		}
		*textPartID = nil
	}

	// Process tool uses
	for _, tool := range tools {
		toolID := tool.ID

		// Get or create state
		state, exists := parser.toolStates[toolID]
		if !exists {
			// Determine parent
			var parentToolCallID *string
			if tool.Name == "Task" {
				parentToolCallID = nil
			} else if sdkParentToolUseID != nil {
				parentToolCallID = sdkParentToolUseID
			} else if tool.ParentToolUseID != nil {
				parentToolCallID = tool.ParentToolUseID
			} else {
				parentToolCallID = parser.GetFallbackParentID()
			}

			state = &ToolStreamState{
				Name:             tool.Name,
				InputStarted:     false,
				InputClosed:      false,
				CallEmitted:      false,
				ParentToolCallID: parentToolCallID,
			}
			parser.toolStates[toolID] = state

			// Track Task tools
			if tool.Name == "Task" {
				parser.TrackTaskTool(toolID, true)
			}
		} else if !state.CallEmitted && sdkParentToolUseID != nil && tool.Name != "Task" {
			// Retroactive parent context from SDK
			state.ParentToolCallID = sdkParentToolUseID
		}

		// Start input if not started
		if !state.InputStarted {
			err := EmitToolInputStart(writer, toolID, tool.Name, state.ParentToolCallID)
			if err != nil {
				return errors.Wrapf(err, "failed to emit tool-input-start")
			}
			state.InputStarted = true
		}

		// Serialize input
		serializedInput := serializeToolInput(tool.Input)
		if serializedInput != "" {
			// Emit delta if needed
			if state.LastSerializedInput == nil {
				// First input - emit full delta if reasonable size
				if len(serializedInput) <= 10000 {
					err := EmitToolInputDelta(writer, toolID, serializedInput)
					if err != nil {
						return errors.Wrapf(err, "failed to emit tool-input-delta")
					}
				}
			} else if len(serializedInput) <= 10000 && len(*state.LastSerializedInput) <= 10000 {
				// Calculate delta
				if strings.HasPrefix(serializedInput, *state.LastSerializedInput) {
					deltaStr := serializedInput[len(*state.LastSerializedInput):]
					if deltaStr != "" {
						err := EmitToolInputDelta(writer, toolID, deltaStr)
						if err != nil {
							return errors.Wrapf(err, "failed to emit tool-input-delta")
						}
					}
				}
			}

			state.LastSerializedInput = &serializedInput
		}
	}

	return nil
}

func handleUserMessage(
	_ context.Context,
	message *ClaudeMessage,
	parser *StreamParser,
	writer http.ResponseWriter,
) error {
	if message.Message == nil || len(message.Message.Content) == 0 {
		return nil
	}

	content := message.Message.Content
	sdkParentToolUseID := message.ParentToolUseID

	// Extract tool results
	results := ExtractToolResults(content)
	for _, result := range results {
		state := parser.toolStates[result.ID]
		toolName := UNKNOWN_TOOL_NAME
		if result.Name != nil {
			toolName = *result.Name
		} else if state != nil {
			toolName = state.Name
		}

		// Create state if missing
		if state == nil {
			var parentToolCallID *string
			if toolName == "Task" {
				parentToolCallID = nil
			} else if sdkParentToolUseID != nil {
				parentToolCallID = sdkParentToolUseID
			} else {
				parentToolCallID = parser.GetFallbackParentID()
			}

			state = &ToolStreamState{
				Name:             toolName,
				InputStarted:     true,
				InputClosed:      true,
				CallEmitted:      false,
				ParentToolCallID: parentToolCallID,
			}
			parser.toolStates[result.ID] = state
		}

		// Ensure tool-call is emitted
		if !state.CallEmitted {
			// If input wasn't closed yet, close it now
			if !state.InputClosed {
				err := EmitToolInputEnd(writer, result.ID)
				if err != nil {
					return errors.Wrapf(err, "failed to emit tool-input-end")
				}
				state.InputClosed = true
			}

			input := ""
			if state.LastSerializedInput != nil {
				input = *state.LastSerializedInput
			}
			err := EmitToolCall(writer, result.ID, toolName, input, state.ParentToolCallID)
			if err != nil {
				return errors.Wrapf(err, "failed to emit tool-call")
			}
			state.CallEmitted = true
		}

		// Remove Task tools from active set
		if toolName == "Task" {
			parser.TrackTaskTool(result.ID, false)
		}

		// Serialize result
		rawResult := serializeToolInput(result.Result)

		// Emit tool-result
		err := EmitToolResult(writer, result.ID, toolName, result.Result, result.IsError, rawResult, false, state.ParentToolCallID)
		if err != nil {
			return errors.Wrapf(err, "failed to emit tool-result")
		}
	}

	// Extract tool errors
	toolErrors := ExtractToolErrors(content)
	for _, toolError := range toolErrors {
		state := parser.toolStates[toolError.ID]
		toolName := UNKNOWN_TOOL_NAME
		if toolError.Name != nil {
			toolName = *toolError.Name
		} else if state != nil {
			toolName = state.Name
		}

		// Create state if missing
		if state == nil {
			var parentToolCallID *string
			if toolName == "Task" {
				parentToolCallID = nil
			} else if sdkParentToolUseID != nil {
				parentToolCallID = sdkParentToolUseID
			} else {
				parentToolCallID = parser.GetFallbackParentID()
			}

			state = &ToolStreamState{
				Name:             toolName,
				InputStarted:     true,
				InputClosed:      true,
				CallEmitted:      false,
				ParentToolCallID: parentToolCallID,
			}
			parser.toolStates[toolError.ID] = state
		}

		// Ensure tool-call is emitted
		if !state.CallEmitted {
			// If input wasn't closed yet, close it now
			if !state.InputClosed {
				err := EmitToolInputEnd(writer, toolError.ID)
				if err != nil {
					return errors.Wrapf(err, "failed to emit tool-input-end")
				}
				state.InputClosed = true
			}

			input := ""
			if state.LastSerializedInput != nil {
				input = *state.LastSerializedInput
			}
			err := EmitToolCall(writer, toolError.ID, toolName, input, state.ParentToolCallID)
			if err != nil {
				return errors.Wrapf(err, "failed to emit tool-call")
			}
			state.CallEmitted = true
		}

		// Remove Task tools from active set
		if toolName == "Task" {
			parser.TrackTaskTool(toolError.ID, false)
		}

		// Serialize error
		rawError := serializeToolInput(toolError.Error)

		// Emit tool-error
		err := EmitToolError(writer, toolError.ID, toolName, rawError, state.ParentToolCallID)
		if err != nil {
			return errors.Wrapf(err, "failed to emit tool-error")
		}
	}

	return nil
}

func serializeToolInput(input any) string {
	if input == nil {
		return ""
	}

	if str, ok := input.(string); ok {
		return str
	}

	// Try to JSON encode
	bytes, err := json.Marshal(input)
	if err != nil {
		return ""
	}

	return string(bytes)
}
