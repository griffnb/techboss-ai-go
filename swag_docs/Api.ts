/* eslint-disable */
/* tslint:disable */
// @ts-nocheck
/*
 * ---------------------------------------------------------------
 * ## THIS FILE WAS GENERATED VIA SWAGGER-TYPESCRIPT-API        ##
 * ##                                                           ##
 * ## AUTHOR: acacode                                           ##
 * ## SOURCE: https://github.com/acacode/swagger-typescript-api ##
 * ---------------------------------------------------------------
 */

export interface AdminAccountJoined {
  created_at?: string;
  created_by_name?: string;
  created_by_urn?: string;
  deleted?: number;
  disabled?: number;
  email?: string;
  email_verified_at_ts?: number;
  external_id?: string;
  first_name?: string;
  hashed_password?: string;
  id?: string;
  is_super_user_session?: number;
  last_login_ts?: number;
  last_name?: string;
  name?: string;
  organization_id?: string;
  organization_name?: string;
  password_updated_at_ts?: number;
  phone?: string;
  properties?: AccountProperties;
  role?: number;
  signup_properties?: AccountSignupProperties;
  status?: number;
  test_user_type?: number;
  updated_at?: string;
  updated_by_name?: string;
  updated_by_urn?: string;
  urn?: string;
}

export interface AccountAccount {
  /** @format date-time */
  created_at: string;
  created_by_urn: string;
  deleted: number;
  disabled: number;
  email: string;
  email_verified_at_ts: number;
  external_id: string;
  first_name: string;
  hashed_password: string;
  /** @format uuid */
  id: string;
  is_super_user_session: number;
  last_login_ts: number;
  last_name: string;
  /** @format uuid */
  organization_id: string;
  password_updated_at_ts: number;
  phone: string;
  properties: AccountProperties;
  role: number;
  signup_properties: AccountSignupProperties;
  status: number;
  test_user_type: number;
  /** @format date-time */
  updated_at: string;
  updated_by_urn: string;
  urn: string;
}

export interface AccountAccountJoined {
  /** @format date-time */
  created_at: string;
  created_by_name: string;
  created_by_urn: string;
  deleted: number;
  disabled: number;
  email: string;
  email_verified_at_ts: number;
  external_id: string;
  first_name: string;
  hashed_password: string;
  /** @format uuid */
  id: string;
  is_super_user_session: number;
  last_login_ts: number;
  last_name: string;
  name: string;
  /** @format uuid */
  organization_id: string;
  organization_name: string;
  password_updated_at_ts: number;
  phone: string;
  properties: AccountProperties;
  role: number;
  signup_properties: AccountSignupProperties;
  status: number;
  test_user_type: number;
  /** @format date-time */
  updated_at: string;
  updated_by_name: string;
  updated_by_urn: string;
  urn: string;
}

export interface AccountAccountJoinedPublic {
  /** @format date-time */
  created_at: string;
  created_by_urn: string;
  deleted: number;
  disabled: number;
  email: string;
  external_id: string;
  first_name: string;
  /** @format uuid */
  id: string;
  is_super_user_session: number;
  last_name: string;
  /** @format uuid */
  organization_id: string;
  phone: string;
  role: number;
  status: number;
  test_user_type: number;
  /** @format date-time */
  updated_at: string;
  updated_by_urn: string;
  urn: string;
}

export interface AccountAccountPublic {
  /** @format date-time */
  created_at: string;
  created_by_urn: string;
  deleted: number;
  disabled: number;
  email: string;
  external_id: string;
  first_name: string;
  /** @format uuid */
  id: string;
  is_super_user_session: number;
  last_name: string;
  /** @format uuid */
  organization_id: string;
  phone: string;
  role: number;
  status: number;
  test_user_type: number;
  /** @format date-time */
  updated_at: string;
  updated_by_urn: string;
  urn: string;
}

export interface AccountAccountWithFeatures {
  billing_plan_feature_set: BillingPlanFeatureSet;
  /** @format uuid */
  billing_plan_id: string;
  billing_plan_level: number;
  billing_plan_name: string;
  billing_plan_price: number;
  /** @format date-time */
  created_at: string;
  created_by_urn: string;
  deleted: number;
  disabled: number;
  email: string;
  email_verified_at_ts: number;
  external_id: string;
  feature_set: BillingPlanFeatureSet;
  feature_set_overrides: BillingPlanMergeableFeatureSet;
  first_name: string;
  hashed_password: string;
  /** @format uuid */
  id: string;
  is_super_user_session: number;
  last_login_ts: number;
  last_name: string;
  /** @format uuid */
  organization_id: string;
  password_updated_at_ts: number;
  phone: string;
  properties: AccountProperties;
  role: number;
  signup_properties: AccountSignupProperties;
  status: number;
  test_user_type: number;
  /** @format date-time */
  updated_at: string;
  updated_by_urn: string;
  urn: string;
}

export interface AccountAccountWithFeaturesPublic {
  /** @format uuid */
  billing_plan_id: string;
  billing_plan_level: number;
  billing_plan_name: string;
  billing_plan_price: number;
  /** @format date-time */
  created_at: string;
  created_by_urn: string;
  deleted: number;
  disabled: number;
  email: string;
  external_id: string;
  feature_set: BillingPlanFeatureSetPublic;
  first_name: string;
  /** @format uuid */
  id: string;
  is_super_user_session: number;
  last_name: string;
  /** @format uuid */
  organization_id: string;
  phone: string;
  role: number;
  status: number;
  test_user_type: number;
  /** @format date-time */
  updated_at: string;
  updated_by_urn: string;
  urn: string;
}

export interface AccountProperties {
  external_user_info?: any;
  invite_key?: string;
  /** @format int64 */
  invite_ts?: number;
  /** @format int64 */
  last_seen?: number;
  verify_email_key?: string;
}

export type AccountPropertiesPublic = object;

export interface AccountSignupProperties {
  /** @format int64 */
  is_oauth: number;
}

export type AccountSignupPropertiesPublic = object;

export interface AccountServiceTestUserInput {
  organization_id?: string;
}

export interface AccountsAPIResponse {
  account?: AccountAccount;
  organization?: OrganizationOrganization;
}

export interface AccountsCheckKeyResponse {
  valid?: boolean;
}

export interface AccountsExistingCheck {
  cf_token?: string;
  email?: string;
}

export interface AccountsExistingResponse {
  exists?: boolean;
}

export interface AccountsPasswordInput {
  current_password?: string;
  password?: string;
  password_confirmation?: string;
}

export interface AccountsPasswordResetInput {
  email?: string;
  hash?: string;
}

export interface AccountsResendVerifyEmailPayload {
  cf_token?: string;
}

export interface AccountsResetPasswordInput {
  password?: string;
  password_confirmation?: string;
  resetKey?: string;
}

export interface AccountsSetPasswordInput {
  password?: string;
  password_confirmation?: string;
  verify?: string;
}

export interface AccountsSignupResponse {
  redirect_url?: string;
  token?: string;
  verification_token?: string;
}

export interface AccountsUpdatePrimaryEmailAddressPayload {
  cf_token?: string;
  email?: string;
}

export interface AccountsVerifyInput {
  email?: string;
  verify?: string;
}

export interface AccountsVerifyResponse {
  redirect_url?: string;
  token?: string;
}

export interface BillingPlanFeatureSet {
  /** @format int64 */
  advanced_analytics?: number;
  /** @format int64 */
  custom_branding?: number;
  /** @format int64 */
  priority_support?: number;
}

export interface BillingPlanFeatureSetPublic {
  /** @format int64 */
  advanced_analytics?: number;
}

export type BillingPlanMergeableFeatureSet = object;

export type BillingPlanMergeableFeatureSetPublic = object;

export interface OrganizationMetaData {
  onboard_answers?: Record<string, any>;
  vector_store_id?: Record<string, string>;
}

export interface OrganizationMetaDataPublic {
  onboard_answers?: Record<string, any>;
  vector_store_id?: Record<string, string>;
}

export interface OrganizationOrganization {
  /** @format uuid */
  billing_plan_price_id: string;
  /** @format date-time */
  created_at: string;
  created_by_urn: string;
  deleted: number;
  disabled: number;
  email_domains: string[];
  external_id: string;
  feature_set_overrides: BillingPlanFeatureSet;
  /** @format uuid */
  id: string;
  meta_data: OrganizationMetaData;
  name: string;
  properties: OrganizationProperties;
  status: number;
  stripe_id: string;
  subdomain: string;
  /** @format date-time */
  updated_at: string;
  updated_by_urn: string;
  urn: string;
}

export interface OrganizationOrganizationPublic {
  /** @format uuid */
  billing_plan_price_id: string;
  /** @format date-time */
  created_at: string;
  created_by_urn: string;
  deleted: number;
  disabled: number;
  /** @format uuid */
  id: string;
  meta_data: OrganizationMetaDataPublic;
  name: string;
  properties: OrganizationPropertiesPublic;
  status: number;
  /** @format date-time */
  updated_at: string;
  updated_by_urn: string;
  urn: string;
}

export interface OrganizationProperties {
  billing_email?: string;
}

export type OrganizationPropertiesPublic = object;

export interface ResponseErrorResponse {
  error?: string;
  /** @default false */
  success?: boolean;
}

export interface ResponseSuccessResponse {
  data?: any;
  /** @default true */
  success?: boolean;
}

export type QueryParamsType = Record<string | number, any>;
export type ResponseFormat = keyof Omit<Body, "body" | "bodyUsed">;

export interface FullRequestParams extends Omit<RequestInit, "body"> {
  /** set parameter to `true` for call `securityWorker` for this request */
  secure?: boolean;
  /** request path */
  path: string;
  /** content type of request body */
  type?: ContentType;
  /** query params */
  query?: QueryParamsType;
  /** format of response (i.e. response.json() -> format: "json") */
  format?: ResponseFormat;
  /** request body */
  body?: unknown;
  /** base url */
  baseUrl?: string;
  /** request cancellation token */
  cancelToken?: CancelToken;
}

export type RequestParams = Omit<
  FullRequestParams,
  "body" | "method" | "query" | "path"
>;

export interface ApiConfig<SecurityDataType = unknown> {
  baseUrl?: string;
  baseApiParams?: Omit<RequestParams, "baseUrl" | "cancelToken" | "signal">;
  securityWorker?: (
    securityData: SecurityDataType | null,
  ) => Promise<RequestParams | void> | RequestParams | void;
  customFetch?: typeof fetch;
}

export interface HttpResponse<D extends unknown, E extends unknown = unknown>
  extends Response {
  data: D;
  error: E;
}

type CancelToken = Symbol | string | number;

export enum ContentType {
  Json = "application/json",
  JsonApi = "application/vnd.api+json",
  FormData = "multipart/form-data",
  UrlEncoded = "application/x-www-form-urlencoded",
  Text = "text/plain",
}

export class HttpClient<SecurityDataType = unknown> {
  public baseUrl: string = "";
  private securityData: SecurityDataType | null = null;
  private securityWorker?: ApiConfig<SecurityDataType>["securityWorker"];
  private abortControllers = new Map<CancelToken, AbortController>();
  private customFetch = (...fetchParams: Parameters<typeof fetch>) =>
    fetch(...fetchParams);

  private baseApiParams: RequestParams = {
    credentials: "same-origin",
    headers: {},
    redirect: "follow",
    referrerPolicy: "no-referrer",
  };

  constructor(apiConfig: ApiConfig<SecurityDataType> = {}) {
    Object.assign(this, apiConfig);
  }

  public setSecurityData = (data: SecurityDataType | null) => {
    this.securityData = data;
  };

  protected encodeQueryParam(key: string, value: any) {
    const encodedKey = encodeURIComponent(key);
    return `${encodedKey}=${encodeURIComponent(typeof value === "number" ? value : `${value}`)}`;
  }

  protected addQueryParam(query: QueryParamsType, key: string) {
    return this.encodeQueryParam(key, query[key]);
  }

  protected addArrayQueryParam(query: QueryParamsType, key: string) {
    const value = query[key];
    return value.map((v: any) => this.encodeQueryParam(key, v)).join("&");
  }

  protected toQueryString(rawQuery?: QueryParamsType): string {
    const query = rawQuery || {};
    const keys = Object.keys(query).filter(
      (key) => "undefined" !== typeof query[key],
    );
    return keys
      .map((key) =>
        Array.isArray(query[key])
          ? this.addArrayQueryParam(query, key)
          : this.addQueryParam(query, key),
      )
      .join("&");
  }

  protected addQueryParams(rawQuery?: QueryParamsType): string {
    const queryString = this.toQueryString(rawQuery);
    return queryString ? `?${queryString}` : "";
  }

  private contentFormatters: Record<ContentType, (input: any) => any> = {
    [ContentType.Json]: (input: any) =>
      input !== null && (typeof input === "object" || typeof input === "string")
        ? JSON.stringify(input)
        : input,
    [ContentType.JsonApi]: (input: any) =>
      input !== null && (typeof input === "object" || typeof input === "string")
        ? JSON.stringify(input)
        : input,
    [ContentType.Text]: (input: any) =>
      input !== null && typeof input !== "string"
        ? JSON.stringify(input)
        : input,
    [ContentType.FormData]: (input: any) => {
      if (input instanceof FormData) {
        return input;
      }

      return Object.keys(input || {}).reduce((formData, key) => {
        const property = input[key];
        formData.append(
          key,
          property instanceof Blob
            ? property
            : typeof property === "object" && property !== null
              ? JSON.stringify(property)
              : `${property}`,
        );
        return formData;
      }, new FormData());
    },
    [ContentType.UrlEncoded]: (input: any) => this.toQueryString(input),
  };

  protected mergeRequestParams(
    params1: RequestParams,
    params2?: RequestParams,
  ): RequestParams {
    return {
      ...this.baseApiParams,
      ...params1,
      ...(params2 || {}),
      headers: {
        ...(this.baseApiParams.headers || {}),
        ...(params1.headers || {}),
        ...((params2 && params2.headers) || {}),
      },
    };
  }

  protected createAbortSignal = (
    cancelToken: CancelToken,
  ): AbortSignal | undefined => {
    if (this.abortControllers.has(cancelToken)) {
      const abortController = this.abortControllers.get(cancelToken);
      if (abortController) {
        return abortController.signal;
      }
      return void 0;
    }

    const abortController = new AbortController();
    this.abortControllers.set(cancelToken, abortController);
    return abortController.signal;
  };

  public abortRequest = (cancelToken: CancelToken) => {
    const abortController = this.abortControllers.get(cancelToken);

    if (abortController) {
      abortController.abort();
      this.abortControllers.delete(cancelToken);
    }
  };

  public request = async <T = any, E = any>({
    body,
    secure,
    path,
    type,
    query,
    format,
    baseUrl,
    cancelToken,
    ...params
  }: FullRequestParams): Promise<HttpResponse<T, E>> => {
    const secureParams =
      ((typeof secure === "boolean" ? secure : this.baseApiParams.secure) &&
        this.securityWorker &&
        (await this.securityWorker(this.securityData))) ||
      {};
    const requestParams = this.mergeRequestParams(params, secureParams);
    const queryString = query && this.toQueryString(query);
    const payloadFormatter = this.contentFormatters[type || ContentType.Json];
    const responseFormat = format || requestParams.format;

    return this.customFetch(
      `${baseUrl || this.baseUrl || ""}${path}${queryString ? `?${queryString}` : ""}`,
      {
        ...requestParams,
        headers: {
          ...(requestParams.headers || {}),
          ...(type && type !== ContentType.FormData
            ? { "Content-Type": type }
            : {}),
        },
        signal:
          (cancelToken
            ? this.createAbortSignal(cancelToken)
            : requestParams.signal) || null,
        body:
          typeof body === "undefined" || body === null
            ? null
            : payloadFormatter(body),
      },
    ).then(async (response) => {
      const r = response as HttpResponse<T, E>;
      r.data = null as unknown as T;
      r.error = null as unknown as E;

      const responseToParse = responseFormat ? response.clone() : response;
      const data = !responseFormat
        ? r
        : await responseToParse[responseFormat]()
            .then((data) => {
              if (r.ok) {
                r.data = data;
              } else {
                r.error = data;
              }
              return r;
            })
            .catch((e) => {
              r.error = e;
              return r;
            });

      if (cancelToken) {
        this.abortControllers.delete(cancelToken);
      }

      if (!response.ok) throw data;
      return data;
    });
  };
}

/**
 * @title No title
 * @contact
 */
export class Api<
  SecurityDataType extends unknown,
> extends HttpClient<SecurityDataType> {
  account = {
    /**
     * @description List Account
     *
     * @tags Account
     * @name AccountList
     * @summary List Account
     * @request GET:/account
     */
    accountList: (
      query?: {
        /** search by q */
        q?: string;
        /**
         * limit results
         * @min 1
         * @max 1000
         * @default 100
         */
        limit?: number;
        /**
         * offset results
         * @min 0
         * @default 0
         */
        offset?: number;
        /**
         * sort results e.g. 'created_at desc'
         * @default "created_at desc"
         */
        order?: string;
        /** filters, see readme */
        filters?: string;
      },
      params: RequestParams = {},
    ) =>
      this.request<
        ResponseSuccessResponse & {
          data?: AccountAccountJoinedPublic[];
        },
        ResponseErrorResponse
      >({
        path: `/account`,
        method: "GET",
        query: query,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Create Account
     *
     * @tags Account
     * @name AccountCreate
     * @summary Create Account
     * @request POST:/account/
     */
    accountCreate: (data: AccountAccountPublic, params: RequestParams = {}) =>
      this.request<
        ResponseSuccessResponse & {
          data?: AccountAccountPublic;
        },
        ResponseErrorResponse
      >({
        path: `/account/`,
        method: "POST",
        body: data,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Checks whether an email address is already registered
     *
     * @tags Account
     * @name CheckCreate
     * @summary Check existing email
     * @request POST:/account/check
     */
    checkCreate: (body: AccountsExistingCheck, params: RequestParams = {}) =>
      this.request<
        ResponseSuccessResponse & {
          data?: AccountsExistingResponse;
        },
        ResponseErrorResponse
      >({
        path: `/account/check`,
        method: "POST",
        body: body,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Get Account
     *
     * @tags Account
     * @name AccountDetail
     * @summary Get Account
     * @request GET:/account/{id}
     */
    accountDetail: (id: string, params: RequestParams = {}) =>
      this.request<
        ResponseSuccessResponse & {
          data?: AccountAccountJoinedPublic;
        },
        ResponseErrorResponse
      >({
        path: `/account/${id}`,
        method: "GET",
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Update Account
     *
     * @tags Account
     * @name AccountUpdate
     * @summary Update Account
     * @request PUT:/account/{id}
     */
    accountUpdate: (
      id: string,
      data: AccountAccountPublic,
      params: RequestParams = {},
    ) =>
      this.request<
        ResponseSuccessResponse & {
          data?: AccountAccountPublic;
        },
        ResponseErrorResponse
      >({
        path: `/account/${id}`,
        method: "PUT",
        body: data,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Marks an account as deleted by setting status to USER_DELETED
     *
     * @tags Account
     * @name AccountDelete
     * @summary Delete account
     * @request DELETE:/account/{id}
     */
    accountDelete: (id: string, params: RequestParams = {}) =>
      this.request<
        ResponseSuccessResponse & {
          data?: AdminAccountJoined;
        },
        ResponseErrorResponse
      >({
        path: `/account/${id}`,
        method: "DELETE",
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Cancels a pending invitation and invalidates the invite session
     *
     * @tags Account
     * @name CancelCreate
     * @summary Cancel invite
     * @request POST:/account/{id}/cancel
     */
    cancelCreate: (id: string, params: RequestParams = {}) =>
      this.request<
        ResponseSuccessResponse & {
          data?: AdminAccountJoined;
        },
        ResponseErrorResponse
      >({
        path: `/account/${id}/cancel`,
        method: "POST",
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Resends invitation email to a pending account
     *
     * @tags Account
     * @name ResendCreate
     * @summary Resend invite
     * @request POST:/account/{id}/resend
     */
    resendCreate: (id: string, params: RequestParams = {}) =>
      this.request<
        ResponseSuccessResponse & {
          data?: AccountAccountJoinedPublic;
        },
        ResponseErrorResponse
      >({
        path: `/account/${id}/resend`,
        method: "POST",
        type: ContentType.Json,
        format: "json",
        ...params,
      }),
  };
  admin = {
    /**
     * @description List Account
     *
     * @tags Account, Admin
     * @name AccountList
     * @summary List Account
     * @request GET:/admin/account
     */
    accountList: (
      query?: {
        /** search by q */
        q?: string;
        /**
         * limit results
         * @min 1
         * @max 1000
         * @default 100
         */
        limit?: number;
        /**
         * offset results
         * @min 0
         * @default 0
         */
        offset?: number;
        /**
         * sort results e.g. 'created_at desc'
         * @default "created_at desc"
         */
        order?: string;
        /** filters, see readme */
        filters?: string;
      },
      params: RequestParams = {},
    ) =>
      this.request<
        ResponseSuccessResponse & {
          data?: AccountAccountJoined[];
        },
        ResponseErrorResponse
      >({
        path: `/admin/account`,
        method: "GET",
        query: query,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Create Account
     *
     * @tags Account, Admin
     * @name AccountCreate
     * @summary Create Account
     * @request POST:/admin/account/
     */
    accountCreate: (data: AccountAccount, params: RequestParams = {}) =>
      this.request<
        ResponseSuccessResponse & {
          data?: AccountAccount;
        },
        ResponseErrorResponse
      >({
        path: `/admin/account/`,
        method: "POST",
        body: data,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Count Account
     *
     * @tags Account, Admin
     * @name AccountCountList
     * @summary Count Account
     * @request GET:/admin/account/count
     */
    accountCountList: (
      query?: {
        /** search by q */
        q?: string;
        /** filters, see readme */
        filters?: string;
      },
      params: RequestParams = {},
    ) =>
      this.request<
        ResponseSuccessResponse & {
          /** @format int64 */
          data?: number;
        },
        ResponseErrorResponse
      >({
        path: `/admin/account/count`,
        method: "GET",
        query: query,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Get Account
     *
     * @tags Account, Admin
     * @name AccountDetail
     * @summary Get Account
     * @request GET:/admin/account/{id}
     */
    accountDetail: (id: string, params: RequestParams = {}) =>
      this.request<
        ResponseSuccessResponse & {
          data?: AccountAccountJoined;
        },
        ResponseErrorResponse
      >({
        path: `/admin/account/${id}`,
        method: "GET",
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Update Account
     *
     * @tags Account, Admin
     * @name AccountUpdate
     * @summary Update Account
     * @request PUT:/admin/account/{id}
     */
    accountUpdate: (
      id: string,
      data: AccountAccount,
      params: RequestParams = {},
    ) =>
      this.request<
        ResponseSuccessResponse & {
          data?: AccountAccount;
        },
        ResponseErrorResponse
      >({
        path: `/admin/account/${id}`,
        method: "PUT",
        body: data,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Creates a test account with the provided details and optionally logs in as that account
     *
     * @tags Account, Admin
     * @name TestUserCreate
     * @summary Create test account
     * @request POST:/admin/testUser
     */
    testUserCreate: (
      body: AccountServiceTestUserInput,
      query?: {
        /** Set to login as the created user */
        login_as?: string;
      },
      params: RequestParams = {},
    ) =>
      this.request<
        ResponseSuccessResponse & {
          data?: AccountAccount;
        },
        ResponseErrorResponse
      >({
        path: `/admin/testUser`,
        method: "POST",
        query: query,
        body: body,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),
  };
  api = {
    /**
     * @description Retrieves account and associated organization data by account ID
     *
     * @tags Account
     * @name AccountDetail
     * @summary Get account with organization
     * @request GET:/api/account/{id}
     */
    accountDetail: (id: string, params: RequestParams = {}) =>
      this.request<
        ResponseSuccessResponse & {
          data?: AccountsAPIResponse;
        },
        ResponseErrorResponse
      >({
        path: `/api/account/${id}`,
        method: "GET",
        type: ContentType.Json,
        format: "json",
        ...params,
      }),
  };
  auth = {
    /**
     * @description Updates the user's primary email address for unverified accounts
     *
     * @tags Account
     * @name EmailUpdate
     * @summary Update primary email
     * @request PUT:/auth/email
     */
    emailUpdate: (
      body: AccountsUpdatePrimaryEmailAddressPayload,
      params: RequestParams = {},
    ) =>
      this.request<
        ResponseSuccessResponse & {
          data?: boolean;
        },
        ResponseErrorResponse
      >({
        path: `/auth/email`,
        method: "PUT",
        body: body,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Updates the user's password after verifying current password
     *
     * @tags Account
     * @name PasswordUpdate
     * @summary Update password
     * @request PUT:/auth/password
     */
    passwordUpdate: (body: AccountsPasswordInput, params: RequestParams = {}) =>
      this.request<
        ResponseSuccessResponse & {
          data?: boolean;
        },
        ResponseErrorResponse
      >({
        path: `/auth/password`,
        method: "PUT",
        body: body,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Sets a password for the first time on an existing account using invite verification
     *
     * @tags Account
     * @name PasswordSetCreate
     * @summary Set password for new account
     * @request POST:/auth/password/set
     */
    passwordSetCreate: (
      body: AccountsSetPasswordInput,
      params: RequestParams = {},
    ) =>
      this.request<
        ResponseSuccessResponse & {
          data?: AccountAccountPublic;
        },
        ResponseErrorResponse
      >({
        path: `/auth/password/set`,
        method: "POST",
        body: body,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Resends the email verification link to the user's email address
     *
     * @tags Account
     * @name VerifyResendCreate
     * @summary Resend verification email
     * @request POST:/auth/verify/resend
     */
    verifyResendCreate: (
      body: AccountsResendVerifyEmailPayload,
      params: RequestParams = {},
    ) =>
      this.request<
        ResponseSuccessResponse & {
          data?: boolean;
        },
        ResponseErrorResponse
      >({
        path: `/auth/verify/resend`,
        method: "POST",
        body: body,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),
  };
  me = {
    /**
     * @description Retrieves the authenticated user's account details with joined data
     *
     * @tags Account
     * @name GetMe
     * @summary Get current user
     * @request GET:/me
     */
    getMe: (params: RequestParams = {}) =>
      this.request<
        ResponseSuccessResponse & {
          data?: AccountAccountWithFeaturesPublic;
        },
        ResponseErrorResponse
      >({
        path: `/me`,
        method: "GET",
        type: ContentType.Json,
        format: "json",
        ...params,
      }),
  };
  password = {
    /**
     * @description Validates whether a password reset key is still valid and active
     *
     * @tags Account
     * @name CheckList
     * @summary Check reset key validity
     * @request GET:/password/check
     */
    checkList: (
      query: {
        /** Reset key to validate */
        key: string;
      },
      params: RequestParams = {},
    ) =>
      this.request<
        ResponseSuccessResponse & {
          data?: AccountsCheckKeyResponse;
        },
        ResponseErrorResponse
      >({
        path: `/password/check`,
        method: "GET",
        query: query,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Resets the user's password using the reset key from email
     *
     * @tags Account
     * @name ResetCreate
     * @summary Reset password
     * @request POST:/password/reset
     */
    resetCreate: (
      body: AccountsResetPasswordInput,
      params: RequestParams = {},
    ) =>
      this.request<
        ResponseSuccessResponse & {
          data?: boolean;
        },
        ResponseErrorResponse
      >({
        path: `/password/reset`,
        method: "POST",
        body: body,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Sends a password reset email with a temporary session key
     *
     * @tags Account
     * @name ResetSendCreate
     * @summary Send password reset email
     * @request POST:/password/reset/send
     */
    resetSendCreate: (
      body: AccountsPasswordResetInput,
      params: RequestParams = {},
    ) =>
      this.request<
        ResponseSuccessResponse & {
          data?: boolean;
        },
        ResponseErrorResponse
      >({
        path: `/password/reset/send`,
        method: "POST",
        body: body,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),
  };
  signup = {
    /**
     * @description Creates a new account with email and password or completes invited user signup
     *
     * @tags Account
     * @name SignupCreate
     * @summary Standard signup
     * @request POST:/signup
     */
    signupCreate: (body: AccountPropertiesPublic, params: RequestParams = {}) =>
      this.request<
        ResponseSuccessResponse & {
          data?: AccountsSignupResponse;
        },
        ResponseErrorResponse
      >({
        path: `/signup`,
        method: "POST",
        body: body,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Completes signup process for OAuth authenticated users
     *
     * @tags Account
     * @name OauthCreate
     * @summary OAuth signup
     * @request POST:/signup/oauth
     */
    oauthCreate: (body: AccountPropertiesPublic, params: RequestParams = {}) =>
      this.request<
        ResponseSuccessResponse & {
          data?: AccountsSignupResponse;
        },
        ResponseErrorResponse
      >({
        path: `/signup/oauth`,
        method: "POST",
        body: body,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),
  };
  verify = {
    /**
     * @description Verifies a user's email address using the verification key and creates a user session
     *
     * @tags Account
     * @name EmailCreate
     * @summary Verify email address
     * @request POST:/verify/email
     */
    emailCreate: (body: AccountsVerifyInput, params: RequestParams = {}) =>
      this.request<
        ResponseSuccessResponse & {
          data?: AccountsVerifyResponse;
        },
        ResponseErrorResponse
      >({
        path: `/verify/email`,
        method: "POST",
        body: body,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Verifies an account invitation using the verification key and creates a user session
     *
     * @tags Account
     * @name InviteCreate
     * @summary Verify account invite
     * @request POST:/verify/invite
     */
    inviteCreate: (body: AccountsVerifyInput, params: RequestParams = {}) =>
      this.request<
        ResponseSuccessResponse & {
          data?: AccountsVerifyResponse;
        },
        ResponseErrorResponse
      >({
        path: `/verify/invite`,
        method: "POST",
        body: body,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),
  };
}
