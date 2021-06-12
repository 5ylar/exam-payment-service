package payment

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/omise/omise-go"
	"github.com/omise/omise-go/operations"
)

type omiseProvider interface {
	CreateSource(createSource operations.CreateSource) (omise.Source, error)
	CreateCharge(createCharge operations.CreateCharge) (omise.Charge, error)
}

type Payment struct {
	oc omiseProvider
	db *sql.DB
}

func New(oc omiseProvider, db *sql.DB) *Payment {
	return &Payment{
		oc,
		db,
	}
}

func (p Payment) CreatePaymentRequest(ctx context.Context, pr PaymentRequest) (PaymentRequestResult, error) {

	currencyS := string(pr.Currency)
	amount := pr.Amount

	// Validation
	switch pr.Currency {
	case CurrencyTHB:
		if amount < ChargeLimitTHBMin {
			return PaymentRequestResult{}, ErrAmountLowerThanChargeLimit
		}

		if amount > ChargeLimitTHBMax {
			return PaymentRequestResult{}, ErrChargeLimitExceeded
		}
	}

	if !pr.Currency.Validate() {
		return PaymentRequestResult{}, ErrInvalidCurrency
	}

	if !pr.SourceType.Validate() {
		return PaymentRequestResult{}, ErrInvalidSourceType
	}

	source, err := p.oc.CreateSource(operations.CreateSource{
		Amount:   amount,
		Currency: currencyS,
		Type:     string(pr.SourceType),
	})
	if err != nil {
		return PaymentRequestResult{}, err
	}

	charge, err := p.oc.CreateCharge(operations.CreateCharge{
		Amount:    amount,
		Currency:  currencyS,
		ReturnURI: pr.ReturnURI,
		Source:    source.ID,
	})
	if err != nil {
		return PaymentRequestResult{}, err
	}

	rs := PaymentRequestResult{
		ChargeID:     charge.ID,
		SourceID:     source.ID,
		AuthorizeURI: charge.AuthorizeURI,
	}

	return rs, nil
}

func (p Payment) GetPaymentStatusWithChargeID(ctx context.Context, chargeID string) (PaymentStatus, error) {
	var q PaymentStatus
	err := p.db.QueryRowContext(ctx, "SELECT status FROM payments WHERE charge_id = ?", chargeID).Scan(&q.Status)
	if err != nil {
		return PaymentStatus{}, err
	}

	return q, nil
}

func (p Payment) HookPaymentEvent(ctx context.Context, event PaymentEvent) error {
	chargeID := event.Data.ID
	sourceID := event.Data.Source.ID
	txnID := event.Data.Transaction
	status := event.Data.Status

	switch event.Key {
	case "charge.create":

		_, err := p.db.ExecContext(
			ctx,
			"INSERT INTO payments (charge_id, source_id, txn_id, status) VALUES (?, ?, ?, ?)",
			chargeID, sourceID, txnID, status,
		)
		if err != nil {
			log.Println("HookPaymentEvent err", err)
			return err
		}
	case "charge.complete":

		_, err := p.db.ExecContext(
			ctx,
			"UPDATE payments SET txn_id = ?, status = ? WHERE charge_id = ?",
			txnID, status, chargeID,
		)
		if err != nil {
			log.Println("HookPaymentEvent err", err)
			return err
		}

	}

	return nil
}

type SourceType string

var (
	SourceTypeInternetBankSCB SourceType = "internet_banking_scb"
)

func (s SourceType) Validate() bool {
	switch s {
	case SourceTypeInternetBankSCB:
		return true
	default:
		return false
	}
}

type Currency string

var (
	CurrencyTHB Currency = "thb"
)

func (c Currency) Validate() bool {
	switch c {
	case CurrencyTHB:
		return true
	default:
		return false
	}
}

type PaymentRequest struct {
	Amount     int64      `json:"amount"`
	Currency   Currency   `json:"currency"`
	ReturnURI  string     `json:"returnUri"`
	SourceType SourceType `json:"sourceType"`
}

type PaymentRequestResult struct {
	ChargeID     string `json:"chargeId"`
	SourceID     string `json:"sourceId"`
	AuthorizeURI string `json:"authorizeUri"`
}

// Capture from Omise hook payload
type PaymentEvent struct {
	CreatedAt time.Time `json:"created_at"`
	Data      struct {
		Amount          int         `json:"amount"`
		AuthorizeURI    string      `json:"authorize_uri"`
		Authorized      bool        `json:"authorized"`
		Branch          interface{} `json:"branch"` // TODO: Unknown data type
		Capturable      bool        `json:"capturable"`
		Capture         bool        `json:"capture"`
		Card            interface{} `json:"card"` // TODO: Unknown data type
		CreatedAt       time.Time   `json:"created_at"`
		Currency        string      `json:"currency"`
		Customer        interface{} `json:"customer"`    // TODO: Unknown data type
		Description     interface{} `json:"description"` // TODO: Unknown data type
		Device          interface{} `json:"device"`      // TODO: Unknown data type
		Disputable      bool        `json:"disputable"`
		Dispute         interface{} `json:"dispute"` // TODO: Unknown data type
		Expired         bool        `json:"expired"`
		ExpiredAt       time.Time   `json:"expired_at"`
		ExpiresAt       time.Time   `json:"expires_at"`
		FailureCode     interface{} `json:"failure_code"`    // TODO: Unknown data type
		FailureMessage  interface{} `json:"failure_message"` // TODO: Unknown data type
		Fee             int         `json:"fee"`
		FeeVat          int         `json:"fee_vat"`
		FundingAmount   int         `json:"funding_amount"`
		FundingCurrency string      `json:"funding_currency"`
		ID              string      `json:"id"`
		Interest        int         `json:"interest"`
		InterestVat     int         `json:"interest_vat"`
		IP              interface{} `json:"ip"`   // TODO: Unknown data type
		Link            interface{} `json:"link"` // TODO: Unknown data type
		Livemode        bool        `json:"livemode"`
		Location        string      `json:"location"`
		Metadata        struct{}    `json:"metadata"`
		Net             int         `json:"net"`
		Object          string      `json:"object"`
		Paid            bool        `json:"paid"`
		PaidAt          time.Time   `json:"paid_at"`
		PlatformFee     struct {
			Amount     interface{} `json:"amount"`     // TODO: Unknown data type
			Fixed      interface{} `json:"fixed"`      // TODO: Unknown data type
			Percentage interface{} `json:"percentage"` // TODO: Unknown data type
		} `json:"platform_fee"`
		Refundable     bool `json:"refundable"`
		RefundedAmount int  `json:"refunded_amount"`
		Refunds        struct {
			Data     []interface{} `json:"data"` // TODO: Unknown data type
			From     time.Time     `json:"from"`
			Limit    int           `json:"limit"`
			Location string        `json:"location"`
			Object   string        `json:"object"`
			Offset   int           `json:"offset"`
			Order    string        `json:"order"`
			To       time.Time     `json:"to"`
			Total    int           `json:"total"`
		} `json:"refunds"`
		ReturnURI  string      `json:"return_uri"`
		Reversed   bool        `json:"reversed"`
		ReversedAt interface{} `json:"reversed_at"` // TODO: Unknown data type
		Reversible bool        `json:"reversible"`
		Schedule   interface{} `json:"schedule"` // TODO: Unknown data type
		Source     struct {
			Amount                   int           `json:"amount"`
			Bank                     interface{}   `json:"bank"`    // TODO: Unknown data type
			Barcode                  interface{}   `json:"barcode"` // TODO: Unknown data type
			ChargeStatus             string        `json:"charge_status"`
			CreatedAt                time.Time     `json:"created_at"`
			Currency                 string        `json:"currency"`
			Discounts                []interface{} `json:"discounts"` // TODO: Unknown data type
			Email                    interface{}   `json:"email"`     // TODO: Unknown data type
			Flow                     string        `json:"flow"`
			ID                       string        `json:"id"`
			InstallmentTerm          interface{}   `json:"installment_term"` // TODO: Unknown data type
			Livemode                 bool          `json:"livemode"`
			Location                 string        `json:"location"`
			MobileNumber             interface{}   `json:"mobile_number"` // TODO: Unknown data type
			Name                     interface{}   `json:"name"`          // TODO: Unknown data type
			Object                   string        `json:"object"`
			PhoneNumber              interface{}   `json:"phone_number"`   // TODO: Unknown data type
			ReceiptAmount            interface{}   `json:"receipt_amount"` // TODO: Unknown data type
			References               interface{}   `json:"references"`     // TODO: Unknown data type
			ScannableCode            interface{}   `json:"scannable_code"` // TODO: Unknown data type
			StoreID                  interface{}   `json:"store_id"`       // TODO: Unknown data type
			StoreName                interface{}   `json:"store_name"`     // TODO: Unknown data type
			TerminalID               interface{}   `json:"terminal_id"`    // TODO: Unknown data type
			Type                     string        `json:"type"`
			ZeroInterestInstallments interface{}   `json:"zero_interest_installments"` // TODO: Unknown data type
		} `json:"source"`
		Status                   string      `json:"status"`
		Terminal                 interface{} `json:"terminal"` // TODO: Unknown data type
		Transaction              string      `json:"transaction"`
		Voided                   bool        `json:"voided"`
		ZeroInterestInstallments bool        `json:"zero_interest_installments"`
	} `json:"data"`
	ID                string        `json:"id"`
	Key               string        `json:"key"`
	Livemode          bool          `json:"livemode"`
	Location          string        `json:"location"`
	Object            string        `json:"object"`
	WebhookDeliveries []interface{} `json:"webhook_deliveries"` // TODO: Unknown data type
}

type PaymentStatus struct {
	Status string `json:"status"`
}
