package types

type WalletSettingsRequest struct {
	Username string `json:"username" binding:"required"`
	Currency string `json:"currency" binding:"required"`
}
