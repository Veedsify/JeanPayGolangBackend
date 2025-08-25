package templates

import "fmt"

func WelcomeTemplate() string {
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
            <p>Your premium payment experience starts here</p>
        </div>
        <div class="content">
            <div class="greeting">Hello {{.UserName}}! 👋</div>
            <div class="message">
                We're absolutely thrilled to welcome you to the JeanPay family! You've just joined thousands of users who trust us with their most important financial transactions.
            </div>
            <div class="card">
                <h3>🚀 What You Can Do</h3>
                <ul>
                    <li>Send and receive money instantly with bank-grade security</li>
                    <li>Track all your transactions in real-time with detailed insights</li>
                    <li>Manage multiple wallets and currencies effortlessly</li>
                    <li>Access premium customer support 24/7</li>
                </ul>
            </div>
            <div class="highlight">
                <p><strong>Your security is our priority.</strong> We use advanced encryption and multi-factor authentication to keep your account safe.</p>
            </div>
            <div class="cta-section">
                <a href="{{.ServerURL}}/activate?token={{.Token}}" class="cta-button">
                    Activate Your Account Now
                </a>
            </div>
            <div class="divider"></div>
            <div class="message">
                Need help getting started? Our support team is here to assist you every step of the way. Simply reply to this email or visit our help center.
            </div>
        </div>
        <div class="footer">
            <div class="footer-logo">JeanPay</div>
            <div class="footer-text">Premium payments made simple</div>
            <div class="footer-text">This email was sent to {{.Email}}</div>
            <div class="footer-text">If you didn't create this account, please ignore this email.</div>
            <div class="footer-links">
                <a href="{{.ServerURL}}/help" class="footer-link">Help Center</a>
                <a href="{{.ServerURL}}/privacy" class="footer-link">Privacy Policy</a>
                <a href="{{.ServerURL}}/terms" class="footer-link">Terms of Service</a>
            </div>
        </div>
    </div>
</body>
</html>`, BaseCss)
}

func WelcomePlainTextTemplate() string {
	return `🎉 Welcome to JeanPay - Your Premium Payment Experience Awaits!
Hello {{.UserName}}!

We're absolutely thrilled to welcome you to the JeanPay family! You've just joined thousands of users who trust us with their most important financial transactions.

🚀 What You Can Do:
• Send and receive money instantly with bank-grade security
• Track all your transactions in real-time with detailed insights
• Manage multiple wallets and currencies effortlessly
• Access premium customer support 24/7

🔒 Your security is our priority. We use advanced encryption and multi-factor authentication to keep your account safe.

To activate your account, please visit: {{.ServerURL}}/activate?token={{.Token}}

Need help getting started? Our support team is here to assist you every step of the way. Simply reply to this email or visit our help center.

Best regards,
The JeanPay Team

---
This email was sent to {{.Email}}
If you didn't create this account, please ignore this email.`
}
