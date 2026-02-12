package jira

import (
	"encoding/json"
	"strings"
)

// ADFDocument represents a top-level Atlassian Document Format document.
type ADFDocument struct {
	Version int       `json:"version"`
	Type    string    `json:"type"`
	Content []ADFNode `json:"content"`
}

// ADFNode represents a node in an ADF document tree.
type ADFNode struct {
	Type    string    `json:"type"`
	Text    string    `json:"text,omitempty"`
	Content []ADFNode `json:"content,omitempty"`
}

// extractPlainText extracts plain text from an ADF JSON field.
// Handles v3 ADF objects, v2 plain strings, and null values.
func extractPlainText(raw json.RawMessage) string {
	if len(raw) == 0 || string(raw) == "null" {
		return ""
	}

	// Try as plain string first (v2 compatibility)
	var str string
	if err := json.Unmarshal(raw, &str); err == nil {
		return str
	}

	// Parse as ADF document
	var doc ADFDocument
	if err := json.Unmarshal(raw, &doc); err != nil {
		return ""
	}

	var buf strings.Builder
	extractTextNodes(&buf, doc.Content)
	return strings.TrimSpace(buf.String())
}

// extractTextNodes recursively walks ADF nodes and collects text content.
func extractTextNodes(buf *strings.Builder, nodes []ADFNode) {
	for _, node := range nodes {
		if node.Type == "text" && node.Text != "" {
			buf.WriteString(node.Text)
		}
		if node.Type == "hardBreak" {
			buf.WriteString("\n")
		}
		if node.Type == "paragraph" && buf.Len() > 0 {
			buf.WriteString("\n")
		}
		if len(node.Content) > 0 {
			extractTextNodes(buf, node.Content)
		}
	}
}
