package templates

import "fmt"

func GeneralEmailTemplate() string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Welcome to JeanPay</title>
    <style>%s</style>
</head>
<body>
    <div class="email-wrapper">
        <div class="header">
            <div class="logo">
                <img src="https://res.cloudinary.com/ds2hdlfvc/image/upload/v1755948663/logo_nf44qm.png" alt="JeanPay Logo" />
            </div>
            <h1>Welcome to JeanPay</h1>
        </div>
        <div class="content">
            <div class="greeting">Hello Admin! 👋</div>
            <div class="message">
                 The're is a new contact request on Jeanpay
            </div>
            <div class="card">
                <h3>🚀 Details</h3>
                <ul>
                    <li>Fullname: {{.FullName}}</li>
                    <li>Email: {{.ContactEmail}}</li>
                    <li>Subject: {{.Subject}}</li>
                    <li>Category: {{.Category}}</li>
                    <li>Message: {{.Message}}</li>

                </ul>
            </div>
            <div class="divider"></div>
        </div>
        <div class="footer">
            <div class="footer-logo">JeanPay</div>
            <div class="footer-text">This email was sent to {{.Email}}</div>
            <div class="footer-text">If you didn't create this account, please ignore this email.</div>
            <div class="footer-links">
                <a href="{{.FrontendURL}}/help" class="footer-link">Help Center</a>
                <a href="{{.FrontendURL}}/privacy" class="footer-link">Privacy Policy</a>
                <a href="{{.FrontendURL}}/terms" class="footer-link">Terms of Service</a>
            </div>
        </div>
    </div>
</body>
</html>`, BaseCss)
}

func GeneralEmailPlainTextTemplate() string {
	return `New Contact Request On JeanPay
Hello Admin!

The're is a new contact request on Jeanpay

New Contact Request On Jeanpay
Fullname: {{.FullName}}
Email: {{.Email}}
Subject: {{.Subject}}
Category: {{.Category}}
Message: {{.Message}}

Best regards,
The JeanPay Team

---
This email was sent to the Admin
If you didn't create this account, please ignore this email.`
}
