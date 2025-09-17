package helpers

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/CrowdShield/go-core/lib/log"
	"github.com/CrowdShield/go-core/lib/oauth"
	"github.com/CrowdShield/go-core/lib/tools"
	"github.com/griffnb/techboss-ai-go/internal/environment"
	"github.com/pkg/errors"
)

// google https://developers.google.com/identity/openid-connect/openid-connect

// Define a structure for the user information we want to retrieve
type UserInfo struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Picture string `json:"picture"`
}

// Function to get user info from the access token and provider
// TODO these session cookies age....
func GetUserInfo(provider, accessToken string) (*UserInfo, error) {
	var endpoint string

	// Choose the appropriate user info endpoint based on the provider
	switch provider {
	case "google":
		endpoint = "https://www.googleapis.com/oauth2/v3/userinfo"
	case "microsoft":
		endpoint = "https://graph.microsoft.com/v1.0/me"
	case "facebook":
		endpoint = "https://graph.facebook.com/me?fields=id,name,email,picture"
	default:
		return nil, errors.Errorf("unsupported provider: %s", provider)
	}

	// Create a new HTTP request with the Authorization header
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	// Perform the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			log.Error(err)
		}
	}()

	// Check if the request was successful
	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("failed to retrieve user info: %s", resp.Status)
	}

	// Read and parse the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var userInfo UserInfo
	err = json.Unmarshal(body, &userInfo)
	if err != nil {
		return nil, errors.WithMessagef(err, "failed to unmarshal user info: %s", string(body))
	}

	return &userInfo, nil
}

// VerifyAccessToken function to verify the access token
func VerifyAccessToken(provider, accessToken string) (bool, error) {
	var endpoint string

	// Choose the appropriate user info endpoint based on the provider
	switch provider {
	case "google":
		endpoint = "https://oauth2.googleapis.com/tokeninfo?access_token=" + accessToken
	case "microsoft":
		endpoint = "https://graph.microsoft.com/v1.0/me"
	case "facebook":
		endpoint = "https://graph.facebook.com/me?fields=id,name,email,picture"
	default:
		return false, errors.Errorf("unsupported provider: %s", provider)
	}

	// Create a new HTTP request with the Authorization header
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return false, errors.WithStack(err)
	}

	// Perform the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, errors.WithStack(err)
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			log.Error(err)
		}
	}()

	// Check if the request was successful
	if resp.StatusCode != http.StatusOK {
		return false, errors.Errorf("failed to retrieve user info: %s", resp.Status)
	}

	// Read and parse the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, errors.WithStack(err)
	}

	var tokenInfo map[string]interface{}
	err = json.Unmarshal(body, &tokenInfo)
	if err != nil {
		return false, errors.WithMessagef(err, "failed to unmarshal user info: %s", string(body))
	}

	if tools.ParseIntI(tokenInfo["expires_in"]) <= 0 {
		return false, errors.New("token has expired")
	}

	if tools.ParseStringI(tokenInfo["aud"]) != environment.GetConfig().Oauth.ClientID {
		return false, errors.New("token was not issued for this client")
	}

	return true, nil
}

// HandleTokenLogin takes a access token from the front and gets a user
func HandleTokenLogin(auth *oauth.Authenticator, _ http.ResponseWriter, req *http.Request) (*oauth.OauthProfile, string, error) {
	token := req.URL.Query().Get("token")

	profile, err := oauth.VerifyIDTokenString(req.Context(), auth.GetDomain(), token)
	if err != nil {
		return nil, "", errors.Wrap(err, "Failed to verify id token")
	}
	if tools.Empty(profile) {
		return nil, "", errors.Errorf("No profile")
	}

	return profile, token, nil
}
