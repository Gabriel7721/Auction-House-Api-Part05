package handlers

import (
	"main/dto"
	"main/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AuctionHandler struct {
	auctionService *services.AuctionService
}

func NewAuctionHandler(auctionService *services.AuctionService) *AuctionHandler {
	return &AuctionHandler{
		auctionService: auctionService,
	}
}

func (h *AuctionHandler) ListAuctions(c *gin.Context) {
	var query dto.AuctionQuery

	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	result, err := h.auctionService.ListAuctions(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *AuctionHandler) GetAuctionDetail(c *gin.Context) {
	auctionID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	auction, err := h.auctionService.GetAuctionDetail(auctionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "auction not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": auction,
	})
}

func (h *AuctionHandler) CreateAuction(c *gin.Context) {
	userID, ok := getCurrentUserID(c)
	if !ok {
		return
	}

	var input dto.CreateAuctionRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	auction, err := h.auctionService.CreateAuction(userID, input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "auction created successfully",
		"data":    auction,
	})
}

func (h *AuctionHandler) UpdateAuction(c *gin.Context) {
	userID, ok := getCurrentUserID(c)
	if !ok {
		return
	}

	userRole := getCurrentUserRole(c)

	auctionID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	var input dto.UpdateAuctionRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	auction, err := h.auctionService.UpdateAuction(userID, userRole, auctionID, input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "auction updated successfully",
		"data":    auction,
	})
}

func (h *AuctionHandler) CancelAuction(c *gin.Context) {
	userID, ok := getCurrentUserID(c)
	if !ok {
		return
	}

	userRole := getCurrentUserRole(c)

	auctionID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	if err := h.auctionService.CancelAuction(userID, userRole, auctionID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "auction cancelled successfully",
	})
}

func (h *AuctionHandler) PlaceBid(c *gin.Context) {
	userID, ok := getCurrentUserID(c)
	if !ok {
		return
	}

	auctionID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	var input dto.PlaceBidRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	bid, auction, err := h.auctionService.PlaceBid(userID, auctionID, input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "bid placed successfully",
		"bid":     bid,
		"auction": auction,
	})
}

func (h *AuctionHandler) GetAuctionBids(c *gin.Context) {
	auctionID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	bids, err := h.auctionService.GetAuctionBids(auctionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": bids,
	})
}

func (h *AuctionHandler) WatchAuction(c *gin.Context) {
	userID, ok := getCurrentUserID(c)
	if !ok {
		return
	}

	auctionID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	if err := h.auctionService.WatchAuction(userID, auctionID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "auction added to watchlist",
	})
}

func (h *AuctionHandler) UnwatchAuction(c *gin.Context) {
	userID, ok := getCurrentUserID(c)
	if !ok {
		return
	}

	auctionID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	if err := h.auctionService.UnwatchAuction(userID, auctionID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "auction removed from watchlist",
	})
}

func (h *AuctionHandler) GetMyWatchlist(c *gin.Context) {
	userID, ok := getCurrentUserID(c)
	if !ok {
		return
	}

	watchlist, err := h.auctionService.GetMyWatchlist(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": watchlist,
	})
}

func (h *AuctionHandler) GetMyAuctions(c *gin.Context) {
	userID, ok := getCurrentUserID(c)
	if !ok {
		return
	}

	auctions, err := h.auctionService.GetMyAuctions(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": auctions,
	})
}

func (h *AuctionHandler) GetMyBids(c *gin.Context) {
	userID, ok := getCurrentUserID(c)
	if !ok {
		return
	}

	bids, err := h.auctionService.GetMyBids(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": bids,
	})
}

func parseIDParam(c *gin.Context, paramName string) (uint, bool) {
	rawID := c.Param(paramName)

	id, err := strconv.ParseUint(rawID, 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid id",
		})
		return 0, false
	}

	return uint(id), true
}

func getCurrentUserID(c *gin.Context) (uint, bool) {
	userIDAny, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "unauthorized",
		})
		return 0, false
	}

	userID, ok := userIDAny.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "invalid user id",
		})
		return 0, false
	}

	return userID, true
}

func getCurrentUserRole(c *gin.Context) string {
	userRoleAny, exists := c.Get("userRole")
	if !exists {
		return "user"
	}

	userRole, ok := userRoleAny.(string)
	if !ok || userRole == "" {
		return "user"
	}

	return userRole
}
