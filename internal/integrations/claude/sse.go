package claude

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

// EmitSSE writes a Server-Sent Event to the response writer and flushes
func EmitSSE(w http.ResponseWriter, event StreamEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return errors.Wrapf(err, "failed to marshal SSE event")
	}

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
func EmitStreamStart(w http.ResponseWriter, warnings []Warning) error {
	event := StreamEvent{
		Type: "stream-start",
		Data: StreamStartEvent{
			Warnings: warnings,
		},
	}
	return EmitSSE(w, event)
}

// EmitTextStart emits a text-start event
func EmitTextStart(w http.ResponseWriter, id string) error {
	event := StreamEvent{
		Type: "text-start",
		Data: TextStartEvent{
			ID: id,
		},
	}
	return EmitSSE(w, event)
}

// EmitTextDelta emits a text-delta event
func EmitTextDelta(w http.ResponseWriter, id string, delta string) error {
	event := StreamEvent{
		Type: "text-delta",
		Data: TextDeltaEvent{
			ID:    id,
			Delta: delta,
		},
	}
	return EmitSSE(w, event)
}

// EmitTextEnd emits a text-end event
func EmitTextEnd(w http.ResponseWriter, id string) error {
	event := StreamEvent{
		Type: "text-end",
		Data: TextEndEvent{
			ID: id,
		},
	}
	return EmitSSE(w, event)
}

// EmitToolInputStart emits a tool-input-start event
func EmitToolInputStart(w http.ResponseWriter, id string, toolName string, parentToolCallID *string) error {
	metadata := make(map[string]any)
	if parentToolCallID != nil {
		metadata["claude-code"] = map[string]any{
			"parentToolCallId": parentToolCallID,
		}
	}

	event := StreamEvent{
		Type: "tool-input-start",
		Data: ToolInputStartEvent{
			ID:               id,
			ToolName:         toolName,
			ProviderExecuted: true,
			Dynamic:          true,
			ProviderMetadata: metadata,
		},
	}
	return EmitSSE(w, event)
}

// EmitToolInputDelta emits a tool-input-delta event
func EmitToolInputDelta(w http.ResponseWriter, id string, delta string) error {
	event := StreamEvent{
		Type: "tool-input-delta",
		Data: ToolInputDeltaEvent{
			ID:    id,
			Delta: delta,
		},
	}
	return EmitSSE(w, event)
}

// EmitToolInputEnd emits a tool-input-end event
func EmitToolInputEnd(w http.ResponseWriter, id string) error {
	event := StreamEvent{
		Type: "tool-input-end",
		Data: ToolInputEndEvent{
			ID: id,
		},
	}
	return EmitSSE(w, event)
}

// EmitToolCall emits a tool-call event
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

	event := StreamEvent{
		Type: "tool-call",
		Data: ToolCallEvent{
			ToolCallID:       toolCallID,
			ToolName:         toolName,
			Input:            input,
			ProviderExecuted: true,
			Dynamic:          true,
			ProviderMetadata: metadata,
		},
	}
	return EmitSSE(w, event)
}

// EmitToolResult emits a tool-result event
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

	event := StreamEvent{
		Type: "tool-result",
		Data: ToolResultEvent{
			ToolCallID:       toolCallID,
			ToolName:         toolName,
			Result:           result,
			IsError:          isError,
			ProviderExecuted: true,
			Dynamic:          true,
			ProviderMetadata: metadata,
		},
	}
	return EmitSSE(w, event)
}

// EmitToolError emits a tool-error event
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

	event := StreamEvent{
		Type: "tool-error",
		Data: ToolErrorEvent{
			ToolCallID:       toolCallID,
			ToolName:         toolName,
			Error:            errorMsg,
			ProviderExecuted: true,
			Dynamic:          true,
			ProviderMetadata: metadata,
		},
	}
	return EmitSSE(w, event)
}

// EmitReasoningStart emits a reasoning-start event
func EmitReasoningStart(w http.ResponseWriter, id string) error {
	event := StreamEvent{
		Type: "reasoning-start",
		Data: ReasoningStartEvent{
			ID: id,
		},
	}
	return EmitSSE(w, event)
}

// EmitReasoningDelta emits a reasoning-delta event
func EmitReasoningDelta(w http.ResponseWriter, id string, delta string) error {
	event := StreamEvent{
		Type: "reasoning-delta",
		Data: ReasoningDeltaEvent{
			ID:    id,
			Delta: delta,
		},
	}
	return EmitSSE(w, event)
}

// EmitReasoningEnd emits a reasoning-end event
func EmitReasoningEnd(w http.ResponseWriter, id string) error {
	event := StreamEvent{
		Type: "reasoning-end",
		Data: ReasoningEndEvent{
			ID: id,
		},
	}
	return EmitSSE(w, event)
}

// EmitResponseMetadata emits a response-metadata event
func EmitResponseMetadata(w http.ResponseWriter, sessionID *string, modelID string, costUSD float64) error {
	data := ResponseMetadataEvent{
		ModelID: modelID,
		CostUSD: costUSD,
	}
	if sessionID != nil {
		data.SessionID = *sessionID
	}

	event := StreamEvent{
		Type: "response-metadata",
		Data: data,
	}
	return EmitSSE(w, event)
}

// EmitFinish emits a finish event with usage stats and finish reason
func EmitFinish(w http.ResponseWriter, finishReason string, usage UsageStats, metadata any) error {
	event := StreamEvent{
		Type: "finish",
		Data: FinishEvent{
			FinishReason: finishReason,
			Usage:        usage,
			Metadata:     metadata,
		},
	}
	return EmitSSE(w, event)
}

// EmitError emits an error event
func EmitError(w http.ResponseWriter, errorMsg string) error {
	event := StreamEvent{
		Type: "error",
		Data: map[string]string{
			"error": errorMsg,
		},
	}
	return EmitSSE(w, event)
}
