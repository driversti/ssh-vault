package agent

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"golang.org/x/crypto/ssh"
)

// EnrollConfig holds the parameters for the enrollment flow.
type EnrollConfig struct {
	HubURL  string
	Token   string
	KeyPath string
	Name    string
}

// enrollResponse matches the hub's response to POST /api/enroll.
type enrollResponse struct {
	DeviceID  string `json:"device_id"`
	Challenge string `json:"challenge"`
}

// verifyResponse matches the hub's response to POST /api/enroll/verify.
type verifyResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// Enroll performs the enrollment flow: sends the public key and token to the hub,
// receives a challenge, signs it, and verifies. Returns the device ID and API token
// path where the config is saved.
func Enroll(cfg EnrollConfig) (*Config, error) {
	// Read and parse the SSH private key
	keyData, err := os.ReadFile(cfg.KeyPath)
	if err != nil {
		return nil, fmt.Errorf("reading SSH key: %w", err)
	}

	signer, err := ssh.ParsePrivateKey(keyData)
	if err != nil {
		return nil, fmt.Errorf("parsing SSH key: %w", err)
	}

	// Get the public key in authorized_keys format
	pubKey := strings.TrimSpace(string(ssh.MarshalAuthorizedKey(signer.PublicKey())))

	// Step 1: POST /api/enroll
	enrollBody, _ := json.Marshal(map[string]string{
		"token":      cfg.Token,
		"public_key": pubKey,
		"name":       cfg.Name,
	})

	resp, err := http.Post(
		strings.TrimRight(cfg.HubURL, "/")+"/api/enroll",
		"application/json",
		bytes.NewReader(enrollBody),
	)
	if err != nil {
		return nil, fmt.Errorf("contacting hub: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("enrollment failed (HTTP %d): %s", resp.StatusCode, body)
	}

	var enrollResp enrollResponse
	if err := json.Unmarshal(body, &enrollResp); err != nil {
		return nil, fmt.Errorf("parsing enrollment response: %w", err)
	}

	// Step 2: Sign the challenge
	challengeBytes, err := hex.DecodeString(enrollResp.Challenge)
	if err != nil {
		return nil, fmt.Errorf("decoding challenge: %w", err)
	}

	sig, err := signer.Sign(nil, challengeBytes)
	if err != nil {
		return nil, fmt.Errorf("signing challenge: %w", err)
	}

	sigHex := hex.EncodeToString(ssh.Marshal(sig))

	// Step 3: POST /api/enroll/verify
	verifyBody, _ := json.Marshal(map[string]string{
		"device_id": enrollResp.DeviceID,
		"signature": sigHex,
	})

	resp2, err := http.Post(
		strings.TrimRight(cfg.HubURL, "/")+"/api/enroll/verify",
		"application/json",
		bytes.NewReader(verifyBody),
	)
	if err != nil {
		return nil, fmt.Errorf("contacting hub for verification: %w", err)
	}
	defer resp2.Body.Close()

	body2, _ := io.ReadAll(resp2.Body)
	if resp2.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("verification failed (HTTP %d): %s", resp2.StatusCode, body2)
	}

	var vResp verifyResponse
	if err := json.Unmarshal(body2, &vResp); err != nil {
		return nil, fmt.Errorf("parsing verify response: %w", err)
	}

	fmt.Printf("%s\n", vResp.Message)

	return &Config{
		HubURL:   cfg.HubURL,
		DeviceID: enrollResp.DeviceID,
	}, nil
}
