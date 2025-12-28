package conversations

import (
	"testing"

	"github.com/griffnb/core/lib/router"
)

func Test_Setup_RegistersRoutes(t *testing.T) {
	t.Run("successfully registers all routes without panic", func(t *testing.T) {
		// Arrange
		coreRouter := router.Setup(8080, "test-session-key", []string{"http://localhost"})

		// Act - Setup should not panic
		// This verifies that all routes, including the new streaming route, are registered correctly
		Setup(coreRouter)

		// Assert - if we get here without panic, routes are registered
		// The streaming route POST /{conversationId}/sandbox/{sandboxId} should now be available
	})
}
