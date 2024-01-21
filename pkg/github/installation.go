package github

import "time"

type Installation struct {
	ID              int    `json:"id"`
	AccessTokensURL string `json:"access_tokens_url"`
	AppID           int    `json:"app_id"`
}

type Token struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}
