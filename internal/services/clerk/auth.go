


func Has(claims *clerk.SessionClaims, featureOrPlan string) bool {


}


type AuthorizationOptions = {
  userId: string | null | undefined;
  orgId: string | null | undefined;
  orgRole: string | null | undefined;
  orgPermissions: string[] | null | undefined;
  factorVerificationAge: [number, number] | null;
  features: string | null | undefined;
  plans: string | null | undefined;
};

/**
 * Checks if a user has the required organization-level authorization.
 * Verifies if the user has the specified role or permission within their organization.
 * @returns null, if unable to determine due to missing data or unspecified role/permission.
 
 type CheckOrgAuthorization = (
  params: { role?: OrganizationCustomRoleKey; permission?: OrganizationCustomPermissionKey },
  options: Pick<AuthorizationOptions, 'orgId' | 'orgRole' | 'orgPermissions'>,
) => boolean | null;
 */
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


const prefixWithOrg = (value: string) => value.replace(/^(org:)*/, 'org:');


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

/*

type CheckBillingAuthorization = (
  params: { feature?: string; plan?: string },
  options: Pick<AuthorizationOptions, 'plans' | 'features'>,
) => boolean | null;
 */

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