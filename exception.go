package payment

import "errors"

const (
	InvalidAmountErrorCode      = -31611
	InvalidParamsErrorCode      = -32602
	ReceiptNotFoundErrorCode    = -31401
	ReceiptAlreadyPaidErrorCode = -31402
	ReceiptExpiredErrorCode     = -31403

	CardNotFoundErrorCode       = -31400
	InvalidFormatTokenErrorCode = -32500
	CardNumberNotFoundCode      = -31300
	CardExpiredCode             = -31301
	P2PIdenticalCardsErrorCode  = -31630

	PaycomServiceNotAvailableCode    = -31001
	ProcessingCenterNotAvailableCode = -31002

	PermissionDeniedCode = -32504
	ParseErrorCode       = -32700
	MethodNotFoundCode   = -32601
	InvalidRequestCode   = -32600
)

var (
	ErrReceiptNotFound    = errors.New("receipt not found")
	ErrReceiptAlreadyPaid = errors.New("receipt already paid")
	ErrReceiptExpired     = errors.New("receipt expired")
	ErrInvalidAmount      = errors.New("invalid amount")
	ErrInvalidParams      = errors.New("invalid parameters")

	ErrCardNotFound       = errors.New("card not found")
	ErrInvalidFormatToken = errors.New("invalid format token")
	ErrCardNumberNotFound = errors.New("card number not found")
	ErrCardExpired        = errors.New("card expired")
	ErrP2PIdenticalCards  = errors.New("similar cards cannot be used for P2P processing")

	ErrPaycomServiceNotAvailable    = errors.New("paycom service not available")
	ErrProcessingCenterNotAvailable = errors.New("processing center not available")

	ErrPermissionDenied = errors.New("permission denied")
	ErrParseError       = errors.New("parse error")
	ErrMethodNotFound   = errors.New("method not found")
	ErrInvalidRequest   = errors.New("invalid request")

	ErrPaymeError              = errors.New("payme error was occurred")
	ErrTimeout                 = errors.New("request timeout exceeded")
	ErrEmptyOrInvalidPaycomID  = errors.New("invalid paycom ID")
	ErrEmptyOrInvalidPaycomKey = errors.New("invalid paycom key")
)

func IsPaymeError(err error) bool {
	switch err {
	case ErrReceiptNotFound, ErrReceiptAlreadyPaid, ErrReceiptExpired,
		ErrInvalidAmount, ErrInvalidParams, ErrCardNotFound, ErrInvalidFormatToken,
		ErrCardNumberNotFound, ErrCardExpired, ErrP2PIdenticalCards,
		ErrPaycomServiceNotAvailable, ErrProcessingCenterNotAvailable,
		ErrPermissionDenied, ErrParseError, ErrMethodNotFound, ErrInvalidRequest:
		return true
	default:
		return false
	}
}

// GetErrorCode extracts the PayMe error code from a custom error.
// It checks if the error is a PayMeError and returns its code.
// Returns the error code as int, or 0 if not a PayMeError.
func GetErrorCode(err error) int {
	switch err {
	case ErrInvalidAmount:
		return InvalidAmountErrorCode
	case ErrInvalidParams:
		return InvalidParamsErrorCode
	case ErrReceiptNotFound:
		return ReceiptNotFoundErrorCode
	case ErrReceiptAlreadyPaid:
		return ReceiptAlreadyPaidErrorCode
	case ErrReceiptExpired:
		return ReceiptExpiredErrorCode
	case ErrCardNotFound:
		return CardNotFoundErrorCode
	case ErrInvalidFormatToken:
		return InvalidFormatTokenErrorCode
	case ErrCardNumberNotFound:
		return CardNumberNotFoundCode
	case ErrCardExpired:
		return CardExpiredCode
	case ErrP2PIdenticalCards:
		return P2PIdenticalCardsErrorCode
	case ErrPaycomServiceNotAvailable:
		return PaycomServiceNotAvailableCode
	case ErrProcessingCenterNotAvailable:
		return ProcessingCenterNotAvailableCode
	case ErrPermissionDenied:
		return PermissionDeniedCode
	case ErrParseError:
		return ParseErrorCode
	case ErrMethodNotFound:
		return MethodNotFoundCode
	case ErrInvalidRequest:
		return InvalidRequestCode
	default:
		return 0
	}
}
