package models

import (
	"time"

	"gorm.io/gorm"
)

type TransactionStatus string

const (
	TransactionPending   TransactionStatus = "pending"
	TransactionCompleted TransactionStatus = "completed"
	TransactionFailed    TransactionStatus = "failed"
)

type TransactionType string

const (
	Deposit    TransactionType = "deposit"
	Withdrawal TransactionType = "withdrawal"
	Conversion TransactionType = "conversion"
	Transfer   TransactionType = "transfer"
)

type TransactionDirection string

const (
	NGNToGHS      TransactionDirection = "NGN-GHS"
	GHSToNGN      TransactionDirection = "GHS-NGN"
	DepositNGN    TransactionDirection = "DEPOSIT-NGN"
	DepositGHS    TransactionDirection = "DEPOSIT-GHS"
	WithdrawalNGN TransactionDirection = "WITHDRAWAL-NGN"
	WithdrawalGHS TransactionDirection = "WITHDRAWAL-GHS"
)

type PaymentType string

const (
	PaymentTypeBank PaymentType = "bank"
	PaymentTypeMomo PaymentType = "momo"
)

type Transaction struct {
	ID                 uint                 `json:"id" gorm:"primarykey"`
	CreatedAt          time.Time            `json:"created_at"`
	UpdatedAt          time.Time            `json:"updated_at"`
	DeletedAt          gorm.DeletedAt       `json:"deleted_at" gorm:"index"`
	Code               string               `json:"code" gorm:"null;index"`
	UserID             uint                 `json:"user_id" gorm:"not null;index"`
	TransactionID      string               `json:"transaction_id" gorm:"not null;uniqueIndex"`
	PaymentType        PaymentType          `json:"payment_type" gorm:"not null"`
	Reason             string               `json:"reason" gorm:"default:''"`
	Status             TransactionStatus    `json:"status" gorm:"default:pending"`
	TransactionType    TransactionType      `json:"transaction_type" gorm:"not null"`
	Reference          string               `json:"reference" gorm:"not null;uniqueIndex"`
	Direction          TransactionDirection `json:"direction" gorm:"not null"`
	Description        string               `json:"description" gorm:"default:''"`
	User               User                 `json:"user" gorm:"not null"`
	TransactionDetails TransactionDetails   `json:"transaction_details" gorm:"foreignKey:TransactionID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

type TransactionDetails struct {
	gorm.Model
	TransactionID   uint    `json:"transaction_id" gorm:"not null;index"`
	RecipientName   string  `json:"recipient_name"`
	AccountNumber   string  `json:"account_number"`
	BankName        string  `json:"bank_name"`
	PhoneNumber     string  `json:"phone_number"`
	Network         string  `json:"network"`
	FromCurrency    string  `json:"from_currency"`
	ToCurrency      string  `json:"to_currency"`
	FromAmount      float64 `json:"from_amount"`
	ToAmount        float64 `json:"to_amount"`
	MethodOfPayment string  `json:"method_of_payment"`
}

func (Transaction) TableName() string {
	return "transactions"
}

func (TransactionDetails) TableName() string {
	return "transaction_details"
}
