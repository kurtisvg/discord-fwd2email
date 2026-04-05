package markdown

import (
	"html/template"
	"strings"
	"testing"
)

func TestToHTML(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  template.HTML
	}{
		// Standard markdown.
		{
			name:  "bold",
			input: "**hello**",
			want:  "<strong>hello</strong>",
		},
		{
			name:  "italic",
			input: "*hello*",
			want:  "<em>hello</em>",
		},
		{
			name:  "bold italic",
			input: "***hello***",
			want:  "<strong><em>hello</em></strong>",
		},
		{
			name:  "strikethrough",
			input: "~~hello~~",
			want:  "<s>hello</s>",
		},
		{
			name:  "inline code",
			input: "use `fmt.Println`",
			want:  "use <code>fmt.Println</code>",
		},
		{
			name:  "code block",
			input: "```\nfmt.Println()\n```",
			want:  "<pre><code>fmt.Println()</code></pre>",
		},
		{
			name:  "code block with language",
			input: "```go\nfmt.Println()\n```",
			want:  "<pre><code>fmt.Println()</code></pre>",
		},
		{
			name:  "markdown link",
			input: "[click here](https://example.com)",
			want:  `<a href="https://example.com">click here</a>`,
		},
		{
			name:  "bare URL",
			input: "check https://example.com for details",
			want:  `check <a href="https://example.com">https://example.com</a> for details`,
		},
		{
			name:  "blockquote",
			input: "> quoted text",
			want:  "<blockquote>quoted text</blockquote>",
		},
		{
			name:  "multi-line blockquote",
			input: "> line one\n> line two",
			want:  "<blockquote>line one</blockquote><br><blockquote>line two</blockquote>",
		},
		{
			name:  "newlines",
			input: "line one\nline two",
			want:  "line one<br>line two",
		},

		// Discord-specific syntax.
		{
			name:  "user mention",
			input: "hey <@123456>",
			want:  "hey <strong>@user:123456</strong>",
		},
		{
			name:  "user mention with nickname",
			input: "hey <@!123456>",
			want:  "hey <strong>@user:123456</strong>",
		},
		{
			name:  "channel mention",
			input: "check <#789012>",
			want:  "check <strong>#channel:789012</strong>",
		},
		{
			name:  "custom emoji",
			input: "nice <:thumbsup:123456>",
			want:  "nice :thumbsup:",
		},
		{
			name:  "animated emoji",
			input: "wow <a:partyblob:789012>",
			want:  "wow :partyblob:",
		},
		{
			name:  "spoiler",
			input: "the answer is ||42||",
			want:  "the answer is 42",
		},
		{
			name:  "spoiler with formatting",
			input: "||**bold spoiler**||",
			want:  "<strong>bold spoiler</strong>",
		},

		// Edge cases.
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "plain text unchanged",
			input: "just plain text",
			want:  "just plain text",
		},
		{
			name:  "html escape",
			input: "<script>alert('xss')</script>",
			want:  "&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;",
		},
		{
			name:  "combined formatting",
			input: "**bold** and *italic* with `code`",
			want:  "<strong>bold</strong> and <em>italic</em> with <code>code</code>",
		},
		{
			name:  "inline code preserves formatting chars",
			input: "`**not bold**`",
			want:  "<code>**not bold**</code>",
		},
		{
			name:  "code block preserves formatting",
			input: "```\n**not bold** *not italic*\n```",
			want:  "<pre><code>**not bold** *not italic*</code></pre>",
		},
		{
			name:  "code block preserves mentions",
			input: "```\n<@123> <#456>\n```",
			want:  "<pre><code>&lt;@123&gt; &lt;#456&gt;</code></pre>",
		},
		{
			name:  "long code block truncated",
			input: "```\n" + strings.Repeat("x", 3500) + "\n```",
			want:  "<pre><code>" + template.HTML(strings.Repeat("x", 3000)) + "\n… (truncated)</code></pre>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := ToHTML(tt.input)
			if got != tt.want {
				t.Errorf("\ninput: %q\nwant:  %q\ngot:   %q", tt.input, tt.want, got)
			}
		})
	}
}
