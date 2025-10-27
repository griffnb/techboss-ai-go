package clerk

import (
	"regexp"
	"slices"
	"strings"

	"github.com/clerk/clerk-sdk-go/v2"
)

// AuthParams represents the parameters for authorization checks
type AuthParams struct {
	Role       *string `json:"role,omitempty"`
	Permission *string `json:"permission,omitempty"`
	Feature    *string `json:"feature,omitempty"`
	Plan       *string `json:"plan,omitempty"`
}

// AuthorizationOptions holds the authorization context

/*
	type AuthorizationOptions = {
	  userId: string | null | undefined;
	  orgId: string | null | undefined;
	  orgRole: string | null | undefined;
	  orgPermissions: string[] | null | undefined;
	  factorVerificationAge: [number, number] | null;
	  features: string | null | undefined;
	  plans: string | null | undefined;
	};
*/
type AuthorizationOptions struct {
	UserID         *string  `json:"userId,omitempty"`
	OrgID          *string  `json:"orgId,omitempty"`
	OrgRole        *string  `json:"orgRole,omitempty"`
	OrgPermissions []string `json:"orgPermissions,omitempty"`
	Features       *string  `json:"features,omitempty"`
	Plans          *string  `json:"plans,omitempty"`
}

// ScopeFeatures represents features split by scope
type ScopeFeatures struct {
	Org  []string
	User []string
}

// Has checks if the user has the specified authorization
func Has(claims *clerk.SessionClaims, params AuthParams) bool {
	if claims == nil {
		return false
	}

	// Build authorization options from claims
	options := buildAuthorizationOptions(claims)

	// Check billing authorization (features/plans)
	if params.Feature != nil || params.Plan != nil {
		if result := checkBillingAuthorization(params, options); result != nil {
			return *result
		}
	}

	// Check organization authorization (roles/permissions)
	if params.Role != nil || params.Permission != nil {
		if result := checkOrgAuthorization(params, options); result != nil {
			return *result
		}
	}

	return false
}

// HasPermission checks if the user has a specific permission
func HasPermission(claims *clerk.SessionClaims, permission string) bool {
	return Has(claims, AuthParams{Permission: &permission})
}

// HasRole checks if the user has a specific role
func HasRole(claims *clerk.SessionClaims, role string) bool {
	return Has(claims, AuthParams{Role: &role})
}

// HasFeature checks if the user has access to a specific feature
func HasFeature(claims *clerk.SessionClaims, feature string) bool {
	return Has(claims, AuthParams{Feature: &feature})
}

// HasPlan checks if the user has access to a specific plan
func HasPlan(claims *clerk.SessionClaims, plan string) bool {
	return Has(claims, AuthParams{Plan: &plan})
}

// buildAuthorizationOptions constructs AuthorizationOptions from claims
func buildAuthorizationOptions(claims *clerk.SessionClaims) AuthorizationOptions {
	options := AuthorizationOptions{}

	if claims.Subject != "" {
		options.UserID = &claims.Subject
	}

	// For now, organization info would need to be populated from elsewhere
	// This is a basic implementation that can be extended based on how
	// organization data is stored in your Clerk setup

	return options
}

// checkOrgAuthorization checks if a user has the required organization-level authorization
/*


/**
 * Checks if a user has the required organization-level authorization.
 * Verifies if the user has the specified role or permission within their organization.
 * @returns null, if unable to determine due to missing data or unspecified role/permission.

 type CheckOrgAuthorization = (
  params: { role?: OrganizationCustomRoleKey; permission?: OrganizationCustomPermissionKey },
  options: Pick<AuthorizationOptions, 'orgId' | 'orgRole' | 'orgPermissions'>,
) => boolean | null;

const checkOrgAuthorization: CheckOrgAuthorization = (params, options) => {
  const { orgId, orgRole, orgPermissions } = options;
  if (!params.role && !params.permission) {
    return null;
  }

  if (!orgId || !orgRole || !orgPermissions) {
    return null;
  }

  if (params.permission) {
    return orgPermissions.includes(prefixWithOrg(params.permission));
  }

  if (params.role) {
    return prefixWithOrg(orgRole) === prefixWithOrg(params.role);
  }
  return null;
};

*/
func checkOrgAuthorization(params AuthParams, options AuthorizationOptions) *bool {
	if params.Role == nil && params.Permission == nil {
		return nil
	}

	if options.OrgID == nil {
		return nil
	}

	if params.Permission != nil {
		if options.OrgRole == nil || len(options.OrgPermissions) == 0 {
			return nil
		}
		for _, perm := range options.OrgPermissions {
			if perm == prefixWithOrg(*params.Permission) {
				result := true
				return &result
			}
		}
		result := false
		return &result
	}

	if params.Role != nil {
		if options.OrgRole == nil {
			return nil
		}
		result := prefixWithOrg(*options.OrgRole) == prefixWithOrg(*params.Role)
		return &result
	}

	return nil
}

// checkBillingAuthorization checks feature and plan authorization

/*
type CheckBillingAuthorization = (

	params: { feature?: string; plan?: string },
	options: Pick<AuthorizationOptions, 'plans' | 'features'>,

) => boolean | null;

	const checkBillingAuthorization: CheckBillingAuthorization = (params, options) => {
	  const { features, plans } = options;

	  if (params.feature && features) {
	    return checkForFeatureOrPlan(features, params.feature);
	  }

	  if (params.plan && plans) {
	    return checkForFeatureOrPlan(plans, params.plan);
	  }
	  return null;
	};
*/
func checkBillingAuthorization(params AuthParams, options AuthorizationOptions) *bool {
	if params.Feature != nil && options.Features != nil {
		result := checkForFeatureOrPlan(*options.Features, *params.Feature)
		return &result
	}

	if params.Plan != nil && options.Plans != nil {
		result := checkForFeatureOrPlan(*options.Plans, *params.Plan)
		return &result
	}

	return nil
}

// checkForFeatureOrPlan checks if a feature or plan is available in the claim
/*

const checkForFeatureOrPlan = (claim: string, featureOrPlan: string) => {
  const { org: orgFeatures, user: userFeatures } = splitByScope(claim);
  const [scope, _id] = featureOrPlan.split(':');
  const id = _id || scope;

  if (scope === 'org') {
    return orgFeatures.includes(id);
  } else if (scope === 'user') {
    return userFeatures.includes(id);
  } else {
    // Since org scoped features will not exist if there is not an active org, merging is safe.
    return [...orgFeatures, ...userFeatures].includes(id);
  }
};

const splitByScope = (fea: string | null | undefined) => {
  const features = fea ? fea.split(',').map(f => f.trim()) : [];

  // TODO: make this more efficient
  return {
    org: features.filter(f => f.split(':')[0].includes('o')).map(f => f.split(':')[1]),
    user: features.filter(f => f.split(':')[0].includes('u')).map(f => f.split(':')[1]),
  };
};

*/
func checkForFeatureOrPlan(claim, featureOrPlan string) bool {
	scopeFeatures := splitByScope(claim)

	parts := strings.Split(featureOrPlan, ":")
	var scope, id string

	if len(parts) >= 2 {
		scope = parts[0]
		id = parts[1]
	} else {
		scope = ""
		id = parts[0]
	}

	switch scope {
	case "org":
		return slices.Contains(scopeFeatures.Org, id)
	case "user":
		return slices.Contains(scopeFeatures.User, id)
	default:
		// Since org scoped features will not exist if there is not an active org, merging is safe
		allFeatures := append(scopeFeatures.Org, scopeFeatures.User...)
		return slices.Contains(allFeatures, id)
	}
}

// splitByScope splits features by their scope (org or user)
func splitByScope(featureString string) ScopeFeatures {
	if featureString == "" {
		return ScopeFeatures{Org: []string{}, User: []string{}}
	}

	features := strings.Split(featureString, ",")
	var orgFeatures, userFeatures []string

	for _, feature := range features {
		feature = strings.TrimSpace(feature)
		parts := strings.Split(feature, ":")
		if len(parts) >= 2 {
			scope := parts[0]
			id := parts[1]

			if strings.Contains(scope, "o") {
				orgFeatures = append(orgFeatures, id)
			} else if strings.Contains(scope, "u") {
				userFeatures = append(userFeatures, id)
			}
		}
	}

	return ScopeFeatures{
		Org:  orgFeatures,
		User: userFeatures,
	}
}

// prefixWithOrg ensures the value is prefixed with "org:"
// const prefixWithOrg = (value: string) => value.replace(/^(org:)*/, 'org:');
func prefixWithOrg(value string) string {
	orgPrefixRegex := regexp.MustCompile(`^(org:)*`)
	return orgPrefixRegex.ReplaceAllString(value, "org:")
}
