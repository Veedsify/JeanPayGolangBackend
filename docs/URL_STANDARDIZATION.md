# URL Standardization Documentation

## Overview

This document outlines the standardization of URL handling across the JeanPay email system to ensure consistency and proper routing of frontend links.

## Changes Made

### Problem
Previously, the email system had inconsistent URL handling:
- Some templates used `ServerURL` pointing to backend server
- Some used `AppUrl` pointing to frontend
- Mixed usage of different environment variables
- Inconsistent function names for environment variable access

### Solution
Standardized all email templates and services to use `FRONTEND_URL` environment variable consistently.

## Implementation Details

### Environment Variables
```env
FRONTEND_URL=https://app.jeanpay.com    # Frontend application URL
SERVER_URL=https://api.jeanpay.com      # Backend API URL (still used for API calls)
```

### EmailService Changes

#### Before
```go
// Mixed usage
data["ServerURL"] = SERVER                                    // Backend URL
data["AppUrl"] = FRONTEND                                     // Frontend URL
data["ServerURL"] = getEnvOrDefault("FRONTEND_URL", "...")    // Inconsistent function
```

#### After
```go
// Consistent usage
data["ServerURL"] = FRONTEND    // Always points to frontend for user-facing links
```

### Template Standardization

All email templates now consistently use `{{.ServerURL}}` which points to the frontend application:

#### Transaction Emails
- `{{.ServerURL}}/dashboard/transactions` - View transaction history
- `{{.ServerURL}}/dashboard/{{.ActionPath}}` - Retry actions (topup, withdraw, transfer, etc.)
- `{{.ServerURL}}/dashboard` - Main dashboard

#### Authentication Emails
- `{{.ServerURL}}/password/reset?token={{.ResetToken}}` - Password reset
- `{{.ServerURL}}/activate?token={{.Token}}` - Account activation

#### Support Links
- `{{.ServerURL}}/support` - Contact support
- `{{.ServerURL}}/help` - Help center
- `{{.ServerURL}}/security` - Security center

### Files Modified

1. **services/EmailService.go**
   - Standardized to use `FRONTEND` variable for all `ServerURL` template data
   - Removed duplicate `getEnvOrDefault` function
   - Fixed password reset email to use consistent URL structure

2. **templates/PasswordReset.go**
   - Changed from `{{.AppUrl}}` to `{{.ServerURL}}`
   - Standardized reset URL path to `/password/reset`

3. **templates/TransactionApproved.go**
   - Already using `{{.ServerURL}}` correctly

4. **templates/TransactionRejected.go**
   - Already using `{{.ServerURL}}` correctly

## URL Routing Structure

### Frontend Routes (FRONTEND_URL)
All user-facing actions in emails should point to frontend:
```
/dashboard                      - Main user dashboard
/dashboard/transactions         - Transaction history
/dashboard/topup               - Topup/deposit page
/dashboard/withdraw            - Withdrawal page
/dashboard/transfer            - Transfer page
/password/reset                - Password reset page
/activate                      - Account activation
/support                       - Support/contact page
/help                         - Help center
/security                     - Security settings
```

### Backend Routes (SERVER_URL)
API calls and webhooks (not used in email templates):
```
/api/auth/*                    - Authentication endpoints
/api/transactions/*            - Transaction API
/api/admin/*                   - Admin API
/webhooks/*                    - Payment webhooks
```

## Benefits

### For Users
- **Consistent Experience**: All email links lead to the correct frontend pages
- **Proper Routing**: No more confusion between backend and frontend URLs
- **Mobile Compatibility**: Frontend routes work properly on mobile devices

### For Developers
- **Maintainability**: Single source of truth for frontend URLs
- **Debugging**: Easier to trace URL-related issues
- **Scalability**: Easy to change frontend domain without updating multiple places

### For Deployment
- **Environment Flexibility**: Easy to configure different URLs for dev/staging/prod
- **CDN Support**: Frontend URLs can point to CDN or static hosting
- **Load Balancing**: Backend and frontend can be scaled independently

## Configuration Examples

### Development
```env
FRONTEND_URL=http://localhost:3000
SERVER_URL=http://localhost:8080
```

### Staging
```env
FRONTEND_URL=https://staging.jeanpay.com
SERVER_URL=https://api-staging.jeanpay.com
```

### Production
```env
FRONTEND_URL=https://app.jeanpay.com
SERVER_URL=https://api.jeanpay.com
```

## Testing

### Email Template Testing
To verify URL standardization:

1. **Send Test Emails**: Trigger transaction approval/rejection emails
2. **Check Links**: Verify all links point to frontend domain
3. **Verify Paths**: Ensure all paths route to correct frontend pages
4. **Cross-Environment**: Test in dev/staging environments

### URL Validation
```bash
# Check environment variables are set correctly
echo $FRONTEND_URL
echo $SERVER_URL

# Verify frontend routes are accessible
curl -I $FRONTEND_URL/dashboard
curl -I $FRONTEND_URL/password/reset
```

## Troubleshooting

### Common Issues

**Links not working in emails**:
- Verify `FRONTEND_URL` environment variable is set
- Check that frontend routes exist and are accessible
- Ensure SSL certificates are valid for production domains

**Mixed HTTP/HTTPS**:
- Use HTTPS for production `FRONTEND_URL`
- Ensure frontend application handles HTTPS properly
- Check for redirect loops between HTTP/HTTPS

**Wrong domain in emails**:
- Verify `FRONTEND_URL` is correctly set in deployment
- Restart email workers after environment changes
- Check email template compilation

### Debugging Commands

```bash
# Check email template rendering
go test ./services -v -run TestEmailTemplates

# Verify environment variables
printenv | grep -E "(FRONTEND_URL|SERVER_URL)"

# Test frontend routes
curl -f $FRONTEND_URL/dashboard || echo "Frontend route not accessible"
```

## Future Considerations

1. **Multi-tenant Support**: Support different frontend URLs per tenant
2. **A/B Testing**: Configure different frontend URLs for testing
3. **Internationalization**: Language-specific frontend URLs
4. **White-labeling**: Custom domains for different brands