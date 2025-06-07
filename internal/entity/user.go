package entity

type User struct {
	// ID 使用者 ID
	ID string `json:"userId,omitzero" gorm:"primaryKey"`

	// UserName 使用者名稱
	UserName string `json:"name,omitzero"`

	// DisplayName 使用者的顯示名稱
	DisplayName string `json:"displayName,omitzero"`

	// Challenge 當次進行 WebAuthn 註冊 / 驗證流程時的使用者 Challenge
	Challenge string `json:"challenge,omitzero"`

	// Credential 使用者的 WebAuthn Credential
	Credential string `json:"credential,oomitzero"`
}

// TableName 設定資料庫表名
func (*User) TableName() string {
	return "user"
}