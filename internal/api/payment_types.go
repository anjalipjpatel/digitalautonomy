package api

import "digitalautonomy/pkg/utils"

type Amount struct {
	Amount   string         `json:"Amount"`
	Currency utils.Currency `json:"Currency"`
}

type Account struct {
	SchemeName     string `json:"SchemeName"`
	Identification string `json:"Identification"`
	Name           string `json:"Name"`
}

type PaymentInitiation struct {
	InstructionIdentification string  `json:"InstructionIdentification"`
	EndToEndIdentification    string  `json:"EndToEndIdentification"`
	InstructedAmount          Amount  `json:"InstructedAmount"`
	CreditorAccount           Account `json:"CreditorAccount"`
}

// 1. POST /domestic-payment-consents
type CreateConsentRequest struct {
	Data ConsentRequestData `json:"Data"`
	Risk struct{}           `json:"Risk"`
}

type ConsentRequestData struct {
	Initiation        PaymentInitiation `json:"Initiation"`
	ReadRefundAccount string            `json:"ReadRefundAccount,omitempty"`
}

// 2. Response from creating consent.
type CreateConsentResponse struct {
	Data ConsentResponseData `json:"Data"`
}

type ConsentResponseData struct {
	ConsentID            string            `json:"ConsentId"`
	CreationDateTime     string            `json:"CreationDateTime"`
	Status               string            `json:"Status"`
	StatusUpdateDateTime string            `json:"StatusUpdateDateTime"`
	Initiation           PaymentInitiation `json:"Initiation"`
}

// 3. POST /domestic-payments
type CreatePaymentRequest struct {
	Data PaymentRequestData `json:"Data"`
	Risk struct{}           `json:"Risk"`
}

type PaymentRequestData struct {
	ConsentID  string            `json:"ConsentId"`
	Initiation PaymentInitiation `json:"Initiation"`
}

// Response from POST /domestic-payments and
// GET /domestic-payments/{DomesticPaymentId}.
type PaymentResponse struct {
	Data PaymentResponseData `json:"Data"`
}

type PaymentResponseData struct {
	DomesticPaymentID    string            `json:"DomesticPaymentId"`
	ConsentID            string            `json:"ConsentId"`
	CreationDateTime     string            `json:"CreationDateTime"`
	Status               string            `json:"Status"`
	StatusUpdateDateTime string            `json:"StatusUpdateDateTime"`
	Initiation           PaymentInitiation `json:"Initiation"`
}
