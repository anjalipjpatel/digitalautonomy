package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func PostJSON(
	ctx context.Context,
	client *http.Client,
	endpoint string,
	payload any,
	result any,
) (int, []byte, error) {
	requestBody, err := json.Marshal(payload)
	if err != nil {
		return 0, nil, fmt.Errorf("marshal request JSON: %w", err)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(requestBody))
	if err != nil {
		return 0, nil, fmt.Errorf("create POST request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")

	response, err := client.Do(request)
	if err != nil {
		return 0, nil, fmt.Errorf("send POST request: %w", err)
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if err != nil {
		return response.StatusCode, nil, fmt.Errorf("read response body: %w", err)
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return response.StatusCode, responseBody, fmt.Errorf(
			"upstream returned HTTP %d: %s",
			response.StatusCode,
			strings.TrimSpace(string(responseBody)),
		)
	}
	if err := json.Unmarshal(responseBody, result); err != nil {
		return response.StatusCode, responseBody, fmt.Errorf("decode response JSON: %w", err)
	}

	return response.StatusCode, responseBody, nil
}
