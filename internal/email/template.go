package email

import "html/template"

type Attachment struct {
	Filename string
	URL      string
	IsImage  bool
}

type MessageData struct {
	AuthorName  string
	AvatarURL   string
	Content     string
	Attachments []Attachment
}

type ForwardData struct {
	ServerName      string
	ChannelName     string
	MessageLink     string
	ContextMessages []MessageData
	TargetMessage   MessageData
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

        {{/* Context messages */}}
        {{range .ContextMessages}}
        <tr><td style="padding:0 24px;">
          <table width="100%" cellpadding="0" cellspacing="0">
            <tr>
              <td width="40" valign="top" style="padding:12px 0 12px 12px;">
                <img src="{{.AvatarURL}}" width="32" height="32" style="border-radius:50%;display:block;" alt="{{.AuthorName}}">
              </td>
              <td style="padding:12px 8px 12px 8px;">
                <p style="margin:0 0 2px 0;font-weight:bold;font-size:14px;color:#111;">{{.AuthorName}}</p>
                {{if .Content}}<p style="margin:0;font-size:14px;color:#333;line-height:1.5;">{{.Content}}</p>{{end}}
                {{range .Attachments}}
                  {{if .IsImage}}<img src="{{.URL}}" style="max-width:100%;border-radius:4px;margin-top:8px;display:block;" alt="{{.Filename}}">
                  {{else}}<p style="margin:4px 0 0 0;font-size:13px;"><a href="{{.URL}}" style="color:#5865F2;text-decoration:none;">📎 {{.Filename}}</a></p>
                  {{end}}
                {{end}}
              </td>
            </tr>
          </table>
        </td></tr>
        {{end}}

        {{/* Target message (highlighted) */}}
        <tr><td style="padding:0 24px;">
          <table width="100%" cellpadding="0" cellspacing="0">
            <tr>
              <td width="40" valign="top" style="padding:12px 0 12px 12px;border-left:3px solid #5865F2;">
                <img src="{{.TargetMessage.AvatarURL}}" width="32" height="32" style="border-radius:50%;display:block;" alt="{{.TargetMessage.AuthorName}}">
              </td>
              <td style="padding:12px 8px 12px 8px;border-left:3px solid #5865F2;">
                <p style="margin:0 0 2px 0;font-weight:bold;font-size:14px;color:#111;">{{.TargetMessage.AuthorName}}</p>
                {{if .TargetMessage.Content}}<p style="margin:0;font-size:14px;color:#333;line-height:1.5;">{{.TargetMessage.Content}}</p>{{end}}
                {{range .TargetMessage.Attachments}}
                  {{if .IsImage}}<img src="{{.URL}}" style="max-width:100%;border-radius:4px;margin-top:8px;display:block;" alt="{{.Filename}}">
                  {{else}}<p style="margin:4px 0 0 0;font-size:13px;"><a href="{{.URL}}" style="color:#5865F2;text-decoration:none;">📎 {{.Filename}}</a></p>
                  {{end}}
                {{end}}
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
