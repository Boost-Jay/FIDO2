package dto

import (
	"fido2/pkg/utils/common"
	"github.com/go-webauthn/webauthn/protocol"
)

type CredentialCreationOptionsRequest struct {
	Username               string                          `json:"username,omitzero"`
	DisplayName            string                          `json:"displayName,omitzero"`
	AuthenticatorSelection protocol.AuthenticatorSelection `json:"authenticatorSelection,omitzero"`
	Attestation            string                          `json:"attestation,omitzero"`
}

type CredentialCreationOptionsResponse struct {
	common.CommonResponse
	protocol.PublicKeyCredentialCreationOptions
}

type AuthenticatorAttestationResponseRequest struct {
	Id                        string                           `json:"id,omitzero"`
	Response                  AuthenticatorAttestationResponse `json:"response,omitzero"`
	GetClientExtensionResults map[string]interface{}           `json:"getClientExtensionResults,omitzero"`
	Type                      string                           `json:"type,omitzero"`
}

type AuthenticatorAttestationResponse struct {
	AttestationObject string `json:"attestationObject,omitzero"`
	ClientDataJSON    string `json:"clientDataJSON,omitzero"`
}