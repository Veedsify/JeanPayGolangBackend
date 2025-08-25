package templates

import "fmt"

func PasswordResetTemplate() string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Password Reset Request</title>
    <style>%s</style>
</head>
<body>
    <div class="email-wrapper">
        <div class="header">
            <div class="logo">
                <img src="https://res.cloudinary.com/ds2hdlfvc/image/upload/v1755948663/logo_nf44qm.png" alt="JeanPay Logo" />
            </div>
            <h1>Password Reset Request</h1>
            <p>Secure your account with a new password</p>
        </div>
        <div class="content">
            <div class="greeting">Password Reset Required 🔐</div>
            <div class="message">
                We received a request to reset the password for your JeanPay account. If this was you, click the button below to create a new secure password.
            </div>
            <div class="highlight">
                <p><strong>Security Notice:</strong> This password reset request was initiated from a new device or location. If this wasn't you, please contact our security team immediately.</p>
            </div>
            <div class="cta-section">
                <a href="{{.ServerURL}}/password/reset?token={{.ResetToken}}" class="cta-button">
                    Reset My Password Securely
                </a>
            </div>
            <div class="card">
                <h3>🛡️ Password Security Tips</h3>
                <ul>
                    <li>Use a combination of uppercase, lowercase, numbers, and symbols</li>
                    <li>Make it at least 12 characters long</li>
                    <li>Avoid using personal information or common words</li>
                    <li>Consider using a password manager for better security</li>
                </ul>
            </div>
            <div class="divider"></div>
            <div class="message">
                <strong>Didn't request this?</strong> If you didn't request a password reset, you can safely ignore this email. Your password will remain unchanged, and your account stays secure.
            </div>
        </div>
        <div class="footer">
            <div class="footer-logo">JeanPay</div>
            <div class="footer-text">Your security is our priority</div>
            <div class="footer-text">This email was sent to {{.Email}}</div>
            <div class="footer-links">
                <a href="{{.ServerURL}}/security" class="footer-link">Security Center</a>
                <a href="{{.ServerURL}}/support" class="footer-link">Contact Support</a>
                <a href="{{.ServerURL}}/help" class="footer-link">Help Center</a>
            </div>
        </div>
    </div>
</body>
</html>`, BaseCss)
}

func PasswordResetPlainTextTemplate() string {
	return `🔐 Secure Password Reset Request - JeanPay
Password Reset Required

We received a request to reset the password for your JeanPay account. If this was you, please visit the link below to create a new secure password.

Reset Link: {{.ServerURL}}/password/reset?token={{.ResetToken}}

⚠️ Security Notice: This password reset request was initiated from a new device or location. If this wasn't you, please contact our security team immediately.

🛡️ Password Security Tips:
• Use a combination of uppercase, lowercase, numbers, and symbols
• Make it at least 12 characters long
• Avoid using personal information or common words
• Consider using a password manager for better security

Didn't request this? If you didn't request a password reset, you can safely ignore this email. Your password will remain unchanged, and your account stays secure.

Best regards,
The JeanPay Security Team

---
This email was sent to {{.Email}}
Your security is our priority.`
}
