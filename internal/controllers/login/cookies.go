package login

import (
	"crypto/hmac"
	"crypto/sha256"
	b64 "encoding/base64"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/CrowdShield/go-core/lib/tools"
	senv "github.com/griffnb/techboss-ai-go/internal/environment"
)

func SendSessionCookie(res http.ResponseWriter, cookieKey string, sessionID string) {
	SendCookies(res, cookieKey, sessionID)
}

func SendOrgCookie(res http.ResponseWriter, orgID string) {
	SendCookies(res, "organization_id", orgID)
}

func SendCookies(res http.ResponseWriter, key, value string) {
	cookieDomain := senv.GetConfig().Server.Domain
	secure := senv.GetConfig().Server.Secure
	var cookieDomainWithDot string
	// Prepend a dot to make the cookie available to all subdomains
	if !strings.HasPrefix(cookieDomain, ".") {
		cookieDomainWithDot = "." + cookieDomain
	} else {
		cookieDomainWithDot = cookieDomain
	}

	sessionCookie := http.Cookie{
		Name:     key,
		Value:    value,
		MaxAge:   60 * 60 * 24 * 30,
		HttpOnly: false, // if set to true, the javascript cant access it
		Expires:  time.Now().Add(30 * 24 * time.Hour),
		Secure:   secure,
		Domain:   cookieDomainWithDot,
		Path:     "/",
	}

	http.SetCookie(res, &sessionCookie)
}

/* Matches Express cookie-parser
exports.sign = function(val, secret){
  if ('string' != typeof val) throw new TypeError("Cookie value must be provided as a string.");
  if ('string' != typeof secret) throw new TypeError("Secret string must be provided.");
  return val + '.' + crypto
    .createHmac('sha256', secret)
    .update(val)
    .digest('base64')
    .replace(/\=+$/, '');
};*/

var VALID_COOKIE_CHARS = regexp.MustCompile(`[^a-zA-Z0-9]+`)

func SignCookie(value, secret string) string {
	sig := hmac.New(sha256.New, []byte(secret))
	sig.Write([]byte(value))

	encoded := b64.StdEncoding.EncodeToString(sig.Sum(nil))

	return tools.BuildString("gs:", value, ".", VALID_COOKIE_CHARS.ReplaceAllString(encoded, ``))
}
