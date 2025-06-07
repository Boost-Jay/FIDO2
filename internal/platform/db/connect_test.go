package db

import (
	"fido2/pkg/utils"
	"testing"
	"time"

	"fido2/internal/entity"
)

func TestConnect_Postgres(t *testing.T) {
	if err := utils.LoadEnv; err != nil {
		t.Fatalf("讀取 .env 失敗: %v", err)
	}

	// （可選）等待容器啟動
	time.Sleep(3 * time.Second)

	// 執行
	Connect()
	conn := GetDB()

	// 斷言
	if conn == nil {
		t.Fatal("資料庫連線為 nil")
	}

	if err := conn.AutoMigrate(&entity.User{}); err != nil {
		t.Fatalf("AutoMigrate 失敗: %v", err)
	}
}