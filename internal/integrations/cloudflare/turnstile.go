package cloudflare

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/CrowdShield/go-core/lib/log"
	"github.com/CrowdShield/go-core/lib/tools"
	"github.com/pkg/errors"
)

func (this *APIClient) ValidateTurnstileResponse(turnstileData string, remoteIP string) (bool, error) {
	//#region format/send request
	var turnstileRequest *http.Request

	url := "https://challenges.cloudflare.com/turnstile/v0/siteverify"

	jsonData := map[string]interface{}{
		"secret":          this.TurnstileKey,
		"response":        turnstileData,
		"idempotency_key": tools.SessionKey(),
		"remoteip":        remoteIP,
	}

	params, err := json.Marshal(jsonData)
	if err != nil {
		return false, errors.WithStack(err)
	}
	turnstileRequest, err = http.NewRequest("POST", url, bytes.NewBuffer(params))
	if err != nil {
		return false, errors.WithStack(err)
	}
	turnstileRequest.Header.Set("Content-Type", "application/json")
	turnstileRequest.Header.Set("X-Requested-With", "XMLHttpRequest")

	client := &http.Client{}
	resp, err := client.Do(turnstileRequest)
	if err != nil {
		return false, errors.WithStack(err)
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			log.Error(err)
		}
	}()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, errors.WithStack(err)
	}

	respData := map[string]any{}

	err = json.Unmarshal(respBody, &respData)
	if err != nil {
		return false, errors.WithStack(err)
	}

	success, ok := respData["success"].(bool)
	if !ok {
		return false, errors.New("Failed to parse success status from turnstile response")
	}

	if !success {
		errorCodes, ok := respData["error-codes"].([]interface{})

		if !ok {
			return false, errors.New("Failed to parse error-codes from turnstile response")
		}

		log.Debug(fmt.Sprintf("Turnstile Response invalid: %v", errorCodes))
		return false, nil
	}

	return true, nil
}
