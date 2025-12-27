package sandbox

import (
	"github.com/griffnb/techboss-ai-go/internal/integrations/modal"
)

type MetaData struct {
	SandboxID string               `json:"sandbox_id"` // Modal sandbox ID
	Status    *modal.SandboxStatus `json:"status"`     // Current status
}
