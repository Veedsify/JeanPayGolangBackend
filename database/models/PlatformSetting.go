package models

type ChartStyle string

const (
	LineChart ChartStyle = "line"
	BarChart  ChartStyle = "bar"
)

type PlatformSetting struct {
	ID                           uint            `json:"id" gorm:"primaryKey"`
	KycEnforcement               bool            `json:"kyc_enforcement" gorm:"default:false"`
	ManualRateOverride           bool            `json:"manual_rate_override" gorm:"default:false"`
	TransactionConfirmationEmail bool            `json:"transaction_confirmation_email" gorm:"default:true"`
	DefaultCurrencyDisplay       DefaultCurrency `json:"default_currency_display" gorm:"default:'NGN'"`
	MinimumTransactionAmount     float64         `json:"minimum_transaction_amount" gorm:"default:100.00"`
	MaximumTransactionAmount     float64         `json:"maximum_transaction_amount" gorm:"default:1000000.00"`
	DailyTransactionLimit        float64         `json:"daily_transaction_limit" gorm:"default:500000.00"`
	ChartStyle                   ChartStyle      `json:"chart_style" gorm:"default:'line'"`
	SendTransactionSuccessEmail  bool            `json:"send_transaction_success_email" gorm:"default:true"`
	SendTransactionDeclineEmail  bool            `json:"send_transaction_decline_email" gorm:"default:true"`
	SendTransactionPendingEmail  bool            `json:"send_transaction_pending_email" gorm:"default:true"`
	SendTransactionRefundEmail   bool            `json:"send_transaction_refund_email" gorm:"default:true"`
	AccountLimitsNotification    bool            `json:"account_limits_notification" gorm:"default:true"`
}

func (PlatformSetting) TableName() string {
	return "platform_settings"
}
