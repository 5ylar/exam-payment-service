package payment

import "errors"

var (
	ErrAmountLowerThanChargeLimit = errors.New("amount is lower than charge limit")
	ErrChargeLimitExceeded        = errors.New("charge limit exceeded")
	ErrInvalidCurrency            = errors.New("invalid currency")
	ErrInvalidSourceType          = errors.New("invalid source type")
)
