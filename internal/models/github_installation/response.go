package github_installation

import "github.com/griffnb/core/lib/sanitize"

// ToPublicJSON converts the model to a sanitized JSON representation for public consumption
func (this *GithubInstallation) ToPublicJSON() any {
	return sanitize.SanitizeModel(this, &Structure{})
}
