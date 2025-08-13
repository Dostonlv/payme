package payment

// Response represents the base JSON-RPC response from PayMe API.
// It contains the standard JSON-RPC fields and optional result or error.
type Response struct {
	Jsonrpc string      `json:"jsonrpc"`
	ID      string      `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *Error      `json:"error,omitempty"`
}

// Error represents a PayMe API error response.
// It contains error message, code, and optional data and origin.
type Error struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
	Data    string `json:"data,omitempty"`
	Origin  string `json:"origin,omitempty"`
}

// Receipt represents a payment receipt in PayMe system.
// It contains all receipt details including amount, status, timestamps, and metadata.
type Receipt struct {
	ID           string           `json:"_id,omitempty"`
	CreateTime   int64            `json:"create_time"`
	PayTime      int64            `json:"pay_time"`
	CancelTime   int64            `json:"cancel_time"`
	State        int              `json:"state"`
	Type         int              `json:"type"`
	External     bool             `json:"external"`
	Operation    int              `json:"operation"`
	Category     *ReceiptCategory `json:"category,omitempty"`
	Error        interface{}      `json:"error"`
	Description  string           `json:"description,omitempty"`
	Detail       *ReceiptDetail   `json:"detail,omitempty"`
	Amount       int64            `json:"amount"`
	Currency     int              `json:"currency"`
	Commission   int64            `json:"commission"`
	Account      []ReceiptAccount `json:"account"`
	Card         interface{}      `json:"card"`
	Creator      interface{}      `json:"creator"`
	Payer        interface{}      `json:"payer"`
	SenderCard   interface{}      `json:"sender_card"`
	Merchant     *ReceiptMerchant `json:"merchant,omitempty"`
	Meta         *ReceiptMeta     `json:"meta,omitempty"`
	ProcessingID interface{}      `json:"processing_id"`
}

// ReceiptCategory represents the category information for a receipt.
// It includes category title, color, icon, and merchant category codes.
type ReceiptCategory struct {
	ID        string                 `json:"_id"`
	Title     map[string]interface{} `json:"title"`
	Color     string                 `json:"color"`
	Sort      int                    `json:"sort"`
	Operation int                    `json:"operation"`
	Indoor    bool                   `json:"indoor"`
	MCC       *ReceiptMCC            `json:"mcc,omitempty"`
	Icon      string                 `json:"icon"`
}

// ReceiptMCC contains merchant category codes for the receipt.
// These codes help classify the type of business transaction.
type ReceiptMCC struct {
	Visa []string `json:"visa"`
}

// ReceiptDetail contains additional details about the receipt.
// It includes discount, shipping, and items information.
type ReceiptDetail struct {
	Discount interface{} `json:"discount"`
	Shipping interface{} `json:"shipping"`
	Items    interface{} `json:"items"`
}

// ReceiptAccount represents account information associated with a receipt.
// It includes account name, title, value, and whether it's the main account.
type ReceiptAccount struct {
	Name  string                 `json:"name"`
	Title map[string]interface{} `json:"title"`
	Value interface{}            `json:"value"`
	Main  bool                   `json:"main"`
}

// ReceiptMerchant contains merchant information for the receipt.
// It includes merchant details, business information, and EPOS terminal data.
type ReceiptMerchant struct {
	ID           string                 `json:"_id"`
	Name         string                 `json:"name"`
	Organization string                 `json:"organization"`
	Address      string                 `json:"address"`
	BusinessID   string                 `json:"business_id"`
	Epos         *ReceiptEpos           `json:"epos"`
	Restrictions interface{}            `json:"restrictions"`
	Date         int64                  `json:"date"`
	Logo         interface{}            `json:"logo"`
	Type         map[string]interface{} `json:"type"`
	Terms        interface{}            `json:"terms"`
}

// ReceiptEpos contains EPOS terminal information for the receipt.
// It includes merchant ID and terminal ID for payment processing.
type ReceiptEpos struct {
	MerchantID string `json:"merchantId"`
	TerminalID string `json:"terminalId"`
}

// ReceiptMeta contains metadata information for the receipt.
// It includes source, owner, and host information.
type ReceiptMeta struct {
	Source string `json:"source"`
	Owner  string `json:"owner"`
	Host   string `json:"host"`
}

// Card represents a payment card in PayMe system.
// It contains card number, expiry date, token, and verification status.
type Card struct {
	Number     string `json:"number"`
	Expire     string `json:"expire"`
	Token      string `json:"token"`
	Recurrent  bool   `json:"recurrent"`
	Verify     bool   `json:"verify"`
	Type       string `json:"type"`
	NumberHash string `json:"number_hash"`
}

// CardData contains basic card information for payment processing.
// It includes card ID and token for authentication.
type CardData struct {
	ID    string `json:"id"`
	Token string `json:"token"`
}

// PaymentData contains payment-related information.
// It includes order ID and associated card data.
type PaymentData struct {
	OrderID  string   `json:"order_id"`
	CardData CardData `json:"card_data"`
}

// PaymentDetails contains complete payment information for merchant transactions.
// It includes client and driver payment data along with amount.
type PaymentDetails struct {
	Client PaymentData `json:"client"`
	Driver PaymentData `json:"driver"`
	Amount int         `json:"amount"`
}

// Account contains account information for receipt creation.
// It includes charge ID, card ID, and reason for the transaction.
type Account struct {
	ID     string `json:"id"`
	CardID string `json:"card_id"`
	Reason string `json:"reason"`
}

// CreateReceiptResponse contains the response from receipts.create method.
// It includes the created receipt details.
type CreateReceiptResponse struct {
	Receipt *Receipt `json:"receipt"`
}

// PayReceiptResponse contains the response from receipts.pay method.
// It includes the payment details and receipt information.
type PayReceiptResponse struct {
	Receipt *Receipt `json:"receipt"`
}

// SendReceiptResponse contains the response from receipts.send method.
// It includes the send confirmation and receipt details.
type SendReceiptResponse struct {
	Receipt *Receipt `json:"receipt"`
}

// CancelReceiptResponse contains the response from receipts.cancel method.
// It includes the cancellation confirmation and receipt details.
type CancelReceiptResponse struct {
	Receipt *Receipt `json:"receipt"`
}

// CheckReceiptResponse contains the response from receipts.check method.
// It includes the receipt status and basic information.
type CheckReceiptResponse struct {
	Receipt *Receipt `json:"receipt"`
}

// GetReceiptResponse contains the response from receipts.get method.
// It includes the complete receipt details and information.
type GetReceiptResponse struct {
	Receipt *Receipt `json:"receipt"`
}

// GetAllReceiptsResponse contains the response from receipts.get_all method.
// It includes a list of receipts within the specified time range.
type GetAllReceiptsResponse struct {
	Receipts []*Receipt `json:"-"`
}

// SetFiscalDataResponse contains the response from receipts.set_fiscal_data method.
// It includes the fiscal data confirmation and receipt details.
type SetFiscalDataResponse struct {
	Receipt *Receipt `json:"receipt"`
}

// ===== TRANSACTION TYPES =====

// Transaction represents a payment transaction
type Transaction struct {
	ID          string `json:"_id"`
	CreateTime  int64  `json:"create_time"`
	PerformTime int64  `json:"perform_time"`
	CancelTime  int64  `json:"cancel_time"`
	State       int    `json:"state"`
	Reason      *int   `json:"reason,omitempty"`
	Amount      int64  `json:"amount"`
	Currency    int    `json:"currency"`
	ReceiptID   string `json:"receipt_id"`
}
