package handlers

import (
	"main/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type NotificationHandler struct {
	notificationService *services.NotificationService
}

func NewNotificationHandler(
	notificationService *services.NotificationService,
) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
	}
}

func (h *NotificationHandler) GetMyNotifications(c *gin.Context) {
	userID, ok := getCurrentUserID(c)
	if !ok {
		return
	}

	notifications, err := h.notificationService.GetMyNotifications(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": notifications,
	})
}

func (h *NotificationHandler) GetUnreadCount(c *gin.Context) {
	userID, ok := getCurrentUserID(c)
	if !ok {
		return
	}

	count, err := h.notificationService.GetUnreadCount(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"unread_count": count,
	})
}

func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	userID, ok := getCurrentUserID(c)
	if !ok {
		return
	}

	notificationID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	notification, err := h.notificationService.MarkAsRead(userID, notificationID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "notification marked as read",
		"data":    notification,
	})
}

func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	userID, ok := getCurrentUserID(c)
	if !ok {
		return
	}

	if err := h.notificationService.MarkAllAsRead(userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "all notifications marked as read",
	})
}
