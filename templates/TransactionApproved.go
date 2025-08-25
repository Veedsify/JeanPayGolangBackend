package templates

import "fmt"

func TransactionApprovedTemplate() string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.TransactionTypeDisplay}} Approved - JeanPay</title>
    <style>%s</style>
</head>
<body>
    <div class="email-wrapper">
        <div class="header">
            <div class="logo">
                <img src="https://res.cloudinary.com/ds2hdlfvc/image/upload/v1755948663/logo_nf44qm.png" alt="JeanPay Logo" />
            </div>
            <h1>{{.TransactionTypeDisplay}} Approved</h1>
            <p>Your {{.TransactionTypeDisplay}} has been processed successfully</p>
        </div>
        <div class="content">
            <div class="greeting">Hello {{.UserName}}! 👋</div>
            <div class="message">
             	{{.StatusMessage}}
            </div>
            <div class="card">
                <h3>Transaction Details</h3>
                <ul>
                    <li><strong>Type:</strong> {{.TransactionTypeDisplay}}</li>
                    <li><strong>Amount:</strong> {{.Amount}}</li>
                    <li><strong>Transaction ID:</strong> <span style="font-family: 'Monaco', 'Menlo', monospace; background-color: #f1f5f9; padding: 2px 4px; border-radius: 4px;">{{.TransactionID}}</span></li>
                    <li><strong>Date & Time:</strong> {{.Date}}</li>
                    <li><strong>Status:</strong> <span style="color: #16a34a;">✅ Approved & Completed</span></li>
                </ul>
            </div>
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
            <div class="highlight">
                <p><strong>{{.HighlightMessage}}</strong> {{.HighlightDescription}}</p>
            </div>
            <div class="cta-section">
                <a href="{{.ServerURL}}/dashboard/transactions" class="cta-button">
                    View Transaction History
                </a>
            </div>
            <div class="divider"></div>
            <div class="message">
                {{.SupportMessage}} Our support team is available 24/7 to assist you with any questions.
            </div>
        </div>
        <div class="footer">
            <div class="footer-logo">JeanPay</div>
            <div class="footer-text">Secure payments made simple</div>
            <div class="footer-text">This email was sent to {{.Email}}</div>
            <div class="footer-links">
                <a href="{{.ServerURL}}/dashboard" class="footer-link">Dashboard</a>
                <a href="{{.ServerURL}}/support" class="footer-link">Support</a>
                <a href="{{.ServerURL}}/help" class="footer-link">Help Center</a>
            </div>
        </div>
    </div>
</body>
</html>`, BaseCss)
}

func TransactionApprovedPlainTextTemplate() string {
	return `✅ {{.TransactionTypeDisplay}} Approved - JeanPay
Hello {{.UserName}}!

Great news! Your {{.TransactionTypeDisplay}} request has been approved and processed successfully. {{.StatusMessage}}

Transaction Details:
📋 Type: {{.TransactionTypeDisplay}}
💰 Amount: {{.Amount}}
🔖 Transaction ID: {{.TransactionID}}
📅 Date & Time: {{.Date}}
✅ Status: Approved & Completed

{{if .ShowRecipientDetails}}{{.RecipientLabel}} Details:
👤 Name: {{.RecipientName}}
{{if .BankName}}🏦 Bank: {{.BankName}}{{end}}
{{if .AccountNumber}}🔢 Account Number: {{.AccountNumber}}{{end}}
{{if .PhoneNumber}}📱 Phone Number: {{.PhoneNumber}}{{end}}
{{if .Network}}📡 Network: {{.Network}}{{end}}

{{end}}{{if .ShowCurrencyDetails}}Currency Details:
💱 From: {{.FromAmount}} {{.FromCurrency}}
💱 To: {{.ToAmount}} {{.ToCurrency}}
📊 Exchange Rate: {{.ExchangeRate}}

{{end}}💡 {{.HighlightMessage}} {{.HighlightDescription}}

{{.SupportMessage}} Our support team is available 24/7 to assist you with any questions.

View your complete transaction history: {{.ServerURL}}/dashboard/transactions

Best regards,
The JeanPay Team

---
This email was sent to {{.Email}}
Secure payments made simple.`
}

// Helper functions to generate dynamic content based on transaction type

func GetApprovedTransactionData(transactionType string, transaction interface{}) map[string]interface{} {
	baseData := map[string]interface{}{
		"ShowRecipientDetails": false,
		"ShowCurrencyDetails":  false,
		"RecipientLabel":       "Recipient",
		"HighlightMessage":     "Transaction completed successfully.",
		"SupportMessage":       "Need help or have questions about this transaction?",
	}

	switch transactionType {
	case "deposit", "topup":
		baseData["TransactionTypeDisplay"] = "Deposit"
		baseData["StatusMessage"] = "The funds have been credited to your wallet and are now available for use."
		baseData["HighlightMessage"] = "Funds are now available in your wallet."
		baseData["HighlightDescription"] = "You can now use these funds for transfers, withdrawals, or currency conversions."
		baseData["SupportMessage"] = "Ready to make your next transaction?"

	case "withdrawal":
		baseData["TransactionTypeDisplay"] = "Withdrawal"
		baseData["StatusMessage"] = "Great News, Your funds have been successfully transferred to your specified account."
		baseData["ShowRecipientDetails"] = true
		baseData["RecipientLabel"] = "Withdrawal Account"
		baseData["HighlightMessage"] = "Funds have been sent to your account."
		baseData["HighlightDescription"] = "Please allow 1-3 business days for the funds to reflect in your account."
		baseData["SupportMessage"] = "If you don't see the funds within the expected timeframe, please contact us."

	case "transfer":
		baseData["TransactionTypeDisplay"] = "Transfer"
		baseData["StatusMessage"] = "Great news! Your Transfer request has been approved and processed successfully."
		baseData["ShowRecipientDetails"] = true
		baseData["RecipientLabel"] = "Recipient"
		baseData["HighlightMessage"] = "Transfer completed successfully."
		baseData["HighlightDescription"] = "The recipient should receive the funds within minutes to a few hours depending on their payment method."
		baseData["SupportMessage"] = "Need to make another transfer?"

	case "conversion":
		baseData["TransactionTypeDisplay"] = "Currency Conversion"
		baseData["StatusMessage"] = "Great news! Your Currency Conversion request has been approved and processed successfully.."
		baseData["ShowCurrencyDetails"] = true
		baseData["HighlightMessage"] = "Currency conversion completed."
		baseData["HighlightDescription"] = "The converted amount is now available in your wallet."
		baseData["SupportMessage"] = "Want to convert more currencies?"

	default:
		baseData["TransactionTypeDisplay"] = "Transaction"
		baseData["StatusMessage"] = "Your transaction has been processed successfully."
		baseData["HighlightMessage"] = "Transaction completed."
		baseData["HighlightDescription"] = "Your request has been processed successfully."
	}

	return baseData
}
