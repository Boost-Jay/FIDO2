package webauthn

import (
	"encoding/json"
	"fido2/internal/entity"
	"fido2/internal/usecase/impl"
	"fido2/pkg/utils"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"strconv"
	"strings"
)

type UserWebAuthn struct {
	*entity.User
}

// NewUserWebAuthn creates a new UserWebAuthn wrapper
func NewUserWebAuthn(user *entity.User) *UserWebAuthn {
	return &UserWebAuthn{User: user}
}

// WebAuthn 介面實作 - 這些方法實現了 webauthn.User 介面

// WebAuthnID 取得使用者的 WebAuthn ID
func (u *UserWebAuthn) WebAuthnID() []byte {
	return []byte(u.UserName)
}

// WebAuthnName 取得使用者的 WebAuthn 名稱
func (u *UserWebAuthn) WebAuthnName() string {
	return u.UserName
}

// WebAuthnDisplayName 取得使用者的 WebAuthn 顯示名稱
func (u *UserWebAuthn) WebAuthnDisplayName() string {
	return u.DisplayName
}

// WebAuthnIcon 取得使用者的圖示 (可選方法)
func (u *UserWebAuthn) WebAuthnIcon() string {
	return ""
}

// WebAuthnCredentials 取得使用者的所有 Credential
func (u *UserWebAuthn) WebAuthnCredentials() []webauthn.Credential {
	return u.getAllCredentials()
}

// CredentialExcludeList WebAuthnCredentialByID 根據 Credential ID 取得對應的 Credential
func (u *UserWebAuthn) CredentialExcludeList() []protocol.CredentialDescriptor {
	var credentialExcludeList []protocol.CredentialDescriptor
	for _, credential := range u.WebAuthnCredentials() {
		descriptor := credential.Descriptor()
		credentialExcludeList = append(credentialExcludeList, descriptor)
	}
	return credentialExcludeList
}

// 輔助方法 - 用於處理 WebAuthn 相關邏輯

// getAllCredentials 取得所有使用者的憑證
func (u *UserWebAuthn) getAllCredentials() []webauthn.Credential {
	credentials := []webauthn.Credential{}

	// 取得所有使用者
	userUseCase := impl.GetUserUseCase()
	allUsers, err := userUseCase.GetUsers()
	if err != nil {
		utils.GetLogger().Errorf("Failed to get users: %v", err)
		return credentials
	}

	// 遍歷所有使用者，處理他們的憑證
	for _, user := range allUsers {
		// 跳過沒有有效憑證的使用者
		if user.Credential == "" || user.Credential == "`{}`" {
			continue
		}

		// 處理憑證字串
		creds := parseCredential(user.Credential)
		credentials = append(credentials, creds...)
	}

	return credentials
}

// parseCredential 解析單一使用者的憑證字串
func parseCredential(credentialStr string) []webauthn.Credential {
	logger := utils.GetLogger()

	// 處理可能帶有反引號的憑證字串
	unquoted := credentialStr
	if strings.HasPrefix(credentialStr, "`") {
		s, err := strconv.Unquote(credentialStr)
		if err != nil {
			logger.Debugf("Unquote failed for credential, using original: %v", err)
			// 移除開頭和結尾的反引號
			if len(credentialStr) > 2 {
				unquoted = credentialStr[1 : len(credentialStr)-1]
			}
		} else {
			unquoted = s
		}
	}

	// 嘗試多種格式解析憑證

	// 嘗試解析為憑證陣列
	var credentials []webauthn.Credential
	if err := json.Unmarshal([]byte(unquoted), &credentials); err == nil {
		return credentials
	}

	// 嘗試解析為單一憑證
	var credential webauthn.Credential
	if err := json.Unmarshal([]byte(unquoted), &credential); err == nil {
		return []webauthn.Credential{credential}
	}

	// 嘗試透過 map 解析
	var credsMap map[string]interface{}
	if err := json.Unmarshal([]byte(unquoted), &credsMap); err == nil {
		credsJson, err := json.Marshal(credsMap)
		if err == nil {
			var cred webauthn.Credential
			if err := json.Unmarshal(credsJson, &cred); err == nil {
				return []webauthn.Credential{cred}
			}
		}
	}

	// 解析失敗時記錄警告
	utils.GetLogger().Warningf("Failed to parse credential, using original: %v", unquoted)
	logger.Warnf("Failed to parse credential string: %s", credentialStr)
	return []webauthn.Credential{}
}