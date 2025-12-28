package conversation

// ConversationStats tracks message counts and token usage for a conversation
type ConversationStats struct {
	MessagesExchanged int   `json:"messages_exchanged"`
	TotalInputTokens  int64 `json:"total_input_tokens"`
	TotalOutputTokens int64 `json:"total_output_tokens"`
	TotalCacheTokens  int64 `json:"total_cache_tokens"`
	// Deprecated: Use TotalInputTokens, TotalOutputTokens, TotalCacheTokens instead
	TotalTokensUsed int64 `json:"total_tokens_used"`
}

// AddTokenUsage accumulates token usage for the conversation
func (s *ConversationStats) AddTokenUsage(inputTokens, outputTokens, cacheTokens int64) {
	s.TotalInputTokens += inputTokens
	s.TotalOutputTokens += outputTokens
	s.TotalCacheTokens += cacheTokens
}

// IncrementMessages increments the message counter
func (s *ConversationStats) IncrementMessages() {
	s.MessagesExchanged++
}
