package helpers

/*


https://developers.cloudflare.com/turnstile/get-started/client-side-rendering/

turnstileData := tools.ParseStringI(data["turnstile_data"])
	turnstileResponseIsValid, unexpectedError := validateTurnstileResponse(turnstileData, req.RemoteAddr)
	if unexpectedError != nil {
		log.Debug(unexpectedError)
		return nil, errors.Errorf("500 Internal Server Error")
	}

	if !turnstileResponseIsValid {
		return nil, errors.Errorf("400 Bad Request")
	}



func validateTurnstileResponse(turnstileData string, remoteIP string) (bool, error) {
	//#region format/send request
	var turnstileRequest *http.Request
	{
		url := "https://challenges.cloudflare.com/turnstile/v0/siteverify"
		secret := environment.GetConfig().GetString("CloudFlareTurnstileSecretKey")
		jsonData := map[string]interface{}{
			"secret":          secret,
			"response":        turnstileData,
			"idempotency_key": tools.GUID(),
			"remoteip":        remoteIP,
		}

		params, err := json.Marshal(jsonData)
		if err != nil {
			return false, err
		}
		turnstileRequest, err = http.NewRequest("POST", url, bytes.NewBuffer(params))
		if err != nil {
			return false, err
		}
		turnstileRequest.Header.Set("Content-Type", "application/json")
		turnstileRequest.Header.Set("X-Requested-With", "XMLHttpRequest")
	}
	//#endregion

	//#region handle response
	{
		client := &http.Client{}
		resp, err := client.Do(turnstileRequest)
		if err != nil {
			return false, err
		}
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return false, err
		}

		respData := map[string]interface{}{}

		err = json.Unmarshal(respBody, &respData)
		if err != nil {
			return false, err
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
	}
	//#endregion

	return true, nil
}
*/
