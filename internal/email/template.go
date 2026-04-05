package email

import "html/template"

type MessageData struct {
	AuthorName string
	Content    string
}

type ForwardData struct {
	ServerName    string
	ChannelName   string
	MessageLink   string
	TargetMessage MessageData
}

var emailTemplate = template.Must(template.New("email").Parse(emailTemplateHTML))

const emailTemplateHTML = `<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="margin:0;padding:0;background-color:#f5f5f5;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;">
  <table width="100%" cellpadding="0" cellspacing="0" style="background-color:#f5f5f5;padding:24px 0;">
    <tr><td align="center">
      <table width="600" cellpadding="0" cellspacing="0" style="background-color:#ffffff;border-radius:8px;overflow:hidden;">

        {{/* Header */}}
        <tr><td style="padding:24px 24px 16px 24px;">
          <p style="margin:0;font-size:16px;color:#333;">
            {{if .ServerName}}Forwarded chat in {{.ServerName}} · #{{.ChannelName}}
            {{else if .ChannelName}}Forwarded chat in #{{.ChannelName}}
            {{else}}Forwarded DM with {{.TargetMessage.AuthorName}}
            {{end}}
          </p>
          <hr style="border:none;border-top:1px solid #e0e0e0;margin-top:16px;">
        </td></tr>

        {{/* Target message */}}
        <tr><td style="padding:0 24px;">
          <table width="100%" cellpadding="0" cellspacing="0">
            <tr>
              <td style="padding:12px 0 12px 12px;border-left:3px solid #5865F2;">
                <p style="margin:0 0 2px 0;font-weight:bold;font-size:14px;color:#111;">{{.TargetMessage.AuthorName}}</p>
                <p style="margin:0;font-size:14px;color:#333;line-height:1.5;">{{.TargetMessage.Content}}</p>
              </td>
            </tr>
          </table>
        </td></tr>

        {{/* CTA button */}}
        <tr><td style="padding:24px;" align="center">
          <a href="{{.MessageLink}}"
             style="display:inline-block;padding:12px 24px;background-color:#5865F2;color:#ffffff;text-decoration:none;border-radius:4px;font-size:14px;font-weight:600;">
            Open in Discord
          </a>
        </td></tr>

      </table>
    </td></tr>
  </table>
</body>
</html>`
