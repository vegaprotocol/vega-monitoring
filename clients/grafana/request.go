package grafana

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
)

func (c *GrafanaClient) GetAny(requestPath string) ([]byte, error) {
	return c.RequestAny("GET", requestPath, url.Values{}, nil)
}

func (c *GrafanaClient) GetJSON(requestPath string, responseStruct interface{}) error {
	return c.RequestJSON("GET", requestPath, url.Values{}, nil, responseStruct)
}

func (c *GrafanaClient) GetPrettyJSON(requestPath string) ([]byte, error) {
	responseBody, err := c.GetAny(requestPath)
	if err != nil {
		return nil, err
	}

	var prettyJSON bytes.Buffer
	err = json.Indent(&prettyJSON, responseBody, "", "\t")
	if err != nil {
		return nil, fmt.Errorf("failed to parse response to %s, %w", requestPath, err)
	}
	return prettyJSON.Bytes(), nil
}

func (c *GrafanaClient) RequestJSON(method, requestPath string, query url.Values, requestBody []byte, responseStruct interface{}) error {
	responseBody, err := c.RequestAny(method, requestPath, query, requestBody)
	if err != nil {
		return err
	}

	err = json.Unmarshal(responseBody, responseStruct)
	if err != nil {
		return err
	}
	return nil
}

func (c *GrafanaClient) RequestAny(method, requestPath string, query url.Values, requestBody []byte) ([]byte, error) {
	// Prepare URL
	url, err := url.Parse(c.apiURL)
	if err != nil {
		return nil, fmt.Errorf("invalid apiURL %s, %w", c.apiURL, err)
	}
	url.Path = path.Join(url.Path, requestPath)
	url.RawQuery = query.Encode()

	// Prepare Request
	req, err := http.NewRequest(method, url.String(), bytes.NewReader(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request, %w", err)
	}
	if c.apiToken != "" {
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.apiToken))
	}
	req.Header.Add("Content-Type", "application/json")

	// Request
	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request failed %s, %w", url.String(), err)
	}

	// Parse response
	defer res.Body.Close()

	bodyContents, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body %s, %w", url.String(), err)
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http failed: %s (%d), %s", url.String(), res.StatusCode, string(bodyContents))
	}

	return bodyContents, nil
}
