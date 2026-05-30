package client

import (
	"bytes"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"scs/internal/identity"
	"scs/internal/protocol"
)

type Client struct {
	addr    string
	baseDir string
}

func New(addr string, baseDir string) *Client {
	return &Client{
		addr:    addr,
		baseDir: baseDir,
	}
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

	var response protocol.InitResponse
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

func (c *Client) Register(req protocol.RegisterRequest) (string, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("marshal register request: %w", err)
	}

	var httpReq *http.Request
	httpReq, err = http.NewRequest(
		http.MethodPost,
		"http://"+c.addr+"/api/register",
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
		body, _ = io.ReadAll(resp.Body)
		return "", fmt.Errorf("ttp register failed: status=%d body=%s", resp.StatusCode, string(body))
	}

	var response protocol.RegisterResponse
	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("decode register response: %w", err)
	}

	if response.Certificate == "" {
		return "", fmt.Errorf("empty certificate in register response")
	}

	return response.Certificate, nil
}

func (c *Client) Authenticate(req protocol.ClientAuthenticateRequest) (protocol.ClientAuthenticateResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return protocol.ClientAuthenticateResponse{}, fmt.Errorf("marshal client authenticate request: %w", err)
	}

	var httpReq *http.Request
	httpReq, err = http.NewRequest(
		http.MethodPost,
		"http://"+c.addr+"/api/authenticate",
		bytes.NewReader(body),
	)
	if err != nil {
		return protocol.ClientAuthenticateResponse{}, fmt.Errorf("create client authenticate request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	var resp *http.Response
	resp, err = http.DefaultClient.Do(httpReq)
	if err != nil {
		return protocol.ClientAuthenticateResponse{}, fmt.Errorf("client authenticate request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return protocol.ClientAuthenticateResponse{}, fmt.Errorf(
			"client authenticate failed: status=%d body=%s",
			resp.StatusCode,
			string(respBody),
		)
	}

	var response protocol.ClientAuthenticateResponse
	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return protocol.ClientAuthenticateResponse{}, fmt.Errorf("decode client authenticate response: %w", err)
	}

	if !response.OK {
		return protocol.ClientAuthenticateResponse{}, fmt.Errorf("authentication rejected: %s", response.Message)
	}

	return response, nil
}

func (c *Client) ReadMessage(msg string) (string, error) {
	sessionKey, err := identity.LoadSessionKey(c.baseDir)
	if err != nil {
		return "", fmt.Errorf("client is not authenticated - missing session key: %w", err)
	}

	var encryptedBody string
	encryptedBody, err = identity.EncryptWithSessionKeyBase64([]byte(msg), sessionKey)
	if err != nil {
		return "", fmt.Errorf("encrypt message: %w", err)
	}

	req := protocol.MessageRequest{
		EncryptedBody: encryptedBody,
	}

	var body []byte
	body, err = json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("marshal message request: %w", err)
	}

	var httpReq *http.Request
	httpReq, err = http.NewRequest(
		http.MethodPost,
		"http://"+c.addr+"/api/message",
		bytes.NewReader(body),
	)
	if err != nil {
		return "", fmt.Errorf("create message request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	var resp *http.Response
	resp, err = http.DefaultClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("server read message request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ = io.ReadAll(resp.Body)
		return "", fmt.Errorf("server read message failed: status=%d body=%s", resp.StatusCode, string(body))
	}

	var response protocol.MessageResponse
	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("decode message response: %w", err)
	}

	var plaintext []byte
	plaintext, err = identity.DecryptWithSessionKeyBase64(response.EncryptedBody, sessionKey)
	if err != nil {
		return "", fmt.Errorf("decrypt message response: %w", err)
	}

	_ = identity.DeleteSessionKey(c.baseDir)

	return string(plaintext), nil
}
