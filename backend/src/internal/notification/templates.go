package notification

import (
	"fmt"
	"strings"
)

// FilingCompletedEmail generates the email content for when a filing is completed
type FilingCompletedEmail struct {
	ClientName  string
	TaxYear     int
	FilingType  string
	TenantName  string
	LoginURL    string
}

// PortalAccessEmail generates the email content for portal magic link
type PortalAccessEmail struct {
	ClientName string
	TenantName string
	PortalURL  string
}

// GenerateFilingCompletedEmail creates HTML and text versions of the filing completed email
func GenerateFilingCompletedEmail(data FilingCompletedEmail) (subject, htmlBody, textBody string) {
	subject = fmt.Sprintf("Your %d Tax Return is Complete", data.TaxYear)

	// HTML version
	htmlBody = fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s</title>
</head>
<body style="margin: 0; padding: 0; font-family: Arial, sans-serif; background-color: #f4f4f4;">
    <table role="presentation" style="width: 100%%; border-collapse: collapse;">
        <tr>
            <td align="center" style="padding: 40px 0;">
                <table role="presentation" style="width: 600px; border-collapse: collapse; background-color: #ffffff; box-shadow: 0 2px 4px rgba(0,0,0,0.1);">
                    <!-- Header -->
                    <tr>
                        <td style="padding: 40px 30px; background-color: #2563eb; text-align: center;">
                            <h1 style="margin: 0; color: #ffffff; font-size: 28px;">Tax Return Completed</h1>
                        </td>
                    </tr>

                    <!-- Body -->
                    <tr>
                        <td style="padding: 40px 30px;">
                            <p style="margin: 0 0 20px 0; font-size: 16px; line-height: 24px; color: #333333;">
                                Dear %s,
                            </p>

                            <p style="margin: 0 0 20px 0; font-size: 16px; line-height: 24px; color: #333333;">
                                Great news! Your <strong>%d %s</strong> tax return has been completed and is ready for your review.
                            </p>

                            <p style="margin: 0 0 20px 0; font-size: 16px; line-height: 24px; color: #333333;">
                                You can access your tax documents and review all details by logging into your account.
                            </p>

                            <!-- CTA Button -->
                            <table role="presentation" style="width: 100%%; margin: 30px 0;">
                                <tr>
                                    <td align="center">
                                        <a href="%s" style="display: inline-block; padding: 14px 40px; background-color: #2563eb; color: #ffffff; text-decoration: none; border-radius: 6px; font-size: 16px; font-weight: bold;">View Your Tax Return</a>
                                    </td>
                                </tr>
                            </table>

                            <p style="margin: 20px 0 0 0; font-size: 14px; line-height: 20px; color: #666666;">
                                If you have any questions or need assistance, please don't hesitate to contact us.
                            </p>
                        </td>
                    </tr>

                    <!-- Footer -->
                    <tr>
                        <td style="padding: 30px; background-color: #f8f9fa; border-top: 1px solid #e5e7eb;">
                            <p style="margin: 0 0 10px 0; font-size: 14px; color: #666666; text-align: center;">
                                Best regards,<br>
                                <strong>%s</strong>
                            </p>
                            <p style="margin: 0; font-size: 12px; color: #999999; text-align: center;">
                                This is an automated message. Please do not reply to this email.
                            </p>
                        </td>
                    </tr>
                </table>
            </td>
        </tr>
    </table>
</body>
</html>
`, subject, data.ClientName, data.TaxYear, data.FilingType, data.LoginURL, data.TenantName)

	// Text version (for email clients that don't support HTML)
	textBody = fmt.Sprintf(`
Dear %s,

Great news! Your %d %s tax return has been completed and is ready for your review.

You can access your tax documents and review all details by logging into your account:
%s

If you have any questions or need assistance, please don't hesitate to contact us.

Best regards,
%s

---
This is an automated message. Please do not reply to this email.
`, data.ClientName, data.TaxYear, data.FilingType, data.LoginURL, data.TenantName)

	// Clean up whitespace
	htmlBody = strings.TrimSpace(htmlBody)
	textBody = strings.TrimSpace(textBody)

	return subject, htmlBody, textBody
}

// GeneratePortalAccessEmail creates HTML and text versions of the portal access email
func GeneratePortalAccessEmail(data PortalAccessEmail) (subject, htmlBody, textBody string) {
	subject = "Access Your Tax Documents Portal"

	// HTML version
	htmlBody = fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s</title>
</head>
<body style="margin: 0; padding: 0; font-family: Arial, sans-serif; background-color: #f4f4f4;">
    <table role="presentation" style="width: 100%%; border-collapse: collapse;">
        <tr>
            <td align="center" style="padding: 40px 0;">
                <table role="presentation" style="width: 600px; border-collapse: collapse; background-color: #ffffff; box-shadow: 0 2px 4px rgba(0,0,0,0.1);">
                    <!-- Header -->
                    <tr>
                        <td style="padding: 40px 30px; background-color: #2563eb; text-align: center;">
                            <h1 style="margin: 0; color: #ffffff; font-size: 28px;">Access Your Portal</h1>
                        </td>
                    </tr>

                    <!-- Body -->
                    <tr>
                        <td style="padding: 40px 30px;">
                            <p style="margin: 0 0 20px 0; font-size: 16px; line-height: 24px; color: #333333;">
                                Dear %s,
                            </p>

                            <p style="margin: 0 0 20px 0; font-size: 16px; line-height: 24px; color: #333333;">
                                You can now access your secure tax documents portal. Click the button below to view your filings, documents, and payment information.
                            </p>

                            <p style="margin: 0 0 20px 0; font-size: 14px; line-height: 20px; color: #666666;">
                                This link is valid for <strong>24 hours</strong> and will automatically log you in.
                            </p>

                            <!-- CTA Button -->
                            <table role="presentation" style="width: 100%%; margin: 30px 0;">
                                <tr>
                                    <td align="center">
                                        <a href="%s" style="display: inline-block; padding: 16px 48px; background-color: #2563eb; color: #ffffff; text-decoration: none; border-radius: 6px; font-size: 16px; font-weight: bold;">Access Your Portal</a>
                                    </td>
                                </tr>
                            </table>

                            <p style="margin: 20px 0 0 0; font-size: 14px; line-height: 20px; color: #666666;">
                                If the button doesn't work, copy and paste this link into your browser:
                            </p>
                            <p style="margin: 10px 0 0 0; font-size: 12px; line-height: 18px; color: #2563eb; word-break: break-all;">
                                %s
                            </p>

                            <div style="margin-top: 30px; padding: 15px; background-color: #fef3c7; border-left: 4px solid #f59e0b; border-radius: 4px;">
                                <p style="margin: 0; font-size: 14px; line-height: 20px; color: #92400e;">
                                    <strong>Security Note:</strong> Never share this link with anyone. If you didn't request this access, please contact us immediately.
                                </p>
                            </div>
                        </td>
                    </tr>

                    <!-- Footer -->
                    <tr>
                        <td style="padding: 30px; background-color: #f8f9fa; border-top: 1px solid #e5e7eb;">
                            <p style="margin: 0 0 10px 0; font-size: 14px; color: #666666; text-align: center;">
                                Best regards,<br>
                                <strong>%s</strong>
                            </p>
                            <p style="margin: 0; font-size: 12px; color: #999999; text-align: center;">
                                This is an automated message. Please do not reply to this email.
                            </p>
                        </td>
                    </tr>
                </table>
            </td>
        </tr>
    </table>
</body>
</html>
`, subject, data.ClientName, data.PortalURL, data.PortalURL, data.TenantName)

	// Text version
	textBody = fmt.Sprintf(`
Dear %s,

You can now access your secure tax documents portal.

Click or copy this link to view your filings, documents, and payment information:
%s

This link is valid for 24 hours and will automatically log you in.

SECURITY NOTE: Never share this link with anyone. If you didn't request this access, please contact us immediately.

Best regards,
%s

---
This is an automated message. Please do not reply to this email.
`, data.ClientName, data.PortalURL, data.TenantName)

	// Clean up whitespace
	htmlBody = strings.TrimSpace(htmlBody)
	textBody = strings.TrimSpace(textBody)

	return subject, htmlBody, textBody
}
