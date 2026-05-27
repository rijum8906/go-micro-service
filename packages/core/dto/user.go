package dto

type UserInfo struct {
	UserID    string
	TokenID   string
	SessionID string
}

type ClientInfo struct {
	DeviceID   string
	UserAgent  string
	IPAddress  string
	ClientType string
	APIVersion string
	SDKVersion string
	RequestID  string
	SessionID  string
	TraceID    string
	Locale     string
}

type AuthTokens struct {
	AccessToken  string
	RefreshToken string
}
