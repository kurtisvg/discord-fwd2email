# Discord Forward-to-Email

Google Chat has "Forward to inbox." Discord doesn't. Now it does.

Right-click any message, select **Apps** > **Forward to inbox**, and a formatted email lands in your Gmail — with context, avatars, and a link back.

<p align="center">
  <img src="EXAMPLE.png" width="500" alt="Example email">
</p>

## Getting started

### 1. Create a Discord application

1. Go to [discord.com/developers/applications](https://discord.com/developers/applications) and click **New Application**
2. Give it a name (e.g. "Forward to Email") and create it
3. On the **General Information** page, copy the **Application ID** — this is your `DISCORD_APP_ID`
4. Copy the **Public Key** — this is your `DISCORD_PUBLIC_KEY` (needed for webhook mode)
5. Go to **Bot** in the left sidebar, click **Reset Token**, and copy it — this is your `DISCORD_TOKEN`
6. Go to **Installation** in the left sidebar and check **User Install**

### 2. Set up Gmail

1. Go to [myaccount.google.com](https://myaccount.google.com) > **Security** > **2-Step Verification** (must be enabled)
2. At the bottom of the 2-Step Verification page, click **App passwords**
3. Create a new app password (name it anything, e.g. "Discord bot")
4. Copy the 16-character password — this is your `GMAIL_APP_PASSWORD`

### 3. Install

<details open>
<summary>Build from source</summary>

```sh
git clone https://github.com/kurtisvg/discord-forward-to-email.git && cd discord-forward-to-email
go build -o fwd2email .
```

</details>

<details>
<summary>Or install directly with <code>go install</code></summary>

```sh
go install github.com/discord-forward-to-email@latest
```

</details>

### 4. Configure

```sh
export DISCORD_TOKEN='your-bot-token'
export DISCORD_APP_ID='your-app-id'
export GMAIL_USER='you@gmail.com'
export GMAIL_APP_PASSWORD='your-app-password'
```

### 5. Run

```sh
./fwd2email -gateway
```

The bot registers its command on startup. Right-click any message > **Apps** > **Forward to inbox**.

### 6. Install the app to your Discord account

1. In the Developer Portal, go to **OAuth2** > **URL Generator**
2. Select the `applications.commands` scope
3. Copy the generated URL and open it in your browser
4. Choose **Install to my account** and authorize

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

**Gateway mode** — connects to Discord via websocket. No public URL, no signature verification. Great for local dev and personal use.

```sh
./fwd2email -gateway
```

**Webhook mode** — runs an HTTP server that receives interaction POSTs from Discord. Requires a public HTTPS URL and the public key for signature verification. This is what you'd use on Cloud Run or similar.

```sh
./fwd2email
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

## Adding the bot to a server

By default, the bot is installed to your user account. It can forward any message you can see, but it can only fetch context messages (the 5 messages before the target) in servers where it's a member.

To add context message support in a server:

1. In the Developer Portal, go to **Installation** and enable **Guild Install**
2. Go to **OAuth2** > **URL Generator**
3. Select scopes: `bot` and `applications.commands`
4. Select bot permissions: **Read Message History**
5. Copy the URL and open it in your browser
6. Select the server you want to add it to

In servers where the bot isn't a member, it gracefully falls back to forwarding just the target message.

## License

MIT
