package serverclient

import (
	"bytes"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"scs/internal/client/httpapi"
	"scs/internal/identity"
	"scs/internal/protocol"
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
	resp, err := http.Get("http://" + c.addr + "/api/init")
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

	var response protocol.Message
	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decode init response: %w", err)
	}

	if response.Type != "TTP_PUBLIC_KEY" {
		return nil, fmt.Errorf("unexpected response type: %s", response.Type)
	}

	keyBase64, ok := response.Body.(string)
	if !ok {
		return nil, fmt.Errorf("invalid response body type: %T", response.Body)
	}

	key, err := identity.ParsePublicKeyFromBase64(keyBase64)
	if err != nil {
		return nil, fmt.Errorf("parse public key: %w", err)
	}

	return key, nil
}

func (c *Client) Register(encryptedID string, authPublicKeyBase64 string) (string, error) {
	req := protocol.RegistrationData{
		ID:            encryptedID,
		AuthPublicKey: authPublicKeyBase64,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("marshal register request: %w", err)
	}

	httpReq, err := http.NewRequest(
		http.MethodPost,
		"http://"+c.addr+"/api/register",
		bytes.NewReader(body),
	)
	if err != nil {
		return "", fmt.Errorf("create register request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("ttp register request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ = io.ReadAll(resp.Body)
		return "", fmt.Errorf("ttp register failed: status=%d body=%s", resp.StatusCode, string(body))
	}

	var response protocol.Message
	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("decode register response: %w", err)
	}

	if response.Type != "CERTIFICATE" {
		return "", fmt.Errorf("unexpected response type: %s", response.Type)
	}

	certificateBase64, ok := response.Body.(string)
	if !ok {
		return "", fmt.Errorf("invalid certificate body type: %T", response.Body)
	}

	return certificateBase64, nil
}

func (c *Client) ReadMessage(msg string) (string, error) {
	fmt.Println("#1")
	req := httpapi.MessageRequest{
		Body: msg,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("marshal message request: %w", err)
	}
	fmt.Println("#2")
	httpReq, err := http.NewRequest(
		http.MethodPost,
		"http://"+c.addr+"/api/message",
		bytes.NewReader(body),
	)
	if err != nil {
		return "", fmt.Errorf("create message request: %w", err)
	}
	fmt.Println("#3")
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("ttp read message request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	fmt.Println("#4")
	fmt.Println(resp.StatusCode)
	if resp.StatusCode != http.StatusOK {
		body, _ = io.ReadAll(resp.Body)
		return "", fmt.Errorf("ttp read message failed: status=%d body=%s", resp.StatusCode, string(body))
	}
	fmt.Println("#5")
	var response httpapi.MessageResponse
	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("decode message response: %w", err)
	}
	fmt.Println("#6")
	return response.Body, nil
}
