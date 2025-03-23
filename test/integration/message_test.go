package integration

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMessageFlow(t *testing.T) {
	ts, cleanup := setupTestServer(t)
	defer cleanup()

	// Register two users for testing
	sender := registerTestUser(t, ts, "sender_"+time.Now().Format("20060102150405"))
	recipient := registerTestUser(t, ts, "recipient_"+time.Now().Format("20060102150405"))

	// Create dummy encrypted message data
	ciphertextKEM := []byte("dummy_kem_data")
	ciphertextMsg := []byte("dummy_message_data")
	nonce := []byte("dummy_nonce")

	// Step 1: Send message from sender to recipient
	sendReq, err := http.NewRequest("POST", ts.URL+"/api/v1/messages/send", NewJSONBody(map[string]interface{}{
		"recipient_pubkey":      recipient.PublicKey,
		"ciphertext_kem":        base64.URLEncoding.EncodeToString(ciphertextKEM),
		"ciphertext_msg":        base64.URLEncoding.EncodeToString(ciphertextMsg),
		"nonce":                 base64.URLEncoding.EncodeToString(nonce),
		"sender_ciphertext_kem": base64.URLEncoding.EncodeToString(ciphertextKEM),
		"sender_ciphertext_msg": base64.URLEncoding.EncodeToString(ciphertextMsg),
		"sender_nonce":          base64.URLEncoding.EncodeToString(nonce),
	}))
	require.NoError(t, err)
	sendReq.Header.Set("Authorization", "Bearer "+sender.Token)

	sendResp, err := ts.Client().Do(sendReq)
	require.NoError(t, err)
	defer sendResp.Body.Close()

	assert.Equal(t, http.StatusCreated, sendResp.StatusCode)

	// Parse the response to get message ID
	var sendResult map[string]interface{}
	ReadJSONBody(t, sendResp.Body, &sendResult)

	// Verify structure
	assert.True(t, sendResult["success"].(bool))
	msgData := sendResult["data"].(map[string]interface{})
	messageID := msgData["message_id"].(string)
	assert.NotEmpty(t, messageID)

	// Step 2: Recipient retrieves their messages
	getReq, err := http.NewRequest("GET", ts.URL+"/api/v1/messages", nil)
	require.NoError(t, err)
	getReq.Header.Set("Authorization", "Bearer "+recipient.Token)

	getResp, err := ts.Client().Do(getReq)
	require.NoError(t, err)
	defer getResp.Body.Close()

	assert.Equal(t, http.StatusOK, getResp.StatusCode)

	var getResult map[string]interface{}
	ReadJSONBody(t, getResp.Body, &getResult)

	// Verify recipient received the message
	assert.True(t, getResult["success"].(bool))
	messagesData := getResult["data"].(map[string]interface{})
	messages := messagesData["messages"].([]interface{})
	assert.GreaterOrEqual(t, len(messages), 1)

	// Verify sender can get conversation
	convReq, err := http.NewRequest("GET", ts.URL+"/api/v1/messages/conversation/"+recipient.PublicKey, nil)
	require.NoError(t, err)
	convReq.Header.Set("Authorization", "Bearer "+sender.Token)

	convResp, err := ts.Client().Do(convReq)
	require.NoError(t, err)
	defer convResp.Body.Close()

	assert.Equal(t, http.StatusOK, convResp.StatusCode)

	var convResult map[string]interface{}
	ReadJSONBody(t, convResp.Body, &convResult)

	// Verify conversation retrieval worked
	assert.True(t, convResult["success"].(bool))
}

// TestUser represents a test user
type TestUser struct {
	Username   string
	Token      string
	PublicKey  string
	PrivateKey string
}

// registerTestUser registers a test user and returns the user info
func registerTestUser(t *testing.T, ts *httptest.Server, username string) *TestUser {
	// Generate test keys
	publicKey := make([]byte, 800)
	encryptedPrivateKey := make([]byte, 1200)
	salt := make([]byte, 16)

	// Register user
	registerResp, err := ts.Client().Post(
		ts.URL+"/api/v1/auth/register",
		"application/json",
		NewJSONBody(map[string]interface{}{
			"username":              username,
			"public_key":            base64.URLEncoding.EncodeToString(publicKey),
			"encrypted_private_key": base64.URLEncoding.EncodeToString(encryptedPrivateKey),
			"salt":                  base64.URLEncoding.EncodeToString(salt),
		}),
	)
	require.NoError(t, err)
	defer registerResp.Body.Close()

	require.Equal(t, http.StatusCreated, registerResp.StatusCode)

	// Parse response
	var registerResult map[string]interface{}
	ReadJSONBody(t, registerResp.Body, &registerResult)

	data := registerResult["data"].(map[string]interface{})
	token := data["access_token"].(string)

	// Get public key
	pubKeyReq, err := http.NewRequest("GET", ts.URL+"/api/v1/keys/public?username="+username, nil)
	require.NoError(t, err)

	pubKeyResp, err := ts.Client().Do(pubKeyReq)
	require.NoError(t, err)
	defer pubKeyResp.Body.Close()

	require.Equal(t, http.StatusOK, pubKeyResp.StatusCode)

	var pubKeyResult map[string]interface{}
	ReadJSONBody(t, pubKeyResp.Body, &pubKeyResult)

	pubKeyData := pubKeyResult["data"].(map[string]interface{})
	publicKeyStr := pubKeyData["public_key"].(string)

	return &TestUser{
		Username:   username,
		Token:      token,
		PublicKey:  publicKeyStr,
		PrivateKey: base64.URLEncoding.EncodeToString(encryptedPrivateKey),
	}
}
