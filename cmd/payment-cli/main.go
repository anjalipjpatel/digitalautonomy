package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type user struct {
	ID        string
	Initials  string
	Name      string
	AccountID string
}

type paymentRequest struct {
	FromUser string `json:"from_user"`
	ToUser   string `json:"to_user"`
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
}

var users = []user{
	{ID: "alice", Initials: "A", Name: "Alice", AccountID: "SIM-ALICE-001"},
	{ID: "bob", Initials: "B", Name: "Bob", AccountID: "SIM-BOB-001"},
}

func main() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Payment demo (simulated; no money moves)")
	fmt.Println("Who is making the payment?")
	fromIndex, err := selectUser(reader, -1)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println("\nWho should receive the payment?")
	toIndex, err := selectUser(reader, fromIndex)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Print("\nAmount in GBP: ")
	amount, err := reader.ReadString('\n')
	if err != nil && len(amount) == 0 {
		fmt.Fprintln(os.Stderr, "could not read amount:", err)
		os.Exit(1)
	}
	amount = strings.TrimSpace(amount)
	if err := validateAmount(amount); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	request := paymentRequest{
		FromUser: users[fromIndex].ID,
		ToUser:   users[toIndex].ID,
		Amount:   amount,
		Currency: "GBP",
	}
	if err := submitPayment(request); err != nil {
		fmt.Fprintln(os.Stderr, "payment simulation failed:", err)
		os.Exit(1)
	}
}

func selectUser(reader *bufio.Reader, excludedIndex int) (int, error) {
	options := make([]int, 0, len(users))
	for i, candidate := range users {
		if i == excludedIndex {
			continue
		}
		options = append(options, i)
		fmt.Printf("  %d. [%s] %s\n", len(options), candidate.Initials, candidate.Name)
	}
	fmt.Print("Choose a number: ")

	input, err := reader.ReadString('\n')
	if err != nil && len(input) == 0 {
		return 0, fmt.Errorf("could not read user selection: %w", err)
	}
	choice, err := strconv.Atoi(strings.TrimSpace(input))
	if err != nil || choice < 1 || choice > len(options) {
		return 0, fmt.Errorf("choose a listed number from 1 to %d", len(options))
	}
	return options[choice-1], nil
}

func validateAmount(amount string) error {
	parts := strings.Split(amount, ".")
	if len(parts) > 2 || parts[0] == "" || !containsOnlyDigits(parts[0]) {
		return fmt.Errorf("enter a positive amount, for example 12.50")
	}
	if len(parts) == 2 && (len(parts[1]) == 0 || len(parts[1]) > 2 || !containsOnlyDigits(parts[1])) {
		return fmt.Errorf("amount must have no more than two decimal places")
	}
	value, err := strconv.ParseFloat(amount, 64)
	if err != nil || value <= 0 {
		return fmt.Errorf("enter a positive amount, for example 12.50")
	}
	return nil
}

func containsOnlyDigits(value string) bool {
	for _, character := range value {
		if character < '0' || character > '9' {
			return false
		}
	}
	return true
}

func submitPayment(payment paymentRequest) error {
	endpoint := os.Getenv("PAYMENTS_URL")
	if endpoint == "" {
		endpoint = "http://localhost:8080/payments"
	}

	requestBody, err := json.Marshal(payment)
	if err != nil {
		return fmt.Errorf("encode payment request: %w", err)
	}
	httpRequest, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(requestBody))
	if err != nil {
		return fmt.Errorf("create API request: %w", err)
	}
	httpRequest.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	response, err := client.Do(httpRequest)
	if err != nil {
		return fmt.Errorf("call Go API at %s: %w", endpoint, err)
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if err != nil {
		return fmt.Errorf("read Go API response: %w", err)
	}

	fmt.Printf("\nGo API response: HTTP %s\n", response.Status)
	var prettyBody bytes.Buffer
	if err := json.Indent(&prettyBody, responseBody, "", "  "); err != nil {
		fmt.Println(string(responseBody))
	} else {
		fmt.Println(prettyBody.String())
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("Go API returned HTTP %s", response.Status)
	}
	return nil
}
