package services

import (
	"bytes"
	"crypto/md5"
	"crypto/sha256"
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

type PaywithnearMethod struct {
	clientID                  string
	clientSecret              string
	gatewayHost               string
	paymentEndpoint           string
	paymentSuccessfulEndpoint string
	depositRepository         *repository.DepositRepository
}

func NewPaywithnearMethod(paymentEndpoint string, clientID string, clientSecret string, gatewayHost string, paymentSuccessfulEndpoint string, depositRepository *repository.DepositRepository) *PaywithnearMethod {
	return &PaywithnearMethod{
		paymentEndpoint:           paymentEndpoint,
		clientID:                  clientID,
		clientSecret:              clientSecret,
		gatewayHost:               gatewayHost,
		paymentSuccessfulEndpoint: paymentSuccessfulEndpoint,
		depositRepository:         depositRepository,
	}
}

func (p *PaywithnearMethod) GetConstraints() PaymentConstraints {
	return PaymentConstraints{MinAmount: 10000, MaxAmount: 1000000}
}

func (p *PaywithnearMethod) BuildPaymentLink(paymentID PaymentID) PaymentLink {
	deposit := p.depositRepository.FindBySlug(string(paymentID))
	if deposit == nil {
		log.Printf("Deposit with slug = %s not found", paymentID)
		return ""
	}

	return PaymentLink("https://go.paywithnear.com/payment/" + deposit.ExternalID)
}

func (p *PaywithnearMethod) GeneratePaymentLink(telegramID int64, amount uint64) (PaymentID, PaymentLink) {
	for {
		paymentID := utils.RandStringBytes(16)
		if nil != p.depositRepository.FindBySlug(paymentID) {
			continue
		}

		externalID, paymentLink := p.retrievePaymentPage(paymentID, amount)

		deposit := &models.Deposit{
			Slug:       paymentID,
			TelegramID: telegramID,
			Method:     "paywithnear",
			Amount:     uint64(float64(amount) / 100 * 1e5),
			Status:     models.PaymentStatusNew,
			CreatedAt:  time.Now(),
			ExternalID: externalID,
		}

		if err := p.depositRepository.Persist(deposit); err != nil {
			log.Println(err.Error())
			return "", ""
		}

		return PaymentID(paymentID), PaymentLink(paymentLink)
	}
}

func (p *PaywithnearMethod) retrievePaymentPage(slug string, amount uint64) (string, string) {
	payload, _ := json.Marshal(map[string]interface{}{
		"name":       "Пополнение баланса @textmoneybot",
		"amount":     float64(amount) / 100,
		"return_url": p.paymentSuccessfulEndpoint + slug,
	})

	req, err := http.NewRequest("POST", p.gatewayHost+"/payment/"+p.clientID+"/page", bytes.NewReader(payload))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	response, _ := ioutil.ReadAll(resp.Body)

	type FormResponse struct {
		PaymentID string `json:"paymentID"`
		URL       string `json:"url"`
	}
	formResponse := &FormResponse{}
	if err := json.Unmarshal(response, formResponse); err != nil {
		log.Println(err)
		return "", ""
	}

	return formResponse.PaymentID, formResponse.URL
}

func (p *PaywithnearMethod) ProcessPayment(request *http.Request) (*models.Deposit, error) {
	data, err := ioutil.ReadAll(request.Body)
	if err != nil {
		return nil, err
	}

	type CallbackPayload struct {
		PaymentID         string `json:"payment_id"`
		Amount            string `json:"amount"`
		AmountUsd         string `json:"amount_usd"`
		ReceivedAmount    string `json:"received_amount"`
		ReceivedAmountUsd string `json:"received_amount_usd"`
		CurrentDatetime   string `json:"current_datetime"`
		Signature         string `json:"signature"`
	}

	payload := &CallbackPayload{}
	if err := json.Unmarshal(data, request); err != nil {
		return nil, err
	}

	if payload.PaymentID == "" || payload.ReceivedAmount == "" || payload.Signature == "" {
		log.Printf("Bad request: %s, %d, %s", payload.PaymentID, payload.ReceivedAmount, payload.Signature)
		return nil, fmt.Errorf("bad request")
	}

	computedSignatureBytes := sha256.Sum256([]byte(fmt.Sprintf(
		"Amount=%s;AmountUsd=%s;CurrentDateTime=%s;PaymentID=%s;ReceivedAmount=%s;ReceivedAmountUsd=%s;SecretKey=%s",
		payload.Amount,
		payload.AmountUsd,
		payload.CurrentDatetime,
		payload.PaymentID,
		payload.ReceivedAmount,
		payload.ReceivedAmountUsd,
		p.clientSecret,
	)))
	computedSignature := hex.EncodeToString(computedSignatureBytes[:])
	if subtle.ConstantTimeCompare([]byte(computedSignature), []byte(payload.Signature)) != 1 {
		log.Printf("Signatures aren't equal: (computed: %s, request: %s)", hex.EncodeToString(computedSignatureBytes[:]), payload.Signature)
		return nil, fmt.Errorf("bad request")
	}

	deposit := p.depositRepository.FindByExternalID("paywithnear", payload.PaymentID)
	if deposit == nil {
		log.Printf("Deposit with external_id = %s (%s) not found", payload.PaymentID, "paywithnear")
		return nil, fmt.Errorf("deposit not found")
	}

	if deposit.Status == models.PaymentStatusNew {
		amount, err := strconv.ParseFloat(payload.ReceivedAmount, 64)
		if err != nil {
			return nil, err
		}
		deposit.Amount = uint64(amount * 1e5)
		deposit.Status = models.PaymentStatusSuccess

		if err := p.depositRepository.Persist(deposit); err != nil {
			log.Printf("Cannot persist deposit with slug = %s: %s", payload.PaymentID, err.Error())
			return nil, fmt.Errorf("cannot persist deposit status")
		}

		return deposit, nil
	}

	return nil, nil
}
