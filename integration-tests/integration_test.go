package integration_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestHealthEndpoint(t *testing.T) {
	resp, err := http.Get("http://banners:8080/health")
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Unexpected status code: got %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestGetUserBanner(t *testing.T) {
	username := fmt.Sprintf("testuser_%d", time.Now().UnixNano())
	password := "testpassword"
	registerURL := "http://banners:8080/v1/auth/register"
	registerBody := fmt.Sprintf(`{"username":"%v","password":"%v","role":"user"}`, username, password)

	registerReq, err := http.NewRequest(http.MethodPost, registerURL, bytes.NewBuffer([]byte(registerBody)))
	if err != nil {
		t.Fatalf("Failed to create register request: %v", err)
	}

	registerReq.Header.Set("Content-Type", "application/json")

	registerResp, err := http.DefaultClient.Do(registerReq)
	if err != nil {
		t.Fatalf("Failed to send register request: %v", err)
	}
	defer registerResp.Body.Close()

	if registerResp.StatusCode != http.StatusCreated {
		t.Errorf("Unexpected status code for register: got %d, want %d", registerResp.StatusCode, http.StatusCreated)
	}

	loginURL := "http://banners:8080/v1/auth/login"
	loginBody := fmt.Sprintf(`{"username":"%v","password":"%v"}`, username, password)

	loginReq, err := http.NewRequest(http.MethodPost, loginURL, bytes.NewBuffer([]byte(loginBody)))
	if err != nil {
		t.Fatalf("Failed to create login request: %v", err)
	}

	loginReq.Header.Set("Content-Type", "application/json")

	loginResp, err := http.DefaultClient.Do(loginReq)
	if err != nil {
		t.Fatalf("Failed to send login request: %v", err)
	}
	defer loginResp.Body.Close()

	if loginResp.StatusCode != http.StatusOK {
		t.Errorf("Unexpected status code for login: got %d, want %d", loginResp.StatusCode, http.StatusOK)
	}

	var loginResponse struct {
		Token string `json:"token"`
	}

	if err := json.NewDecoder(loginResp.Body).Decode(&loginResponse); err != nil {
		t.Fatalf("Failed to decode login response: %v", err)
	}

	userBannerURL := "http://banners:8080/v1/user_banner"

	req, err := http.NewRequest(http.MethodGet, userBannerURL, nil)
	if err != nil {
		t.Fatalf("Failed to create user banner request: %v", err)
	}

	q := req.URL.Query()
	q.Add("tag_id", "1")
	q.Add("feature_id", "2")
	req.URL.RawQuery = q.Encode()
	req.Header.Set("Authorization", "Bearer "+loginResponse.Token)

	userBannerResp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to send user banner request: %v", err)
	}
	defer userBannerResp.Body.Close()

	if userBannerResp.StatusCode != http.StatusNotFound {
		t.Errorf("Unexpected status code for user banner: got %d, want %d", userBannerResp.StatusCode, http.StatusNotFound)
	}
}
