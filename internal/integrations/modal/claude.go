package modal

import (
	"context"

	"github.com/modal-labs/libmodal/modal-go"
	"github.com/pkg/errors"
)

func (this *APIClient) ExecClaude(
	ctx context.Context,
	sb *modal.Sandbox,
	prompt string,
) (*modal.ContainerProcess, error) {
	secrets, err := this.client.Secrets.FromMap(ctx, map[string]string{
		"ANTHROPIC_API_KEY":       "sk-xxxx",
		"AWS_BEDROCK_API_KEY":     "ABSKxxxx",
		"CLAUDE_CODE_USE_BEDROCK": "1",
		"AWS_REGION":              "us-east-1", // or your preferred region
	}, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get secrets for Sandbox %s", sb.SandboxID)
	}

	cmd := []string{"claude", "-c", "-p", prompt, "--dangerously-skip-permissions", "--output-format", "stream-json", "--verbose"}

	claude, err := sb.Exec(ctx, cmd, &modal.SandboxExecParams{
		PTY:     true, // Adding a PTY is important, since Claude requires it!
		Secrets: []*modal.Secret{secrets},
		Workdir: "/repo",
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to execute command in Sandbox %s", sb.SandboxID)
	}

	return claude, nil
}
