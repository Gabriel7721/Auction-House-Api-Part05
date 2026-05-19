package handlers

import (
	"main/dto"
	"main/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PushTokenHandler struct {
	pushTokenService        *services.PushTokenService
	pushNotificationService *services.PushNotificationService
}

func NewPushTokenHandler(
	pushTokenService *services.PushTokenService,
	pushNotificationService *services.PushNotificationService,
) *PushTokenHandler {
	return &PushTokenHandler{
		pushTokenService:        pushTokenService,
		pushNotificationService: pushNotificationService,
	}
}

func (h *PushTokenHandler) RegisterPushToken(c *gin.Context) {
	userID, ok := getCurrentUserID(c)
	if !ok {
		return
	}

	var input dto.RegisterPushTokenRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	_, err := h.pushTokenService.RegisterPushToken(userID, input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "push token registered successfully",
	})
}

func (h *PushTokenHandler) UnregisterPushToken(c *gin.Context) {
	userID, ok := getCurrentUserID(c)
	if !ok {
		return
	}

	var input dto.UnregisterPushTokenRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}
	if err := h.pushTokenService.UnregisterPushToken(userID, input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "push token unregistered successfully",
	})
}

func (h *PushTokenHandler) SendTestPushNotification(c *gin.Context) {
	userID, ok := getCurrentUserID(c)
	if !ok {
		return
	}

	err := h.pushNotificationService.SendToUser(
		userID,
		"Auction House Push Test",
		"Your real device push notification pipeline is connected.",
		map[string]any{
			"screen": "notifications",
			"type":   "push_test",
		},
	)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "test push notification sent",
	})
}
