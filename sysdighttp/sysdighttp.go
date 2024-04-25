package sysdighttp

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type SysdigRequestConfig struct {
	Method      string
	Path        string
	Headers     map[string]string
	Params      map[string]interface{}
	JSON        interface{}
	Data        map[string]string
	Auth        [2]string
	Verify      bool
	Stream      bool
	MaxRetries  int
	BaseDelay   int
	MaxDelay    int
	Timeout     int
	ApiEndpoint string
	SecureToken string
}

func DefaultSysdigRequestConfig(apiEndpoint string, secureToken string) SysdigRequestConfig {
	return SysdigRequestConfig{
		Method:      "GET",
		Verify:      false,
		MaxRetries:  1,
		BaseDelay:   5,
		MaxDelay:    60,
		Timeout:     600,
		ApiEndpoint: apiEndpoint,
		SecureToken: secureToken,
	}
}

//goland:noinspection GoBoolExpressions
func SysdigRequest(logger *logrus.Logger, SysdigRequest SysdigRequestConfig) (*http.Response, error) {
	retries := 0
	var resp *http.Response
	var err error

	for retries <= SysdigRequest.MaxRetries {
		resp, err = makeRequest(&SysdigRequest)
		if err != nil {
			logger.Errorf("Error on HTTP request: %v", err)
			// Implement retry logic as necessary
			time.Sleep(time.Duration(SysdigRequest.BaseDelay) * time.Second)
			retries++
			continue
		}

		if resp.StatusCode >= 400 { // Check for HTTP error codes
			respBody, _ := io.ReadAll(resp.Body)
			_ = resp.Body.Close()
			logger.Infof("Received HTTP status code: %d", resp.StatusCode)
			logger.Infof("Response body: %s", string(respBody))
			return resp, fmt.Errorf("HTTP request failed with status code: %d", resp.StatusCode)
		}

		logger.Debugf("Received HTTP status code: %d", resp.StatusCode)
		return resp, nil
	}

	// Manually create an HTTP response with a 503 status code if all retries fail
	logger.Errorf("Failed to fetch data from %s after %d retries.", SysdigRequest.ApiEndpoint, SysdigRequest.MaxRetries)
	return &http.Response{
		Status:     "503 Service Unavailable",
		StatusCode: http.StatusServiceUnavailable,
		Body:       io.NopCloser(bytes.NewBufferString("Service is unavailable after retries.")),
	}, fmt.Errorf("service unavailable after %d retries", SysdigRequest.MaxRetries)
}

// makeRequest is a helper function to execute the HTTP request
func makeRequest(config *SysdigRequestConfig) (*http.Response, error) {
	u, err := url.Parse(fmt.Sprintf("%s%s", config.ApiEndpoint, config.Path))
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %v", err)
	}

	// Prepare query parameters
	params := u.Query() // get current URL encoded parameters
	for k, v := range config.Params {
		switch value := v.(type) {
		case int:
			params.Add(k, strconv.Itoa(value))
		case string:
			params.Add(k, value)
		default:
			// Handle unexpected types if necessary, or ignore them
		}
	}
	u.RawQuery = params.Encode() // re-encode the parameters into the URL

	var requestBody io.Reader
	if config.JSON != nil {
		byteData, err := json.Marshal(config.JSON)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal JSON data: %v", err)
		}
		requestBody = bytes.NewBuffer(byteData)
	}

	req, err := http.NewRequest(config.Method, u.String(), requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	for key, value := range config.Headers {
		req.Header.Set(key, value)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.SecureToken))

	// Set up client and transport
	client := &http.Client{
		Timeout: time.Duration(config.Timeout) * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: !config.Verify},
		},
	}
	return client.Do(req)
}

func ResponseBodyToJson(resp *http.Response, target interface{}) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	err = json.Unmarshal(body, target)
	if err != nil {
		return err
	}

	return nil
}
