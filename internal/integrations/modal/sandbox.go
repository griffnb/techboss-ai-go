// https://github.com/modal-labs/libmodal/blob/main/modal-go/examples/sandbox-cloud-bucket/main.go
package modal

import (
	"bufio"
	"context"
	"fmt"
	"log"

	"github.com/modal-labs/libmodal/modal-go"
)

func main() {
	ctx := context.Background()
	mc, err := modal.NewClient()
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	app, err := mc.Apps.FromName(ctx, "libmodal-example", &modal.AppFromNameParams{CreateIfMissing: true})
	if err != nil {
		log.Fatalf("Failed to get or create App: %v", err)
	}

	image := mc.Images.FromRegistry("alpine:3.21", nil).DockerfileCommands([]string{
		"RUN apk add --no-cache bash curl git libgcc libstdc++ ripgrep",
		"RUN curl -fsSL https://claude.ai/install.sh | bash",
		"ENV PATH=/root/.local/bin:$PATH USE_BUILTIN_RIPGREP=0",
	}, nil)

	// standard volume
	volume, err := mc.Volumes.FromName(ctx, "libmodal-example-volume", &modal.VolumeFromNameParams{
		CreateIfMissing: true,
	})
	if err != nil {
		log.Fatalf("Failed to create Volume: %v", err)
	}

	secret, err := mc.Secrets.FromName(ctx, "libmodal-aws-bucket-secret", nil)
	if err != nil {
		log.Fatalf("Failed to get Secret: %v", err)
	}

	// S3 bucket mount
	keyPrefix := "data/"
	cloudBucketMount, err := mc.CloudBucketMounts.New("my-s3-bucket", &modal.CloudBucketMountParams{
		Secret:    secret,
		KeyPrefix: &keyPrefix,
		ReadOnly:  true,
	})
	if err != nil {
		log.Fatalf("Failed to create Cloud Bucket Mount: %v", err)
	}

	sb, err := mc.Sandboxes.Create(ctx, app, image, &modal.SandboxCreateParams{
		Command: []string{
			"sh",
			"-c",
			"echo 'Hello from writer Sandbox!' > /mnt/volume/message.txt",
		},
		Volumes: map[string]*modal.Volume{
			"/mnt/volume": volume,
		},
		CloudBucketMounts: map[string]*modal.CloudBucketMount{
			"/mnt/s3-bucket": cloudBucketMount,
		},
	})
	if err != nil {
		log.Fatalf("Failed to create writer Sandbox: %v", err)
	}
	fmt.Printf("Writer Sandbox: %s\n", sb.SandboxID)
	defer func() {
		if err := sb.Terminate(context.Background()); err != nil {
			log.Fatalf("Failed to terminate Sandbox %s: %v", sb.SandboxID, err)
		}
	}()

	claudeCmd := []string{
		"claude",
		"-p",
		"Summarize what this repository is about. Don't modify any code or files.",
	}
	fmt.Println("\nRunning command:", claudeCmd)

	claudeSecret, err := mc.Secrets.FromName(ctx, "libmodal-anthropic-secret", &modal.SecretFromNameParams{
		RequiredKeys: []string{"ANTHROPIC_API_KEY"},
	})
	if err != nil {
		log.Fatalf("Failed to get secret: %v", err)
	}

	claude, err := sb.Exec(ctx, claudeCmd, &modal.SandboxExecParams{
		PTY:     true, // Adding a PTY is important, since Claude requires it!
		Secrets: []*modal.Secret{claudeSecret},
		Workdir: "/repo",
	})
	if err != nil {
		log.Fatalf("Failed to execute claude command: %v", err)
	}

	scanner := bufio.NewScanner(claude.Stdout)
	for scanner.Scan() {
		//line := scanner.Text()

		// Write the line to the response
		//_, err := fmt.Fprintf(responseWriter, "%s\n", line)
		//if err != nil {
		//	return errors.Wrap(err, "failed to write streaming response")
		//}
		//
		//// Flush the response
		//flusher.Flush()
		//
		//// Check if the stream is done
		//if line == "data: [DONE]" {
		//	break
		//}
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading claude output: %v", err)
	}

	_, err = claude.Wait(ctx)
	if err != nil {
		log.Fatalf("Claude command failed: %v", err)
	}

	credentials, err := sb.CreateConnectToken(ctx, &modal.SandboxCreateConnectTokenParams{UserMetadata: "user_id=xxxx"})
	if err != nil {
		log.Fatalf("Failed to create connect token: %v", err)
	}
	fmt.Printf("Writer connect token: %s\n", credentials.Token)
	exitCode, err := sb.Wait(ctx)
	if err != nil {
		log.Fatalf("Failed to wait for writer Sandbox: %v", err)
	}
	fmt.Printf("Writer finished with exit code: %d\n", exitCode)
}
