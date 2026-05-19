package dto

type RegisterPushTokenRequest struct {
	ExpoPushToken string `json:"expo_push_token" binding:"required"`
	Platform      string `json:"platform"`
}

type UnregisterPushTokenRequest struct {
	ExpoPushToken string `json:"expo_push_token" binding:"required"`
}
