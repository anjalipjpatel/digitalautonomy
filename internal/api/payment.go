package api

import (
	"bytes"
	"digitalautonomy/pkg/utils"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

type paymentRequest struct {
	FromUser string         `json:"from_user"`
	ToUser   string         `json:"to_user"`
	Amount   string         `json:"amount"`
	Currency utils.Currency `json:"currency"`
}

func NewPaymentHandler(paymentAPIURL string) http.HandlerFunc {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "unsupported method", http.StatusMethodNotAllowed)
			return
		}

		var payment paymentRequest
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
		if err := decoder.Decode(&payment); err != nil {
			http.Error(w, "invalid JSON request body", http.StatusBadRequest)
			return
		}

		if payment.FromUser == "" || payment.ToUser == "" || payment.Amount == "" || payment.Currency == "" {
			http.Error(w, "from_user, to_user, amount, currency are required fields", http.StatusBadRequest)
			return
		}

		reference := fmt.Sprintf("demo-%d", time.Now().UnixNano())
		initiation := PaymentInitiation{
			InstructionIdentification: reference,
			EndToEndIdentification:    reference,
			InstructedAmount: Amount{
				Amount:   payment.Amount,
				Currency: payment.Currency,
			},
			CreditorAccount: Account{
				SchemeName:     "UK.OBIE.SortCodeAccountNumber",
				Identification: payment.ToUser,
				Name:           payment.ToUser,
			},
		}
		consentRequest := CreateConsentRequest{
			Data: ConsentRequestData{
				Initiation:        initiation,
				ReadRefundAccount: "No",
			},
			Risk: struct{}{},
		}

		baseURL := strings.TrimRight(paymentAPIURL, "/")

		var consentResponse CreateConsentResponse
		consentStatus, consentBody, err := utils.PostJSON(
			r.Context(),
			client,
			baseURL+"/domestic-payment-consents",
			consentRequest,
			&consentResponse,
		)
		if err != nil {
			log.Printf("Mockoon consent response: HTTP %d %s", consentStatus, consentBody)
			http.Error(w, "consent request to Mockoon failed", http.StatusBadGateway)
			return
		}
		log.Printf("Mockoon consent response: HTTP %d %s", consentStatus, consentBody)

		consentID := consentResponse.Data.ConsentID
		if consentID == "" {
			http.Error(w, "Mockoon consent response has an empty ConsentID", http.StatusBadGateway)
			return
		}
		if consentResponse.Data.Status != "AUTH" {
			http.Error(
				w,
				fmt.Sprintf("payment consent is not authorised (status %q)", consentResponse.Data.Status),
				http.StatusBadGateway,
			)
			return
		}

		paymentRequest := CreatePaymentRequest{
			Data: PaymentRequestData{
				ConsentID:  consentID,
				Initiation: initiation,
			},
			Risk: struct{}{},
		}

		var paymentResponse PaymentResponse
		paymentStatus, paymentBody, err := utils.PostJSON(
			r.Context(),
			client,
			baseURL+"/domestic-payments",
			paymentRequest,
			&paymentResponse,
		)
		if err != nil {
			log.Printf("Mockoon payment response: HTTP %d %s", paymentStatus, paymentBody)
			http.Error(w, "payment request to Mockoon failed", http.StatusBadGateway)
			return
		}
		if paymentResponse.Data.Status == "" || paymentResponse.Data.Status == "RJCT" {
			http.Error(
				w,
				fmt.Sprintf("Mockoon did not accept the payment (status %q)", paymentResponse.Data.Status),
				http.StatusBadGateway,
			)
			return
		}

		var prettyResponse bytes.Buffer
		if err := json.Indent(&prettyResponse, paymentBody, "", "	"); err != nil {
			log.Printf("Mockoon payment response: HTTP %d %s", paymentStatus, paymentBody)
		} else {
			log.Printf("Mockoon payment response: HTTP %d\n%s", paymentStatus, &prettyResponse)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(struct {
			Consent CreateConsentResponse `json:"consent"`
			Payment PaymentResponse       `json:"payment"`
		}{
			Consent: consentResponse,
			Payment: paymentResponse,
		})
	}
}
