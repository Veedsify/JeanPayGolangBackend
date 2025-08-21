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
	Theme                        string          `json:"theme" gorm:"default:'light'"`
	EmailNotifications           bool            `json:"email_notifications" gorm:"default:true"`
	SMSNotifications             bool            `json:"sms_notifications" gorm:"default:false"`
	PushNotifications            bool            `json:"push_notifications" gorm:"default:true"`
	EnforceTwoFactor             bool            `json:"enforce_two_factor" gorm:"default:false"`
	SessionTimeoutMinutes        float64         `json:"session_timeout_minutes" gorm:"default:30"`
	PasswordExpiryDays           float64         `json:"password_expiry_days" gorm:"default:90"`
	SendTransactionSuccessEmail  bool            `json:"send_transaction_success_email" gorm:"default:true"`
	SendTransactionDeclineEmail  bool            `json:"send_transaction_decline_email" gorm:"default:true"`
	SendTransactionPendingEmail  bool            `json:"send_transaction_pending_email" gorm:"default:true"`
	SendTransactionRefundEmail   bool            `json:"send_transaction_refund_email" gorm:"default:true"`
	AccountLimitsNotification    bool            `json:"account_limits_notification" gorm:"default:true"`
}

func (PlatformSetting) TableName() string {
	return "platform_settings"
}
