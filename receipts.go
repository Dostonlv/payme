package payment

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// ===== RECEIPT CONSTANTS =====

const (
	PayForOrderReasonID = "6"
	Description         = "Merchant transaction for order - %s"
)

// CreateMerchantReceipt creates a merchant receipt with dynamic account field mapping.
// It uses the client's RequisiteName configuration to set the account identifier.
// This method is useful when the account field name varies between different systems.
// Returns the created receipt ID as string or an error.
func (c *Client) CreateMerchantReceipt(ctx context.Context, data PaymentDetails) (string, error) {
	requestID := fmt.Sprintf("ReceiptsCreate:MerchantTransaction:%s", data.Client.OrderID)

	amountInTiyin := FromSomToTiyin(data.Amount)

	receiptParams := map[string]interface{}{
		"amount": amountInTiyin,
		"account": map[string]interface{}{
			c.RequisiteName: data.Client.OrderID,
			"card_id":       data.Client.CardData.ID,
			"reason":        PayForOrderReasonID, // payment for order
		},
		"description": fmt.Sprintf(Description, data.Client.OrderID),
	}

	resp, err := c.sendRequest(ctx, requestID, "receipts.create", receiptParams, false)
	if err != nil {
		return "", fmt.Errorf("failed receipts create (error - %v request-id - %s response - %v)", err, requestID, resp)
	}

	var result CreateReceiptResponse
	if resp.Result != nil {
		resultBytes, _ := json.Marshal(resp.Result)
		if err := json.Unmarshal(resultBytes, &result); err != nil {
			return "", fmt.Errorf("result unmarshal error: %w", err)
		}
	}

	createdReceiptsID := result.Receipt.ID

	if c.Logger != nil {
		c.Logger.Printf("receipts created for order - %v request-id - %s transaction-id - %s", data.Client.OrderID, requestID, createdReceiptsID)
	}

	return createdReceiptsID, nil
}

// PayMerchantReceipt processes payment for an existing merchant receipt.
// It uses the receipt ID and card token to complete the payment.
// Returns the paid receipt ID as string or an error.
func (c *Client) PayMerchantReceipt(ctx context.Context, data PaymentDetails, createdReceiptsID string) (string, error) {

	requestID := fmt.Sprintf("ReceiptsPay:%s", data.Client.OrderID)

	receiptParams := map[string]interface{}{
		"id":    createdReceiptsID,
		"token": data.Client.CardData.Token,
	}

	resp, err := c.sendRequest(ctx, requestID, "receipts.pay", receiptParams, false)
	if err != nil {
		return "", fmt.Errorf("failed receipts pay (error - %v request-id - %s receipts-id %s response - %v)", err, requestID, createdReceiptsID, resp.Error)
	}

	var result PayReceiptResponse
	if resp.Result != nil {
		resultBytes, _ := json.Marshal(resp.Result)
		if err := json.Unmarshal(resultBytes, &result); err != nil {
			return "", fmt.Errorf("result unmarshal error: %w", err)
		}
	}

	paidReceiptsID := result.Receipt.ID

	if c.Logger != nil {
		c.Logger.Printf("receipts paid for order - %v request-id - %s transaction-id - %s", data.Client.OrderID, requestID, paidReceiptsID)
	}

	return paidReceiptsID, nil
}

// CreateAndPayMerchantReceipt creates a merchant receipt and immediately processes payment.
// This is a convenience method that combines CreateMerchantReceipt and PayMerchantReceipt.
// Returns the final receipt ID as string or an error.
func (c *Client) CreateAndPayMerchantReceipt(ctx context.Context, data PaymentDetails) (string, error) {

	createdReceiptsID, err := c.CreateMerchantReceipt(ctx, data)
	if err != nil {
		return "", err
	}

	paidReceiptsID, err := c.PayMerchantReceipt(ctx, data, createdReceiptsID)
	if err != nil {
		return "", err
	}

	return paidReceiptsID, nil
}

// GetReceiptStatus retrieves the current state of a receipt.
// It calls CheckReceipt internally and returns the state value.
// Returns the receipt state as int or -1 if error occurs.
func (c *Client) GetReceiptStatus(ctx context.Context, receiptID string) (int, error) {
	resp, err := c.CheckReceipt(ctx, receiptID)
	if err != nil {
		return -1, err
	}

	return resp.Receipt.State, nil
}

// IsReceiptPaid checks if a receipt has been paid.
// It checks if the receipt state equals 1 (Paid).
// Returns true if paid, false otherwise.
func (c *Client) IsReceiptPaid(ctx context.Context, receiptID string) (bool, error) {
	state, err := c.GetReceiptStatus(ctx, receiptID)
	if err != nil {
		return false, err
	}

	return state == 1, nil // 1 = Paid
}

// IsReceiptCanceled checks if a receipt has been canceled.
// It checks if the receipt state equals -1 (Canceled).
// Returns true if canceled, false otherwise.
func (c *Client) IsReceiptCanceled(ctx context.Context, receiptID string) (bool, error) {
	state, err := c.GetReceiptStatus(ctx, receiptID)
	if err != nil {
		return false, err
	}

	return state == -1, nil // -1 = Canceled
}

// IsReceiptExpired checks if a receipt has expired.
// It checks if the receipt state equals -2 (Expired).
// Returns true if expired, false otherwise.
func (c *Client) IsReceiptExpired(ctx context.Context, receiptID string) (bool, error) {
	state, err := c.GetReceiptStatus(ctx, receiptID)
	if err != nil {
		return false, err
	}

	return state == -2, nil // -2 = Expired
}

// CreateMultipleReceipts creates multiple receipts in a single call.
// It processes each receipt in the slice and returns a list of created receipt IDs.
// Returns a slice of receipt IDs and any error that occurred.
func (c *Client) CreateMultipleReceipts(ctx context.Context, receipts []map[string]interface{}) ([]string, error) {
	var receiptIDs []string

	for _, receipt := range receipts {
		amount, ok := receipt["amount"].(int64)
		if !ok {
			continue
		}

		account, ok := receipt["account"].(map[string]interface{})
		if !ok {
			continue
		}

		description, _ := receipt["description"].(string)
		detail, _ := receipt["detail"].(map[string]interface{})

		resp, err := c.CreateReceipt(ctx, amount, account, description, detail)
		if err != nil {
			continue
		}

		receiptIDs = append(receiptIDs, resp.Receipt.ID)
	}

	return receiptIDs, nil
}

// CancelMultipleReceipts cancels multiple receipts in a single call.
// It attempts to cancel each receipt and logs any errors that occur.
// Returns nil if all receipts were processed, regardless of individual success.
func (c *Client) CancelMultipleReceipts(ctx context.Context, receiptIDs []string) error {
	for _, receiptID := range receiptIDs {
		_, err := c.CancelReceipt(ctx, receiptID)
		if err != nil {

			if c.Logger != nil {
				c.Logger.Printf("Failed to cancel receipt %s: %v", receiptID, err)
			}
		}
	}

	return nil
}

// GetReceiptsByDateRange retrieves receipts within a specific date range.
// It converts time.Time to Unix timestamp and calls GetAllReceipts.
// Returns GetAllReceiptsResponse with receipts in the specified range.
func (c *Client) GetReceiptsByDateRange(ctx context.Context, from, to time.Time, limit int) (*GetAllReceiptsResponse, error) {
	fromTimestamp := from.UnixMilli()
	toTimestamp := to.UnixMilli()

	return c.GetAllReceipts(ctx, fromTimestamp, toTimestamp, limit)
}

// GetReceiptsByState retrieves receipts filtered by their state.
// Since PayMe API doesn't directly support state filtering, it fetches all receipts
// and filters them locally by the specified state.
// Returns GetAllReceiptsResponse with filtered receipts.
func (c *Client) GetReceiptsByState(ctx context.Context, state int, limit int) (*GetAllReceiptsResponse, error) {
	from := time.Now().AddDate(0, -1, 0).UnixMilli()
	to := time.Now().UnixMilli()

	resp, err := c.GetAllReceipts(ctx, from, to, limit)
	if err != nil {
		return nil, err
	}

	var filteredReceipts []*Receipt
	for _, receipt := range resp.Receipts {
		if receipt.State == state {
			filteredReceipts = append(filteredReceipts, receipt)
		}
	}

	return &GetAllReceiptsResponse{
		Receipts: filteredReceipts,
	}, nil
}

// GetReceiptsByAmountRange retrieves receipts filtered by amount range.
// Since PayMe API doesn't directly support amount filtering, it fetches all receipts
// and filters them locally by the specified amount range.
// Returns GetAllReceiptsResponse with filtered receipts.
func (c *Client) GetReceiptsByAmountRange(ctx context.Context, minAmount, maxAmount int64, limit int) (*GetAllReceiptsResponse, error) {
	from := time.Now().AddDate(0, -1, 0).UnixMilli()
	to := time.Now().UnixMilli()

	resp, err := c.GetAllReceipts(ctx, from, to, limit)
	if err != nil {
		return nil, err
	}

	var filteredReceipts []*Receipt
	for _, receipt := range resp.Receipts {
		if receipt.Amount >= minAmount && receipt.Amount <= maxAmount {
			filteredReceipts = append(filteredReceipts, receipt)
		}
	}

	return &GetAllReceiptsResponse{
		Receipts: filteredReceipts,
	}, nil
}
