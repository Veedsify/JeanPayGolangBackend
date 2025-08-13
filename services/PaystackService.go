package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"os"
)

type PaystackConfig struct {
	PublicKey string
	SecretKey string
	BaseURL   string
}
type PaystackService struct {
	config *PaystackConfig
}

type InitializeNewTransactionPayload struct {
	TransactionId string `json:"transaction_id"`
	Email         string `json:"email"`
	Amount        int    `json:"amount"`
	Currency      string `json:"currency"`
}

type InitilizeNewTransactionResponse struct {
	Status  bool
	Message string
	Data    struct {
		AuthorizationUrl string
		Reference        string
	}
}

func NewPaystackService(config *PaystackConfig) *PaystackService {
	return &PaystackService{
		config: config,
	}
}

func NewPaystackConfigFromEnv() (*PaystackService, error) {
	publicKey := os.Getenv("PAYSTACK_PUBLIC_KEY")
	if publicKey == "" {
		return nil, errors.New("PAYSTACK_PUBLIC_KEY environment variable not set")
	}

	secretKey := os.Getenv("PAYSTACK_SECRET_KEY")
	if secretKey == "" {
		return nil, errors.New("PAYSTACK_SECRET_KEY environment variable not set")
	}

	baseURL := os.Getenv("PAYSTACK_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.paystack.co"
	}

	return NewPaystackService(&PaystackConfig{
		PublicKey: publicKey,
		SecretKey: secretKey,
		BaseURL:   baseURL,
	}), nil
}

func (s *PaystackService) InitializeNewTransaction(trx InitializeNewTransactionPayload) (InitilizeNewTransactionResponse, error) { // Prepare request payload
	payload := map[string]any{
		"reference": trx.TransactionId,
		"email":     trx.Email,
		"amount":    trx.Amount * 100, // Paystack expects amount in kobo (smallest currency unit)
		"currency":  trx.Currency,
	}

	// Convert payload to JSON
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return InitilizeNewTransactionResponse{}, err
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", s.config.BaseURL+"/transaction/initialize", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return InitilizeNewTransactionResponse{}, err
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.config.SecretKey)

	// Make the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return InitilizeNewTransactionResponse{}, err
	}
	defer resp.Body.Close()

	// Parse response
	var response InitilizeNewTransactionResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return InitilizeNewTransactionResponse{}, err
	}
	return response, nil
}

func (s *PaystackService) GetTransactionDetails(transactionID string) (bool, error) {
	// Implementation for fetching transaction details from Paystack
	// This is a placeholder; actual implementation will depend on the Paystack API
	return false, nil
}
