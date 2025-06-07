package dto

import (
	"fido2/pkg/utils/common"
	"github.com/go-webauthn/webauthn/protocol"
)

type CredentialGetOptionsRequest struct {
	Username         string `json:"username,omitzero"`
	UserVerification string `json:"userVerification,omitzero"`
}

type CredentialGetOptionsResponse struct {
	common.CommonResponse
	protocol.PublicKeyCredentialRequestOptions
}

type AuthenticatorAssertionResponseRequest struct {
	Id                        string                         `json:"id,omitzero"`
	Response                  AuthenticatorAssertionResponse `json:"response,omitzero"`
	GetClientExtensionResults map[string]interface{}         `json:"getClientExtensionResults,omitzero"`
	Type                      string                         `json:"type,omitzero"`
}

type AuthenticatorAssertionResponse struct {
	AuthenticatorData string `json:"authenticatorData,omitzero"`
	ClientDataJSON    string `json:"clientDataJSON,omitzero"`
	Signature         string `json:"signature,omitzero"`
	UserHandle        string `json:"userHandle,omitzero"`
}