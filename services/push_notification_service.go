package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"main/config"
	"main/repository"
	"net/http"
	"strings"
	"time"
)

type PushNotificationService struct {
	pushTokenRepo *repository.DevicePushTokenRepository
	httpClient    *http.Client
}

func NewPushNotificationService(
	pushTokenRepo *repository.DevicePushTokenRepository,
) *PushNotificationService {
	return &PushNotificationService{
		pushTokenRepo: pushTokenRepo,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

type expoPushMessage struct {
	To        string         `json:"to"`
	Title     string         `json:"title"`
	Body      string         `json:"body"`
	Sound     string         `json:"sound,omitempty"`
	Priority  string         `json:"priority,omitempty"`
	ChannelID string         `json:"channelId,omitempty"`
	Data      map[string]any `json:"data,omitempty"`
}
type expoPushTicketResponse struct {
	Data   []expoPushTicket `json:"data"`
	Errors []expoPushError  `json:"errors"`
}
type expoPushTicket struct {
	Status  string         `json:"status"`
	ID      string         `json:"id,omitempty"`
	Message string         `json:"message,omitempty"`
	Details map[string]any `json:"details,omitempty"`
}
type expoPushError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (s *PushNotificationService) SendToUser(
	userID uint,
	title string,
	body string,
	data map[string]any,
) error {
	tokens, err := s.pushTokenRepo.FindActiveByUserID(userID)
	if err != nil {
		return err
	}

	if len(tokens) == 0 {
		return nil
	}

	messages := make([]expoPushMessage, 0, len(tokens))

	for _, token := range tokens {
		messages = append(messages, expoPushMessage{
			To:        token.ExpoPushToken,
			Title:     title,
			Body:      body,
			Sound:     "auction_alert.wav",
			Priority:  "high",
			ChannelID: "auction-alerts",
			Data:      data,
		})
	}

	return s.sendMessages(messages)
}

func (s *PushNotificationService) sendMessages(
	messages []expoPushMessage,
) error {
	if len(messages) == 0 {
		return nil
	}

	payload, err := json.Marshal(messages)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(
		http.MethodPost,
		config.ExpoPushAPIURL,
		bytes.NewBuffer(payload),
	)
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	req.Header.Set("Content-Type", "application/json")

	res, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	responseBody, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf(
			"expo push request failed with status %d: %s",
			res.StatusCode,
			string(responseBody),
		)
	}

	var expoResponse expoPushTicketResponse

	if err := json.Unmarshal(responseBody, &expoResponse); err != nil {
		return err
	}

	if len(expoResponse.Errors) > 0 {
		var serverErrors []string

		for _, item := range expoResponse.Errors {
			serverErrors = append(
				serverErrors,
				fmt.Sprintf("%s: %s", item.Code, item.Message),
			)
		}

		return errors.New(strings.Join(serverErrors, "; "))
	}

	var ticketErrors []string
	var hasSuccess bool

	for _, ticket := range expoResponse.Data {
		if ticket.Status == "ok" {
			hasSuccess = true
			continue
		}

		ticketErrors = append(ticketErrors, ticket.Message)
	}

	if len(ticketErrors) > 0 {
		log.Println("expo push ticket errors:", strings.Join(ticketErrors, "; "))
	}

	if !hasSuccess && len(ticketErrors) > 0 {
		return errors.New(strings.Join(ticketErrors, "; "))
	}

	return nil
}
