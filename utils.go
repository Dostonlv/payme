package payment

import (
	"encoding/json"
	"fmt"
	"time"
)

var (
	TestEndpoint       = "https://checkout.test.paycom.uz/api"
	ProductionEndpoint = "https://checkout.paycom.uz/api"
)

// ===== PAYME CURRENCY CODES =====

const (
	CurrencyUZS = 860 // Uzbekistan Som
	CurrencyUSD = 840 // US Dollar
	CurrencyEUR = 978 // Euro
)

// SomToTiyin converts Uzbek som to tiyin (smallest currency unit).
// PayMe API expects amounts in tiyin, not som.
// Returns the amount in tiyin as int64.
func SomToTiyin(som float64) int64 {
	return int64(som * 100)
}

// TiyinToSom converts tiyin (smallest currency unit) to Uzbek som.
// Useful for displaying amounts in human-readable format.
// Returns the amount in som as float64.
func TiyinToSom(tiyin int64) float64 {
	return float64(tiyin) / 100
}

func FromSomToTiyin(amount int) int {
	return amount * 100
}

func FromTiyinToSom(amount int) int {
	return amount / 100
}

func GetCurrentTimestamp() int64 {
	return time.Now().UnixMilli()
}

// FormatTimestamp converts Unix timestamp to human-readable date format.
// Converts milliseconds timestamp to "YYYY-MM-DD HH:MM:SS" format.
// Returns a formatted date string.
func FormatTimestamp(timestamp int64) string {
	return time.UnixMilli(timestamp).Format("2006-01-02 15:04:05")
}

func ParseTimestamp(timestampStr string) (int64, error) {
	t, err := time.Parse("2006-01-02 15:04:05", timestampStr)
	if err != nil {
		return 0, err
	}
	return t.UnixMilli(), nil
}

func ValidateAmount(amount int64) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}
	if amount > 999999999999 {
		return ErrInvalidAmount
	}
	return nil
}

func ValidateCardToken(token string) error {
	if token == "" {
		return ErrInvalidFormatToken
	}
	if len(token) < 10 || len(token) > 100 {
		return ErrInvalidFormatToken
	}
	return nil
}

func ValidateReceiptID(id string) error {
	if id == "" {
		return ErrReceiptNotFound
	}
	if len(id) < 5 || len(id) > 100 {
		return ErrReceiptNotFound
	}
	return nil
}

// GenerateRequestID creates a unique request identifier for PayMe API calls.
// It combines a prefix with a UUID to ensure uniqueness across requests.
// Returns a string in format "prefix-uuid".
func GenerateRequestID(prefix string) string {
	return fmt.Sprintf("%s:%d", prefix, time.Now().UnixNano())
}

func GenerateReceiptID(chargeID string) string {
	return fmt.Sprintf("receipt_%s_%d", chargeID, time.Now().Unix())
}

// IsValidCurrency validates if the provided currency code is supported.
// Currently supports UZS (860), USD (840), and EUR (978).
// Returns true if currency is valid, false otherwise.
func IsValidCurrency(currency int) bool {
	validCurrencies := []int{
		CurrencyUZS, CurrencyUSD, CurrencyEUR,
	}
	for _, valid := range validCurrencies {
		if valid == currency {
			return true
		}
	}
	return false
}

func IsValidReceiptState(state int) bool {
	validStates := []int{0, 1, -1, -2} // Created, Paid, Canceled, Expired
	for _, valid := range validStates {
		if valid == state {
			return true
		}
	}
	return false
}

func SafeUnmarshal(data []byte, v interface{}) error {
	if len(data) == 0 {
		return fmt.Errorf("empty data")
	}
	return json.Unmarshal(data, v)
}

func SafeMarshal(v interface{}) ([]byte, error) {
	if v == nil {
		return nil, fmt.Errorf("nil value")
	}
	return json.Marshal(v)
}

func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func MaskCardNumber(number string) string {
	if len(number) < 8 {
		return "****"
	}
	return number[:4] + "****" + number[len(number)-4:]
}

// FormatAmount formats a monetary amount with appropriate currency symbol.
// Converts tiyin to som for UZS, and formats with currency symbols.
// Returns a formatted string like "100.00 сум" or "$10.00".
func FormatAmount(amount int64, currency int) string {
	currencySymbol := "сум"
	switch currency {
	case CurrencyUSD:
		currencySymbol = "$"
	case CurrencyEUR:
		currencySymbol = "€"
	}

	formattedAmount := fmt.Sprintf("%.2f", TiyinToSom(amount))
	return formattedAmount + " " + currencySymbol
}
