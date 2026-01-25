package claude

// StreamEvent represents any event in the Claude stream
type StreamEvent struct {
	Type string `json:"type"`
	Data any    `json:"data,omitempty"`
}

// StreamStartEvent is emitted first with warnings
type StreamStartEvent struct {
	Warnings []Warning `json:"warnings,omitempty"`
}

// Warning represents a warning message in the stream
type Warning struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// TextStartEvent is emitted when text content begins
type TextStartEvent struct {
	ID string `json:"id"`
}

// TextDeltaEvent is emitted for each text chunk
type TextDeltaEvent struct {
	ID    string `json:"id"`
	Delta string `json:"delta"`
}

// TextEndEvent is emitted when text content completes
type TextEndEvent struct {
	ID string `json:"id"`
}

// ToolInputStartEvent is emitted when tool input begins
type ToolInputStartEvent struct {
	ID               string         `json:"id"`
	ToolName         string         `json:"toolName"`
	ProviderExecuted bool           `json:"providerExecuted"`
	Dynamic          bool           `json:"dynamic"`
	ProviderMetadata map[string]any `json:"providerMetadata,omitempty"`
}

// ToolInputDeltaEvent is emitted for each tool input chunk
type ToolInputDeltaEvent struct {
	ID    string `json:"id"`
	Delta string `json:"delta"`
}

// ToolInputEndEvent is emitted when tool input completes
type ToolInputEndEvent struct {
	ID string `json:"id"`
}

// ToolCallEvent is emitted when a tool is invoked
type ToolCallEvent struct {
	ToolCallID       string         `json:"toolCallId"`
	ToolName         string         `json:"toolName"`
	Input            string         `json:"input"`
	ProviderExecuted bool           `json:"providerExecuted"`
	Dynamic          bool           `json:"dynamic"`
	ProviderMetadata map[string]any `json:"providerMetadata,omitempty"`
}

// ToolResultEvent is emitted when a tool result is received
type ToolResultEvent struct {
	ToolCallID       string         `json:"toolCallId"`
	ToolName         string         `json:"toolName"`
	Result           any            `json:"result"`
	IsError          bool           `json:"isError"`
	ProviderExecuted bool           `json:"providerExecuted"`
	Dynamic          bool           `json:"dynamic"`
	ProviderMetadata map[string]any `json:"providerMetadata,omitempty"`
}

// ToolErrorEvent is emitted when a tool execution fails
type ToolErrorEvent struct {
	ToolCallID       string         `json:"toolCallId"`
	ToolName         string         `json:"toolName"`
	Error            string         `json:"error"`
	ProviderExecuted bool           `json:"providerExecuted"`
	Dynamic          bool           `json:"dynamic"`
	ProviderMetadata map[string]any `json:"providerMetadata,omitempty"`
}

// ReasoningStartEvent is emitted when extended thinking begins
type ReasoningStartEvent struct {
	ID string `json:"id"`
}

// ReasoningDeltaEvent is emitted for each reasoning chunk
type ReasoningDeltaEvent struct {
	ID    string `json:"id"`
	Delta string `json:"delta"`
}

// ReasoningEndEvent is emitted when extended thinking completes
type ReasoningEndEvent struct {
	ID string `json:"id"`
}

// ResponseMetadataEvent contains session and model metadata
type ResponseMetadataEvent struct {
	SessionID string  `json:"sessionId,omitempty"`
	ModelID   string  `json:"modelId,omitempty"`
	CostUSD   float64 `json:"costUsd,omitempty"`
}

// FinishEvent is the final event with usage stats and finish reason
type FinishEvent struct {
	FinishReason string     `json:"finishReason"`
	Usage        UsageStats `json:"usage"`
	Metadata     any        `json:"metadata,omitempty"`
}

// UsageStats represents token usage in AI SDK v6 format
type UsageStats struct {
	InputTokens  InputTokens  `json:"inputTokens"`
	OutputTokens OutputTokens `json:"outputTokens"`
	Raw          any          `json:"raw,omitempty"`
}

// InputTokens represents input token breakdown
type InputTokens struct {
	Total      int64 `json:"total"`
	NoCache    int64 `json:"noCache"`
	CacheRead  int64 `json:"cacheRead"`
	CacheWrite int64 `json:"cacheWrite"`
}

// OutputTokens represents output token breakdown
type OutputTokens struct {
	Total     int64  `json:"total"`
	Text      *int64 `json:"text,omitempty"`
	Reasoning *int64 `json:"reasoning,omitempty"`
}

// ClaudeUsage represents raw usage data from Claude SDK
type ClaudeUsage struct {
	InputTokens         *int64 `json:"input_tokens"`
	OutputTokens        *int64 `json:"output_tokens"`
	CacheCreationTokens *int64 `json:"cache_creation_input_tokens"`
	CacheReadTokens     *int64 `json:"cache_read_input_tokens"`
}

// ToolStreamState tracks the streaming lifecycle for a single tool invocation
type ToolStreamState struct {
	Name                string
	LastSerializedInput *string
	InputStarted        bool
	InputClosed         bool
	CallEmitted         bool
	ParentToolCallID    *string
}

// ClaudeToolUse represents a tool_use content block from Claude
type ClaudeToolUse struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	Input           any     `json:"input"`
	ParentToolUseID *string `json:"parent_tool_use_id,omitempty"`
}

// ClaudeToolResult represents a tool_result content block
type ClaudeToolResult struct {
	ID      string  `json:"tool_use_id"`
	Name    *string `json:"name,omitempty"`
	Result  any     `json:"content"`
	IsError bool    `json:"is_error"`
}

// ClaudeToolError represents a tool_error content block
type ClaudeToolError struct {
	ID    string  `json:"tool_use_id"`
	Name  *string `json:"name,omitempty"`
	Error any     `json:"error"`
}

// ClaudeMessage represents a message from the Claude SDK stream
type ClaudeMessage struct {
	Type             string              `json:"type"` // "stream_event", "assistant", "user", "result", "system"
	Subtype          *string             `json:"subtype,omitempty"`
	Message          *MessageContent     `json:"message,omitempty"`
	Event            *StreamEventDetails `json:"event,omitempty"`
	ParentToolUseID  *string             `json:"parent_tool_use_id,omitempty"`
	SessionID        *string             `json:"session_id,omitempty"`
	Usage            *ClaudeUsage        `json:"usage,omitempty"`
	TotalCostUSD     *float64            `json:"total_cost_usd,omitempty"`
	DurationMS       *int64              `json:"duration_ms,omitempty"`
	IsError          bool                `json:"is_error"`
	Result           *string             `json:"result,omitempty"`
	StructuredOutput any                 `json:"structured_output,omitempty"`
}

// MessageContent represents the content field in assistant/user messages
type MessageContent struct {
	Content []ContentBlock `json:"content"`
}

// ContentBlock represents a content block (text, tool_use, tool_result, thinking)
type ContentBlock struct {
	Type            string  `json:"type"`
	Text            *string `json:"text,omitempty"`
	ID              *string `json:"id,omitempty"`
	Name            *string `json:"name,omitempty"`
	Input           any     `json:"input,omitempty"`
	ToolUseID       *string `json:"tool_use_id,omitempty"`
	Content         any     `json:"content,omitempty"`
	IsError         *bool   `json:"is_error,omitempty"`
	Error           any     `json:"error,omitempty"`
	Thinking        *string `json:"thinking,omitempty"`
	ParentToolUseID *string `json:"parent_tool_use_id,omitempty"`
}

// StreamEventDetails represents the event field in stream_event messages
type StreamEventDetails struct {
	Type         string        `json:"type"` // "content_block_start", "content_block_delta", "content_block_stop"
	Index        *int          `json:"index,omitempty"`
	ContentBlock *ContentBlock `json:"content_block,omitempty"`
	Delta        *ContentDelta `json:"delta,omitempty"`
}

// ContentDelta represents delta updates in streaming
type ContentDelta struct {
	Type        string  `json:"type"` // "text_delta", "input_json_delta", "thinking_delta"
	Text        *string `json:"text,omitempty"`
	PartialJSON *string `json:"partial_json,omitempty"`
	Thinking    *string `json:"thinking,omitempty"`
}
