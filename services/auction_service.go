package services

import (
	"errors"
	"fmt"
	"log"
	"main/dto"
	"main/models"
	"main/repository"
	"strings"
	"time"

	"gorm.io/gorm"
)

type AuctionService struct {
	db                      *gorm.DB
	auctionRepo             *repository.AuctionRepository
	productRepo             *repository.ProductRepository
	bidRepo                 *repository.BidRepository
	watchlistRepo           *repository.WatchlistRepository
	categoryRepo            *repository.CategoryRepository
	notificationRepo        *repository.NotificationRepository
	pushNotificationService *PushNotificationService
}

type pushNotificationJob struct {
	UserID uint
	Title  string
	Body   string
	Data   map[string]any
}

func NewAuctionService(
	db *gorm.DB,
	auctionRepo *repository.AuctionRepository,
	productRepo *repository.ProductRepository,
	bidRepo *repository.BidRepository,
	watchlistRepo *repository.WatchlistRepository,
	categoryRepo *repository.CategoryRepository,
	notificationRepo *repository.NotificationRepository,
	pushNotificationService *PushNotificationService,
) *AuctionService {
	return &AuctionService{
		db:                      db,
		auctionRepo:             auctionRepo,
		productRepo:             productRepo,
		bidRepo:                 bidRepo,
		watchlistRepo:           watchlistRepo,
		categoryRepo:            categoryRepo,
		notificationRepo:        notificationRepo,
		pushNotificationService: pushNotificationService,
	}
}

func (s *AuctionService) CreateAuction(sellerID uint, input dto.CreateAuctionRequest) (*models.Auction, error) {
	if err := validateCreateAuctionInput(input); err != nil {
		return nil, err
	}

	_, err := s.categoryRepo.FindByID(input.CategoryID)
	if err != nil {
		return nil, errors.New("category not found")
	}

	var createdAuction models.Auction

	err = s.db.Transaction(func(tx *gorm.DB) error {
		product := models.Product{
			SellerID:    sellerID,
			CategoryID:  input.CategoryID,
			Title:       strings.TrimSpace(input.Title),
			Description: strings.TrimSpace(input.Description),
			Images:      models.StringArray(input.Images),
			Condition:   models.ProductCondition(input.Condition),
			Status:      models.ProductStatusActive,
		}

		if err := s.productRepo.CreateTx(tx, &product); err != nil {
			return err
		}

		auction := models.Auction{
			ProductID:    product.ID,
			SellerID:     sellerID,
			StartPrice:   input.StartPrice,
			ReservePrice: input.ReservePrice,
			BuyNowPrice:  input.BuyNowPrice,
			CurrentPrice: input.StartPrice,
			BidIncrement: input.BidIncrement,
			StartsAt:     input.StartsAt,
			EndsAt:       input.EndsAt,
			Status:       resolveAuctionStatus(input.StartsAt, input.EndsAt),
		}

		if err := s.auctionRepo.CreateTx(tx, &auction); err != nil {
			return err
		}

		createdAuction = auction

		return nil
	})

	if err != nil {
		return nil, err
	}

	return s.auctionRepo.FindByID(createdAuction.ID)
}

func (s *AuctionService) ListAuctions(query dto.AuctionQuery) (*dto.PaginatedResponse, error) {
	if err := s.ActivateScheduledAuctions(); err != nil {
		return nil, err
	}

	if err := s.EndExpiredAuctions(); err != nil {
		return nil, err
	}

	if query.Page <= 0 {
		query.Page = 1
	}

	if query.Limit <= 0 {
		query.Limit = 10
	}

	if query.Limit > 100 {
		query.Limit = 100
	}

	auctions, total, err := s.auctionRepo.List(query)
	if err != nil {
		return nil, err
	}

	return &dto.PaginatedResponse{
		Data:       auctions,
		Page:       query.Page,
		Limit:      query.Limit,
		Total:      total,
		TotalPages: repository.CalculateTotalPages(total, query.Limit),
	}, nil
}

func (s *AuctionService) GetAuctionDetail(auctionID uint) (*models.Auction, error) {
	if err := s.ActivateScheduledAuctions(); err != nil {
		return nil, err
	}

	if err := s.FinalizeAuctionIfExpired(auctionID); err != nil {
		return nil, err
	}

	return s.auctionRepo.FindByID(auctionID)
}

func (s *AuctionService) UpdateAuction(
	userID uint,
	userRole string,
	auctionID uint,
	input dto.UpdateAuctionRequest,
) (*models.Auction, error) {
	auction, err := s.auctionRepo.FindByID(auctionID)
	if err != nil {
		return nil, errors.New("auction not found")
	}

	if err := s.FinalizeAuctionIfExpired(auctionID); err != nil {
		return nil, err
	}

	auction, err = s.auctionRepo.FindByID(auctionID)
	if err != nil {
		return nil, errors.New("auction not found")
	}

	if !canManageAuction(userID, userRole, auction.SellerID) {
		return nil, errors.New("you do not have permission to update this auction")
	}

	if auction.Status == models.AuctionStatusEnded {
		return nil, errors.New("ended auction cannot be updated")
	}

	if auction.Status == models.AuctionStatusCancelled {
		return nil, errors.New("cancelled auction cannot be updated")
	}

	if err := validateUpdateAuctionInput(auction, input); err != nil {
		return nil, err
	}

	if input.Title != nil {
		auction.Product.Title = strings.TrimSpace(*input.Title)
	}

	if input.Description != nil {
		auction.Product.Description = strings.TrimSpace(*input.Description)
	}

	if input.CategoryID != nil {
		_, err := s.categoryRepo.FindByID(*input.CategoryID)
		if err != nil {
			return nil, errors.New("category not found")
		}

		auction.Product.CategoryID = *input.CategoryID
	}

	if input.Condition != nil {
		condition := strings.ToLower(strings.TrimSpace(*input.Condition))
		auction.Product.Condition = models.ProductCondition(condition)
	}

	if input.Images != nil {
		auction.Product.Images = models.StringArray(input.Images)
	}

	if input.ReservePrice != nil {
		auction.ReservePrice = input.ReservePrice
	}

	if input.BuyNowPrice != nil {
		auction.BuyNowPrice = input.BuyNowPrice
	}

	if input.BidIncrement != nil {
		auction.BidIncrement = *input.BidIncrement
	}

	if input.StartsAt != nil {
		auction.StartsAt = *input.StartsAt
	}

	if input.EndsAt != nil {
		auction.EndsAt = *input.EndsAt
	}

	auction.Status = resolveAuctionStatus(auction.StartsAt, auction.EndsAt)

	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.productRepo.UpdateTx(tx, &auction.Product); err != nil {
			return err
		}

		if err := s.auctionRepo.UpdateTx(tx, auction); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return s.auctionRepo.FindByID(auction.ID)
}

func (s *AuctionService) CancelAuction(userID uint, userRole string, auctionID uint) error {
	auction, err := s.auctionRepo.FindByID(auctionID)
	if err != nil {
		return errors.New("auction not found")
	}

	if !canManageAuction(userID, userRole, auction.SellerID) {
		return errors.New("you do not have permission to cancel this auction")
	}

	if auction.Status == models.AuctionStatusEnded {
		return errors.New("ended auction cannot be cancelled")
	}

	if auction.Status == models.AuctionStatusCancelled {
		return errors.New("auction is already cancelled")
	}

	auction.Status = models.AuctionStatusCancelled
	auction.Product.Status = models.ProductStatusCancelled

	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.productRepo.UpdateTx(tx, &auction.Product); err != nil {
			return err
		}

		return s.auctionRepo.UpdateTx(tx, auction)
	})
}

func (s *AuctionService) PlaceBid(
	userID uint,
	auctionID uint,
	input dto.PlaceBidRequest,
) (*models.Bid, *models.Auction, error) {
	if input.Amount <= 0 {
		return nil, nil, errors.New("bid amount must be greater than 0")
	}

	var createdBid models.Bid
	var updatedAuction models.Auction
	var businessErr error
	var pushJobs []pushNotificationJob

	err := s.db.Transaction(func(tx *gorm.DB) error {
		auction, err := s.auctionRepo.FindByIDForUpdate(tx, auctionID)
		if err != nil {
			return errors.New("auction not found")
		}

		now := time.Now()

		if auction.Status == models.AuctionStatusCancelled {
			businessErr = errors.New("cancelled auction cannot receive bids")
			return nil
		}

		if auction.Status == models.AuctionStatusEnded {
			businessErr = errors.New("ended auction cannot receive bids")
			return nil
		}

		if now.Before(auction.StartsAt) {
			businessErr = errors.New("auction has not started")
			return nil
		}

		if !now.Before(auction.EndsAt) {
			if err := s.finalizeLockedAuction(tx, auction, &pushJobs); err != nil {
				return err
			}

			businessErr = errors.New("auction has ended")
			return nil
		}

		if auction.SellerID == userID {
			businessErr = errors.New("seller cannot bid on their own auction")
			return nil
		}

		highestBid, err := s.bidRepo.FindHighestByAuctionIDTx(tx, auctionID)
		if err != nil {
			return err
		}

		var previousHighestBid *models.Bid
		if highestBid != nil {
			previousHighestBid = highestBid
		}

		minAmount := auction.StartPrice

		if highestBid != nil {
			minAmount = auction.CurrentPrice + auction.BidIncrement
		}

		if input.Amount < minAmount {
			businessErr = fmt.Errorf("bid amount must be at least %.2f", minAmount)
			return nil
		}

		bid := models.Bid{
			AuctionID: auctionID,
			UserID:    userID,
			Amount:    input.Amount,
		}

		if err := s.bidRepo.CreateTx(tx, &bid); err != nil {
			return err
		}

		if previousHighestBid != nil && previousHighestBid.UserID != userID {
			productTitle := auction.Product.Title
			if productTitle == "" {
				productTitle = "this auction"
			}

			outbidNotification := buildNotification(
				previousHighestBid.UserID,
				models.NotificationTypeOutbid,
				fmt.Sprintf("You have been outbid on \"%s\"", productTitle),
				fmt.Sprintf(
					"A higher bid of %.2f has been placed on this auction.",
					input.Amount,
				),
			)

			if err := s.notificationRepo.CreateTx(tx, &outbidNotification); err != nil {
				return err
			}

			pushJobs = append(pushJobs, pushNotificationJob{
				UserID: previousHighestBid.UserID,
				Title:  outbidNotification.Title,
				Body:   outbidNotification.Message,
				Data:   buildAuctionPushData(outbidNotification, auction.ID),
			})
		}

		auction.CurrentPrice = input.Amount
		auction.Status = models.AuctionStatusLive
		auction.Product.Status = models.ProductStatusActive

		if auction.BuyNowPrice != nil && input.Amount >= *auction.BuyNowPrice {
			winnerID := userID

			auction.Status = models.AuctionStatusEnded
			auction.WinnerID = &winnerID
			auction.Product.Status = models.ProductStatusSold

			productTitle := auction.Product.Title
			if productTitle == "" {
				productTitle = "this auction"
			}

			winnerNotification := buildNotification(
				userID,
				models.NotificationTypeAuctionWon,
				fmt.Sprintf("Congratulations! You won \"%s\"", productTitle),
				fmt.Sprintf(
					"Your winning bid was %.2f.",
					input.Amount,
				),
			)

			if err := s.notificationRepo.CreateTx(tx, &winnerNotification); err != nil {
				return err
			}

			pushJobs = append(pushJobs, pushNotificationJob{
				UserID: userID,
				Title:  winnerNotification.Title,
				Body:   winnerNotification.Message,
				Data:   buildAuctionPushData(winnerNotification, auction.ID),
			})

			sellerNotification := buildNotification(
				auction.SellerID,
				models.NotificationTypeAuctionEnded,
				fmt.Sprintf("Your auction \"%s\" has ended", productTitle),
				fmt.Sprintf(
					"The auction ended immediately through Buy Now at %.2f.",
					input.Amount,
				),
			)

			if err := s.notificationRepo.CreateTx(tx, &sellerNotification); err != nil {
				return err
			}
		}

		if err := s.productRepo.UpdateTx(tx, &auction.Product); err != nil {
			return err
		}

		if err := s.auctionRepo.UpdateTx(tx, auction); err != nil {
			return err
		}

		createdBid = bid
		updatedAuction = *auction

		return nil
	})

	if err != nil {
		return nil, nil, err
	}

	if businessErr != nil {
		s.dispatchPushJobs(pushJobs)
		return nil, nil, businessErr
	}
	s.dispatchPushJobs(pushJobs)

	fullAuction, err := s.auctionRepo.FindByID(updatedAuction.ID)
	if err == nil {
		updatedAuction = *fullAuction
	}

	return &createdBid, &updatedAuction, nil
}

func (s *AuctionService) GetAuctionBids(auctionID uint) ([]models.Bid, error) {
	_, err := s.auctionRepo.FindByID(auctionID)
	if err != nil {
		return nil, errors.New("auction not found")
	}

	return s.bidRepo.FindByAuctionID(auctionID)
}

func (s *AuctionService) WatchAuction(userID uint, auctionID uint) error {
	_, err := s.auctionRepo.FindByID(auctionID)
	if err != nil {
		return errors.New("auction not found")
	}

	return s.watchlistRepo.Add(userID, auctionID)
}

func (s *AuctionService) UnwatchAuction(userID uint, auctionID uint) error {
	_, err := s.auctionRepo.FindByID(auctionID)
	if err != nil {
		return errors.New("auction not found")
	}

	return s.watchlistRepo.Remove(userID, auctionID)
}

func (s *AuctionService) GetMyWatchlist(userID uint) ([]models.Watchlist, error) {
	return s.watchlistRepo.FindByUserID(userID)
}

func (s *AuctionService) GetMyAuctions(userID uint) ([]models.Auction, error) {
	if err := s.ActivateScheduledAuctions(); err != nil {
		return nil, err
	}

	if err := s.EndExpiredAuctions(); err != nil {
		return nil, err
	}

	return s.auctionRepo.FindBySellerID(userID)
}

func (s *AuctionService) GetMyBids(userID uint) ([]models.Bid, error) {
	if err := s.EndExpiredAuctions(); err != nil {
		return nil, err
	}

	return s.bidRepo.FindByUserID(userID)
}

func (s *AuctionService) FinalizeAuctionIfExpired(auctionID uint) error {
	var pushJobs []pushNotificationJob

	err := s.db.Transaction(func(tx *gorm.DB) error {
		auction, err := s.auctionRepo.FindByIDForUpdate(tx, auctionID)
		if err != nil {
			return err
		}

		if auction.Status == models.AuctionStatusEnded ||
			auction.Status == models.AuctionStatusCancelled {
			return nil
		}

		if time.Now().Before(auction.EndsAt) {
			return nil
		}

		return s.finalizeLockedAuction(tx, auction, &pushJobs)
	})

	if err != nil {
		return err
	}

	s.dispatchPushJobs(pushJobs)
	return nil
}

func (s *AuctionService) EndExpiredAuctions() error {
	auctions, err := s.auctionRepo.FindExpiredLiveAuctions()
	if err != nil {
		return err
	}

	for _, auction := range auctions {
		if err := s.FinalizeAuctionIfExpired(auction.ID); err != nil {
			return err
		}
	}

	return nil
}

func (s *AuctionService) finalizeLockedAuction(
	tx *gorm.DB,
	auction *models.Auction,
	pushJobs *[]pushNotificationJob,
) error {
	highestBid, err := s.bidRepo.FindHighestByAuctionIDTx(tx, auction.ID)
	if err != nil {
		return err
	}

	auction.Status = models.AuctionStatusEnded

	productTitle := auction.Product.Title
	if productTitle == "" {
		productTitle = "this auction"
	}

	if highestBid != nil {
		winnerID := highestBid.UserID

		auction.WinnerID = &winnerID
		auction.CurrentPrice = highestBid.Amount
		auction.Product.Status = models.ProductStatusSold

		winnerNotification := buildNotification(
			highestBid.UserID,
			models.NotificationTypeAuctionWon,
			fmt.Sprintf("Congratulations! You won \"%s\"", productTitle),
			fmt.Sprintf(
				"Your winning bid was %.2f.",
				highestBid.Amount,
			),
		)

		if err := s.notificationRepo.CreateTx(tx, &winnerNotification); err != nil {
			return err
		}

		*pushJobs = append(*pushJobs, pushNotificationJob{
			UserID: highestBid.UserID,
			Title:  winnerNotification.Title,
			Body:   winnerNotification.Message,
			Data:   buildAuctionPushData(winnerNotification, auction.ID),
		})

		sellerNotification := buildNotification(
			auction.SellerID,
			models.NotificationTypeAuctionEnded,
			fmt.Sprintf("Your auction \"%s\" has ended", productTitle),
			fmt.Sprintf(
				"The auction ended with a winning bid of %.2f.",
				highestBid.Amount,
			),
		)

		if err := s.notificationRepo.CreateTx(tx, &sellerNotification); err != nil {
			return err
		}

		*pushJobs = append(*pushJobs, pushNotificationJob{
			UserID: auction.SellerID,
			Title:  sellerNotification.Title,
			Body:   sellerNotification.Message,
			Data:   buildAuctionPushData(sellerNotification, auction.ID),
		})

	} else {
		auction.Product.Status = models.ProductStatusActive

		sellerNotification := buildNotification(
			auction.SellerID,
			models.NotificationTypeAuctionEnded,
			fmt.Sprintf("Your auction \"%s\" has ended", productTitle),
			"No bids were placed before the auction ended.",
		)

		if err := s.notificationRepo.CreateTx(tx, &sellerNotification); err != nil {
			return err
		}

		*pushJobs = append(*pushJobs, pushNotificationJob{
			UserID: auction.SellerID,
			Title:  sellerNotification.Title,
			Body:   sellerNotification.Message,
			Data:   buildAuctionPushData(sellerNotification, auction.ID),
		})
	}

	if auction.Product.ID != 0 {
		if err := s.productRepo.UpdateTx(tx, &auction.Product); err != nil {
			return err
		}
	}

	return s.auctionRepo.UpdateTx(tx, auction)
}

func validateCreateAuctionInput(input dto.CreateAuctionRequest) error {
	if strings.TrimSpace(input.Title) == "" {
		return errors.New("title is required")
	}

	if strings.TrimSpace(input.Description) == "" {
		return errors.New("description is required")
	}

	if input.CategoryID == 0 {
		return errors.New("category_id is required")
	}

	condition := strings.ToLower(strings.TrimSpace(input.Condition))
	if !isValidProductCondition(condition) {
		return errors.New("condition must be one of: new, used, refurbished")
	}

	if input.StartPrice <= 0 {
		return errors.New("start_price must be greater than 0")
	}

	if input.BidIncrement <= 0 {
		return errors.New("bid_increment must be greater than 0")
	}

	if !input.EndsAt.After(input.StartsAt) {
		return errors.New("ends_at must be after starts_at")
	}

	if !input.EndsAt.After(time.Now()) {
		return errors.New("cannot create auction that ends in the past")
	}

	if input.ReservePrice != nil && *input.ReservePrice < input.StartPrice {
		return errors.New("reserve_price must be greater than or equal to start_price")
	}

	if input.BuyNowPrice != nil && *input.BuyNowPrice < input.StartPrice {
		return errors.New("buy_now_price must be greater than or equal to start_price")
	}

	return nil
}

func validateUpdateAuctionInput(auction *models.Auction, input dto.UpdateAuctionRequest) error {
	if input.Title != nil && strings.TrimSpace(*input.Title) == "" {
		return errors.New("title cannot be empty")
	}

	if input.Description != nil && strings.TrimSpace(*input.Description) == "" {
		return errors.New("description cannot be empty")
	}

	if input.Condition != nil {
		condition := strings.ToLower(strings.TrimSpace(*input.Condition))

		if !isValidProductCondition(condition) {
			return errors.New("condition must be one of: new, used, refurbished")
		}
	}

	if input.BidIncrement != nil && *input.BidIncrement <= 0 {
		return errors.New("bid_increment must be greater than 0")
	}

	startsAt := auction.StartsAt
	endsAt := auction.EndsAt

	if input.StartsAt != nil {
		startsAt = *input.StartsAt
	}

	if input.EndsAt != nil {
		endsAt = *input.EndsAt
	}

	if !endsAt.After(startsAt) {
		return errors.New("ends_at must be after starts_at")
	}

	if input.EndsAt != nil && !endsAt.After(time.Now()) {
		return errors.New("ends_at cannot be in the past")
	}

	if input.ReservePrice != nil && *input.ReservePrice < auction.StartPrice {
		return errors.New("reserve_price must be greater than or equal to start_price")
	}

	if input.BuyNowPrice != nil && *input.BuyNowPrice < auction.StartPrice {
		return errors.New("buy_now_price must be greater than or equal to start_price")
	}

	return nil
}

func resolveAuctionStatus(startsAt time.Time, endsAt time.Time) models.AuctionStatus {
	now := time.Now()

	if now.Before(startsAt) {
		return models.AuctionStatusScheduled
	}

	if now.Before(endsAt) {
		return models.AuctionStatusLive
	}

	return models.AuctionStatusEnded
}

func isValidProductCondition(condition string) bool {
	switch condition {
	case string(models.ProductConditionNew),
		string(models.ProductConditionUsed),
		string(models.ProductConditionRefurbished):
		return true
	default:
		return false
	}
}

func canManageAuction(userID uint, userRole string, sellerID uint) bool {
	if userRole == string(models.RoleAdmin) {
		return true
	}

	return userID == sellerID
}

func (s *AuctionService) ActivateScheduledAuctions() error {
	now := time.Now()
	return s.auctionRepo.ActivateReadyScheduledAuctions(now)
}

func buildNotification(
	userID uint,
	notificationType models.NotificationType,
	title string,
	message string,
) models.Notification {
	return models.Notification{
		UserID:  userID,
		Type:    notificationType,
		Title:   title,
		Message: message,
		IsRead:  false,
	}
}

func buildAuctionPushData(
	notification models.Notification,
	auctionID uint,
) map[string]any {
	return map[string]any{
		"screen":            "auction_detail",
		"auction_id":        auctionID,
		"notification_id":   notification.ID,
		"notification_type": string(notification.Type),
	}
}

func (s *AuctionService) dispatchPushJobs(
	jobs []pushNotificationJob,
) {
	if s.pushNotificationService == nil || len(jobs) == 0 {
		return
	}

	for _, job := range jobs {
		err := s.pushNotificationService.SendToUser(
			job.UserID,
			job.Title,
			job.Body,
			job.Data,
		)

		if err != nil {
			log.Printf(
				"failed to send push notification to user %d: %v",
				job.UserID,
				err,
			)
		}
	}
}
