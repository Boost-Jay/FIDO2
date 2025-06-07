package controller

import (
	"encoding/base64"
	"encoding/json"
	"fido2/internal/dto"
	"fido2/internal/entity"
	wAuth "fido2/internal/platform/webauthn"
	"fido2/internal/usecase"
	"fido2/pkg/utils"
	"fido2/pkg/utils/common"
	"github.com/gin-gonic/gin"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"net/http"
)

var assertionSessionData *webauthn.SessionData

type AuthController struct {
	UserUC usecase.UserUseCase
}

func NewAuthController(u usecase.UserUseCase) *AuthController {
	return &AuthController{UserUC: u}
}

// StartAssertionHandler Credential Get Options
// WebAuthn 產生登入資訊的請求
func (c *AuthController) StartAssertionHandler(ctx *gin.Context) {
	utils.GetLogger().Info("StartAssertionHandler called")

	var request *dto.CredentialGetOptionsRequest

	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(
			http.StatusBadRequest,
			common.CommonResponse{
				Status:       "failed",
				ErrorMessage: "failed to parse request body, error: " + err.Error(),
			},
		)
		return
	}

	foundUser, err := c.UserUC.GetUserByUsername(request.Username)
	if err != nil {
		ctx.JSON(
			http.StatusInternalServerError,
			common.CommonResponse{
				Status:       "failed",
				ErrorMessage: "failed to get user by name, error: " + err.Error(),
			},
		)
		return
	}

	utils.GetLogger().Infof("Found user: %s", foundUser.Challenge)

	webauthnUser := wAuth.NewUserWebAuthn(foundUser)

	authenticatorSelection := func(options *protocol.PublicKeyCredentialRequestOptions) {
		options.UserVerification = protocol.UserVerificationRequirement(request.UserVerification)
	}

	options, sessionData, err := wAuth.WebAuthn.BeginLogin(webauthnUser, authenticatorSelection)
	if err != nil {
		ctx.JSON(
			http.StatusInternalServerError,
			common.CommonResponse{
				Status:       "failed",
				ErrorMessage: "failed to begin login, error: " + err.Error(),
			},
		)
		return
	}

	sessionData.Challenge = base64.RawStdEncoding.EncodeToString([]byte(sessionData.Challenge))

	utils.GetLogger().Infof("Session data: %+v", sessionData)

	assertionSessionData = sessionData
	// 更新使用者 Challenge 並呼叫 UpdateUser
	if err = c.UserUC.UpdateUser(
		foundUser, &entity.User{
			Challenge: options.Response.Challenge.String(),
		},
	); err != nil {
		ctx.JSON(
			http.StatusInternalServerError,
			common.CommonResponse{
				Status:       "failed",
				ErrorMessage: "failed to update user, error: " + err.Error(),
			},
		)
		return
	}

	utils.GetLogger().Infof("Updated user challenge: %s", foundUser.Challenge)

	ctx.JSON(
		http.StatusOK,
		dto.CredentialGetOptionsResponse{
			CommonResponse: common.CommonResponse{
				Status:       "ok",
				ErrorMessage: "",
			},
			PublicKeyCredentialRequestOptions: options.Response,
		},
	)
}

// FinishAssertionHandler Authenticator Assertion Response
// WebAuthn 驗證登入資訊的請求
func (c *AuthController) FinishAssertionHandler(ctx *gin.Context) {
	utils.GetLogger().Info("FinishAssertionHandler called")

	var request *dto.AuthenticatorAssertionResponseRequest

	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(
			http.StatusBadRequest,
			common.CommonResponse{
				Status:       "failed",
				ErrorMessage: "failed to parse request body, error: " + err.Error(),
			},
		)
		return
	}

	utils.GetLogger().Infof("Request: %+v", request)

	// 將請求物件序列化為 JSON 字串紀錄
	if reqBodyBytes, err := json.Marshal(request); err != nil {
		utils.GetLogger().Errorf("failed to marshal request: %v", err)
	} else {
		utils.GetLogger().Infof("Request body: %s", string(reqBodyBytes))
	}

	authenticatorClientDataJSON, err := base64.RawURLEncoding.DecodeString(request.Response.ClientDataJSON)
	if err != nil {
		ctx.JSON(
			http.StatusBadRequest,
			common.CommonResponse{
				Status:       "failed",
				ErrorMessage: "failed to decode clientDataJSON, error: " + err.Error(),
			},
		)
		return
	}

	authenticatorData, err := base64.RawURLEncoding.DecodeString(request.Response.AuthenticatorData)
	if err != nil {
		ctx.JSON(
			http.StatusBadRequest,
			common.CommonResponse{
				Status:       "failed",
				ErrorMessage: "failed to decode authenticatorData, error: " + err.Error(),
			},
		)
		return
	}

	authenticatorSignature, err := base64.RawURLEncoding.DecodeString(request.Response.Signature)
	if err != nil {
		ctx.JSON(
			http.StatusBadRequest,
			common.CommonResponse{
				Status:       "failed",
				ErrorMessage: "failed to decode signature, error: " + err.Error(),
			},
		)
		return
	}

	authenticatorUserHandle, err := base64.RawURLEncoding.DecodeString(request.Response.UserHandle)
	if err != nil {
		ctx.JSON(
			http.StatusBadRequest,
			common.CommonResponse{
				Status:       "failed",
				ErrorMessage: "failed to decode userHandle, error: " + err.Error(),
			},
		)
		return
	}

	var clientDataJSON map[string]interface{}
	if err := json.Unmarshal(authenticatorClientDataJSON, &clientDataJSON); err != nil {
		ctx.JSON(
			http.StatusBadRequest,
			common.CommonResponse{
				Status:       "failed",
				ErrorMessage: "failed to unmarshal clientDataJSON, error: " + err.Error(),
			},
		)
		return
	}

	utils.GetLogger().Infof("ClientDataJSON: %+v", clientDataJSON["challenge"].(string))

	challenge, ok := clientDataJSON["challenge"].(string)
	if !ok {
		ctx.JSON(
			http.StatusBadRequest,
			common.CommonResponse{
				Status:       "failed",
				ErrorMessage: "challenge not found",
			},
		)
		return
	}

	if challenge != assertionSessionData.Challenge {
		ctx.JSON(
			http.StatusBadRequest,
			common.CommonResponse{
				Status:       "failed",
				ErrorMessage: "challenge mismatch",
			},
		)
		return
	}

	decodedChallenge, err := base64.RawURLEncoding.DecodeString(challenge)
	if err != nil {
		ctx.JSON(
			http.StatusBadRequest,
			common.CommonResponse{
				Status:       "failed",
				ErrorMessage: "failed to decode challenge, error: " + err.Error(),
			},
		)
		return
	}
	challenge = string(decodedChallenge)

	foundUser, err := c.UserUC.GetUserByChallenge(challenge)
	if err != nil {
		ctx.JSON(
			http.StatusInternalServerError,
			common.CommonResponse{
				Status:       "failed",
				ErrorMessage: "failed to get user by challenge, error: " + err.Error(),
			},
		)
		return
	}

	credentialRawID, err := utils.DecodeCredentialRawID(request.Id)
	if err != nil {
		ctx.JSON(
			http.StatusBadRequest,
			common.CommonResponse{
				Status:       "failed",
				ErrorMessage: "failed to decode credential ID, error: " + err.Error(),
			},
		)
		return
	}

	utils.GetLogger().Infof("Found user: %s", foundUser.TableName())

	car := protocol.CredentialAssertionResponse{
		PublicKeyCredential: protocol.PublicKeyCredential{
			Credential: protocol.Credential{
				ID:   request.Id,
				Type: request.Type,
			},
			RawID:                  protocol.URLEncodedBase64(credentialRawID),
			ClientExtensionResults: request.GetClientExtensionResults,
		},
		AssertionResponse: protocol.AuthenticatorAssertionResponse{
			AuthenticatorResponse: protocol.AuthenticatorResponse{
				ClientDataJSON: protocol.URLEncodedBase64(authenticatorClientDataJSON),
			},
			AuthenticatorData: protocol.URLEncodedBase64(authenticatorData),
			Signature:         protocol.URLEncodedBase64(authenticatorSignature),
			UserHandle:        protocol.URLEncodedBase64(authenticatorUserHandle),
		},
	}
	pca, err := car.Parse()
	if err != nil {
		ctx.JSON(
			http.StatusInternalServerError,
			common.CommonResponse{
				Status:       "failed",
				ErrorMessage: "failed to parse assertion response, error: " + err.Error(),
			},
		)
		return
	}

	utils.GetLogger().Infof("Parsed PublicKeyCredential: %+v", pca)

	webauthnUser := wAuth.NewUserWebAuthn(foundUser)

	if _, err := wAuth.WebAuthn.ValidateLogin(webauthnUser, *assertionSessionData, pca); err != nil {
		ctx.JSON(
			http.StatusInternalServerError,
			common.CommonResponse{
				Status:       "failed",
				ErrorMessage: "failed to validate login, error: " + err.Error(),
			},
		)
		return
	}

	utils.GetLogger().Infof("User %s logged in successfully", foundUser.TableName())

	ctx.JSON(
		http.StatusOK,
		common.CommonResponse{
			Status:       "ok",
			ErrorMessage: "",
		},
	)
}