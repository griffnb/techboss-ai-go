package accounts

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/griffnb/core/lib/log"
	"github.com/griffnb/core/lib/router/request"
	"github.com/griffnb/core/lib/router/response"
	"github.com/griffnb/core/lib/session"
	"github.com/griffnb/core/lib/tools"
	"github.com/griffnb/techboss-ai-go/internal/models/account"
	"github.com/griffnb/techboss-ai-go/internal/services/email_sender"
	"github.com/pkg/errors"
)

type PasswordResetInput struct {
	Email string `json:"email"`
	Hash  string `json:"hash"`
}

// openSendResetPasswordEmail sends a password reset email to the user
//
//	@Public
//	@Summary		Send password reset email
//	@Description	Sends a password reset email with a temporary session key
//	@Tags			accounts
//	@Accept			json
//	@Produce		json
//	@Param			body	body		PasswordResetInput	true	"Email and verification hash"
//	@Success		200		{object}	response.SuccessResponse{data=bool}
//	@Failure		400		{object}	response.ErrorResponse
//	@Failure		403		{object}	response.ErrorResponse
//	@Failure		404		{object}	response.ErrorResponse
//	@Router			/password/reset/send [post]
func openSendResetPasswordEmail(
	_ http.ResponseWriter,
	req *http.Request,
) (bool, int, error) {
	data, err := request.GetJSONPostAs[*PasswordResetInput](req)
	if err != nil {
		return response.PublicBadRequestError[bool]()
	}

	if tools.Empty(data.Email) {
		return response.PublicNotFoundError[bool]()
	}

	if tools.Empty(data.Hash) {
		return response.PublicNotFoundError[bool]()
	}

	// simple way to deter bots till we can get something better
	resetEmailHash, err := resetEmailHash(data.Email)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[bool]()
	}

	if resetEmailHash != data.Hash {
		return response.PublicNotFoundError[bool]()
	}

	accountObj, err := account.GetByEmail(req.Context(), data.Email)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[bool]()
	}
	if tools.Empty(accountObj) {
		return response.PublicNotFoundError[bool]()
	}

	if accountObj.Disabled.Bool() {
		return response.PublicCustomError[bool]("Account is disabled", http.StatusForbidden)
	}

	userSession := session.New(tools.ParseStringI(req.Context().Value("ip"))).WithUser(accountObj)
	err = userSession.SaveWithExpiration(time.Now().Add(10 * time.Minute).Unix())
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[bool]()
	}

	err = email_sender.SendResetPasswordEmail(accountObj, userSession.Key)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[bool]()
	}

	return response.Success[bool](true)
}

type ResetPasswordInput struct {
	ResetKey        string `json:"resetKey"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"password_confirmation"`
}

// openResetPassword resets the user's password using a reset key
//
//	@Public
//	@Summary		Reset password
//	@Description	Resets the user's password using the reset key from email
//	@Tags			accounts
//	@Accept			json
//	@Produce		json
//	@Param			body	body		ResetPasswordInput	true	"Password reset details"
//	@Success		200		{object}	response.SuccessResponse{data=bool}
//	@Failure		400		{object}	response.ErrorResponse
//	@Failure		403		{object}	response.ErrorResponse
//	@Failure		404		{object}	response.ErrorResponse
//	@Router			/password/reset [post]
func openResetPassword(_ http.ResponseWriter, req *http.Request) (bool, int, error) {
	data, err := request.GetJSONPostAs[*ResetPasswordInput](req)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[bool]()
	}

	if tools.Empty(data.ResetKey) || tools.Empty(data.Password) || tools.Empty(data.ConfirmPassword) {
		return response.PublicBadRequestError[bool]()
	}

	sessionObj := session.Load(data.ResetKey)

	if tools.Empty(sessionObj) {
		log.ErrorContext(errors.Errorf("session not found for key: %s", data.ResetKey), req.Context())
		return response.PublicNotFoundError[bool]()
	}

	accountObj, err := account.Get(req.Context(), sessionObj.User.ID())
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[bool]()
	}
	if tools.Empty(accountObj) {
		log.ErrorContext(errors.Errorf("account not found for key: %s account:%s", data.ResetKey, sessionObj.User.ID()), req.Context())
		return response.PublicNotFoundError[bool]()
	}

	if accountObj.Disabled.Bool() {
		return response.PublicCustomError[bool]("Account is disabled", http.StatusForbidden)
	}

	if data.Password != data.ConfirmPassword {
		return response.PublicCustomError[bool]("Passwords do not match", http.StatusBadRequest)
	}

	accountObj.Set("password", data.Password)
	accountObj.PasswordUpdatedAtTS.Set(time.Now().Unix())
	err = accountObj.Save(sessionObj.User)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[bool]()
	}

	bgCtx := context.WithoutCancel(req.Context())
	go func(accountObj *account.Account) {
		err := email_sender.SendPasswordWasResetEmail(accountObj)
		if err != nil {
			log.ErrorContext(err, bgCtx)
		}
	}(accountObj)

	return response.Success[bool](true)
}

type CheckKeyResponse struct {
	Valid bool `json:"valid"`
}

// openCheckKey validates if a reset key is still valid
//
//	@Summary		Check reset key validity
//	@Description	Validates whether a password reset key is still valid and active
//	@Tags			Account
//	@Accept			json
//	@Produce		json
//	@Param			key	query		string	true	"Reset key to validate"
//	@Success		200	{object}	response.SuccessResponse{data=CheckKeyResponse}
//	@Failure		400	{object}	response.ErrorResponse
//	@Router			/password/check [get]
func openCheckKey(_ http.ResponseWriter, req *http.Request) (*CheckKeyResponse, int, error) {
	key := req.URL.Query().Get("key")

	if tools.Empty(key) {
		return response.PublicBadRequestError[*CheckKeyResponse]()
	}

	sessionObj := session.Load(key)

	if tools.Empty(sessionObj) {
		return response.Success(&CheckKeyResponse{Valid: false})
	}

	accountObj, err := account.Get(req.Context(), sessionObj.User.ID())
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[*CheckKeyResponse]()
	}
	if tools.Empty(accountObj) {
		log.ErrorContext(errors.Errorf("account not found for key: %s account:%s", key, sessionObj.User.ID()), req.Context())
		return response.Success(&CheckKeyResponse{Valid: false})
	}

	return response.Success(&CheckKeyResponse{Valid: true})
}

// tmp thing to hash email for now
func reverseString(s string) string {
	// Convert the string to a slice of runes.
	runes := []rune(s)

	// Reverse the runes.
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}

	// Convert the slice of runes back into a string and return it.
	return string(runes)
}

// tmp thing to hash email for now
func resetEmailHash(email string) (string, error) {
	reverseEmail := reverseString(email)

	// Create a new SHA1 instance
	h := sha256.New()

	// Write the input string to the SHA1 instance
	_, err := io.WriteString(h, reverseEmail)
	if err != nil {
		return "", errors.WithStack(err)
	}

	// Calculate the SHA1 checksum
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
