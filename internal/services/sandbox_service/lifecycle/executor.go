package lifecycle

import (
	"context"
	"time"

	"github.com/griffnb/core/lib/log"
)

// ExecuteHook runs a lifecycle hook with logging and duration tracking.
// The hook itself determines criticality by its return value:
//   - Returning an error = critical failure, caller should handle
//   - Returning nil = success or non-critical failure (hook swallowed error)
//
// If hook is nil, this function returns nil immediately (no-op).
//
// Satisfies:
//   - Requirement 6.3: Hook execution with logging
//   - Requirement 6.4: Hook error propagation
//   - Requirement 9.1-9.3: Critical vs non-critical hook handling
//   - Requirement 10.4: Log hook execution with duration
//   - Design Phase 5.2: ExecuteHook implementation
func ExecuteHook(
	ctx context.Context,
	hookName string,
	hook HookFunc,
	hookData *HookData,
) error {
	if hook == nil {
		return nil // No hook registered, skip
	}

	startTime := time.Now()
	log.Infof("[Lifecycle Hook] Starting %s for conversation=%s sandbox=%s",
		hookName,
		hookData.ConversationID,
		hookData.SandboxInfo.SandboxID)

	err := hook(ctx, hookData)

	duration := time.Since(startTime)
	if err != nil {
		log.Errorf(err, "[Lifecycle Hook] %s failed after %v for conversation=%s sandbox=%s",
			hookName,
			duration,
			hookData.ConversationID,
			hookData.SandboxInfo.SandboxID)
		return err
	}

	log.Infof("[Lifecycle Hook] %s completed successfully in %v for conversation=%s sandbox=%s",
		hookName,
		duration,
		hookData.ConversationID,
		hookData.SandboxInfo.SandboxID)
	return nil
}
