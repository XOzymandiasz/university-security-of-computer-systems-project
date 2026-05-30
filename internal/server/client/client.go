package client

import (
	"bytes"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"scs/internal/shared/identity"
	protocol2 "scs/internal/shared/protocol"
)

type RegisterRequest struct {
	ID        string `json:"id"`
	PublicKey string `json:"public_key"`
}

type Client struct {
	addr string
}

func New(addr string) *Client {
	return &Client{addr: addr}
}

func (c *Client) Init() (*rsa.PublicKey, error) {
	resp, err := http.Get(c.url("/api/init"))
	if err != nil {
		return nil, fmt.Errorf("ttp init request: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {

		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ttp init failed: status=%d body=%s", resp.StatusCode, string(body))
	}

	var response protocol2.InitResponse
	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decode init response: %w", err)
	}

	keyBase64 := response.TTPEncPublicKey

	var key *rsa.PublicKey
	key, err = identity.ParsePublicKeyFromBase64(keyBase64)
	if err != nil {
		return nil, fmt.Errorf("parse public key: %w", err)
	}

	return key, nil
}

func (c *Client) Register(req protocol2.RegisterRequest) (string, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("marshal register request: %w", err)
	}

	var httpReq *http.Request
	httpReq, err = http.NewRequest(
		http.MethodPost,
		c.url("/api/register"),
		bytes.NewReader(body),
	)
	if err != nil {
		return "", fmt.Errorf("create register request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	var resp *http.Response
	resp, err = http.DefaultClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("ttp register request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ttp register failed: status=%d body=%s", resp.StatusCode, string(respBody))
	}

	var response protocol2.RegisterResponse
	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("decode register response: %w", err)
	}

	if response.Certificate == "" {
		return "", fmt.Errorf("empty certificate in register response")
	}

	return response.Certificate, nil
}

func (c *Client) Authenticate(req protocol2.AuthenticateRequest) (protocol2.AuthenticateResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return protocol2.AuthenticateResponse{}, fmt.Errorf("marshal ttp authenticate request: %w", err)
	}

	var httpReq *http.Request
	httpReq, err = http.NewRequest(
		http.MethodPost,
		c.url("/api/authenticate"),
		bytes.NewReader(body),
	)
	if err != nil {
		return protocol2.AuthenticateResponse{}, fmt.Errorf("create ttp authenticate request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	var resp *http.Response
	resp, err = http.DefaultClient.Do(httpReq)
	if err != nil {
		return protocol2.AuthenticateResponse{}, fmt.Errorf("ttp authenticate request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return protocol2.AuthenticateResponse{}, fmt.Errorf(
			"ttp authenticate failed: status=%d body=%s",
			resp.StatusCode,
			string(respBody),
		)
	}

	var response protocol2.AuthenticateResponse
	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return protocol2.AuthenticateResponse{}, fmt.Errorf("decode ttp authenticate response: %w", err)
	}

	if !response.OK {
		return protocol2.AuthenticateResponse{}, fmt.Errorf("authentication rejected: %s", response.Message)
	}

	if response.EncryptedSessionKeyForClient == "" {
		return protocol2.AuthenticateResponse{}, fmt.Errorf("empty encrypted session key for client")
	}

	if response.EncryptedSessionKeyForServer == "" {
		return protocol2.AuthenticateResponse{}, fmt.Errorf("empty encrypted session key for server")
	}

	return response, nil
}

func (c *Client) url(path string) string {
	return "http://" + c.addr + path
}
