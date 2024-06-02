package api

import (
	fiberapi "github.com/fruitspace/FiberAPI"
	"github.com/fruitspace/FiberAPI/api/ent"
	"github.com/fruitspace/FiberAPI/models/structs"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/exp/slices"
)

type PaymentsAPI struct {
	*ent.API
}

func (api *PaymentsAPI) Register(router fiber.Router) error {
	router.Get("/", api.GetPayments) // get payments
	router.Post("/", api.Create)     //create payment
	return nil
}

// PaymentsCreate creates invoice for specified merchant
// @Tags Payments
// @Summary Creates invoice for specified merchant
// @Accept json
// @Produce json
// @Param Authorization header string true "User token"
// @Param data body structs.APIPaymentRequest true "All fields are required"
// @Success 200 {object} structs.APIPaymentResponse
// @Failure 500 {object} structs.APIError
// @Failure 403 {object} structs.APIError
// @Router /payments [post]
func (api *PaymentsAPI) Create(c *fiber.Ctx) error {
	acc := api.AccountProvider.New()
	if !api.PerformAuth_(c, acc) {
		return c.Status(403).JSON(structs.NewAPIError("Unauthorized"))
	}
	var data structs.APIPaymentRequest
	if c.BodyParser(&data) != nil {
		return c.Status(500).JSON(structs.NewAPIError("Invalid request"))
	}
	if !slices.Contains(fiberapi.ValidMerchants, data.Merchant) {
		return c.Status(500).JSON(structs.NewAPIError("Invalid merchant"))
	}
	if data.Amount < 20 || data.Amount > 100000 {
		return c.Status(500).JSON(structs.NewAPIError("Invalid amount (>100K or <20)"))
	}
	tx := api.PaymentProvider.CreateInvoice(acc.Data().UID, data.Amount, acc.Data().Email, data.Merchant)
	if !tx.IsActive {
		return c.Status(500).JSON(structs.NewAPIError("Unable to create transaction", "tr_create"))
	}
	return c.JSON(structs.APIPaymentResponse{
		APIBasicSuccess: structs.NewAPIBasicResponse("Success"),
		Url:             tx.GoPayURL,
	})
}

// PaymentsGet returns list of user payments
// @Tags Payments
// @Summary Returns list of user payments
// @Accept json
// @Produce json
// @Param Authorization header string true "User token"
// @Success 200 {object} structs.APIPaymentListResponse
// @Failure 500 {object} structs.APIError
// @Failure 403 {object} structs.APIError
// @Router /payments [get]
func (api *PaymentsAPI) GetPayments(c *fiber.Ctx) error {
	acc := api.AccountProvider.New()
	if !api.PerformAuth_(c, acc) {
		return c.Status(403).JSON(structs.NewAPIError("Unauthorized"))
	}

	transactions := api.PaymentProvider.GetPaymentsForUID(acc.Data().UID, false)
	return c.JSON(structs.APIPaymentListResponse{
		APIBasicSuccess: structs.NewAPIBasicResponse("Success"),
		Transactions:    transactions,
	})
}
