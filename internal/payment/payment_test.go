package payment

import (
	"context"
	"database/sql"
	mockOmiseProvider "exam-payment-service/internal/payment/mocks/omiseprovider"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang/mock/gomock"
	"github.com/omise/omise-go"
	"github.com/omise/omise-go/operations"
	"github.com/stretchr/testify/assert"
)

func TestCreatePaymentRequest(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name            string
		amount          int64
		currency        Currency
		sourceType      SourceType
		returnURI       string
		sourceID        string
		chargeID        string
		authorizeURI    string
		expectedError   error
		expectedResult  PaymentRequestResult
		errorValidation bool
	}{
		{
			name:            "Amount lower than charge limit",
			amount:          1000,
			currency:        CurrencyTHB,
			sourceType:      SourceTypeInternetBankSCB,
			returnURI:       "https://example.com",
			sourceID:        "source_xxx",
			chargeID:        "charge_xxx",
			authorizeURI:    "https://example.com/pay",
			expectedError:   ErrAmountLowerThanChargeLimit,
			expectedResult:  PaymentRequestResult{},
			errorValidation: true,
		},
		{
			name:            "Amount more than charge limit",
			amount:          20000000,
			currency:        CurrencyTHB,
			sourceType:      SourceTypeInternetBankSCB,
			returnURI:       "https://example.com",
			sourceID:        "source_xxx",
			chargeID:        "charge_xxx",
			authorizeURI:    "https://example.com/pay",
			expectedError:   ErrChargeLimitExceeded,
			expectedResult:  PaymentRequestResult{},
			errorValidation: true,
		},
		{
			name:          "Success",
			amount:        20000,
			currency:      CurrencyTHB,
			sourceType:    SourceTypeInternetBankSCB,
			returnURI:     "https://example.com",
			sourceID:      "source_xxx",
			chargeID:      "charge_xxx",
			authorizeURI:  "https://example.com/pay",
			expectedError: nil,
			expectedResult: PaymentRequestResult{
				SourceID:     "source_xxx",
				ChargeID:     "charge_xxx",
				AuthorizeURI: "https://example.com/pay",
			},
		},
		{
			name:            "Invalid currency value",
			amount:          20000,
			currency:        Currency("test"),
			sourceType:      SourceTypeInternetBankSCB,
			returnURI:       "https://example.com",
			sourceID:        "source_xxx",
			chargeID:        "charge_xxx",
			authorizeURI:    "https://example.com/pay",
			expectedError:   ErrInvalidCurrency,
			expectedResult:  PaymentRequestResult{},
			errorValidation: true,
		},
		{
			name:            "Invalid source type value",
			amount:          20000,
			currency:        CurrencyTHB,
			sourceType:      SourceType("test"),
			returnURI:       "https://example.com",
			sourceID:        "source_xxx",
			chargeID:        "charge_xxx",
			authorizeURI:    "https://example.com/pay",
			expectedError:   ErrInvalidSourceType,
			expectedResult:  PaymentRequestResult{},
			errorValidation: true,
		},
	}

	t.Parallel()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockCtl := gomock.NewController(t)

			op := mockOmiseProvider.NewMockOmiseProvider(mockCtl)

			if !tc.errorValidation {
				op.EXPECT().CreateSource(operations.CreateSource{
					Amount:   tc.amount,
					Currency: string(tc.currency),
					Type:     string(tc.sourceType),
				}).Return(omise.Source{
					ID: tc.sourceID,
				}, nil)

				returnCharge := omise.Charge{
					AuthorizeURI: tc.authorizeURI,
				}
				returnCharge.ID = tc.chargeID
				op.EXPECT().CreateCharge(operations.CreateCharge{
					Amount:    tc.amount,
					Currency:  string(tc.currency),
					ReturnURI: tc.returnURI,
					Source:    tc.sourceID,
				}).Return(returnCharge, nil)
			}

			p := New(op, nil)

			result, err := p.CreatePaymentRequest(ctx, PaymentRequest{
				Amount:     tc.amount,
				Currency:   tc.currency,
				ReturnURI:  tc.returnURI,
				SourceType: tc.sourceType,
			})

			assert.Equal(t, tc.expectedError, err)
			assert.Equal(t, tc.expectedResult, result)

		})
	}

}

func TestGetTxnStatusWithChargeID(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name           string
		chargeID       string
		expectedError  error
		expectedResult PaymentStatus
		addRow         bool
	}{
		{
			name:     "Found",
			chargeID: "charge_xxx",
			expectedResult: PaymentStatus{
				Status: "successful",
			},
			addRow: true,
		},
		{
			name:           "Not found",
			chargeID:       "charge_xxx",
			expectedResult: PaymentStatus{},
			expectedError:  sql.ErrNoRows,
		},
	}

	t.Parallel()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			db, mock, err := sqlmock.New()
			if err != nil {
				t.Error(err)
			}
			defer db.Close()

			if tc.addRow {
				rows := sqlmock.NewRows([]string{"status"}).AddRow(tc.expectedResult.Status)
				mock.ExpectQuery("SELECT status FROM payments WHERE charge_id = ?").WithArgs(tc.chargeID).WillReturnRows(rows)
			} else {
				mock.ExpectQuery("SELECT status FROM payments WHERE charge_id = ?").WithArgs(tc.chargeID).WillReturnError(tc.expectedError)
			}

			p := New(nil, db)

			result, err := p.GetPaymentStatusWithChargeID(ctx, tc.chargeID)

			assert.Equal(t, tc.expectedError, err)
			assert.Equal(t, tc.expectedResult, result)

		})
	}

}

func TestHookPaymentEvent(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name          string
		event         PaymentEvent
		expectedError error
	}{
		{
			name: "Created",
			event: func() PaymentEvent {
				p := PaymentEvent{
					Key: "charge.create",
				}
				p.Data.ID = "charge_xxx"
				p.Data.Source.ID = "source_xxx"
				p.Data.Status = "pending"
				return p
			}(),
		},
		{
			name: "Success",
			event: func() PaymentEvent {
				p := PaymentEvent{
					Key: "charge.complete",
				}
				p.Data.ID = "charge_xxx"
				p.Data.Source.ID = "source_xxx"
				p.Data.Transaction = "transaction_xxx"
				p.Data.Status = "successful"
				return p
			}(),
		},
		{
			name: "Failed",
			event: func() PaymentEvent {
				p := PaymentEvent{
					Key: "charge.complete",
				}
				p.Data.ID = "charge_xxx"
				p.Data.Source.ID = "source_xxx"
				p.Data.Status = "failed"
				return p
			}(),
		},
		{
			name: "Not matched key",
			event: func() PaymentEvent {
				p := PaymentEvent{
					Key: "charge.something",
				}
				return p
			}(),
		},
	}

	t.Parallel()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var db *sql.DB

			if tc.event.Key == "charge.create" || tc.event.Key == "charge.complete" {
				var (
					mock sqlmock.Sqlmock
					err  error
				)
				db, mock, err = sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
				if err != nil {
					t.Error(err)
				}
				defer db.Close()

				switch tc.event.Key {
				case "charge.create":
					mock.ExpectExec("INSERT INTO payments (charge_id, source_id, txn_id, status) VALUES (?, ?, ?, ?)").
						WithArgs(tc.event.Data.ID, tc.event.Data.Source.ID, tc.event.Data.Transaction, tc.event.Data.Status).
						WillReturnResult(sqlmock.NewResult(1, 1))
				case "charge.complete":
					mock.ExpectExec("UPDATE payments SET txn_id = ?, status = ? WHERE charge_id = ?").
						WithArgs(tc.event.Data.Transaction, tc.event.Data.Status, tc.event.Data.ID).
						WillReturnResult(sqlmock.NewResult(1, 1))
				}
			}

			p := New(nil, db)

			err := p.HookPaymentEvent(ctx, tc.event)

			assert.Equal(t, tc.expectedError, err)

		})
	}

}
