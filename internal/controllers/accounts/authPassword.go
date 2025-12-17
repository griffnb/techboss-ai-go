package accounts

import (
	"fmt"
	"net/http"
	"time"

	"github.com/griffnb/core/lib/log"
	"github.com/griffnb/core/lib/router/request"
	"github.com/griffnb/core/lib/router/response"
	"github.com/griffnb/core/lib/session"
	"github.com/griffnb/core/lib/tools"
	"github.com/griffnb/techboss-ai-go/internal/controllers/helpers"
	"github.com/griffnb/techboss-ai-go/internal/environment"
	"github.com/griffnb/techboss-ai-go/internal/integrations/cloudflare"
	"github.com/griffnb/techboss-ai-go/internal/models/account"
	"github.com/griffnb/techboss-ai-go/internal/services/account_service"
	"github.com/griffnb/techboss-ai-go/internal/services/email_sender"
	"github.com/pkg/errors"
)

type PasswordInput struct {
	CurrentPassword      string `json:"current_password"`
	Password             string `json:"password"`
	PasswordConfirmation string `json:"password_confirmation"`
}

// updatePassword allows users to update their password
//
//	@Public
//	@Summary		Update password
//	@Description	Updates the user's password after verifying current password
//	@Tags			Account
//	@Accept			json
//	@Produce		json
//	@Param			body	body	PasswordInput	true	"Password update details"
//	@Success		200	{object}	response.SuccessResponse{data=bool}
//	@Failure		400	{object}	response.ErrorResponse
//	@Router			/auth/password [put]
func updatePassword(_ http.ResponseWriter, req *http.Request) (bool, int, error) {
	//if helpers.IsSuperUpdate(req) {
	//	return response.PublicCustomError[*account.Account]("not allowed to update as super user", http.StatusBadRequest)
	//}
	userSession := request.GetReqSession(req)
	accountObj := helpers.GetLoadedUser(req)

	data, err := request.GetJSONPostAs[*PasswordInput](req)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[bool]()
	}

	if !account.VerifyPassword(accountObj.HashedPassword.Get(), data.CurrentPassword, accountObj.ID()) {
		return response.PublicCustomError[bool]("Current password is incorrect", http.StatusBadRequest)
	}

	if data.Password != data.PasswordConfirmation {
		return response.PublicCustomError[bool]("Passwords do not match", http.StatusBadRequest)
	}

	accountObj.Set("password", data.Password)
	accountObj.PasswordUpdatedAtTS.Set(time.Now().Unix())
	err = accountObj.Save(userSession.User)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[bool]()
	}

	/*
		bgCtx := context.WithoutCancel(req.Context())
		go func() {
			err := emailsender.SendPasswordWasResetEmail(accountObj)
			if err != nil {
				log.ErrorContext(err, bgCtx)
			}
		}()
	*/

	return response.Success(true)
}

type SetPasswordInput struct {
	Verify               string `json:"verify"`
	Password             string `json:"password"`
	PasswordConfirmation string `json:"password_confirmation"`
}

// setPassword sets a password for the first time on an existing account
//
//	@Public
//	@Summary		Set password for new account
//	@Description	Sets a password for the first time on an existing account using invite verification
//	@Tags			Account
//	@Accept			json
//	@Produce		json
//	@Param			body	body	SetPasswordInput	true	"Password setup details"
//	@Success		200	{object}	response.SuccessResponse{data=account.Account}
//	@Failure		400	{object}	response.ErrorResponse
//	@Router			/auth/password/set [post]
func setPassword(_ http.ResponseWriter, req *http.Request) (*account.Account, int, error) {
	//if helpers.IsSuperUpdate(req) {
	//	return response.PublicCustomErrorV2[*account.Account]("not allowed to update as super user", http.StatusBadRequest)
	//}
	userSession := request.GetReqSession(req)
	accountObj, err := account.Get(req.Context(), userSession.User.ID())
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[*account.Account]()
	}

	accountProps, err := accountObj.Properties.Get()
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[*account.Account]()
	}

	// TODO - more statuses should probably block against this but its a quick fix for now
	if accountObj.Status.Get() == account.STATUS_ACTIVE {
		return response.PublicCustomError[*account.Account](
			"Account is not in a state to set a password",
			http.StatusBadRequest,
		)
	}

	data, err := request.GetJSONPostAs[*SetPasswordInput](req)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[*account.Account]()
	}

	if data.Verify != accountProps.InviteKey {
		return response.PublicCustomError[*account.Account](
			"Invalid invite link",
			http.StatusBadRequest,
		)
	}

	if data.Password != data.PasswordConfirmation {
		return response.PublicCustomError[*account.Account]("Passwords do not match", http.StatusBadRequest)
	}

	accountObj.Set("password", data.Password)
	accountObj.PasswordUpdatedAtTS.Set(time.Now().Unix())
	err = accountObj.Save(userSession.User)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[*account.Account]()
	}

	return response.Success[*account.Account](accountObj)
}

type ResendVerifyEmailPayload struct {
	CFToken string `json:"cf_token"`
}

// authResendVerifyEmail resends the verification email to the user
//
//	@Public
//	@Summary		Resend verification email
//	@Description	Resends the email verification link to the user's email address
//	@Tags			Account
//	@Accept			json
//	@Produce		json
//	@Param			body	body	ResendVerifyEmailPayload	true	"Cloudflare token"
//	@Success		200	{object}	response.SuccessResponse{data=bool}
//	@Failure		400	{object}	response.ErrorResponse
//	@Failure		429	{object}	response.ErrorResponse
//	@Router			/auth/verify/resend [post]
func authResendVerifyEmail(_ http.ResponseWriter, req *http.Request) (bool, int, error) {
	//if helpers.IsSuperUpdate(req) {
	//	return response.PublicCustomError[bool]("not allowed to update as super user", http.StatusBadRequest)
	//}

	accountObj := helpers.GetLoadedUser(req)

	locked, err := environment.GetCache().Incr(tools.Sha1(fmt.Sprintf("%s%s", accountObj.Email.Get(), time.Now().Round(time.Minute).String())))
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[bool]()
	}
	if locked > 1 {
		// If we are locked, we return an error
		log.ErrorContext(errors.Errorf("Too many requests for resend verify email"), req.Context())
		return response.PublicCustomError[bool]("Too many requests for resend verify email", http.StatusTooManyRequests)
	}

	if accountObj.IsEmailVerified() {
		return response.PublicCustomError[bool]("Email is already verified", http.StatusBadRequest)
	}

	payload, err := request.GetJSONPostAs[*ResendVerifyEmailPayload](req)
	if err != nil {
		return response.PublicCustomError[bool]("Payload error", http.StatusBadRequest)
	}

	if tools.Empty(payload.CFToken) && !tools.Empty(cloudflare.Client().TurnstileKey) {
		return response.PublicCustomError[bool]("Cloudflare token is required", http.StatusBadRequest)
	}

	if !tools.Empty(cloudflare.Client().TurnstileKey) {
		resp, err := cloudflare.Client().
			ValidateTurnstileResponse(tools.ParseStringI(payload.CFToken), req.RemoteAddr)
		if err != nil {
			log.ErrorContext(err, req.Context())
			return response.PublicBadRequestError[bool]()
		}

		if !resp {
			return response.PublicBadRequestError[bool]()
		}
	}
	err = email_sender.SendVerifyEmail(&accountObj.Account)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[bool]()
	}

	return response.Success(true)
}

type UpdatePrimaryEmailAddressPayload struct {
	Email   string `json:"email"`
	CFToken string `json:"cf_token"`
}

// updatePrimaryEmailAddress updates the user's primary email address
//
//	@Public
//	@Summary		Update primary email
//	@Description	Updates the user's primary email address for unverified accounts
//	@Tags			Account
//	@Accept			json
//	@Produce		json
//	@Param			body	body	UpdatePrimaryEmailAddressPayload	true	"Email update details"
//	@Success		200	{object}	response.SuccessResponse{data=bool}
//	@Failure		400	{object}	response.ErrorResponse
//	@Router			/auth/email [put]
func updatePrimaryEmailAddress(_ http.ResponseWriter, req *http.Request) (bool, int, error) {
	if helpers.IsSuperUpdate(req) {
		return response.PublicCustomError[bool]("not allowed to update as super user", http.StatusBadRequest)
	}

	accountObj := helpers.GetLoadedUser(req)

	if accountObj.IsEmailVerified() {
		return response.PublicCustomError[bool]("Email is already verified", http.StatusBadRequest)
	}

	payload, err := request.GetJSONPostAs[*UpdatePrimaryEmailAddressPayload](req)
	if err != nil {
		return response.PublicCustomError[bool]("Payload error", http.StatusBadRequest)
	}

	if tools.Empty(payload.CFToken) && !tools.Empty(cloudflare.Client().TurnstileKey) {
		return response.PublicCustomError[bool]("Cloudflare token is required", http.StatusBadRequest)
	}

	if !tools.Empty(cloudflare.Client().TurnstileKey) {
		resp, err := cloudflare.Client().
			ValidateTurnstileResponse(tools.ParseStringI(payload.CFToken), req.RemoteAddr)
		if err != nil {
			log.ErrorContext(err, req.Context())
			return response.PublicBadRequestError[bool]()
		}

		if !resp {
			return response.PublicBadRequestError[bool]()
		}
	}

	if tools.Empty(payload.Email) {
		return response.PublicCustomError[bool]("Email address is required", http.StatusBadRequest)
	}

	if tools.IsEqual(accountObj.Email.Get(), payload.Email) {
		return response.Success[bool](true)
	}

	err = account_service.UpdatePrimaryEmailAddressForUnverified(req.Context(), &accountObj.Account, payload.Email)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[bool]()
	}

	accProps, err := accountObj.Properties.Get()
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[bool]()
	}

	// Unset the previous email to prevent abuse with changing emails
	if !tools.Empty(accProps.VerifyEmailKey) {
		verifySession := session.Load(accProps.VerifyEmailKey)
		if !tools.Empty(verifySession) {
			err := verifySession.Invalidate()
			if err != nil {
				log.ErrorContext(err, req.Context())
				// Don't return, we can still send the email
				// The session will expire after 2 days regardless
			}
		}
	}

	err = email_sender.SendVerifyEmail(&accountObj.Account)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[bool]()
	}

	return response.Success(true)
}
