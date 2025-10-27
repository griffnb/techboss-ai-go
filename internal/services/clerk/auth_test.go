package clerk

import (
	"testing"
)

func TestPrefixWithOrg(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"team_settings:manage", "org:team_settings:manage"},
		{"org:team_settings:manage", "org:team_settings:manage"},
		{"org:org:team_settings:manage", "org:team_settings:manage"},
		{"free", "org:free"},
		{"org:free", "org:free"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := prefixWithOrg(tt.input)
			if result != tt.expected {
				t.Errorf("prefixWithOrg(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSplitByScope(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected ScopeFeatures
	}{
		{
			name:  "should split org and user features",
			input: "org:feature1,user:feature2,org:feature3",
			expected: ScopeFeatures{
				Org:  []string{"feature1", "feature3"},
				User: []string{"feature2"},
			},
		},
		{
			name:  "should handle empty string",
			input: "",
			expected: ScopeFeatures{
				Org:  []string{},
				User: []string{},
			},
		},
		{
			name:  "should handle features with spaces",
			input: " org:feature1 , user:feature2 ",
			expected: ScopeFeatures{
				Org:  []string{"feature1"},
				User: []string{"feature2"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitByScope(tt.input)
			if !slicesEqual(result.Org, tt.expected.Org) {
				t.Errorf("splitByScope(%q).Org = %v, want %v", tt.input, result.Org, tt.expected.Org)
			}
			if !slicesEqual(result.User, tt.expected.User) {
				t.Errorf("splitByScope(%q).User = %v, want %v", tt.input, result.User, tt.expected.User)
			}
		})
	}
}

func TestCheckForFeatureOrPlan(t *testing.T) {
	tests := []struct {
		name          string
		claim         string
		featureOrPlan string
		expected      bool
	}{
		{
			name:          "should find org scoped feature",
			claim:         "org:feature1,user:feature2",
			featureOrPlan: "org:feature1",
			expected:      true,
		},
		{
			name:          "should find user scoped feature",
			claim:         "org:feature1,user:feature2",
			featureOrPlan: "user:feature2",
			expected:      true,
		},
		{
			name:          "should find feature without scope",
			claim:         "org:feature1,user:feature2",
			featureOrPlan: "feature1",
			expected:      true,
		},
		{
			name:          "should not find missing feature",
			claim:         "org:feature1,user:feature2",
			featureOrPlan: "feature3",
			expected:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := checkForFeatureOrPlan(tt.claim, tt.featureOrPlan)
			if result != tt.expected {
				t.Errorf("checkForFeatureOrPlan(%q, %q) = %v, want %v", tt.claim, tt.featureOrPlan, result, tt.expected)
			}
		})
	}
}

func TestCheckOrgAuthorization(t *testing.T) {
	tests := []struct {
		name     string
		params   AuthParams
		options  AuthorizationOptions
		expected *bool
	}{
		{
			name:   "should allow matching permission",
			params: AuthParams{Permission: stringPtr("team_settings:manage")},
			options: AuthorizationOptions{
				OrgID:          stringPtr("org1"),
				OrgRole:        stringPtr("admin"),
				OrgPermissions: []string{"org:team_settings:manage", "org:team_settings:read"},
			},
			expected: boolPtr(true),
		},
		{
			name:   "should deny non-matching permission",
			params: AuthParams{Permission: stringPtr("billing:manage")},
			options: AuthorizationOptions{
				OrgID:          stringPtr("org1"),
				OrgRole:        stringPtr("admin"),
				OrgPermissions: []string{"org:team_settings:manage", "org:team_settings:read"},
			},
			expected: boolPtr(false),
		},
		{
			name:   "should allow matching role",
			params: AuthParams{Role: stringPtr("admin")},
			options: AuthorizationOptions{
				OrgID:          stringPtr("org1"),
				OrgRole:        stringPtr("admin"),
				OrgPermissions: []string{},
			},
			expected: boolPtr(true),
		},
		{
			name:   "should return nil when missing org context",
			params: AuthParams{Permission: stringPtr("team_settings:manage")},
			options: AuthorizationOptions{
				OrgID: nil,
			},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := checkOrgAuthorization(tt.params, tt.options)
			if (result == nil) != (tt.expected == nil) {
				t.Errorf("checkOrgAuthorization() = %v, want %v", result, tt.expected)
				return
			}
			if result != nil && *result != *tt.expected {
				t.Errorf("checkOrgAuthorization() = %v, want %v", *result, *tt.expected)
			}
		})
	}
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}

func slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
