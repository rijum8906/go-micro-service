package request

type GenerateScopedTokenRequest struct {
	Scope         string        `json:"scope"    binding:"required"`
	Authorization Authorization `json:"authorization" binding:"required"`
	Metadata      Metadata      `json:"metadata" binding:"required"`
}
