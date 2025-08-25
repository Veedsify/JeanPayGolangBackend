package templates

import "fmt"

func TwoFactorAuthenticationTemplate() string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Two-Factor Authentication</title>
    <style>%s</style>
</head>
<body>
    <div class="email-wrapper">
        <div class="header">
            <div class="logo">
                <img src="https://res.cloudinary.com/ds2hdlfvc/image/upload/v1755948663/logo_nf44qm.png" alt="JeanPay Logo" />
            </div>
            <h1>Security Verification</h1>
            <p>Your account security is our priority</p>
        </div>
        <div class="content">
            <div class="greeting">Hello {{.UserName}}! 👋</div>
            <div class="message">
                We've detected a login attempt to your JeanPay account. To ensure your security, please use the verification code below to complete your login.
            </div>
            <div class="code-section">
                <div class="code-label">Your Verification Code</div>
                <div class="verification-code">{{.VerificationCode}}</div>
                <div class="message">Enter this code in your JeanPay app or browser</div>
            </div>
            <div class="highlight">
                <p><strong>Security Notice:</strong> If you didn't attempt to log in to your JeanPay account, please ignore this email and consider changing your password immediately. Never share this code with anyone.</p>
            </div>
            <div class="divider"></div>
            <div class="message">
                <strong>Need Help?</strong> If you're having trouble logging in or didn't request this code, please contact our security team immediately. We're here to help keep your account safe.
            </div>
        </div>
        <div class="footer">
            <div class="footer-logo">JeanPay</div>
            <div class="footer-text">Protecting your financial security</div>
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

func TwoFactorAuthenticationPlainTemplate() string {
	return `🔐 Your JeanPay Security Code - Expires in 10 Minutes
Hello {{.UserName}}!

We've detected a login attempt to your JeanPay account. To ensure your security, please use the verification code below to complete your login.

Your Verification Code: {{.VerificationCode}}

Enter this code in your JeanPay app or browser.

🛡️ Security Notice: If you didn't attempt to log in to your JeanPay account, please ignore this email and consider changing your password immediately. Never share this code with anyone.

Need Help? If you're having trouble logging in or didn't request this code, please contact our security team immediately. We're here to help keep your account safe.

Best regards,
The JeanPay Security Team

---
This email was sent to {{.Email}}
Protecting your financial security.`
}
