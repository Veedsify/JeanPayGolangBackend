package templates

import "fmt"

// NewDepositAdminTemplate returns the HTML template for new deposit admin notifications
func NewDepositAdminTemplate() string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>New Deposit Notification</title>
    <style>%s</style>
</head>
<body>
    <div class="email-wrapper">
        <div class="header">
            <div class="logo">
                <img src="https://res.cloudinary.com/ds2hdlfvc/image/upload/v1755948663/logo_nf44qm.png" alt="JeanPay Logo" />
            </div>
            <h1>💰 New Deposit Alert</h1>
        </div>

        <div class="content">
            <div class="greeting">Hello Admin!</div>

            <div class="message">
                A new <strong>deposit</strong> transaction has been initiated and requires your approval.
            </div>

            <div class="card">
                <h3>Deposit Details</h3>
                <ul>
                    <li><strong>Transaction ID:</strong> {{.TransactionID}}</li>
                    <li><strong>User:</strong> {{.UserName}} ({{.UserEmail}})</li>
                    <li><strong>Amount:</strong> {{.Amount}}</li>
                    <li><strong>Currency:</strong> {{.Currency}}</li>
                    <li><strong>Date:</strong> {{.Date}}</li>
                    {{if .Description}}<li><strong>Description:</strong> {{.Description}}</li>{{end}}
                </ul>
            </div>

            <div class="cta-section">
                <a href="{{.ServerURL}}/transactions/{{.TransactionID}}" class="cta-button">Review Deposit</a>
            </div>

            <div class="message">
                Please review this deposit transaction for approval to ensure smooth processing.
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

// NewDepositAdminPlainTextTemplate returns the plain text template for new deposit admin notifications
func NewDepositAdminPlainTextTemplate() string {
	return `
💰 NEW DEPOSIT ALERT

Hello Admin!

A new deposit transaction has been initiated and requires your approval.

DEPOSIT DETAILS
===============
Transaction ID: {{.TransactionID}}
User: {{.UserName}} ({{.UserEmail}})
Amount: {{.Amount}}
Currency: {{.Currency}}
Date: {{.Date}}
{{if .Description}}Description: {{.Description}}{{end}}

Please review this deposit transaction for approval to ensure smooth processing.

Review Deposit: {{.ServerURL}}/admin/transactions/{{.TransactionID}}

Best regards,
The JeanPay Team

---
This is an automated notification from JeanPay.
`
}
