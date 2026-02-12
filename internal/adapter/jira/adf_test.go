package jira

import (
	"encoding/json"
	"testing"
)

func TestExtractPlainText(t *testing.T) {
	tests := []struct {
		name  string
		input json.RawMessage
		want  string
	}{
		{
			name:  "nil input",
			input: nil,
			want:  "",
		},
		{
			name:  "null JSON",
			input: json.RawMessage(`null`),
			want:  "",
		},
		{
			name:  "empty bytes",
			input: json.RawMessage{},
			want:  "",
		},
		{
			name:  "plain string v2 compat",
			input: json.RawMessage(`"hello world"`),
			want:  "hello world",
		},
		{
			name:  "empty string",
			input: json.RawMessage(`""`),
			want:  "",
		},
		{
			name: "simple ADF paragraph",
			input: json.RawMessage(`{
				"version": 1,
				"type": "doc",
				"content": [
					{
						"type": "paragraph",
						"content": [
							{"type": "text", "text": "Hello world"}
						]
					}
				]
			}`),
			want: "Hello world",
		},
		{
			name: "multiple ADF paragraphs",
			input: json.RawMessage(`{
				"version": 1,
				"type": "doc",
				"content": [
					{
						"type": "paragraph",
						"content": [
							{"type": "text", "text": "First paragraph"}
						]
					},
					{
						"type": "paragraph",
						"content": [
							{"type": "text", "text": "Second paragraph"}
						]
					}
				]
			}`),
			want: "First paragraph\nSecond paragraph",
		},
		{
			name: "ADF with bullet list",
			input: json.RawMessage(`{
				"version": 1,
				"type": "doc",
				"content": [
					{
						"type": "paragraph",
						"content": [
							{"type": "text", "text": "Requirements:"}
						]
					},
					{
						"type": "bulletList",
						"content": [
							{
								"type": "listItem",
								"content": [
									{
										"type": "paragraph",
										"content": [
											{"type": "text", "text": "Login page"}
										]
									}
								]
							},
							{
								"type": "listItem",
								"content": [
									{
										"type": "paragraph",
										"content": [
											{"type": "text", "text": "Token refresh"}
										]
									}
								]
							}
						]
					}
				]
			}`),
			want: "Requirements:\nLogin page\nToken refresh",
		},
		{
			name: "ADF with text marks ignored",
			input: json.RawMessage(`{
				"version": 1,
				"type": "doc",
				"content": [
					{
						"type": "paragraph",
						"content": [
							{"type": "text", "text": "Use "},
							{"type": "text", "text": "golang.org/x/oauth2", "marks": [{"type": "code"}]},
							{"type": "text", "text": " for auth."}
						]
					}
				]
			}`),
			want: "Use golang.org/x/oauth2 for auth.",
		},
		{
			name: "ADF with hard break",
			input: json.RawMessage(`{
				"version": 1,
				"type": "doc",
				"content": [
					{
						"type": "paragraph",
						"content": [
							{"type": "text", "text": "Line 1"},
							{"type": "hardBreak"},
							{"type": "text", "text": "Line 2"}
						]
					}
				]
			}`),
			want: "Line 1\nLine 2",
		},
		{
			name:  "empty ADF document",
			input: json.RawMessage(`{"version": 1, "type": "doc", "content": []}`),
			want:  "",
		},
		{
			name:  "invalid JSON",
			input: json.RawMessage(`not valid json`),
			want:  "",
		},
		{
			name: "ADF with heading",
			input: json.RawMessage(`{
				"version": 1,
				"type": "doc",
				"content": [
					{
						"type": "heading",
						"attrs": {"level": 2},
						"content": [
							{"type": "text", "text": "Section Title"}
						]
					},
					{
						"type": "paragraph",
						"content": [
							{"type": "text", "text": "Content here"}
						]
					}
				]
			}`),
			want: "Section Title\nContent here",
		},
		{
			name: "ADF with nested inline text",
			input: json.RawMessage(`{
				"version": 1,
				"type": "doc",
				"content": [
					{
						"type": "paragraph",
						"content": [
							{"type": "text", "text": "Start "},
							{"type": "text", "text": "middle"},
							{"type": "text", "text": " end"}
						]
					}
				]
			}`),
			want: "Start middle end",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractPlainText(tt.input)
			if got != tt.want {
				t.Errorf("extractPlainText() = %q, want %q", got, tt.want)
			}
		})
	}
}
