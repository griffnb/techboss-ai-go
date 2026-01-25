package claude

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

// EmitSSE writes a Server-Sent Event to the response writer and flushes
// The event parameter should be a struct with a "type" field at the top level
func EmitSSE(w http.ResponseWriter, event any) error {
	data, err := json.Marshal(event)
	if err != nil {
		return errors.Wrapf(err, "failed to marshal SSE event")
	}

	fmt.Println("[EmitSSE] data:", string(data))

	_, err = fmt.Fprintf(w, "data: %s\n\n", data)
	if err != nil {
		return errors.Wrapf(err, "failed to write SSE event")
	}

	// Flush immediately for real-time streaming
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}

	return nil
}

// EmitStreamStart emits a stream-start event
// Format: { type: 'stream-start', warnings: [] }
func EmitStreamStart(w http.ResponseWriter, warnings []Warning) error {
	// Ensure warnings is never nil (AI SDK expects array, not null)
	if warnings == nil {
		warnings = []Warning{}
	}

	event := struct {
		Type     string    `json:"type"`
		Warnings []Warning `json:"warnings"` // Always include warnings field
	}{
		Type:     "stream-start",
		Warnings: warnings,
	}
	return EmitSSE(w, event)
}

// EmitTextStart emits a text-start event
// Note: This is an extended event not in base AI SDK spec but required by Vercel AI SDK
func EmitTextStart(w http.ResponseWriter, id string) error {
	event := struct {
		Type string `json:"type"`
		ID   string `json:"id"`
	}{
		Type: "text-start",
		ID:   id,
	}
	return EmitSSE(w, event)
}

// EmitTextDelta emits a text-delta event
// Note: This is an extended event not in base AI SDK spec but required by Vercel AI SDK
func EmitTextDelta(w http.ResponseWriter, id string, delta string) error {
	event := struct {
		Type  string `json:"type"`
		ID    string `json:"id"`
		Delta string `json:"delta"`
	}{
		Type:  "text-delta",
		ID:    id,
		Delta: delta,
	}
	return EmitSSE(w, event)
}

// EmitTextEnd emits a text-end event
// Note: This is an extended event not in base AI SDK spec but required by Vercel AI SDK
func EmitTextEnd(w http.ResponseWriter, id string) error {
	event := struct {
		Type string `json:"type"`
		ID   string `json:"id"`
	}{
		Type: "text-end",
		ID:   id,
	}
	return EmitSSE(w, event)
}

// EmitToolInputStart emits a tool-input-start event (matches claude-code reference)
// Format: { type: 'tool-input-start', id, toolName, providerExecuted, dynamic, providerMetadata }
func EmitToolInputStart(w http.ResponseWriter, id string, toolName string, parentToolCallID *string) error {
	metadata := make(map[string]any)
	claudeCodeMeta := map[string]any{}
	if parentToolCallID != nil {
		claudeCodeMeta["parentToolCallId"] = *parentToolCallID
	} else {
		claudeCodeMeta["parentToolCallId"] = nil
	}
	metadata["claude-code"] = claudeCodeMeta

	event := struct {
		Type             string         `json:"type"`
		ID               string         `json:"id"`
		ToolName         string         `json:"toolName"`
		ProviderExecuted bool           `json:"providerExecuted"`
		Dynamic          bool           `json:"dynamic"`
		ProviderMetadata map[string]any `json:"providerMetadata,omitempty"`
	}{
		Type:             "tool-input-start",
		ID:               id,
		ToolName:         toolName,
		ProviderExecuted: true,
		Dynamic:          true,
		ProviderMetadata: metadata,
	}
	return EmitSSE(w, event)
}

// EmitToolInputDelta emits a tool-input-delta event (matches claude-code reference)
// Format: { type: 'tool-input-delta', id, delta }
func EmitToolInputDelta(w http.ResponseWriter, id string, delta string) error {
	event := struct {
		Type  string `json:"type"`
		ID    string `json:"id"`
		Delta string `json:"delta"`
	}{
		Type:  "tool-input-delta",
		ID:    id,
		Delta: delta,
	}
	return EmitSSE(w, event)
}

// EmitToolInputEnd emits a tool-input-end event (matches claude-code reference)
// Format: { type: 'tool-input-end', id }
func EmitToolInputEnd(w http.ResponseWriter, id string) error {
	event := struct {
		Type string `json:"type"`
		ID   string `json:"id"`
	}{
		Type: "tool-input-end",
		ID:   id,
	}
	return EmitSSE(w, event)
}

// EmitToolCall emits a tool-call event (matches claude-code reference implementation)
// Format: { type: 'tool-call', toolCallId, toolName, input, providerExecuted, dynamic, providerMetadata }
func EmitToolCall(w http.ResponseWriter, toolCallID string, toolName string, input string, parentToolCallID *string) error {
	metadata := make(map[string]any)
	claudeCodeMeta := map[string]any{
		"rawInput": input,
	}
	if parentToolCallID != nil {
		claudeCodeMeta["parentToolCallId"] = *parentToolCallID
	} else {
		claudeCodeMeta["parentToolCallId"] = nil
	}
	metadata["claude-code"] = claudeCodeMeta

	event := struct {
		Type             string         `json:"type"`
		ToolCallID       string         `json:"toolCallId"`
		ToolName         string         `json:"toolName"`
		Input            string         `json:"input"`
		ProviderExecuted bool           `json:"providerExecuted"`
		Dynamic          bool           `json:"dynamic"`
		ProviderMetadata map[string]any `json:"providerMetadata,omitempty"`
	}{
		Type:             "tool-call",
		ToolCallID:       toolCallID,
		ToolName:         toolName,
		Input:            input,
		ProviderExecuted: true,
		Dynamic:          true,
		ProviderMetadata: metadata,
	}
	return EmitSSE(w, event)
}

// EmitToolResult emits a tool-result event
// Format: { type: 'tool-result', toolCallId, toolName, result, ... }
func EmitToolResult(
	w http.ResponseWriter,
	toolCallID string,
	toolName string,
	result any,
	isError bool,
	rawResult string,
	rawResultTruncated bool,
	parentToolCallID *string,
) error {
	metadata := make(map[string]any)
	claudeCodeMeta := map[string]any{
		"rawResult":          rawResult,
		"rawResultTruncated": rawResultTruncated,
	}
	if parentToolCallID != nil {
		claudeCodeMeta["parentToolCallId"] = *parentToolCallID
	} else {
		claudeCodeMeta["parentToolCallId"] = nil
	}
	metadata["claude-code"] = claudeCodeMeta

	event := struct {
		Type             string         `json:"type"`
		ToolCallID       string         `json:"toolCallId"`
		ToolName         string         `json:"toolName"`
		Result           any            `json:"result"`
		IsError          bool           `json:"isError"`
		ProviderExecuted bool           `json:"providerExecuted"`
		Dynamic          bool           `json:"dynamic"`
		ProviderMetadata map[string]any `json:"providerMetadata,omitempty"`
	}{
		Type:             "tool-result",
		ToolCallID:       toolCallID,
		ToolName:         toolName,
		Result:           result,
		IsError:          isError,
		ProviderExecuted: true,
		Dynamic:          true,
		ProviderMetadata: metadata,
	}
	return EmitSSE(w, event)
}

// EmitToolError emits a tool-error event
// Format: { type: 'tool-error', toolCallId, toolName, error, ... }
func EmitToolError(w http.ResponseWriter, toolCallID string, toolName string, errorMsg string, parentToolCallID *string) error {
	metadata := make(map[string]any)
	claudeCodeMeta := map[string]any{
		"rawError": errorMsg,
	}
	if parentToolCallID != nil {
		claudeCodeMeta["parentToolCallId"] = *parentToolCallID
	} else {
		claudeCodeMeta["parentToolCallId"] = nil
	}
	metadata["claude-code"] = claudeCodeMeta

	event := struct {
		Type             string         `json:"type"`
		ToolCallID       string         `json:"toolCallId"`
		ToolName         string         `json:"toolName"`
		Error            string         `json:"error"`
		ProviderExecuted bool           `json:"providerExecuted"`
		Dynamic          bool           `json:"dynamic"`
		ProviderMetadata map[string]any `json:"providerMetadata,omitempty"`
	}{
		Type:             "tool-error",
		ToolCallID:       toolCallID,
		ToolName:         toolName,
		Error:            errorMsg,
		ProviderExecuted: true,
		Dynamic:          true,
		ProviderMetadata: metadata,
	}
	return EmitSSE(w, event)
}

// EmitReasoningStart emits a reasoning-start event
// Note: This is an extended event not in base AI SDK spec
func EmitReasoningStart(w http.ResponseWriter, id string) error {
	event := struct {
		Type string `json:"type"`
		ID   string `json:"id"`
	}{
		Type: "reasoning-start",
		ID:   id,
	}
	return EmitSSE(w, event)
}

// EmitReasoningDelta emits a reasoning-delta event
// Note: This is an extended event not in base AI SDK spec
func EmitReasoningDelta(w http.ResponseWriter, id string, delta string) error {
	event := struct {
		Type  string `json:"type"`
		ID    string `json:"id"`
		Delta string `json:"delta"`
	}{
		Type:  "reasoning-delta",
		ID:    id,
		Delta: delta,
	}
	return EmitSSE(w, event)
}

// EmitReasoningEnd emits a reasoning-end event
// Note: This is an extended event not in base AI SDK spec
func EmitReasoningEnd(w http.ResponseWriter, id string) error {
	event := struct {
		Type string `json:"type"`
		ID   string `json:"id"`
	}{
		Type: "reasoning-end",
		ID:   id,
	}
	return EmitSSE(w, event)
}

// EmitResponseMetadata emits a response-metadata event
// Format: { type: 'response-metadata', sessionId, modelId, costUsd, ... }
func EmitResponseMetadata(w http.ResponseWriter, sessionID *string, modelID string, costUSD float64) error {
	event := struct {
		Type      string  `json:"type"`
		SessionID string  `json:"sessionId,omitempty"`
		ModelID   string  `json:"modelId,omitempty"`
		CostUSD   float64 `json:"costUsd,omitempty"`
	}{
		Type:    "response-metadata",
		ModelID: modelID,
		CostUSD: costUSD,
	}
	if sessionID != nil {
		event.SessionID = *sessionID
	}
	return EmitSSE(w, event)
}

// EmitFinish emits a finish event with usage stats and finish reason
// Format: { type: 'finish', finishReason, usage, metadata, ... }
func EmitFinish(w http.ResponseWriter, finishReason string, usage UsageStats, metadata any) error {
	event := struct {
		Type         string     `json:"type"`
		FinishReason string     `json:"finishReason"`
		Usage        UsageStats `json:"usage"`
		Metadata     any        `json:"metadata,omitempty"`
	}{
		Type:         "finish",
		FinishReason: finishReason,
		Usage:        usage,
		Metadata:     metadata,
	}
	return EmitSSE(w, event)
}

// EmitError emits an error event
// Format: { type: 'error', error: 'message' }
func EmitError(w http.ResponseWriter, errorMsg string) error {
	event := struct {
		Type  string `json:"type"`
		Error string `json:"error"`
	}{
		Type:  "error",
		Error: errorMsg,
	}
	return EmitSSE(w, event)
}
