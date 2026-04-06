# PRD: Discord Forward-to-Inbox Bot

**Author:** [Your Name]
**Last Updated:** April 5, 2026
**Status:** Draft

---

## 1. Problem Statement

Important messages in Discord get buried in fast-moving channels. There's no native way to push a Discord message into an email inbox where it can be tracked, searched, or acted on alongside other work.

Google Chat solved this with "Forward to inbox" — right-click a message, select it from the menu, and a formatted email lands in your Gmail with surrounding context and a link back. Discord has no equivalent.

**This bot brings that exact workflow to Discord:** right-click a message → Apps → "Forward to inbox" → an email appears in your inbox with the message, surrounding context, author avatars, and a deep link back to the conversation.


## 2. User Story

As a Discord user, I want to right-click any message and forward it to my email so that I can act on it later from my inbox.


## 3. UX Flow

1. Right-click a message in any server, DM, or thread.
2. Select **Apps → "Forward to inbox"** from the context menu.
3. A brief "thinking..." indicator appears (ephemeral, only visible to you) while the bot works.
4. The bot extracts the target message from the interaction payload and attempts to fetch the 5 previous messages for context.
5. The bot sends a formatted HTML email to my inbox.
6. The "thinking..." indicator is replaced with a confirmation: *"✉️ Forwarded to you@gmail.com (with 5 messages of context)"*
7. The email arrives in my inbox within a few seconds, containing the conversation snippet and an "Open in Discord" button linking back to the original message.

No one else in the channel sees that anything happened. If the bot can't fetch context messages (e.g., it doesn't have channel access in that server), it forwards just the target message and notes this in the confirmation.


## 4. Functional Requirements

### 4.1 Discord Message Command

The bot registers a **Message Command** (a.k.a. context menu command) named "Forward to inbox." This is not a slash command — it appears in the right-click → Apps menu on any message, which matches the Google Chat UX exactly.

Discord's API for this is an application command with `type: 3` (MESSAGE).

### 4.2 Message Fetching

When the command is triggered, the interaction payload includes the full target message data (content, author, attachments, timestamp) — no API call needed for the target itself.

The bot then attempts to fetch up to 5 preceding messages for context via the REST API:

```
GET /channels/{channel_id}/messages?before={message_id}&limit=5
```

**Graceful degradation:** Since the bot is user-installed (not added as a member to each server), it may not have Read Message History permission in every channel. If the context fetch fails (403 Forbidden), the bot falls back to forwarding just the target message alone — which is still useful. The ephemeral confirmation indicates whether context was included:

- ✉️ Forwarded to you@gmail.com (with 5 messages of context)
- ✉️ Forwarded to you@gmail.com (target message only — no channel access for context)

The email includes context messages (oldest first) followed by the target message at the bottom. For each message, the bot captures:

| Field | Source |
|-------|--------|
| Author display name | `message.author.global_name` or `message.author.username` |
| Author avatar | `https://cdn.discordapp.com/avatars/{user_id}/{avatar_hash}.png` |
| Message text | `message.content` (already markdown-formatted) |
| Attachments | `message.attachments[]` — include URLs for images/files |
| Message link | Constructed: `https://discord.com/channels/{guild_id}/{channel_id}/{message_id}` |
| Server name | Fetched from `GET /guilds/{guild_id}` (can be cached) |
| Channel name | Fetched from `GET /channels/{channel_id}` |
| Timestamp | `message.timestamp` (ISO 8601) |

### 4.3 Email Composition

The bot sends an HTML email via SMTP that mirrors Google Chat's forwarded email format:

```
Subject: [Discord] Forwarded chat in #support

┌─────────────────────────────────────────┐
│                                         │
│  Forwarded chat in Acme Corp · #support │
│  ─────────────────────────────────────  │
│                                         │
│  [avatar]  Alice Chen                   │
│            Has anyone else seen this?   │
│                                         │
│  [avatar]  Bob Smith                    │
│            Yeah I hit that yesterday    │
│                                         │
│  [avatar]  Alice Chen                   │
│            What endpoint are you        │
│            hitting?                     │
│                                         │
│  [avatar]  Bob Smith                    │
│            /api/v1/users — the bulk     │
│            fetch                        │
│                                         │
│  [avatar]  Alice Chen                   │
│            That one's limited to        │
│            30/min actually              │
│                                         │
│  [avatar]  Jane Doe           ← target  │
│            Can someone help with the    │
│            API rate limit issue? I'm    │
│            seeing 429s after ~50 reqs.  │
│                                         │
│  ┌──────────────────┐                   │
│  │ Open in Discord  │  ← deep link     │
│  └──────────────────┘                   │
│                                         │
└─────────────────────────────────────────┘
```

**Format details:**

- **Header:** "Forwarded chat in {Server Name} · #{channel-name}"
- **Messages:** Each with a circular avatar (loaded from Discord CDN), bold author name, and message text. The target message is visually highlighted with a subtle left border or background tint so it stands out from context messages.
- **CTA button:** "Open in Discord" in Discord blurple (`#5865F2`), linking to the target message.
- **Attachments:** Embedded as clickable links or inline images below the relevant message.
- **No individual timestamps** on messages — keeps it clean. The email's own timestamp provides "when."

### 4.4 Email Delivery

SMTP via Gmail using an app password. The email is sent from me to me.

```
Transport: smtp.gmail.com:587 (TLS)
Auth:      Gmail address + app password (stored as env var)
From:      you@gmail.com
To:        you@gmail.com
```

### 4.5 Ephemeral Confirmation

The bot immediately defers the interaction with an ephemeral "thinking..." state (satisfying Discord's 3-second deadline), then edits the deferred response once the email is sent.

**Success (with context):**
> ✉️ Forwarded to you@gmail.com (with 5 messages of context)

**Success (target only, context fetch failed):**
> ✉️ Forwarded to you@gmail.com (target message only — no channel access for context)

**Failure:**
> ❌ Failed to forward — check bot logs.


## 5. Technical Design

### 5.1 Architecture

```
User right-clicks message
        │
        ▼
Discord sends POST to
interactions endpoint
        │
        ▼
┌──────────────────────────────┐
│  Cloud Run container         │
│  (HTTP server, scales to 0)  │
│                              │
│  1. Verify Ed25519 signature │
│  2. Respond with deferred    │
│     ephemeral reply — must   │
│     be within 3 seconds      │
│  3. Extract target message   │
│     from interaction payload │
│  4. Fetch up to 5 messages   │
│     before target (REST)     │
│  5. Fetch server/channel     │
│     names                    │
│  6. Render HTML email        │
│  7. Send via SMTP            │
│  8. Edit deferred reply      │
│     with ✓ or ❌              │
└──────────────────────────────┘
        │
        ▼
Gmail SMTP → your inbox
```

**Why webhook mode:** The bot is deployed on Cloud Run, which scales to zero when idle and spins up on incoming HTTP requests. A WebSocket gateway connection (the other option) requires a persistent process, which doesn't work on scale-to-zero platforms. In webhook mode, Discord sends a POST request to the bot's interactions endpoint URL for each command invocation. The bot verifies the request signature (`Ed25519`), handles the interaction, and the container can shut down afterward.

**Interaction flow:** Discord's webhook interactions have a nuance with deferred replies. The initial HTTP response to Discord must be sent within 3 seconds and contains the deferred acknowledgment. The bot then does the actual work (fetch messages, send email) and uses the Discord REST API to edit the deferred response with the final result. This means the handler needs to return the deferred response immediately, then continue processing in the background before the request completes — or use a goroutine to do the async work after sending the HTTP response.

### 5.2 Stack

| Component | Choice | Rationale |
|-----------|--------|-----------|
| Runtime | Go | Single static binary, minimal container image, fast cold starts on Cloud Run |
| HTTP server | `net/http` (stdlib) | Receives Discord interaction webhooks. No framework needed for a single endpoint. |
| Discord interactions | `discordgo` (for types and REST client) | Provides struct definitions for interaction payloads, message types, and a REST client for fetching messages and editing deferred replies. The gateway/WebSocket features are not used. |
| Signature verification | `crypto/ed25519` (stdlib) | Discord requires Ed25519 signature verification on all incoming webhook requests |
| Email | `net/smtp` (stdlib) | Built-in SMTP support. Requires manually constructing the MIME message (Content-Type, headers, HTML body), but avoids external dependencies for a straightforward use case. |
| HTML templating | `html/template` (stdlib) | Built-in, secure by default (auto-escapes), sufficient for email rendering |
| Hosting | Google Cloud Run | Scales to zero when idle — effectively free for a personal bot used a few times per day. Spins up on each Discord interaction POST. |
| Config | Environment variables | `DISCORD_TOKEN`, `DISCORD_PUBLIC_KEY`, `GMAIL_USER`, `GMAIL_APP_PASSWORD` |

### 5.3 Bot Permissions & Intents

| Requirement | Value |
|-------------|-------|
| OAuth2 scopes | `bot`, `applications.commands` |
| Bot permissions | Read Message History (used for context fetch; bot functions without it) |
| Gateway intents | N/A — the bot uses webhook mode, not a gateway connection |
| Privileged intents | N/A |
| Interactions endpoint URL | Registered in Discord Developer Portal, points to the Cloud Run service URL (e.g., `https://forward-bot-xyz.run.app/interactions`) |
| Installation | User-installable app (installed to your account, usable in any server or DM) |

**Permission model:** The bot is installed to the user's account, not to individual servers. This means it can receive interactions (and the target message) everywhere, but REST API calls to fetch context messages require the bot to have Read Message History in that channel — which it won't have in servers where it isn't a guild member. The bot handles this gracefully: if the context fetch returns a 403, it forwards just the target message and notes this in the ephemeral reply. This is an acceptable trade-off for the convenience of not needing server admin permission anywhere.

### 5.4 Registering the Message Command

Command registration is done via a one-time setup script (or on first deploy) using the Discord REST API. This is separate from the webhook server itself:

```go
// One-time setup script
s, _ := discordgo.New("Bot " + token)
_, err := s.ApplicationCommandCreate(appID, "", &discordgo.ApplicationCommand{
    Name: "Forward to inbox",
    Type: discordgo.MessageApplicationCommand, // type 3
})
```

This registers the command globally, making it appear in right-click → Apps for every message. The command only needs to be registered once — it persists until explicitly deleted.

### 5.5 HTML Email Template

The email body is a self-contained HTML document with inline CSS (for email client compatibility), rendered using Go's `html/template`. Key implementation notes:

- Use `<table>` layouts — email clients don't reliably support flexbox/grid.
- Inline all CSS — `<style>` blocks are stripped by many clients.
- Avatar images reference Discord's CDN directly (`<img src="https://cdn.discordapp.com/avatars/...">`). Gmail will proxy these automatically.
- The "Open in Discord" button is an `<a>` tag styled as a button, not a `<button>` element.
- The target message row gets a subtle left border (`border-left: 3px solid #5865F2`) to distinguish it from context.

### 5.6 Discord Markdown → HTML Conversion

`message.content` contains Discord-flavored markdown and custom syntax that needs to be converted to HTML for the email. The conversion rules for v1:

| Discord syntax | HTML output |
|----------------|-------------|
| `**bold**` | `<strong>bold</strong>` |
| `*italic*` | `<em>italic</em>` |
| `` `inline code` `` | `<code>inline code</code>` |
| ` ```code block``` ` | `<pre><code>code block</code></pre>` |
| `> quote` | `<blockquote>quote</blockquote>` |
| `~~strikethrough~~` | `<s>strikethrough</s>` |
| `[text](url)` | `<a href="url">text</a>` |
| Bare URLs | Wrapped in `<a>` tags |
| `<@user_id>` (user mention) | Render as `@username` in bold (resolve from message data if available, otherwise use raw ID) |
| `<#channel_id>` (channel mention) | Render as `#channel-name` (resolve if possible, otherwise use raw ID) |
| `<:name:id>` (custom emoji) | Render as `:name:` plain text |
| `||spoiler||` | Strip spoiler tags, render content as plain text |
| Newlines | Convert to `<br>` |

This can be implemented as a simple regex-based converter — no need for a full markdown parser. A Go library like `gomarkdown` could be used for the standard markdown subset, with custom pre-processing for Discord-specific syntax (mentions, emoji, spoilers).


## 6. Edge Cases

| Scenario | Handling |
|----------|----------|
| Message has no text (image/embed only) | Include attachment URLs. Show "📎 [filename]" as the message content. |
| Message is in a DM (no server) | Header becomes "Forwarded DM with {username}" instead of server/channel. Message link uses `@me` as the guild ID. |
| Message is in a thread | Use thread name as channel. Link goes to the thread message. |
| Message has a very long code block | Truncate at ~3000 chars in the email. Full content available via the "Open in Discord" link. |
| Bot lacks channel access for context fetch (403) | Forward the target message alone (which is always available from the interaction payload). Note in ephemeral reply: "target message only — no channel access for context." |
| Fewer than 5 context messages available | The target message is near the top of a channel or thread. The `?before=` endpoint returns fewer results — the template renders however many were returned (0–5) without assuming exactly 5. |
| SMTP send fails | Edit deferred reply with error message. Log details to console. |
| Discord-specific markdown (spoilers, custom emoji) | Convert per the rules in section 5.6. |
| User has no avatar set | Use Discord's default avatar URL based on `(user_id >> 22) % 6`. |


## 7. Setup Steps

1. **Create a Discord application** at discord.com/developers/applications. Copy the **Application ID** and **Public Key**.
2. **Create a bot user**, copy the token → `DISCORD_TOKEN` env var. Copy the public key → `DISCORD_PUBLIC_KEY` env var.
3. **Enable "User Install"** in the app's Installation settings so it's installable to your account without needing server admin.
4. **Generate a Gmail app password** at myaccount.google.com → Security → App Passwords → `GMAIL_APP_PASSWORD` env var.
5. **Deploy to Cloud Run** — build a container image, deploy it, and note the service URL (e.g., `https://forward-bot-xyz.run.app`).
6. **Set the Interactions Endpoint URL** in the Discord Developer Portal → General Information → paste `https://forward-bot-xyz.run.app/interactions`. Discord will send a verification ping — the bot must respond correctly for this to save.
7. **Register the message command** — run the one-time setup script (section 5.4) from your local machine.
8. **Install the app** to your Discord account via the OAuth2 URL.
9. Right-click any message → Apps → "Forward to inbox" → done.


## 8. Future Possibilities (Out of Scope for v1)

- **Configurable context window** — slash command to change how many previous messages are included.
- **Multiple email targets** — forward to different emails depending on the server.
- **Forward to other services** — Notion, Linear, Todoist via their APIs.
- **Multi-user support** — let others install the bot with their own email config. Would require a database for per-user settings and an email relay service.
- **Digest mode** — queue forwards and send as a single daily digest email.


## 9. Reference: Google Chat's "Forward to Inbox"

This bot is modeled directly on Google Chat's native feature:

- **Trigger:** "Forward to inbox" appears in the message context menu (via `⋮` or right-click), alongside "Quote in reply," "Star," "Pin to board," and "Copy message link."
- **Email format:** Header reads "Forwarded chat in room {Room Name}" (room name is a link), followed by ~5 messages — each with a circular avatar, bold author name, and message text. No timestamps on individual messages.
- **CTA:** A single "Open in Google Chat" button at the bottom.
- **Delivery:** Sent silently to the user's own Gmail inbox. No compose window.
- **Reference:** [Google Support: Forward a message from Chat to Gmail](https://support.google.com/chat/answer/9757415)