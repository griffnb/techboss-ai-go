package claude

// ConvertClaudeCodeUsage converts Claude SDK usage to AI SDK v6 format
func ConvertClaudeCodeUsage(usage *ClaudeUsage) UsageStats {
	if usage == nil {
		return UsageStats{
			InputTokens: InputTokens{
				Total:      0,
				NoCache:    0,
				CacheRead:  0,
				CacheWrite: 0,
			},
			OutputTokens: OutputTokens{
				Total: 0,
			},
			Raw: nil,
		}
	}

	// Helper function to safely dereference int64 pointers
	derefInt64 := func(ptr *int64) int64 {
		if ptr == nil {
			return 0
		}
		return *ptr
	}

	inputTokens := derefInt64(usage.InputTokens)
	cacheWrite := derefInt64(usage.CacheCreationTokens)
	cacheRead := derefInt64(usage.CacheReadTokens)
	outputTokens := derefInt64(usage.OutputTokens)

	return UsageStats{
		InputTokens: InputTokens{
			Total:      inputTokens + cacheWrite + cacheRead,
			NoCache:    inputTokens,
			CacheRead:  cacheRead,
			CacheWrite: cacheWrite,
		},
		OutputTokens: OutputTokens{
			Total: outputTokens,
		},
		Raw: usage,
	}
}
