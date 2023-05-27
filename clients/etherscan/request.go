package etherscan

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func (c *EtherscanClient) sendRequest(args map[string]string, resultPayload any) error {
	// Prepare URL
	url := fmt.Sprintf("%s?apikey=%s", c.apiURL, c.apiKey)
	for key, value := range args {
		url = fmt.Sprintf("%s&%s=%s", url, key, value)
	}

	// Rate Limiter
	if err := c.rateLimiter.Wait(context.Background()); err != nil {
		return fmt.Errorf("failed rate limit sending request to Etherscan. %w", err)
	}

	// Send Request
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed Etherscan request with args: %+v. %w", args, err)
	}
	defer resp.Body.Close()
	if err = json.NewDecoder(resp.Body).Decode(resultPayload); err != nil {
		return fmt.Errorf("failed to parse Etherscan response. %w", err)
	}

	return nil
}
