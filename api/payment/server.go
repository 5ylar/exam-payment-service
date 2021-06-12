package payment

import (
	"database/sql"
	"exam-payment-service/internal/payment"
	"exam-payment-service/pkg/fiberhelper"
	"log"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

func Start(address string, payment *payment.Payment) {

	s := server{
		payment,
	}

	f := fiber.New()

	p := f.Group("/payments")

	p.Post("/", s.createPayment)
	p.Get("/charges/:chargeID/status", s.GetPaymentStatusWithChargeID)

	// TODO: middleware for verify event source (maybe using ip  address)
	f.Post("/webhook/omise", s.omiseWebhook)

	if err := f.Listen(address); err != nil {
		log.Println("Fiber listen error", err)
	}
}

type server struct {
	payment *payment.Payment
}

func (s server) createPayment(c *fiber.Ctx) error {
	var b payment.PaymentRequest

	if err := c.BodyParser(&b); err != nil {
		log.Println("BodyParser error", err)
		return fiberhelper.HandleErrorJSONResp(
			c,
			http.StatusBadRequest,
			"invalid request payload",
		)
	}

	result, err := s.payment.CreatePaymentRequest(c.Context(), b)
	if err != nil {
		log.Println("CreatePaymentRequest error", err)

		code := http.StatusInternalServerError
		message := "internal server error"

		if err == payment.ErrInvalidCurrency || err == payment.ErrInvalidSourceType {
			code = http.StatusBadRequest
			message = err.Error()
		}

		return fiberhelper.HandleErrorJSONResp(
			c,
			code,
			message,
		)
	}

	return c.Status(200).JSON(result)
}

func (s server) GetPaymentStatusWithChargeID(c *fiber.Ctx) error {
	chargeID := c.Params("chargeID", "")
	if len(chargeID) == 0 {
		return fiberhelper.HandleErrorJSONResp(
			c,
			http.StatusBadRequest,
			"require charge id",
		)
	}

	resp, err := s.payment.GetPaymentStatusWithChargeID(c.Context(), chargeID)
	if err != nil {

		log.Println("GetPaymentStatusWithChargeID error", err)

		code := http.StatusInternalServerError
		message := "internal server error"

		if err == sql.ErrNoRows {
			code = http.StatusBadRequest
			message = "not found"
		}

		return fiberhelper.HandleErrorJSONResp(
			c,
			code,
			message,
		)
	}

	return c.Status(200).JSON(resp)
}

func (s server) omiseWebhook(c *fiber.Ctx) error {
	var b payment.PaymentEvent

	if err := c.BodyParser(&b); err != nil {
		log.Println("BodyParser error", err)
		return fiberhelper.HandleErrorJSONResp(
			c,
			http.StatusBadRequest,
			"require transaction id",
		)

	}

	if err := s.payment.HookPaymentEvent(c.Context(), b); err != nil {
		return fiberhelper.HandleErrorJSONResp(
			c,
			http.StatusInternalServerError,
			"internal server error",
		)
	}

	return c.Status(200).JSON(nil)
}
