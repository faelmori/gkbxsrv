package sync

type ProductInput struct {
	ProductID uint `json:"product_id" binding:"required"`
}

type ProductsInput struct {
	ProductIDs []uint `json:"product_ids" binding:"required"`
}

type BasicAuth struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type OAuth2 struct {
	AccessToken string `json:"access_token" binding:"required"`
}

type AuthOptions struct {
	Type      string    `json:"type" binding:"required"`
	BasicAuth BasicAuth `json:"basic_auth" binding:"omitempty"`
	APIKey    string    `json:"api_key" binding:"omitempty"`
	OAuth2    OAuth2    `json:"oauth2" binding:"omitempty"`
	JWT       string    `json:"jwt" binding:"omitempty"`
}

type Session struct {
	ExpireDuration  float64 `json:"expire_duration" binding:"omitempty"`
	RateLimit       int     `json:"rate_limit" binding:"omitempty"`
	RateLimitPeriod float64 `json:"rate_limit_period" binding:"omitempty"`
	TimeOutDuration float64 `json:"timeout_duration" binding:"omitempty"`
	Concurrency     int     `json:"concurrency" binding:"default=1"`
	LastSync        string  `json:"last_sync" binding:"omitempty"`
	IsAuthenticated bool    `json:"is_authenticated" binding:"default=false"`
	AuthOptions     AuthOptions
}
