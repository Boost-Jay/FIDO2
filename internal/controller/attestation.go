package controller

import (
	"encoding/base64"
	"encoding/json"
	"fido2/internal/dto"
	"fido2/internal/entity"
	wAuth "fido2/internal/platform/webauthn"
	"fido2/pkg/utils"
	"fido2/pkg/utils/common"
	"github.com/gin-gonic/gin"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
	"net/http"
)

var attestationSessionData *webauthn.SessionData

// StartAttestationHandler Credential Creation Options
// WebAuthn 產生註冊資訊的請求
func (c *AuthController) StartAttestationHandler(ctx *gin.Context) {

	utils.GetLogger().Info("StartAttestationHandler called")

	var request *dto.CredentialCreationOptionsRequest

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

	user := &entity.User{
		ID:          uuid.New().String(),
		UserName:    request.Username,
		DisplayName: request.DisplayName,
	}

	webauthnUser := wAuth.NewUserWebAuthn(user)
	//excludeCredentialsOption := webauthn.WithExclusions(webauthnUser.CredentialExcludeList())
	//authenticatorSelectionOption := webauthn.WithAuthenticatorSelection(request.AuthenticatorSelection)
	//attestationOption := webauthn.WithConveyancePreference(protocol.ConveyancePreference(request.Attestation))

	opts := func(options *protocol.PublicKeyCredentialCreationOptions) {
		options.CredentialExcludeList = webauthnUser.CredentialExcludeList()
		options.AuthenticatorSelection = request.AuthenticatorSelection
		options.Attestation = protocol.ConveyancePreference(request.Attestation)
	}

	options, sessionData, err := wAuth.WebAuthn.BeginRegistration(webauthnUser, opts)

	if err != nil {
		utils.GetLogger().Error("begin registration failed, error: ", err.Error())
		ctx.JSON(
			http.StatusInternalServerError,
			common.CommonResponse{
				Status:       "failed",
				ErrorMessage: "failed to create credential creation options, error: " + err.Error(),
			},
		)
		return
	}

	utils.GetLogger().Infof("Credential Creation Options: %+v", options)

	user.Challenge = options.Response.Challenge.String()
	user.Credential = "`" + "{}" + "`"

	if err := c.UserUC.CreateUser(user); err != nil {
		ctx.JSON(
			http.StatusInternalServerError,
			common.CommonResponse{
				Status:       "failed",
				ErrorMessage: "failed to create user, error: " + err.Error(),
			},
		)
		return
	}

	sessionData.Challenge = base64.RawStdEncoding.EncodeToString([]byte(sessionData.Challenge))

	utils.GetLogger().Infof("Created user: %+v", user)

	attestationSessionData = sessionData

	ctx.JSON(
		http.StatusOK,
		dto.CredentialCreationOptionsResponse{
			CommonResponse: common.CommonResponse{
				Status:       "ok",
				ErrorMessage: "",
			},
			PublicKeyCredentialCreationOptions: options.Response,
		},
	)
}

// FinishAttestationHandler Authenticator Attestation Response
// WebAuthn 驗證註冊資訊的請求
func (c *AuthController) FinishAttestationHandler(ctx *gin.Context) {
	utils.GetLogger().Info("Processing attestation response")

	var request *dto.AuthenticatorAttestationResponseRequest

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

	utils.GetLogger().Infof("Client Data JSON: %+v", clientDataJSON)

	challenge, ok := clientDataJSON["challenge"].(string)
	if !ok || challenge == "" {
		ctx.JSON(
			http.StatusBadRequest,
			common.CommonResponse{
				Status:       "failed",
				ErrorMessage: "challenge is missing or not a string",
			},
		)
		return
	}

	if challenge != attestationSessionData.Challenge {
		ctx.JSON(
			http.StatusBadRequest,
			common.CommonResponse{
				Status:       "failed",
				ErrorMessage: "challenge mismatch",
			},
		)
		return
	} else {

		decodedChallenge, err := base64.RawURLEncoding.DecodeString(challenge)

		if err != nil {
			ctx.JSON(
				http.StatusBadRequest,
				common.CommonResponse{
					Status:       "failed",
					ErrorMessage: "failed to decode challenge, error: " + err.Error(),
				},
			)
		}

		challenge = string(decodedChallenge)

		// decodedChallenge 在 err != nil 時不會有值
		//if decodedChallenge, err := base64.RawURLEncoding.DecodeString(challenge); err != nil {
		//	ctx.JSON(
		//		http.StatusBadRequest,
		//		common.CommonResponse{
		//			Status:       "failed",
		//			ErrorMessage: "failed to decode challenge, error: " + err.Error(),
		//		},
		//	)
		//	challenge = string(decodedChallenge)
		//}
	}

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

	utils.GetLogger().Infof("Found user: %+v", foundUser)

	authenticatorAttestationObject, err := base64.RawURLEncoding.DecodeString(request.Response.AttestationObject)
	if err != nil {
		ctx.JSON(
			http.StatusBadRequest,
			common.CommonResponse{
				Status:       "failed",
				ErrorMessage: "failed to decode attestationObject, error: " + err.Error(),
			},
		)
		return
	}

	ccr := protocol.CredentialCreationResponse{
		PublicKeyCredential: protocol.PublicKeyCredential{
			Credential: protocol.Credential{
				ID:   request.Id,
				Type: request.Type,
			},
			RawID:                  []byte(request.Id),
			ClientExtensionResults: request.GetClientExtensionResults,
		},
		AttestationResponse: protocol.AuthenticatorAttestationResponse{
			AttestationObject: protocol.URLEncodedBase64(authenticatorAttestationObject),
			AuthenticatorResponse: protocol.AuthenticatorResponse{
				ClientDataJSON: authenticatorClientDataJSON,
			},
		},
	}

	pcc, err := ccr.Parse()
	if err != nil {
		ctx.JSON(
			http.StatusBadRequest,
			common.CommonResponse{
				Status:       "failed",
				ErrorMessage: "failed to parse credential creation response, error: " + err.Error(),
			},
		)
		return
	}
	utils.GetLogger().Infof("Parsed PublicKeyCredential: %+v", pcc)

	// 將 domain.User 包裝為 WebAuthn User
	webauthnUser := wAuth.NewUserWebAuthn(foundUser)

	credential, err := wAuth.WebAuthn.CreateCredential(webauthnUser, *attestationSessionData, pcc)

	if err != nil {
		ctx.JSON(
			http.StatusInternalServerError,
			common.CommonResponse{
				Status:       "failed",
				ErrorMessage: "failed to create credential, error: " + err.Error(),
			},
		)
		return
	}

	utils.GetLogger().Infof("Created credential: %+v", credential)

	credentialJSON, err := json.Marshal(credential)
	if err != nil {
		ctx.JSON(
			http.StatusInternalServerError,
			common.CommonResponse{
				Status:       "failed",
				ErrorMessage: "failed to marshal credential, error: " + err.Error(),
			},
		)
		return
	}

	utils.GetLogger().Infof("Credential JSON: %s", string(credentialJSON))

	// 更新使用者憑證後再呼叫 UpdateUser
	if err = c.UserUC.UpdateUser(
		foundUser, entity.User{
			Credential: "`" + string(credentialJSON) + "`",
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

	utils.GetLogger().Infof("User %s registered successfully with credential ID: %s", foundUser.ID, request.Id)

	ctx.JSON(
		http.StatusOK,
		common.CommonResponse{
			Status:       "ok",
			ErrorMessage: "",
		},
	)
}