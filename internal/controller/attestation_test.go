package controller

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fido2/internal/dto"
	"fido2/internal/entity"
	mocks "fido2/internal/usecase"
	"fido2/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"testing"
)

// 測試成功流程
func TestStartAttestationHandler_Success(t *testing.T) {
	mockUC := mocks.NewMockUserUseCase(t)
	c := NewAuthController(mockUC)

	// input 輸入
	reqBody := dto.CredentialCreationOptionsRequest{
		Username:    "testuser",
		DisplayName: "Test User",
		AuthenticatorSelection: protocol.AuthenticatorSelection{
			AuthenticatorAttachment: protocol.Platform,
			RequireResidentKey:      utils.Bool(false),
			UserVerification:        protocol.VerificationPreferred,
		},
		Attestation: "direct",
	}
	body, _ := json.Marshal(reqBody)

	// 期望 mock CreateUser 被呼叫，並回傳 nil（成功）
	mockUC.EXPECT().
		CreateUser(mock.AnythingOfType("*domain.User")).
		Return(nil)

	// Gin context
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request, _ = http.NewRequest("POST", "/attestation/options", bytes.NewBuffer(body))
	ctx.Request.Header.Set("Content-Type", "application/json")

	// Act
	c.StartAttestationHandler(ctx)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	// 你可視需求解析 response 內容
}

// 測試 CreateUser 失敗流程
func TestStartAttestationHandler_CreateUserFail(t *testing.T) {
	mockUC := mocks.NewMockUserUseCase(t)
	c := NewAuthController(mockUC)

	reqBody := dto.CredentialCreationOptionsRequest{
		Username:    "testuser",
		DisplayName: "Test User",
		AuthenticatorSelection: protocol.AuthenticatorSelection{
			AuthenticatorAttachment: protocol.Platform,
			RequireResidentKey:      utils.Bool(false),
			UserVerification:        protocol.VerificationPreferred,
		},
		Attestation: "direct",
	}
	body, _ := json.Marshal(reqBody)

	mockUC.EXPECT().
		CreateUser(mock.AnythingOfType("*domain.User")).
		Return(errors.New("db error"))

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request, _ = http.NewRequest("POST", "/attestation/options", bytes.NewBuffer(body))
	ctx.Request.Header.Set("Content-Type", "application/json")

	c.StartAttestationHandler(ctx)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestFinishAttestationHandler_Success(t *testing.T) {
	mockUC := mocks.NewMockUserUseCase(t)
	c := NewAuthController(mockUC)

	// 假設 attestationSessionData 全局變數先設好
	attestationSessionData = &webauthn.SessionData{
		Challenge: "test-challenge",
	}

	// 模擬輸入
	clientData := map[string]interface{}{
		"challenge": "test-challenge",
	}
	clientDataJSON, _ := json.Marshal(clientData)

	req := dto.AuthenticatorAttestationResponseRequest{
		Id: "fakeid",
		Response: dto.AuthenticatorAttestationResponse{
			AttestationObject: base64.RawURLEncoding.EncodeToString([]byte("fake")),
			ClientDataJSON:    base64.RawURLEncoding.EncodeToString(clientDataJSON),
		},
		GetClientExtensionResults: map[string]interface{}{},
		Type:                      "public-key",
	}
	body, _ := json.Marshal(req)

	// 模擬找到 user
	mockUC.EXPECT().
		GetUserByChallenge("test-challenge").
		Return(&entity.User{
			ID:          "1",
			UserName:    "testuser",
			DisplayName: "Test User",
		}, nil)

	// 模擬 UpdateUser 成功
	mockUC.EXPECT().
		UpdateUser(mock.AnythingOfType("*domain.User")).
		Return(nil)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request, _ = http.NewRequest("POST", "/attestation/result", bytes.NewBuffer(body))
	ctx.Request.Header.Set("Content-Type", "application/json")

	c.FinishAttestationHandler(ctx)

	assert.Equal(t, http.StatusOK, w.Code)
}

// 測試 GetUserByChallenge 失敗
func TestFinishAttestationHandler_GetUserByChallengeFail(t *testing.T) {
	mockUC := mocks.NewMockUserUseCase(t)
	c := NewAuthController(mockUC)

	attestationSessionData = &webauthn.SessionData{
		Challenge: "test-challenge",
	}
	clientData := map[string]interface{}{
		"challenge": "test-challenge",
	}
	clientDataJSON, _ := json.Marshal(clientData)

	req := dto.AuthenticatorAttestationResponseRequest{
		Id: "fakeid",
		Response: dto.AuthenticatorAttestationResponse{
			AttestationObject: base64.RawURLEncoding.EncodeToString([]byte("fake")),
			ClientDataJSON:    base64.RawURLEncoding.EncodeToString(clientDataJSON),
		},
		GetClientExtensionResults: map[string]interface{}{},
		Type:                      "public-key",
	}
	body, _ := json.Marshal(req)

	mockUC.EXPECT().
		GetUserByChallenge("test-challenge").
		Return(nil, errors.New("not found"))

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request, _ = http.NewRequest("POST", "/attestation/result", bytes.NewBuffer(body))
	ctx.Request.Header.Set("Content-Type", "application/json")

	c.FinishAttestationHandler(ctx)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}