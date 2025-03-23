package integration

import (
	"context"
	"encoding/base64"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthFlow(t *testing.T) {
	ts, cleanup := setupTestServer(t)
	defer cleanup()

	// Generate test keys on the client side
	publicKey := make([]byte, 800)
	encryptedPrivateKey := make([]byte, 1200)
	salt := make([]byte, 16)

	// Step 1: Register a new user
	username := "testuser_" + time.Now().Format("20060102150405")
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

	assert.Equal(t, http.StatusCreated, registerResp.StatusCode)

	// Parse the response to get token
	var registerResult map[string]interface{}
	ReadJSONBody(t, registerResp.Body, &registerResult)

	// Verify structure
	assert.True(t, registerResult["success"].(bool))
	data := registerResult["data"].(map[string]interface{})
	assert.NotEmpty(t, data["access_token"])
	assert.Equal(t, "Bearer", data["token_type"])
	assert.Greater(t, int(data["expires_in"].(float64)), 0)

	// Extract token
	token := data["access_token"].(string)

	// Step 2: Use the token to access a protected endpoint
	req, err := http.NewRequest("GET", ts.URL+"/api/v1/keys/private", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	privKeyResp, err := ts.Client().Do(req)
	require.NoError(t, err)
	defer privKeyResp.Body.Close()

	assert.Equal(t, http.StatusOK, privKeyResp.StatusCode)

	// Step 3: Logout
	logoutReq, err := http.NewRequest("POST", ts.URL+"/api/v1/auth/logout", nil)
	require.NoError(t, err)
	logoutReq.Header.Set("Authorization", "Bearer "+token)

	logoutResp, err := ts.Client().Do(logoutReq)
	require.NoError(t, err)
	defer logoutResp.Body.Close()

	assert.Equal(t, http.StatusOK, logoutResp.StatusCode)

	// Step 4: Try to use the token after logout (should fail)
	invalidReq, err := http.NewRequest("GET", ts.URL+"/api/v1/keys/private", nil)
	require.NoError(t, err)
	invalidReq.Header.Set("Authorization", "Bearer "+token)

	invalidResp, err := ts.Client().Do(invalidReq)
	require.NoError(t, err)
	defer invalidResp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, invalidResp.StatusCode)

	// Clean up test user
	ctx := context.Background()
	err = testEnv.UserRepo.Delete(ctx, testEnv.Security.HashUsername(username))
	require.NoError(t, err)
}
