package github_installations

import (
	"net/http"

	"github.com/griffnb/techboss-ai-go/internal/models/github_installation"
)

// Test helpers to expose internal functions for testing

// AdminIndex exposes adminIndex for testing
func AdminIndex(w http.ResponseWriter, req *http.Request) ([]*github_installation.GithubInstallationJoined, int, error) {
	return adminIndex(w, req)
}

// AdminGet exposes adminGet for testing
func AdminGet(w http.ResponseWriter, req *http.Request) (*github_installation.GithubInstallationJoined, int, error) {
	return adminGet(w, req)
}

// AdminCreate exposes adminCreate for testing
func AdminCreate(w http.ResponseWriter, req *http.Request) (*github_installation.GithubInstallation, int, error) {
	return adminCreate(w, req)
}

// AdminUpdate exposes adminUpdate for testing
func AdminUpdate(w http.ResponseWriter, req *http.Request) (*github_installation.GithubInstallationJoined, int, error) {
	return adminUpdate(w, req)
}

// AdminCount exposes adminCount for testing
func AdminCount(w http.ResponseWriter, req *http.Request) (int64, int, error) {
	return adminCount(w, req)
}
