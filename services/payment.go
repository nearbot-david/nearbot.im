package services

import (
	"bytes"
	"crypto/md5"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/mazanax/moneybot/models"
	"github.com/mazanax/moneybot/repository"
	"github.com/mazanax/moneybot/utils"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
)

type PaymentID string
type PaymentLink string

type PaymentConstraints struct {
	MinAmount uint64
	MaxAmount uint64
}

type PaymentMethod interface {
	GetConstraints() PaymentConstraints
	GeneratePaymentLink(telegramID int64, amount uint64) (PaymentID, PaymentLink)
	BuildPaymentLink(paymentID PaymentID) PaymentLink
	ProcessPayment(request *http.Request) (*models.Deposit, error)
}

type GatewayPaymentMethod struct {
	paymentEndpoint           string
	paymentSuccessfulEndpoint string
	clientID                  string
	secretKey                 string
	gatewayHost               string
	depositRepository         *repository.DepositRepository
}

func NewGatewayPaymentMethod(
	paymentEndpoint string,
	paymentSuccessfulEndpoint string,
	clientID string,
	secretKey string,
	gatewayHost string,
	depositRepository *repository.DepositRepository,
) *GatewayPaymentMethod {
	return &GatewayPaymentMethod{
		paymentEndpoint:           paymentEndpoint,
		paymentSuccessfulEndpoint: paymentSuccessfulEndpoint,
		clientID:                  clientID,
		secretKey:                 secretKey,
		gatewayHost:               gatewayHost,
		depositRepository:         depositRepository,
	}
}

func (g *GatewayPaymentMethod) GetConstraints() PaymentConstraints {
	return PaymentConstraints{MinAmount: 10000, MaxAmount: 1000000}
}

func (g *GatewayPaymentMethod) BuildPaymentLink(paymentID PaymentID) PaymentLink {
	return PaymentLink(g.paymentEndpoint + string(paymentID))
}

func (g *GatewayPaymentMethod) GeneratePaymentLink(telegramID int64, amount uint64) (PaymentID, PaymentLink) {
	for {
		paymentID := utils.RandStringBytes(16)
		if nil == g.depositRepository.FindBySlug(paymentID) {
			deposit := &models.Deposit{
				Slug:       paymentID,
				TelegramID: telegramID,
				Method:     "yoomoney1",
				Amount:     amount * 100,
				Status:     models.PaymentStatusNew,
				CreatedAt:  time.Now(),
			}

			if err := g.depositRepository.Persist(deposit); err != nil {
				log.Println(err.Error())
				return "", ""
			}

			return PaymentID(paymentID), g.BuildPaymentLink(PaymentID(paymentID))
		}
	}
}

func (g *GatewayPaymentMethod) GetPaymentForm(txID string, amount uint64) string {
	payload, _ := json.Marshal(map[string]interface{}{
		"client_id":   g.clientID,
		"method":      "yoomoney1",
		"amount":      amount,
		"tx_id":       txID,
		"success_url": g.paymentSuccessfulEndpoint + txID,
	})

	req, err := http.NewRequest("POST", g.gatewayHost+"/payment-form", bytes.NewReader(payload))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	response, _ := ioutil.ReadAll(resp.Body)

	type FormResponse struct {
		Data string `json:"data"`
	}
	formResponse := &FormResponse{}

	if err := json.Unmarshal(response, formResponse); err != nil {
		log.Println(err)
		return ""
	}

	return formResponse.Data
}

func (g *GatewayPaymentMethod) ProcessPayment(request *http.Request) (*models.Deposit, error) {
	txID := request.FormValue("uuid")
	amount, err := strconv.Atoi(request.FormValue("amount"))
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("cannot process entity")
	}
	signature := request.FormValue("signature")

	if txID == "" || amount == 0 || signature == "" {
		log.Printf("Bad request: %s, %d, %s", txID, amount, signature)
		return nil, fmt.Errorf("bad request")
	}

	computedSignatureBytes := md5.Sum([]byte(fmt.Sprintf("UUID:%s,Amount:%d,SecretKey:%s", txID, amount, g.secretKey)))
	computedSignature := hex.EncodeToString(computedSignatureBytes[:])
	if subtle.ConstantTimeCompare([]byte(computedSignature), []byte(signature)) != 1 {
		log.Printf("Signatures aren't equal: (computed: %s, request: %s)", hex.EncodeToString(computedSignatureBytes[:]), signature)
		return nil, fmt.Errorf("bad request")
	}

	deposit := g.depositRepository.FindBySlug(txID)
	if deposit == nil || deposit.Amount > uint64(amount*100) {
		log.Printf("Deposit with slug = %s not found", txID)
		return nil, fmt.Errorf("deposit not found")
	}

	deposit.Status = models.PaymentStatusSuccess
	if err := g.depositRepository.Persist(deposit); err != nil {
		log.Printf("Cannot persist deposit with slug = %s: %s", txID, err.Error())
		return nil, fmt.Errorf("cannot persist deposit status")
	}

	return deposit, nil
}

type CentAppMethod struct {
	token string
}

func NewCentAppMethod(token string) *CentAppMethod {
	return &CentAppMethod{
		token: token,
	}
}

func (c *CentAppMethod) GetConstraints() PaymentConstraints {
	return PaymentConstraints{MinAmount: 10000, MaxAmount: 1000000}
}

func (c *CentAppMethod) GeneratePaymentLink(telegramID int64, amount uint64) (PaymentID, PaymentLink) {
	return "", ""
}

func (c *CentAppMethod) ProcessPayment(request *http.Request) (*models.Deposit, error) {
	return nil, nil
}

func (c *CentAppMethod) BuildPaymentLink(paymentID PaymentID) PaymentLink {
	return PaymentLink(paymentID)
}
