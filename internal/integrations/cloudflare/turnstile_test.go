package cloudflare_test

import (
	"testing"

	"github.com/griffnb/techboss-ai-go/internal/common/system_testing"
	"github.com/griffnb/techboss-ai-go/internal/integrations/cloudflare"
)

func init() {
	system_testing.BuildSystem()
}

const (
	ALWAYS_PASS  = "1x0000000000000000000000000000000AA"
	ALWAYS_FAIL  = "2x0000000000000000000000000000000AA"
	ALWAYS_ERROR = "3x0000000000000000000000000000000AA"
)

func TestValidateTurnstileResponse(t *testing.T) {
	if !cloudflare.Configured() {
		t.Skip("Cloudflare client is not configured, skipping test")
		return
	}

	cloudflare.Client().TurnstileKey = ALWAYS_PASS

	resp, err := cloudflare.Client().ValidateTurnstileResponse("test", "127.0.0.1")
	if err != nil {
		t.Fatal(err)
	}

	if !resp {
		t.Fatal("expected true")
	}
}
