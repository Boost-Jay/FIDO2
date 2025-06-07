package controller

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fido2/internal/dto"
	"fido2/internal/entity"
	mocks "fido2/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStartAssertionHandler_Success(t *testing.T) {
	// Arrange
	mockUC := mocks.NewMockUserUseCase(t)
	c := NewAuthController(mockUC)

	// 建立 input
	reqBody := dto.CredentialGetOptionsRequest{
		Username:         "testuser",
		UserVerification: "preferred",
	}
	body, _ := json.Marshal(reqBody)

	// 模擬 usecase 行為
	mockUC.EXPECT().
		GetUserByUsername("testuser").
		Return(&entity.User{
			ID:          "1",
			UserName:    "testuser",
			DisplayName: "Test User",
			Challenge:   "",
		}, nil)

	mockUC.EXPECT().
		UpdateUser(mock.Anything).
		Return(nil)

	// gin context
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request, _ = http.NewRequest("POST", "/assertion/options", bytes.NewBuffer(body))
	ctx.Request.Header.Set("Content-Type", "application/json")

	// Act
	c.StartAssertionHandler(ctx)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	// 你可以根據 response 再 parse body 做進一步驗證
}

func TestStartAssertionHandler_UserNotFound(t *testing.T) {
	mockUC := mocks.NewMockUserUseCase(t)
	c := NewAuthController(mockUC)

	reqBody := dto.CredentialGetOptionsRequest{
		Username:         "no_such_user",
		UserVerification: "preferred",
	}
	body, _ := json.Marshal(reqBody)

	mockUC.EXPECT().
		GetUserByUsername("no_such_user").
		Return(nil, errors.New("not found"))

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request, _ = http.NewRequest("POST", "/assertion/options", bytes.NewBuffer(body))
	ctx.Request.Header.Set("Content-Type", "application/json")

	c.StartAssertionHandler(ctx)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestFinishAssertionHandler_Success(t *testing.T) {
	mockUC := mocks.NewMockUserUseCase(t)
	c := NewAuthController(mockUC)

	// 設定全域 assertionSessionData（模擬 Challenge）
	assertionSessionData = &webauthn.SessionData{
		Challenge: "test-challenge",
	}

	// clientDataJSON payload 構造
	clientData := map[string]interface{}{
		"challenge": "test-challenge",
	}
	clientDataJSON, _ := json.Marshal(clientData)
	encodedClientData := base64.RawURLEncoding.EncodeToString(clientDataJSON)

	// request input
	req := dto.AuthenticatorAssertionResponseRequest{
		Id: "credid",
		Response: dto.AuthenticatorAssertionResponse{
			ClientDataJSON:    encodedClientData,
			AuthenticatorData: base64.RawURLEncoding.EncodeToString([]byte("data")),
			Signature:         base64.RawURLEncoding.EncodeToString([]byte("sig")),
			UserHandle:        base64.RawURLEncoding.EncodeToString([]byte("user-handle")),
		},
		GetClientExtensionResults: map[string]interface{}{},
		Type:                      "public-key",
	}
	body, _ := json.Marshal(req)

	// mock UserUseCase
	mockUC.EXPECT().
		GetUserByChallenge("test-challenge").
		Return(&entity.User{
			ID:          "1",
			UserName:    "testuser",
			DisplayName: "Test User",
		}, nil)

	// 這邊其實不需要 UpdateUser, 因為 handler 沒有呼叫，但如果有要 mock

	// gin context
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request, _ = http.NewRequest("POST", "/assertion/result", bytes.NewBuffer(body))
	ctx.Request.Header.Set("Content-Type", "application/json")

	// Act
	c.FinishAssertionHandler(ctx)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestFinishAssertionHandler_ChallengeMismatch(t *testing.T) {
	mockUC := mocks.NewMockUserUseCase(t)
	c := NewAuthController(mockUC)

	assertionSessionData = &webauthn.SessionData{
		Challenge: "correct-challenge",
	}

	// clientData challenge 與 session 不同
	clientData := map[string]interface{}{
		"challenge": "wrong-challenge",
	}
	clientDataJSON, _ := json.Marshal(clientData)
	encodedClientData := base64.RawURLEncoding.EncodeToString(clientDataJSON)

	req := dto.AuthenticatorAssertionResponseRequest{
		Id: "credid",
		Response: dto.AuthenticatorAssertionResponse{
			ClientDataJSON:    encodedClientData,
			AuthenticatorData: base64.RawURLEncoding.EncodeToString([]byte("data")),
			Signature:         base64.RawURLEncoding.EncodeToString([]byte("sig")),
			UserHandle:        base64.RawURLEncoding.EncodeToString([]byte("user-handle")),
		},
		GetClientExtensionResults: map[string]interface{}{},
		Type:                      "public-key",
	}
	body, _ := json.Marshal(req)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request, _ = http.NewRequest("POST", "/assertion/result", bytes.NewBuffer(body))
	ctx.Request.Header.Set("Content-Type", "application/json")

	c.FinishAssertionHandler(ctx)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestFinishAssertionHandler_UserNotFound(t *testing.T) {
	mockUC := mocks.NewMockUserUseCase(t)
	c := NewAuthController(mockUC)

	assertionSessionData = &webauthn.SessionData{
		Challenge: "test-challenge",
	}

	clientData := map[string]interface{}{
		"challenge": "test-challenge",
	}
	clientDataJSON, _ := json.Marshal(clientData)
	encodedClientData := base64.RawURLEncoding.EncodeToString(clientDataJSON)

	req := dto.AuthenticatorAssertionResponseRequest{
		Id: "credid",
		Response: dto.AuthenticatorAssertionResponse{
			ClientDataJSON:    encodedClientData,
			AuthenticatorData: base64.RawURLEncoding.EncodeToString([]byte("data")),
			Signature:         base64.RawURLEncoding.EncodeToString([]byte("sig")),
			UserHandle:        base64.RawURLEncoding.EncodeToString([]byte("user-handle")),
		},
		GetClientExtensionResults: map[string]interface{}{},
		Type:                      "public-key",
	}
	body, _ := json.Marshal(req)

	// mock UserUseCase 查無資料
	mockUC.EXPECT().
		GetUserByChallenge("test-challenge").
		Return(nil, errors.New("not found"))

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request, _ = http.NewRequest("POST", "/assertion/result", bytes.NewBuffer(body))
	ctx.Request.Header.Set("Content-Type", "application/json")

	c.FinishAssertionHandler(ctx)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestFinishAssertionHandler_InvalidBase64(t *testing.T) {
	mockUC := mocks.NewMockUserUseCase(t)
	c := NewAuthController(mockUC)

	assertionSessionData = &webauthn.SessionData{
		Challenge: "test-challenge",
	}

	req := dto.AuthenticatorAssertionResponseRequest{
		Id: "credid",
		Response: dto.AuthenticatorAssertionResponse{
			ClientDataJSON:    "!!!invalidbase64!!!",
			AuthenticatorData: base64.RawURLEncoding.EncodeToString([]byte("data")),
			Signature:         base64.RawURLEncoding.EncodeToString([]byte("sig")),
			UserHandle:        base64.RawURLEncoding.EncodeToString([]byte("user-handle")),
		},
		GetClientExtensionResults: map[string]interface{}{},
		Type:                      "public-key",
	}
	body, _ := json.Marshal(req)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request, _ = http.NewRequest("POST", "/assertion/result", bytes.NewBuffer(body))
	ctx.Request.Header.Set("Content-Type", "application/json")

	c.FinishAssertionHandler(ctx)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}