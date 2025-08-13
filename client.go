package payment

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// Client is the main struct for interacting with PayMe API.
// This struct handles all PayMe API method calls, HTTP requests,
// authentication, and response parsing.
type Client struct {
	// headers for authentication
	Headers xAuthHeaders
	// base url
	BaseURL string
	// http client
	HTTPClient http.Client
	// logger
	Logger *log.Logger
	// timeout
	Timeout time.Duration
	// is test mode
	IsTestMode bool
	// requisite name like charge_id, order_id, id you given to requisite title in payme dashboard
	RequisiteName string
}

// ClientConfig contains configuration parameters for creating a PayMe client.
// This struct includes all necessary parameters like PayMe ID, key,
// test mode, and other settings.
type ClientConfig struct {
	// Merchant ID
	PaymeID string `json:"payme_id"`
	// payme key
	PaymeKey string `json:"payme_key"`
	// is test mode
	IsTestMode bool `json:"is_test_mode"`
	// requisite name like charge_id, order_id, id you given to requisite title in payme dashboard
	RequisiteName string `json:"requisite_name"`
	// logger
	Logger *log.Logger `json:"logger"`
	// http client
	HTTPClient http.Client `json:"http_client"`
	// base url
	BaseURL string `json:"base_url"`
	// timeout default 30 seconds
	Timeout time.Duration `json:"timeout"`
}

// xAuthHeaders contains authentication headers for PayMe API.
// This struct stores PayMe ID and key used in HTTP requests.
type xAuthHeaders struct {
	paymeID  string
	paymeKey string
}

// NewClient creates a new PayMe client instance with the provided configuration.
// It validates the config, sets default values, and initializes the client.
// Returns a pointer to Client and any error that occurred during initialization.
func NewClient(config ClientConfig) (*Client, error) {
	err := config.validate()
	if err != nil {
		return nil, err
	}

	// Default timeout
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	// Default requisite name
	if config.RequisiteName == "" {
		config.RequisiteName = "id"
	}

	// Default base URL based on test mode
	if config.BaseURL == "" {
		if config.IsTestMode {
			config.BaseURL = TestEndpoint
		} else {
			config.BaseURL = ProductionEndpoint
		}
	}

	// Default HTTP client
	if config.HTTPClient.Timeout == 0 {
		config.HTTPClient.Timeout = config.Timeout
	}

	client := &Client{
		HTTPClient:    config.HTTPClient,
		BaseURL:       config.BaseURL,
		Logger:        config.Logger,
		Headers:       getXAuthHeaders(config.PaymeID, config.PaymeKey),
		Timeout:       config.Timeout,
		IsTestMode:    config.IsTestMode,
		RequisiteName: config.RequisiteName,
	}

	return client, nil
}

// validate checks if the ClientConfig contains valid parameters.
// It ensures PayMe ID and key are not empty.
// Returns an error if validation fails.
func (c ClientConfig) validate() error {
	if c.PaymeID == "" {
		return ErrEmptyOrInvalidPaycomID
	}
	if c.PaymeKey == "" {
		return ErrEmptyOrInvalidPaycomKey
	}

	return nil
}

// getXAuthHeaders creates authentication headers from PayMe ID and key.
// Returns an xAuthHeaders struct with the provided credentials.
func getXAuthHeaders(paymeID, paymeKey string) xAuthHeaders {
	return xAuthHeaders{paymeID: paymeID, paymeKey: paymeKey}
}

// sendRequest sends HTTP requests to PayMe API.
// It handles request creation, authentication headers, timeout, and response parsing.
// Returns a Response struct and any error that occurred.
func (c *Client) sendRequest(
	ctx context.Context,
	requestID, method string,
	params interface{},
	withID bool,
	timeout ...time.Duration,
) (*Response, error) {
	var requestTimeout time.Duration

	if len(timeout) > 0 {
		requestTimeout = timeout[0]
	} else {
		requestTimeout = c.Timeout
	}

	// Create a context with the specified timeout.
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	data := map[string]interface{}{
		"id":     requestID,
		"method": method,
		"params": params,
	}

	requestBody, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("json marshal error: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.BaseURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("request creation error: %w", err)
	}

	// Set headers
	if withID {
		req.Header.Set("X-Auth", c.Headers.paymeID)
	} else {
		req.Header.Set("X-Auth", fmt.Sprintf("%s:%s", c.Headers.paymeID, c.Headers.paymeKey))
	}

	req.Header.Set("Content-Type", "application/json")

	// Send request
	response, err := c.HTTPClient.Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, ErrTimeout
		}
		return nil, fmt.Errorf("http request error: %w", err)
	}
	defer response.Body.Close()

	// Read response body
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("response body read error: %w", err)
	}

	// Parse response
	var responseJson Response
	err = json.Unmarshal(responseBody, &responseJson)
	if err != nil {
		return nil, fmt.Errorf("json unmarshal error: %w", err)
	}

	// Handle error response with payme specific error codes
	responseJson, err = c.handleErrorResponse(responseJson)
	if err != nil {
		if c.Logger != nil {
			c.Logger.Printf("PayMe error response - %v, error - %v", responseJson.Error, err)
		}
	}

	return &responseJson, err
}

// handleErrorResponse processes PayMe API error responses.
// It maps PayMe error codes to custom error types.
// Returns the original response and a custom error if applicable.
func (c *Client) handleErrorResponse(responseJson Response) (Response, error) {
	var paymeError error

	if responseJson.Error == nil {
		return responseJson, nil
	}

	errorCode := responseJson.Error.Code

	switch errorCode {
	case InvalidAmountErrorCode:
		paymeError = ErrInvalidAmount
	case InvalidParamsErrorCode:
		paymeError = ErrInvalidParams
	case CardNotFoundErrorCode:
		paymeError = ErrCardNotFound
	case InvalidFormatTokenErrorCode:
		paymeError = ErrInvalidFormatToken
	case CardNumberNotFoundCode:
		paymeError = ErrCardNotFound
	case CardExpiredCode:
		paymeError = ErrCardExpired
	case ProcessingCenterNotAvailableCode:
		paymeError = ErrProcessingCenterNotAvailable
	case PaycomServiceNotAvailableCode:
		paymeError = ErrPaycomServiceNotAvailable
	case ReceiptNotFoundErrorCode:
		paymeError = ErrReceiptNotFound
	case ReceiptAlreadyPaidErrorCode:
		paymeError = ErrReceiptAlreadyPaid
	case ReceiptExpiredErrorCode:
		paymeError = ErrReceiptExpired
	default:
		if errorCode != 0 {
			paymeError = ErrPaymeError
		}
	}

	return responseJson, paymeError
}

// CreateReceipt creates a new payment receipt in PayMe system.
// It validates the amount and sends a request to receipts.create method.
// Returns CreateReceiptResponse with receipt details or an error.
func (c *Client) CreateReceipt(ctx context.Context, amount int64, account map[string]interface{}, description string, detail map[string]interface{}) (*CreateReceiptResponse, error) {
	// Validation
	if err := ValidateAmount(amount); err != nil {
		return nil, err
	}

	requestID := GenerateRequestID("ReceiptsCreate")

	receiptParams := map[string]interface{}{
		"amount":      amount,
		"account":     account,
		"description": description,
		"detail":      detail,
	}

	resp, err := c.sendRequest(ctx, requestID, "receipts.create", receiptParams, false)
	if err != nil {
		return nil, err
	}

	// Parse result
	var result CreateReceiptResponse
	if resp.Result != nil {
		resultBytes, _ := json.Marshal(resp.Result)
		if err := json.Unmarshal(resultBytes, &result); err != nil {
			return nil, fmt.Errorf("result unmarshal error: %w", err)
		}
	}

	return &result, nil
}

// PayReceipt processes payment for an existing receipt.
// It validates receipt ID and card token, then sends a request to receipts.pay method.
// Returns PayReceiptResponse with payment details or an error.
func (c *Client) PayReceipt(ctx context.Context, receiptID, token string) (*PayReceiptResponse, error) {
	// Validation
	if err := ValidateReceiptID(receiptID); err != nil {
		return nil, err
	}
	if err := ValidateCardToken(token); err != nil {
		return nil, err
	}

	requestID := GenerateRequestID("ReceiptsPay")

	receiptParams := map[string]interface{}{
		"id":    receiptID,
		"token": token,
	}

	resp, err := c.sendRequest(ctx, requestID, "receipts.pay", receiptParams, false)
	if err != nil {
		return nil, err
	}

	// Parse result
	var result PayReceiptResponse
	if resp.Result != nil {
		resultBytes, _ := json.Marshal(resp.Result)
		if err := json.Unmarshal(resultBytes, &result); err != nil {
			return nil, fmt.Errorf("result unmarshal error: %w", err)
		}
	}

	return &result, nil
}

// SendReceipt sends a receipt to the customer.
// It validates receipt ID and sends a request to receipts.send method.
// Returns SendReceiptResponse with send details or an error.
func (c *Client) SendReceipt(ctx context.Context, receiptID string) (*SendReceiptResponse, error) {
	// Validation
	if err := ValidateReceiptID(receiptID); err != nil {
		return nil, err
	}

	requestID := GenerateRequestID("ReceiptsSend")

	receiptParams := map[string]interface{}{
		"id": receiptID,
	}

	resp, err := c.sendRequest(ctx, requestID, "receipts.send", receiptParams, false)
	if err != nil {
		return nil, err
	}

	// Parse result
	var result SendReceiptResponse
	if resp.Result != nil {
		resultBytes, _ := json.Marshal(resp.Result)
		if err := json.Unmarshal(resultBytes, &result); err != nil {
			return nil, fmt.Errorf("result unmarshal error: %w", err)
		}
	}

	return &result, nil
}

// CancelReceipt cancels an existing receipt.
// It validates receipt ID and sends a request to receipts.cancel method.
// Returns CancelReceiptResponse with cancellation details or an error.
func (c *Client) CancelReceipt(ctx context.Context, receiptID string) (*CancelReceiptResponse, error) {
	// Validation
	if err := ValidateReceiptID(receiptID); err != nil {
		return nil, err
	}

	requestID := GenerateRequestID("ReceiptsCancel")

	receiptParams := map[string]interface{}{
		"id": receiptID,
	}

	resp, err := c.sendRequest(ctx, requestID, "receipts.cancel", receiptParams, false)
	if err != nil {
		return nil, err
	}

	// Parse result
	var result CancelReceiptResponse
	if resp.Result != nil {
		resultBytes, _ := json.Marshal(resp.Result)
		if err := json.Unmarshal(resultBytes, &result); err != nil {
			return nil, fmt.Errorf("result unmarshal error: %w", err)
		}
	}

	return &result, nil
}

// CheckReceipt checks the status of an existing receipt.
// It validates receipt ID and sends a request to receipts.check method.
// Returns CheckReceiptResponse with receipt status or an error.
func (c *Client) CheckReceipt(ctx context.Context, receiptID string) (*CheckReceiptResponse, error) {
	// Validation
	if err := ValidateReceiptID(receiptID); err != nil {
		return nil, err
	}

	requestID := GenerateRequestID("ReceiptsCheck")

	receiptParams := map[string]interface{}{
		"id": receiptID,
	}

	resp, err := c.sendRequest(ctx, requestID, "receipts.check", receiptParams, false)
	if err != nil {
		return nil, err
	}

	// Parse result
	var result CheckReceiptResponse
	if resp.Result != nil {
		resultBytes, _ := json.Marshal(resp.Result)
		if err := json.Unmarshal(resultBytes, &result); err != nil {
			return nil, fmt.Errorf("result unmarshal error: %w", err)
		}
	}

	return &result, nil
}

// GetReceipt retrieves detailed information about an existing receipt.
// It validates receipt ID and sends a request to receipts.get method.
// Returns GetReceiptResponse with receipt details or an error.
func (c *Client) GetReceipt(ctx context.Context, receiptID string) (*GetReceiptResponse, error) {
	// Validation
	if err := ValidateReceiptID(receiptID); err != nil {
		return nil, err
	}

	requestID := GenerateRequestID("ReceiptsGet")

	receiptParams := map[string]interface{}{
		"id": receiptID,
	}

	resp, err := c.sendRequest(ctx, requestID, "receipts.get", receiptParams, false)
	if err != nil {
		return nil, err
	}

	// Parse result
	var result GetReceiptResponse
	if resp.Result != nil {
		resultBytes, _ := json.Marshal(resp.Result)
		if err := json.Unmarshal(resultBytes, &result); err != nil {
			return nil, fmt.Errorf("result unmarshal error: %w", err)
		}
	}

	return &result, nil
}

// GetAllReceipts retrieves multiple receipts within a specified time range.
// It sends a request to receipts.get_all method with time parameters.
// Returns GetAllReceiptsResponse with receipt list or an error.
func (c *Client) GetAllReceipts(ctx context.Context, from, to int64, count int) (*GetAllReceiptsResponse, error) {
	requestID := GenerateRequestID("ReceiptsGetAll")

	receiptParams := map[string]interface{}{
		"from":  from,
		"to":    to,
		"count": count,
	}

	resp, err := c.sendRequest(ctx, requestID, "receipts.get_all", receiptParams, false)
	if err != nil {
		return nil, err
	}

	// Parse result
	var result GetAllReceiptsResponse
	if resp.Result != nil {

		if c.Logger != nil {
			resultBytes, _ := json.Marshal(resp.Result)
			c.Logger.Printf("GetAllReceipts response: %s", string(resultBytes))
		}

		resultBytes, _ := json.Marshal(resp.Result)
		if err := json.Unmarshal(resultBytes, &result.Receipts); err != nil {
			return nil, fmt.Errorf("result unmarshal error: %w", err)
		}
	}

	return &result, nil
}

// SetFiscalData sets fiscal data for an existing receipt.
// It validates receipt ID and sends a request to receipts.set_fiscal_data method.
// Returns SetFiscalDataResponse with fiscal data details or an error.
func (c *Client) SetFiscalData(ctx context.Context, receiptID string, fiscalData map[string]interface{}) (*SetFiscalDataResponse, error) {
	// Validation
	if err := ValidateReceiptID(receiptID); err != nil {
		return nil, err
	}

	requestID := GenerateRequestID("ReceiptsSetFiscalData")

	receiptParams := map[string]interface{}{
		"id":          receiptID,
		"fiscal_data": fiscalData,
	}

	resp, err := c.sendRequest(ctx, requestID, "receipts.set_fiscal_data", receiptParams, false)
	if err != nil {
		return nil, err
	}

	// Parse result
	var result SetFiscalDataResponse
	if resp.Result != nil {
		resultBytes, _ := json.Marshal(resp.Result)
		if err := json.Unmarshal(resultBytes, &result); err != nil {
			return nil, fmt.Errorf("result unmarshal error: %w", err)
		}
	}

	return &result, nil
}

// ===== ADVANCED RECEIPT METHODS =====

// CreateAndPayReceipt creates a receipt and immediately processes payment.
// This is a convenience method that combines CreateReceipt and PayReceipt.
// Returns PayReceiptResponse with payment details or an error.
func (c *Client) CreateAndPayReceipt(ctx context.Context, amount int64, account map[string]interface{}, description string, token string) (*PayReceiptResponse, error) {
	// Create receipt
	createResp, err := c.CreateReceipt(ctx, amount, account, description, nil)
	if err != nil {
		return nil, fmt.Errorf("create receipt error: %w", err)
	}

	// Pay receipt
	payResp, err := c.PayReceipt(ctx, createResp.Receipt.ID, token)
	if err != nil {
		return nil, fmt.Errorf("pay receipt error: %w", err)
	}

	return payResp, nil
}
