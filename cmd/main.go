package main

import (
	"fido2/internal/platform/db"
	"fido2/internal/platform/webauthn"
	"fido2/internal/router"
	"fido2/pkg/utils"
)

func init() {
	if err := utils.LoadEnv(); err != nil {
		utils.GetLogger().Errorf("讀取 .env 失敗: %v", err)
	}
}

func main() {
	db.Connect()
	webauthn.NewRPServer()
	router.InitRouter()
}