package templates

import "fmt"

func TransactionRejectedTemplate() string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.TransactionTypeDisplay}} Rejected - JeanPay</title>
    <style>%s</style>
</head>
<body>
    <div class="email-wrapper">
        <div class="header">
            <div class="logo">
                <img src="https://res.cloudinary.com/ds2hdlfvc/image/upload/v1755948663/logo_nf44qm.png" alt="JeanPay Logo" />
            </div>
            <h1>{{.TransactionTypeDisplay}} Rejected</h1>
            <p>Unfortunately, your {{.TransactionTypeDisplay}} could not be processed</p>
        </div>
        <div class="content">
            <div class="greeting">Hello {{.UserName}}! 👋</div>
            <div class="message">
                We're sorry to inform you that your {{.TransactionTypeDisplay}} request could not be processed at this time. {{.StatusMessage}}
            </div>
            <div class="card">
                <h3>Transaction Details</h3>
                <ul>
                    <li><strong>Type:</strong> {{.TransactionTypeDisplay}}</li>
                    <li><strong>Amount:</strong> {{.Amount}}</li>
                    <li><strong>Transaction ID:</strong> <span style="font-family: 'Monaco', 'Menlo', monospace; background-color: #f1f5f9; padding: 2px 4px; border-radius: 4px;">{{.TransactionID}}</span></li>
                    <li><strong>Date & Time:</strong> {{.Date}}</li>
                    <li><strong>Status:</strong> <span style="color: #dc2626;">❌ Rejected</span></li>
                </ul>
            </div>
            {{if .Reason}}
            <div class="card" style="border-left: 4px solid #dc2626; background-color: #fef2f2;">
                <h3 style="color: #dc2626;">Reason for Rejection</h3>
                <p style="color: #7f1d1d; margin: 0;">{{.Reason}}</p>
            </div>
            {{end}}
            {{if .ShowRecipientDetails}}
            <div class="card">
                <h3>{{.RecipientLabel}} Details</h3>
                <ul>
                    <li><strong>Name:</strong> {{.RecipientName}}</li>
                    {{if .BankName}}<li><strong>Bank:</strong> {{.BankName}}</li>{{end}}
                    {{if .AccountNumber}}<li><strong>Account Number:</strong> {{.AccountNumber}}</li>{{end}}
                    {{if .PhoneNumber}}<li><strong>Phone Number:</strong> {{.PhoneNumber}}</li>{{end}}
                    {{if .Network}}<li><strong>Network:</strong> {{.Network}}</li>{{end}}
                </ul>
            </div>
            {{end}}
            {{if .ShowCurrencyDetails}}
            <div class="card">
                <h3>Currency Details</h3>
                <ul>
                    <li><strong>From:</strong> {{.FromAmount}} {{.FromCurrency}}</li>
                    <li><strong>To:</strong> {{.ToAmount}} {{.ToCurrency}}</li>
                    <li><strong>Exchange Rate:</strong> {{.ExchangeRate}}</li>
                </ul>
            </div>
            {{end}}
            <div class="highlight" style="background-color: #fef3c7; border-left: 4px solid #f59e0b;">
                <p style="color: #92400e;"><strong>{{.HighlightMessage}}</strong> {{.HighlightDescription}}</p>
            </div>
            <div class="cta-section">
                <a href="{{.ServerURL}}/dashboard/{{.ActionPath}}" class="cta-button">
                    {{.ActionButtonText}}
                </a>
            </div>
            <div class="divider"></div>
            <div class="message">
                {{.SupportMessage}} Our support team is available 24/7 to help you resolve this issue and process your transaction successfully.
            </div>
        </div>
        <div class="footer">
            <div class="footer-logo">JeanPay</div>
            <div class="footer-text">We're here to help you succeed</div>
            <div class="footer-text">This email was sent to {{.Email}}</div>
            <div class="footer-links">
                <a href="{{.ServerURL}}/dashboard" class="footer-link">Dashboard</a>
                <a href="{{.ServerURL}}/support" class="footer-link">Contact Support</a>
                <a href="{{.ServerURL}}/help" class="footer-link">Help Center</a>
            </div>
        </div>
    </div>
</body>
</html>`, BaseCss)
}

func TransactionRejectedPlainTextTemplate() string {
	return `❌ {{.TransactionTypeDisplay}} Rejected - JeanPay
Hello {{.UserName}}!

We're sorry to inform you that your {{.TransactionTypeDisplay}} request could not be processed at this time. {{.StatusMessage}}

Transaction Details:
📋 Type: {{.TransactionTypeDisplay}}
💰 Amount: {{.Amount}}
🔖 Transaction ID: {{.TransactionID}}
📅 Date & Time: {{.Date}}
❌ Status: Rejected

{{if .Reason}}Reason for Rejection:
⚠️  {{.Reason}}

{{end}}{{if .ShowRecipientDetails}}{{.RecipientLabel}} Details:
👤 Name: {{.RecipientName}}
{{if .BankName}}🏦 Bank: {{.BankName}}{{end}}
{{if .AccountNumber}}🔢 Account Number: {{.AccountNumber}}{{end}}
{{if .PhoneNumber}}📱 Phone Number: {{.PhoneNumber}}{{end}}
{{if .Network}}📡 Network: {{.Network}}{{end}}

{{end}}{{if .ShowCurrencyDetails}}Currency Details:
💱 From: {{.FromAmount}} {{.FromCurrency}}
💱 To: {{.ToAmount}} {{.ToCurrency}}
📊 Exchange Rate: {{.ExchangeRate}}

{{end}}⚠️  {{.HighlightMessage}} {{.HighlightDescription}}

{{.SupportMessage}} Our support team is available 24/7 to help you resolve this issue and process your transaction successfully.

{{.ActionButtonText}}: {{.ServerURL}}/dashboard/{{.ActionPath}}

Best regards,
The JeanPay Support Team

---
This email was sent to {{.Email}}
We're here to help you succeed.`
}

// Helper functions to generate dynamic content based on transaction type for rejected transactions

func GetRejectedTransactionData(transactionType string, transaction interface{}) map[string]interface{} {
	baseData := map[string]interface{}{
		"ShowRecipientDetails": false,
		"ShowCurrencyDetails":  false,
		"RecipientLabel":       "Recipient",
		"HighlightMessage":     "What happens next?",
		"SupportMessage":       "Need help understanding why this transaction was rejected?",
		"ActionPath":           "transactions",
		"ActionButtonText":     "View Transaction History",
	}

	switch transactionType {
	case "deposit", "topup":
		baseData["TransactionTypeDisplay"] = "Deposit"
		baseData["StatusMessage"] = "The funds were not added to your wallet due to verification issues."
		baseData["HighlightMessage"] = "You can try again."
		baseData["HighlightDescription"] = "Please review your payment method details and ensure all information is correct before submitting a new deposit request."
		baseData["SupportMessage"] = "Having trouble with your deposit?"
		baseData["ActionPath"] = "topup"
		baseData["ActionButtonText"] = "Try Deposit Again"

	case "withdrawal":
		baseData["TransactionTypeDisplay"] = "Withdrawal"
		baseData["StatusMessage"] = "Your withdrawal request could not be completed."
		baseData["ShowRecipientDetails"] = true
		baseData["RecipientLabel"] = "Withdrawal Account"
		baseData["HighlightMessage"] = "Your account balance remains unchanged."
		baseData["HighlightDescription"] = "Please verify your withdrawal details and ensure you have sufficient funds before trying again."
		baseData["SupportMessage"] = "Need help with your withdrawal?"
		baseData["ActionPath"] = "withdraw"
		baseData["ActionButtonText"] = "Try Withdrawal Again"

	case "transfer":
		baseData["TransactionTypeDisplay"] = "Transfer"
		baseData["StatusMessage"] = "Your money transfer could not be processed."
		baseData["ShowRecipientDetails"] = true
		baseData["RecipientLabel"] = "Recipient"
		baseData["HighlightMessage"] = "No funds have been deducted."
		baseData["HighlightDescription"] = "Please verify the recipient details and your account balance before attempting another transfer."
		baseData["SupportMessage"] = "Need assistance with your transfer?"
		baseData["ActionPath"] = "transfer"
		baseData["ActionButtonText"] = "Try Transfer Again"

	case "conversion":
		baseData["TransactionTypeDisplay"] = "Currency Conversion"
		baseData["StatusMessage"] = "Your currency conversion request was not successful."
		baseData["ShowCurrencyDetails"] = true
		baseData["HighlightMessage"] = "Your original currency balance is intact."
		baseData["HighlightDescription"] = "Please check current exchange rates and your available balance before trying again."
		baseData["SupportMessage"] = "Questions about currency conversion?"
		baseData["ActionPath"] = "transactions"
		baseData["ActionButtonText"] = "View Transaction History"

	default:
		baseData["TransactionTypeDisplay"] = "Transaction"
		baseData["StatusMessage"] = "Your transaction request could not be processed."
		baseData["HighlightMessage"] = "No changes were made to your account."
		baseData["HighlightDescription"] = "Please review the transaction details and try again."
	}

	return baseData
}

// Common rejection reasons and their user-friendly messages
func GetUserFriendlyRejectionReason(reason string, transactionType string) string {
	commonReasons := map[string]string{
		"insufficient_funds":       "Insufficient funds in your account to complete this transaction.",
		"invalid_recipient":        "The recipient details provided are invalid or incomplete.",
		"account_verification":     "Your account requires additional verification to process this transaction.",
		"daily_limit_exceeded":     "This transaction exceeds your daily transaction limit.",
		"suspicious_activity":      "This transaction was flagged by our security system for review.",
		"invalid_payment_method":   "The payment method provided is invalid or expired.",
		"recipient_account_closed": "The recipient's account is closed or inactive.",
		"network_error":            "A network error occurred while processing your transaction.",
		"compliance_check_failed":  "This transaction did not pass our compliance verification.",
		"currency_not_supported":   "The currency conversion is not currently supported.",
	}

	if friendlyReason, exists := commonReasons[reason]; exists {
		return friendlyReason
	}

	// If no specific reason found, return the original reason
	if reason != "" {
		return reason
	}

	// Default message based on transaction type
	switch transactionType {
	case "deposit", "topup":
		return "Unable to verify the payment method or source of funds."
	case "withdrawal":
		return "Unable to process withdrawal to the specified account."
	case "transfer":
		return "Unable to complete the transfer to the specified recipient."
	case "conversion":
		return "Unable to process the currency conversion at this time."
	default:
		return "The transaction could not be processed due to verification issues."
	}
}
