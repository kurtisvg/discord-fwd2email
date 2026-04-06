# Discord Forward-to-Email

Google Chat has "Forward to inbox." Discord doesn't. Now it does.

Right-click any message, select **Apps** > **Forward to inbox**, and a formatted email lands in your Gmail — with context, avatars, and a link back.

![Example email](EXAMPLE.png)

## Getting started

Set your environment variables:

```sh
export DISCORD_TOKEN='your-bot-token'
export DISCORD_APP_ID='your-app-id'
export GMAIL_USER='you@gmail.com'
export GMAIL_APP_PASSWORD='your-app-password'
```

Then run it:

```sh
# Run directly (requires Go 1.24+)
go run github.com/discord-forward-to-email@latest -gateway

# Or clone and build from source
git clone https://github.com/kurtisvg/discord-forward-to-email.git
cd discord-forward-to-email
go run . -gateway
```

The bot registers its command on startup. Right-click any message > **Apps** > **Forward to inbox**.

## Discord setup

1. Create an app at [discord.com/developers/applications](https://discord.com/developers/applications)
2. **General Information** — copy the Application ID
3. **Bot** — create a bot, copy the token
4. **Installation** — enable User Install

That's it for basic usage. The bot will work in any server or DM, forwarding the target message.

**Want context messages?** The 5 messages before the target require the bot to be a server member with Read Message History. Enable Guild Install, then install the bot to your server with the `bot` + `applications.commands` scopes. Servers where the bot isn't a member gracefully fall back to target-only.

## Gmail setup

Generate an app password at [myaccount.google.com](https://myaccount.google.com) > Security > App Passwords. You need this because the bot sends email via SMTP, not the Gmail API.

## Configuration

Everything is configurable via flags or environment variables. Flags take precedence.

```
-discord-token     / DISCORD_TOKEN        Bot token (required)
-discord-app-id    / DISCORD_APP_ID       Application ID (required)
-discord-public-key / DISCORD_PUBLIC_KEY  Public key (webhook mode only)
-gmail-user        / GMAIL_USER           Gmail address (required)
-gmail-app-password / GMAIL_APP_PASSWORD  App password (required)
-host              / HOST                 Server host (default: all interfaces)
-port              / PORT                 Server port (default: 8080)
-gateway                                  Use websocket mode instead of webhooks
```

## Running modes

**Gateway mode** — connects to Discord via websocket. No public URL, no signature verification. Great for local dev.

```sh
go run . -gateway
```

**Webhook mode** — runs an HTTP server that receives interaction POSTs from Discord. Requires a public HTTPS URL and the public key for signature verification. This is what you'd use on Cloud Run or similar.

```sh
go run .
# Then set your Interactions Endpoint URL in the Discord Developer Portal
# to https://your-domain/interactions
```

## What you get

Each forwarded email includes:

- Up to 5 messages of context (oldest first), with the target highlighted
- Author names and avatars
- Discord markdown rendered as HTML (bold, italic, code, links, etc.)
- Attachments (images inline, files as links)
- An "Open in Discord" button linking back to the exact message
- Thread and channel names in the header

## License

MIT
