package environment

import (
	"github.com/griffnb/core/lib/config"
)

const (
	QUEUE_THROTTLES = "throttles"
)

type Config struct {
	config.DefaultConfig
	InternalAPIKey string        `json:"internal_api_key"`
	Email          *Email        `json:"email"`
	Encryption     *Encryption   `json:"encryption"`
	AIKeys         *AIKeys       `json:"ai_keys"`
	Cloudflare     *Cloudflare   `json:"cloudflare"`
	Sendpulse      *Sendpulse    `json:"sendpulse"`
	Stripe         *StripeConfig `json:"stripe"`
}

type Cloudflare struct {
	TurnstileKey string `json:"turnstile_key"`
	APIKey       string `json:"api_key"`
	AccountID    string `json:"account_id"`
}

type AIKeys struct {
	OpenAI struct {
		APIKey string `json:"api_key"`
	} `json:"openai"`
	Gemini struct {
		APIKey string `json:"api_key"`
	} `json:"gemini"`
	Azure struct {
		APIKey   string `json:"api_key"`
		Endpoint string `json:"endpoint"`
	} `json:"azure"`
	Anthropic struct {
		APIKey string `json:"api_key"`
	} `json:"anthropic"`
}

type Encryption struct {
	Ciphers       map[string]string `json:"ciphers"`
	CurrentCipher string            `json:"current_cipher"`
}

type Email struct {
	Provider string `json:"provider"`
	From     string `json:"from"`
	DevEmail string `json:"dev_email"`
	SMTP     *SMTP  `json:"smtp"`
	SES      *SES   `json:"ses"`
}

// SMTP
type SMTP struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	UserName string `json:"username"`
	Password string `json:"password"`
}

type SES struct {
	Region string `json:"region"`
	Key    string `json:"key"`
	Skey   string `json:"skey"`
}

type Sendpulse struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	WebhookKey   string `json:"webhook_key"`
}

type StripeConfig struct {
	WebhookKey string `json:"webhook_key"`
	SecretKey  string `json:"secret_key"`
	PublicKey  string `json:"public_key"`
}
