package claude

import (
	"testing"
)

func Test_ConvertClaudeCodeUsage(t *testing.T) {
	tests := []struct {
		name     string
		usage    *ClaudeUsage
		expected UsageStats
	}{
		{
			name: "all fields populated",
			usage: &ClaudeUsage{
				InputTokens:         intPtr(1000),
				OutputTokens:        intPtr(500),
				CacheCreationTokens: intPtr(200),
				CacheReadTokens:     intPtr(300),
			},
			expected: UsageStats{
				InputTokens: InputTokens{
					Total:      1500, // 1000 + 200 + 300
					NoCache:    1000,
					CacheRead:  300,
					CacheWrite: 200,
				},
				OutputTokens: OutputTokens{
					Total: 500,
				},
				Raw: &ClaudeUsage{
					InputTokens:         intPtr(1000),
					OutputTokens:        intPtr(500),
					CacheCreationTokens: intPtr(200),
					CacheReadTokens:     intPtr(300),
				},
			},
		},
		{
			name: "nil cache fields",
			usage: &ClaudeUsage{
				InputTokens:         intPtr(1000),
				OutputTokens:        intPtr(500),
				CacheCreationTokens: nil,
				CacheReadTokens:     nil,
			},
			expected: UsageStats{
				InputTokens: InputTokens{
					Total:      1000,
					NoCache:    1000,
					CacheRead:  0,
					CacheWrite: 0,
				},
				OutputTokens: OutputTokens{
					Total: 500,
				},
				Raw: &ClaudeUsage{
					InputTokens:         intPtr(1000),
					OutputTokens:        intPtr(500),
					CacheCreationTokens: nil,
					CacheReadTokens:     nil,
				},
			},
		},
		{
			name: "zero values",
			usage: &ClaudeUsage{
				InputTokens:         intPtr(0),
				OutputTokens:        intPtr(0),
				CacheCreationTokens: intPtr(0),
				CacheReadTokens:     intPtr(0),
			},
			expected: UsageStats{
				InputTokens: InputTokens{
					Total:      0,
					NoCache:    0,
					CacheRead:  0,
					CacheWrite: 0,
				},
				OutputTokens: OutputTokens{
					Total: 0,
				},
				Raw: &ClaudeUsage{
					InputTokens:         intPtr(0),
					OutputTokens:        intPtr(0),
					CacheCreationTokens: intPtr(0),
					CacheReadTokens:     intPtr(0),
				},
			},
		},
		{
			name: "large token counts",
			usage: &ClaudeUsage{
				InputTokens:         intPtr(1000000),
				OutputTokens:        intPtr(500000),
				CacheCreationTokens: intPtr(200000),
				CacheReadTokens:     intPtr(300000),
			},
			expected: UsageStats{
				InputTokens: InputTokens{
					Total:      1500000, // 1000000 + 200000 + 300000
					NoCache:    1000000,
					CacheRead:  300000,
					CacheWrite: 200000,
				},
				OutputTokens: OutputTokens{
					Total: 500000,
				},
				Raw: &ClaudeUsage{
					InputTokens:         intPtr(1000000),
					OutputTokens:        intPtr(500000),
					CacheCreationTokens: intPtr(200000),
					CacheReadTokens:     intPtr(300000),
				},
			},
		},
		{
			name:  "nil usage pointer",
			usage: nil,
			expected: UsageStats{
				InputTokens: InputTokens{
					Total:      0,
					NoCache:    0,
					CacheRead:  0,
					CacheWrite: 0,
				},
				OutputTokens: OutputTokens{
					Total: 0,
				},
				Raw: (*ClaudeUsage)(nil),
			},
		},
		{
			name: "only input tokens",
			usage: &ClaudeUsage{
				InputTokens:         intPtr(1000),
				OutputTokens:        nil,
				CacheCreationTokens: nil,
				CacheReadTokens:     nil,
			},
			expected: UsageStats{
				InputTokens: InputTokens{
					Total:      1000,
					NoCache:    1000,
					CacheRead:  0,
					CacheWrite: 0,
				},
				OutputTokens: OutputTokens{
					Total: 0,
				},
				Raw: &ClaudeUsage{
					InputTokens:         intPtr(1000),
					OutputTokens:        nil,
					CacheCreationTokens: nil,
					CacheReadTokens:     nil,
				},
			},
		},
		{
			name: "only output tokens",
			usage: &ClaudeUsage{
				InputTokens:         nil,
				OutputTokens:        intPtr(500),
				CacheCreationTokens: nil,
				CacheReadTokens:     nil,
			},
			expected: UsageStats{
				InputTokens: InputTokens{
					Total:      0,
					NoCache:    0,
					CacheRead:  0,
					CacheWrite: 0,
				},
				OutputTokens: OutputTokens{
					Total: 500,
				},
				Raw: &ClaudeUsage{
					InputTokens:         nil,
					OutputTokens:        intPtr(500),
					CacheCreationTokens: nil,
					CacheReadTokens:     nil,
				},
			},
		},
		{
			name: "only cache tokens",
			usage: &ClaudeUsage{
				InputTokens:         nil,
				OutputTokens:        nil,
				CacheCreationTokens: intPtr(200),
				CacheReadTokens:     intPtr(300),
			},
			expected: UsageStats{
				InputTokens: InputTokens{
					Total:      500, // 200 + 300
					NoCache:    0,
					CacheRead:  300,
					CacheWrite: 200,
				},
				OutputTokens: OutputTokens{
					Total: 0,
				},
				Raw: &ClaudeUsage{
					InputTokens:         nil,
					OutputTokens:        nil,
					CacheCreationTokens: intPtr(200),
					CacheReadTokens:     intPtr(300),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertClaudeCodeUsage(tt.usage)

			// Check input tokens
			if result.InputTokens.Total != tt.expected.InputTokens.Total {
				t.Errorf("InputTokens.Total mismatch: expected %d, got %d", tt.expected.InputTokens.Total, result.InputTokens.Total)
			}
			if result.InputTokens.NoCache != tt.expected.InputTokens.NoCache {
				t.Errorf("InputTokens.NoCache mismatch: expected %d, got %d", tt.expected.InputTokens.NoCache, result.InputTokens.NoCache)
			}
			if result.InputTokens.CacheRead != tt.expected.InputTokens.CacheRead {
				t.Errorf("InputTokens.CacheRead mismatch: expected %d, got %d", tt.expected.InputTokens.CacheRead, result.InputTokens.CacheRead)
			}
			if result.InputTokens.CacheWrite != tt.expected.InputTokens.CacheWrite {
				t.Errorf("InputTokens.CacheWrite mismatch: expected %d, got %d", tt.expected.InputTokens.CacheWrite, result.InputTokens.CacheWrite)
			}

			// Check output tokens
			if result.OutputTokens.Total != tt.expected.OutputTokens.Total {
				t.Errorf("OutputTokens.Total mismatch: expected %d, got %d", tt.expected.OutputTokens.Total, result.OutputTokens.Total)
			}

			// Check raw usage is set correctly
			if tt.usage == nil {
				if result.Raw != nil {
					t.Error("Raw should be nil when usage is nil")
				}
			} else {
				if result.Raw == nil {
					t.Error("Raw should not be nil when usage is not nil")
				}
			}
		})
	}
}

// intPtr is a helper function to create int64 pointers for test data
func intPtr(i int64) *int64 {
	return &i
}
