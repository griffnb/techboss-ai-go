package claude

import (
	"testing"
)

func Test_NewStreamParser_createsEmptyMaps(t *testing.T) {
	// Act
	parser := NewStreamParser()

	// Assert
	if parser == nil {
		t.Fatal("Expected parser to be non-nil")
	}
	if parser.toolStates == nil {
		t.Error("Expected toolStates to be initialized")
	}
	if parser.toolBlocksByIndex == nil {
		t.Error("Expected toolBlocksByIndex to be initialized")
	}
	if parser.toolInputAccumulators == nil {
		t.Error("Expected toolInputAccumulators to be initialized")
	}
	if parser.textBlocksByIndex == nil {
		t.Error("Expected textBlocksByIndex to be initialized")
	}
	if parser.activeTaskTools == nil {
		t.Error("Expected activeTaskTools to be initialized")
	}
	if len(parser.toolStates) != 0 {
		t.Errorf("Expected toolStates to be empty, got %d items", len(parser.toolStates))
	}
	if len(parser.toolBlocksByIndex) != 0 {
		t.Errorf("Expected toolBlocksByIndex to be empty, got %d items", len(parser.toolBlocksByIndex))
	}
	if len(parser.toolInputAccumulators) != 0 {
		t.Errorf("Expected toolInputAccumulators to be empty, got %d items", len(parser.toolInputAccumulators))
	}
	if len(parser.textBlocksByIndex) != 0 {
		t.Errorf("Expected textBlocksByIndex to be empty, got %d items", len(parser.textBlocksByIndex))
	}
	if len(parser.activeTaskTools) != 0 {
		t.Errorf("Expected activeTaskTools to be empty, got %d items", len(parser.activeTaskTools))
	}
}

func Test_CloseToolInput_onlyClosesIfInputStartedAndNotClosed(t *testing.T) {
	tests := []struct {
		name           string
		setup          func(*StreamParser, string)
		expectClosed   bool
		expectNoChange bool
	}{
		{
			name: "closes when input started and not closed",
			setup: func(p *StreamParser, toolID string) {
				p.toolStates[toolID] = &ToolStreamState{
					InputStarted: true,
					InputClosed:  false,
				}
			},
			expectClosed:   true,
			expectNoChange: false,
		},
		{
			name: "does not close when input not started",
			setup: func(p *StreamParser, toolID string) {
				p.toolStates[toolID] = &ToolStreamState{
					InputStarted: false,
					InputClosed:  false,
				}
			},
			expectClosed:   false,
			expectNoChange: true,
		},
		{
			name: "does not close when already closed",
			setup: func(p *StreamParser, toolID string) {
				p.toolStates[toolID] = &ToolStreamState{
					InputStarted: true,
					InputClosed:  true,
				}
			},
			expectClosed:   true,
			expectNoChange: true,
		},
		{
			name: "does nothing when tool state does not exist",
			setup: func(_ *StreamParser, _ string) {
				// Don't add any state
			},
			expectClosed:   false,
			expectNoChange: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			parser := NewStreamParser()
			toolID := "tool-123"
			tt.setup(parser, toolID)

			// Act
			parser.CloseToolInput(toolID)

			// Assert
			if tt.expectNoChange {
				if state, exists := parser.toolStates[toolID]; exists {
					if state.InputClosed != tt.expectClosed {
						t.Errorf("Expected InputClosed=%v, got %v", tt.expectClosed, state.InputClosed)
					}
				}
			} else {
				state := parser.toolStates[toolID]
				if state == nil {
					t.Fatal("Expected state to be non-nil")
				}
				if state.InputClosed != tt.expectClosed {
					t.Errorf("Expected InputClosed=%v, got %v", tt.expectClosed, state.InputClosed)
				}
			}
		})
	}
}

func Test_EmitToolCall_closesInputFirstAndMarksEmitted(t *testing.T) {
	tests := []struct {
		name              string
		setup             func(*StreamParser, string)
		expectInputClosed bool
		expectCallEmitted bool
	}{
		{
			name: "emits call and closes input when not yet emitted",
			setup: func(p *StreamParser, toolID string) {
				p.toolStates[toolID] = &ToolStreamState{
					InputStarted: true,
					InputClosed:  false,
					CallEmitted:  false,
				}
			},
			expectInputClosed: true,
			expectCallEmitted: true,
		},
		{
			name: "does not emit again when already emitted",
			setup: func(p *StreamParser, toolID string) {
				p.toolStates[toolID] = &ToolStreamState{
					InputStarted: true,
					InputClosed:  true,
					CallEmitted:  true,
				}
			},
			expectInputClosed: true,
			expectCallEmitted: true,
		},
		{
			name: "emits call even when input not started",
			setup: func(p *StreamParser, toolID string) {
				p.toolStates[toolID] = &ToolStreamState{
					InputStarted: false,
					InputClosed:  false,
					CallEmitted:  false,
				}
			},
			expectInputClosed: false,
			expectCallEmitted: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			parser := NewStreamParser()
			toolID := "tool-123"
			tt.setup(parser, toolID)

			// Act
			parser.EmitToolCall(toolID)

			// Assert
			state := parser.toolStates[toolID]
			if state == nil {
				t.Fatal("Expected state to be non-nil")
			}
			if state.InputClosed != tt.expectInputClosed {
				t.Errorf("Expected InputClosed=%v, got %v", tt.expectInputClosed, state.InputClosed)
			}
			if state.CallEmitted != tt.expectCallEmitted {
				t.Errorf("Expected CallEmitted=%v, got %v", tt.expectCallEmitted, state.CallEmitted)
			}
		})
	}
}

func Test_FinalizeToolCalls_emitsAllPendingToolsAndClearsState(t *testing.T) {
	// Arrange
	parser := NewStreamParser()

	// Add multiple tool states
	parser.toolStates["tool-1"] = &ToolStreamState{
		InputStarted: true,
		InputClosed:  false,
		CallEmitted:  false,
	}
	parser.toolStates["tool-2"] = &ToolStreamState{
		InputStarted: true,
		InputClosed:  true,
		CallEmitted:  false,
	}
	parser.toolStates["tool-3"] = &ToolStreamState{
		InputStarted: true,
		InputClosed:  true,
		CallEmitted:  true, // Already emitted
	}

	parser.toolBlocksByIndex[0] = "tool-1"
	parser.toolBlocksByIndex[1] = "tool-2"
	parser.toolInputAccumulators["tool-1"] = "input1"
	parser.textBlocksByIndex[0] = "text-1"

	// Act
	parser.FinalizeToolCalls()

	// Assert - all maps should be cleared
	if len(parser.toolStates) != 0 {
		t.Errorf("Expected toolStates to be cleared, got %d items", len(parser.toolStates))
	}
	if len(parser.toolBlocksByIndex) != 0 {
		t.Errorf("Expected toolBlocksByIndex to be cleared, got %d items", len(parser.toolBlocksByIndex))
	}
	if len(parser.toolInputAccumulators) != 0 {
		t.Errorf("Expected toolInputAccumulators to be cleared, got %d items", len(parser.toolInputAccumulators))
	}
	if len(parser.textBlocksByIndex) != 0 {
		t.Errorf("Expected textBlocksByIndex to be cleared, got %d items", len(parser.textBlocksByIndex))
	}
}

func Test_GetFallbackParentID_returnsIDWhenExactlyOneTaskToolActive(t *testing.T) {
	tests := []struct {
		name           string
		setup          func(*StreamParser)
		expectParentID *string
	}{
		{
			name: "returns nil when no Task tools active",
			setup: func(_ *StreamParser) {
				// No active Task tools
			},
			expectParentID: nil,
		},
		{
			name: "returns tool ID when exactly one Task tool active",
			setup: func(p *StreamParser) {
				p.activeTaskTools["task-tool-1"] = struct{}{}
			},
			expectParentID: stringPtr("task-tool-1"),
		},
		{
			name: "returns nil when two Task tools active",
			setup: func(p *StreamParser) {
				p.activeTaskTools["task-tool-1"] = struct{}{}
				p.activeTaskTools["task-tool-2"] = struct{}{}
			},
			expectParentID: nil,
		},
		{
			name: "returns nil when three Task tools active",
			setup: func(p *StreamParser) {
				p.activeTaskTools["task-tool-1"] = struct{}{}
				p.activeTaskTools["task-tool-2"] = struct{}{}
				p.activeTaskTools["task-tool-3"] = struct{}{}
			},
			expectParentID: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			parser := NewStreamParser()
			tt.setup(parser)

			// Act
			result := parser.GetFallbackParentID()

			// Assert
			if tt.expectParentID == nil {
				if result != nil {
					t.Errorf("Expected nil, got %v", *result)
				}
			} else {
				if result == nil {
					t.Error("Expected non-nil result")
				} else if *result != *tt.expectParentID {
					t.Errorf("Expected %s, got %s", *tt.expectParentID, *result)
				}
			}
		})
	}
}

func Test_TrackTaskTool_addsAndRemovesFromActiveTaskTools(t *testing.T) {
	tests := []struct {
		name     string
		toolID   string
		isTask   bool
		setup    func(*StreamParser)
		expected int
	}{
		{
			name:   "adds Task tool to activeTaskTools",
			toolID: "task-1",
			isTask: true,
			setup: func(_ *StreamParser) {
				// Start empty
			},
			expected: 1,
		},
		{
			name:   "removes non-Task tool from activeTaskTools",
			toolID: "task-1",
			isTask: false,
			setup: func(p *StreamParser) {
				p.activeTaskTools["task-1"] = struct{}{}
			},
			expected: 0,
		},
		{
			name:   "does not add non-Task tool",
			toolID: "regular-tool",
			isTask: false,
			setup: func(_ *StreamParser) {
				// Start empty
			},
			expected: 0,
		},
		{
			name:   "adds second Task tool",
			toolID: "task-2",
			isTask: true,
			setup: func(p *StreamParser) {
				p.activeTaskTools["task-1"] = struct{}{}
			},
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			parser := NewStreamParser()
			tt.setup(parser)

			// Act
			parser.TrackTaskTool(tt.toolID, tt.isTask)

			// Assert
			if len(parser.activeTaskTools) != tt.expected {
				t.Errorf("Expected %d active task tools, got %d", tt.expected, len(parser.activeTaskTools))
			}
			if tt.isTask {
				_, exists := parser.activeTaskTools[tt.toolID]
				if !exists {
					t.Errorf("Expected tool %s to be in activeTaskTools", tt.toolID)
				}
			}
		})
	}
}

// Helper function for creating string pointers
func stringPtr(s string) *string {
	return &s
}
