package domain

type ParseTokenRequest struct {
	Token string `json:"token" binding:"required"`
}

type ParseTokenResponse struct {
	UserID            string `json:"user_id"`
	UserType          int32  `json:"user_type"`
	ExpireTimeSeconds int64  `json:"expire_time_seconds"`
}
