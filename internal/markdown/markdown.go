package markdown

import (
	"html"
	"html/template"
	"regexp"
	"strconv"
	"strings"
)

var (
	// Extraction patterns — matched first to protect content from later passes.
	codeBlockRe  = regexp.MustCompile("(?s)```(?:\\w*\n)?(.*?)```")
	inlineCodeRe = regexp.MustCompile("`([^`]+)`")
	linkRe       = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)

	// Discord-specific syntax (matched after HTML escaping, so angle brackets are &lt;/&gt;).
	userMentionRe    = regexp.MustCompile(`&lt;@!?(\d+)&gt;`)
	channelMentionRe = regexp.MustCompile(`&lt;#(\d+)&gt;`)
	customEmojiRe    = regexp.MustCompile(`&lt;a?:(\w+):\d+&gt;`)
	spoilerRe        = regexp.MustCompile(`\|\|(.+?)\|\|`)

	// Inline formatting.
	boldItalicRe    = regexp.MustCompile(`\*\*\*(.+?)\*\*\*`)
	boldRe          = regexp.MustCompile(`\*\*(.+?)\*\*`)
	italicRe        = regexp.MustCompile(`\*(.+?)\*`)
	strikethroughRe = regexp.MustCompile(`~~(.+?)~~`)

	// Links and structure.
	bareURLRe    = regexp.MustCompile(`(https?://[^\s<]+)`)
	blockquoteRe = regexp.MustCompile(`(?m)^&gt; (.+)$`)
)

const maxCodeBlockLen = 3000

// ToHTML converts Discord-flavored markdown to safe HTML.
func ToHTML(s string) template.HTML {
	// Escape HTML first so user content can't inject tags.
	s = html.EscapeString(s)

	// Extract code blocks, inline code, and links to protect from later passes.
	var placeholders []string

	s = codeBlockRe.ReplaceAllStringFunc(s, func(match string) string {
		inner := strings.TrimSpace(codeBlockRe.FindStringSubmatch(match)[1])
		if len(inner) > maxCodeBlockLen {
			inner = inner[:maxCodeBlockLen] + "\n… (truncated)"
		}
		p := placeholder(len(placeholders))
		placeholders = append(placeholders, "<pre><code>"+inner+"</code></pre>")
		return p
	})

	s = inlineCodeRe.ReplaceAllStringFunc(s, func(match string) string {
		inner := inlineCodeRe.FindStringSubmatch(match)[1]
		p := placeholder(len(placeholders))
		placeholders = append(placeholders, "<code>"+inner+"</code>")
		return p
	})

	s = linkRe.ReplaceAllStringFunc(s, func(match string) string {
		parts := linkRe.FindStringSubmatch(match)
		p := placeholder(len(placeholders))
		placeholders = append(placeholders, `<a href="`+parts[2]+`">`+parts[1]+`</a>`)
		return p
	})

	// Discord-specific syntax.
	s = userMentionRe.ReplaceAllString(s, "<strong>@user:$1</strong>")
	s = channelMentionRe.ReplaceAllString(s, "<strong>#channel:$1</strong>")
	s = customEmojiRe.ReplaceAllString(s, ":$1:")
	s = spoilerRe.ReplaceAllString(s, "$1")

	// Inline formatting — bold+italic before bold before italic.
	s = boldItalicRe.ReplaceAllString(s, "<strong><em>$1</em></strong>")
	s = boldRe.ReplaceAllString(s, "<strong>$1</strong>")
	s = italicRe.ReplaceAllString(s, "<em>$1</em>")
	s = strikethroughRe.ReplaceAllString(s, "<s>$1</s>")

	// Bare URLs.
	s = bareURLRe.ReplaceAllString(s, `<a href="$1">$1</a>`)

	// Block quotes.
	s = blockquoteRe.ReplaceAllString(s, "<blockquote>$1</blockquote>")

	// Newlines.
	s = strings.ReplaceAll(s, "\n", "<br>")

	// Restore all placeholders.
	for i, val := range placeholders {
		s = strings.Replace(s, placeholder(i), val, 1)
	}

	return template.HTML(s)
}

// placeholder returns a null-byte-delimited token that temporarily replaces
// extracted content (code blocks, inline code, links) so that later regex
// passes don't modify it. Placeholders are restored at the end of ToHTML.
func placeholder(i int) string {
	return "\x00PH" + strconv.Itoa(i) + "\x00"
}
