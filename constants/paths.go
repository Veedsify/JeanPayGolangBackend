package constants

// API Path constants
const (
	// Base API path
	APIBase = "/api"

	ProtectedBase = "/protected"

	// Auth paths
	AuthBase                = "/auth"
	AuthLogin               = "/login"
	AuthVerify              = "/verify"
	AuthPasswordReset       = "/password-reset"
	AuthResetPassWordVerify = "/reset-password-verify"
	AuthSignup              = "/register"
	AuthLogout              = "/logout"
	AuthResetPassword       = "/reset-password"
	AuthVerifyOtp           = "/verify-otp"

	// User paths
	UserBase     = "/user"
	UserRetrieve = "/retrieve"

	// Users paths
	UsersBase          = "/users"
	UsersGet           = "/all"
	UsersAdd           = "/add"
	UsersUpdate        = "/update"
	UsersDelete        = "/delete/:id"
	UsersProfile       = "/profile"
	UsersUpdateProfile = "/profile/update"

	// Dashboard paths
	DashboardBase     = "/dashboard"
	DashboardOverview = "/overview"
	DashboardStats    = "/stats"

	// Wallet paths
	WalletBase               = "/wallet"
	WalletBalance            = "/balance"
	WalletTopUp              = "/topup"
	WalletTopUpDetails       = "/topup/:id"
	WalletWithdraw           = "/withdraw"
	WalletHistory            = "/history"
	WalletUpdateAfterPayment = "/update-after-payment"

	// Convert paths
	ConvertBase      = "/convert"
	ConvertExchange  = "/exchange"
	ConvertRates     = "/rates"
	ConvertCalculate = "/calculate"

	// Transaction paths
	TransactionsBase         = "/transactions"
	TransactionsAll          = "/all"
	TransactionsUserHistory  = "/history"
	TransactionsDetails      = "/details/:id"
	TransactionsUpdateStatus = "/status/:id"
	TransactionsNew          = "/new"

	// Notification paths
	NotificationsBase        = "/notifications"
	NotificationsAll         = "/all"
	NotificationsMarkRead    = "/mark-read/:id"
	NotificationsMarkAllRead = "/mark-all-read"

	// Settings paths
	SettingsBase             = "/settings"
	SettingsUpdate           = "/update"
	SettingsProfilePicture   = "/profile-picture"
	SettingsWallet           = "/wallet"
	SettingsChangePassword   = "/change-password"
	SettingsPreferences      = "/preferences"
	SettingsProfile          = "/profile"
	SettingsSecurity         = "/security"
	SettingsNotifications    = "/notifications"
	SettingsTwoFactor        = "/securitytwo-factor/qr"
	SettingsTwoFactorEnable  = "/security/two-factor/enable"
	SettingsTwoFactorDisable = "/security/two-factor/disable"

	// Admin paths
	AdminBase      = "/admin"
	AdminLogin     = "/login"
	AdminDashboard = "/dashboard"

	// Admin transaction paths
	AdminTransactionsBase     = "/transactions"
	AdminTransactionsAll      = "/all"
	AdminTransactionsDetails  = "/details/:id"
	AdminTransactionsApprove  = "/approve/:id"
	AdminTransactionsReject   = "/reject/:id"
	AdminTransactionsStatus   = "/status/:id"
	AdminTransactionsOverview = "/overview"

	// Admin user paths
	AdminUsersBase    = "/users"
	AdminUsersAll     = "/all"
	AdminUsersDetails = "/details/:id"
	AdminUsersBlock   = "/block/:id"
	AdminUsersUnblock = "/unblock/:id"
	AdminUsersDelete  = "/delete/:id"

	// Admin rates paths
	AdminRatesBase    = "/rates"
	AdminRatesGet     = "/get"
	AdminRatesAdd     = "/add"
	AdminRatesUpdate  = "/update"
	AdminRatesHistory = "/history"

	// Admin logs paths
	AdminLogsBase          = "/logs"
	AdminLogsAll           = "/all"
	AdminLogsNotifications = "/notifications"
	AdminLogsAudit         = "/audit"

	// Webhook paths
	WebhooksBase     = "/webhooks"
	WebhooksPaystack = "/paystack"
	WebhooksMomo     = "/momo"
)

// GetFullPath combines base API path with specific path
func GetFullPath(path string) string {
	return APIBase + path
}

// Auth full paths
func GetAuthPath(subPath string) string {
	return APIBase + AuthBase + subPath
}

// User full paths
func GetUserPath(subPath string) string {
	return APIBase + UsersBase + subPath
}

// Dashboard full paths
func GetDashboardPath(subPath string) string {
	return APIBase + DashboardBase + subPath
}

// Wallet full paths
func GetWalletPath(subPath string) string {
	return APIBase + WalletBase + subPath
}

// Convert full paths
func GetConvertPath(subPath string) string {
	return APIBase + ConvertBase + subPath
}

// Transaction full paths
func GetTransactionPath(subPath string) string {
	return APIBase + TransactionsBase + subPath
}

// Notification full paths
func GetNotificationPath(subPath string) string {
	return APIBase + NotificationsBase + subPath
}

// Settings full paths
func GetSettingsPath(subPath string) string {
	return APIBase + SettingsBase + subPath
}

// Admin full paths
func GetAdminPath(subPath string) string {
	return APIBase + AdminBase + subPath
}

// Webhook full paths
func GetWebhookPath(subPath string) string {
	return APIBase + WebhooksBase + subPath
}
