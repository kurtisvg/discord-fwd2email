# Discord Forward-to-Email

Right-click any Discord message and forward it to your email inbox. Get a formatted email with surrounding context, author avatars, and a link back to the conversation.

![Example email](EXAMPLE.png)

## How it works

1. Right-click a message in any server, DM, or thread
2. Select **Apps > Forward to inbox**
3. A formatted HTML email arrives in your Gmail with up to 5 messages of context and an "Open in Discord" link

## Setup

### 1. Create a Discord application

1. Go to [discord.com/developers/applications](https://discord.com/developers/applications) and create a new application
2. Copy the **Application ID** and **Public Key** from General Information
3. Go to **Bot**, create a bot user, and copy the **Token**
4. Go to **Installation**, enable **User Install** (and optionally **Guild Install** for context message access in servers)

### 2. Generate a Gmail app password

1. Go to [myaccount.google.com](https://myaccount.google.com) > Security > App Passwords
2. Generate a new app password for "Mail"

### 3. Run the bot

All config can be set via flags or environment variables:

| Flag | Env var | Required | Description |
|------|---------|----------|-------------|
| `-discord-token` | `DISCORD_TOKEN` | Yes | Bot token |
| `-discord-app-id` | `DISCORD_APP_ID` | Yes | Application ID |
| `-discord-public-key` | `DISCORD_PUBLIC_KEY` | Webhook mode | Public key for signature verification |
| `-gmail-user` | `GMAIL_USER` | Yes | Gmail address |
| `-gmail-app-password` | `GMAIL_APP_PASSWORD` | Yes | Gmail app password |
| `-host` | `HOST` | No | HTTP server host (default: all interfaces) |
| `-port` | `PORT` | No | HTTP server port (default: 8080) |
| `-gateway` | | No | Use gateway (websocket) mode |

**Gateway mode** (local development, no public URL needed):

```sh
export DISCORD_TOKEN='...'
export DISCORD_APP_ID='...'
export GMAIL_USER='you@gmail.com'
export GMAIL_APP_PASSWORD='...'
go run . -gateway
```

**Webhook mode** (production, requires public HTTPS URL):

```sh
export DISCORD_TOKEN='...'
export DISCORD_APP_ID='...'
export DISCORD_PUBLIC_KEY='...'
export GMAIL_USER='you@gmail.com'
export GMAIL_APP_PASSWORD='...'
go run .
```

Then set the **Interactions Endpoint URL** in the Discord Developer Portal to `https://your-domain/interactions`.

### 4. Install the app

Generate an OAuth2 URL in the Developer Portal with the `applications.commands` scope (add `bot` scope with Read Message History permission for server installs) and open it to install the app to your account.

## Context messages

The bot fetches up to 5 messages before the target for context. This requires the bot to be a **server member** with Read Message History permission. In servers where the bot isn't a member, it forwards just the target message (which is always available from the interaction payload).

## License

MIT
