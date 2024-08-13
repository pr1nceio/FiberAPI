package providers

import (
	"encoding/json"
	"errors"
	gorm "github.com/cradio/gormx"
	"github.com/fruitspace/FiberAPI/models/structs"
	"github.com/fruitspace/FiberAPI/utils"
	"github.com/fruitspace/schemas/db/go/db"
	"net/http"
	"net/url"
	"strconv"
)

type PaymentProvider struct {
	db           *gorm.DB
	paymentsHost string
}

func NewPaymentProvider(db *gorm.DB, paymentsHost string) *PaymentProvider {
	return &PaymentProvider{db: db, paymentsHost: paymentsHost}
}

func (p *PaymentProvider) CountUnfinishedPaymentsForUID(uid int) int {
	// Count unfinished payments for user
	var count int64
	p.db.Model(db.Transaction{}).Where(db.Transaction{IsActive: true, UID: uid}).Count(&count)
	return int(count)
}

func (p *PaymentProvider) GetPaymentsForUID(uid int, unfinished bool) []*db.Transaction {
	var transactions []*db.Transaction

	p.db.Model(db.Transaction{}).Where(db.Transaction{UID: uid, IsActive: unfinished}).Find(&transactions)

	for _, t := range transactions {
		t.Method = p.GetPaymentMethodName(t.Method)
	}
	return transactions
}

func (p *PaymentProvider) GetPaymentMethodName(method string) string {
	// Get payment method name
	switch method {
	case "bank_card":
		return "Банковская карта (ЮКасса)"
	case "yoo_money":
		return "ЮMoney (ЮКасса)"
	case "qiwi":
		return "Qiwi (ЮКасса)"
	case "sberbank":
		return "Сбербанк (ЮКасса)"
	case "alfabank":
		return "Альфа-Банк (ЮКасса)"
	case "tinkoff_bank":
		return "Тинькофф (ЮКасса)"
	case "sbp":
		return "СБП"
	case "mobile_balance":
		return "Мобильный баланс (ЮКасса)"
	case "cash":
		return "Наличные (ЮКасса)"

	case "qw":
		return "Qiwi (PayOk)"
	case "ya":
		return "ЮMoney (PayOk)"
	case "wm":
		return "WebMoney (PayOk)"
	case "pr":
		return "Payeer (PayOk)"
	case "cd":
		return "Банковская карта (PayOk)"
	case "pm":
		return "PerfectMoney (PayOk)"
	case "ad":
		return "AdvCash (PayOk)"
	case "mg":
		return "Мегафон (PayOk)"
	case "bt":
		return "Bitcoin (PayOk)"
	case "th":
		return "[USDT] Tether (PayOk)"
	case "lt":
		return "[LTC] Litecoin (PayOk)"
	case "dg":
		return "DogeCoin (PayOk)"

	case "qq":
		return "Qiwi/Карта"
	default:
		return "Неизвестный метод"
	}
}

// CreateInvoice requests Payments Service for merchant=[qiwi,enot,yookassa]
func (p *PaymentProvider) CreateInvoice(uid int, amount float64, email string, merchant string) *db.Transaction {
	// Create invoice
	resp, err := http.PostForm(p.paymentsHost+"/create/"+merchant, url.Values{
		"uid":    []string{strconv.Itoa(uid)},
		"amount": []string{strconv.FormatFloat(amount, 'f', 2, 64)},
		"email":  []string{email},
	})
	if utils.Should(err) != nil {
		return nil
	}
	defer resp.Body.Close()
	var res structs.PaymentResponse
	utils.Should(json.NewDecoder(resp.Body).Decode(&res))
	if res.Status != "ok" {
		utils.Should(errors.New(res.Message))
		utils.SendMessageDiscord(res.Message)
	}
	return &db.Transaction{Amount: amount, GoPayURL: res.Url, IsActive: true, Method: "none"}
}

func (p *PaymentProvider) SpendMoney(uid int, amount float64) *structs.PaymentResponse {
	resp, err := http.PostForm(p.paymentsHost+"/internal/buy", url.Values{
		"uid":    []string{strconv.Itoa(uid)},
		"amount": []string{strconv.FormatFloat(amount, 'f', 2, 64)},
	})
	if utils.Should(err) != nil {
		return &structs.PaymentResponse{Status: "error", Message: "Internal error"}
	}
	defer resp.Body.Close()
	var res structs.PaymentResponse
	utils.Should(json.NewDecoder(resp.Body).Decode(&res))
	return &res
}
