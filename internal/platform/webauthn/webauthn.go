package webauthn

import (
	"fido2/pkg/utils"

	"github.com/go-webauthn/webauthn/webauthn"
)

var WebAuthn *webauthn.WebAuthn

// NewRPServer 使用指定的配置初始化並返回 WebAuthn RP 伺服器
func NewRPServer() {
	// 建立 WebAuthn 配置
	wConfig := &webauthn.Config{
		RPDisplayName: "my RP server",
		RPID:          "da47-106-107-249-202.ngrok-free.app",
		RPOrigins:     []string{"https://da47-106-107-249-202.ngrok-free.app"},
	}

	webAuthn, err := webauthn.New(wConfig)
	if err != nil {
		utils.GetLogger().Fatalf("Failed to initialize WebAuthn RP server: %v", err)
	} else {
		utils.GetLogger().Info("WebAuthn RP server initialized successfully")
	}

	WebAuthn = webAuthn
}