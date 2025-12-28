package conversation

type ConversationStats struct {
	MessagesExchanged int   `json:"messages_exchanged"`
	TotalTokensUsed   int64 `json:"total_tokens_used"`
}
