package claude

// StreamParser manages the state of streaming tool calls and text blocks
type StreamParser struct {
	toolStates            map[string]*ToolStreamState
	toolBlocksByIndex     map[int]string
	toolInputAccumulators map[string]string
	textBlocksByIndex     map[int]string
	activeTaskTools       map[string]struct{}
}

// NewStreamParser creates a new StreamParser with initialized maps
func NewStreamParser() *StreamParser {
	return &StreamParser{
		toolStates:            make(map[string]*ToolStreamState),
		toolBlocksByIndex:     make(map[int]string),
		toolInputAccumulators: make(map[string]string),
		textBlocksByIndex:     make(map[int]string),
		activeTaskTools:       make(map[string]struct{}),
	}
}

// CloseToolInput closes tool input if started and not already closed
func (p *StreamParser) CloseToolInput(toolID string) {
	state, exists := p.toolStates[toolID]
	if !exists {
		return
	}

	if state.InputStarted && !state.InputClosed {
		state.InputClosed = true
	}
}

// EmitToolCall emits tool call if not already emitted, closes input first
func (p *StreamParser) EmitToolCall(toolID string) {
	state, exists := p.toolStates[toolID]
	if !exists {
		return
	}

	if state.CallEmitted {
		return
	}

	// Close input first
	p.CloseToolInput(toolID)

	// Mark as emitted
	state.CallEmitted = true
}

// FinalizeToolCalls emits all pending tool calls and clears state
func (p *StreamParser) FinalizeToolCalls() {
	// Emit all pending tools
	for toolID := range p.toolStates {
		p.EmitToolCall(toolID)
	}

	// Clear all maps
	p.toolStates = make(map[string]*ToolStreamState)
	p.toolBlocksByIndex = make(map[int]string)
	p.toolInputAccumulators = make(map[string]string)
	p.textBlocksByIndex = make(map[int]string)
}

// GetFallbackParentID returns parent tool ID if exactly one Task tool is active
func (p *StreamParser) GetFallbackParentID() *string {
	if len(p.activeTaskTools) != 1 {
		return nil
	}

	// Return the single active task tool ID
	for toolID := range p.activeTaskTools {
		return &toolID
	}

	return nil
}

// TrackTaskTool tracks or untracks Task tools in activeTaskTools
func (p *StreamParser) TrackTaskTool(toolID string, isTask bool) {
	if isTask {
		p.activeTaskTools[toolID] = struct{}{}
	} else {
		delete(p.activeTaskTools, toolID)
	}
}
