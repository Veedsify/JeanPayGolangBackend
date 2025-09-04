package templates

import "fmt"

// NewTransactionAdminTemplate returns the HTML template for new transaction admin notifications
func NewTransactionAdminTemplate() string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>New Transaction Notification</title>
    <style>%s</style>
</head>
<body>
    <div class="email-wrapper">
        <div class="header">
            <div class="logo">
                <img src="https://res.cloudinary.com/ds2hdlfvc/image/upload/v1755948663/logo_nf44qm.png" alt="JeanPay Logo" />
            </div>
            <h1>🔔 New Transaction Alert</h1>
        </div>

        <div class="content">
            <div class="greeting">Hello Admin!</div>

            <div class="message">
                A new <strong>{{.TransactionType}}</strong> transaction has been initiated and requires your attention.
            </div>

            <div class="card">
                <h3>Transaction Details</h3>
                <ul>
                    <li><strong>Transaction ID:</strong> {{.TransactionID}}</li>
                    <li><strong>User:</strong> {{.UserName}} ({{.UserEmail}})</li>
                    <li><strong>Amount:</strong> {{.Amount}}</li>
                    <li><strong>Type:</strong> {{.TransactionType}}</li>
                    <li><strong>Date:</strong> {{.Date}}</li>
                    {{if .RecipientName}}<li><strong>Recipient:</strong> {{.RecipientName}}</li>{{end}}
                    {{if .BankName}}<li><strong>Bank:</strong> {{.BankName}}</li>{{end}}
                    {{if .AccountNumber}}<li><strong>Account:</strong> {{.AccountNumber}}</li>{{end}}
                </ul>
            </div>

            <div class="cta-section">
                <a href="{{.ServerURL}}/transactions/{{.TransactionID}}" class="cta-button">Review Transaction</a>
            </div>

            <div class="message">
                Please review this transaction promptly to ensure smooth processing.
            </div>
        </div>

        <div class="footer">
            <div class="footer-logo">JeanPay</div>
            <div class="footer-text">Admin Notification System</div>
            <div class="footer-text">This is an automated notification from JeanPay.</div>
        </div>
    </div>
</body>
</html>`, BaseCss)
}

// NewTransactionAdminPlainTextTemplate returns the plain text template for new transaction admin notifications
func NewTransactionAdminPlainTextTemplate() string {
	return `
🔔 NEW TRANSACTION ALERT

Hello Admin!

A new {{.TransactionType}} transaction has been initiated and requires your attention.

TRANSACTION DETAILS
==================
Transaction ID: {{.TransactionID}}
User: {{.UserName}} ({{.UserEmail}})
Amount: {{.Amount}}
Type: {{.TransactionType}}
Date: {{.Date}}
{{if .RecipientName}}Recipient: {{.RecipientName}}{{end}}
{{if .BankName}}Bank: {{.BankName}}{{end}}
{{if .AccountNumber}}Account: {{.AccountNumber}}{{end}}

Please review this transaction promptly to ensure smooth processing.

Review Transaction: {{.ServerURL}}/admin/transactions/{{.TransactionID}}

Best regards,
The JeanPay Team

---
This is an automated notification from JeanPay.
`
}
