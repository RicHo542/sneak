package jira

import "strings"

// adfToText converts an Atlassian Document Format (ADF) node tree into
// plain text. Jira Cloud returns description/comment bodies as ADF rather
// than plain strings; sneak only ever displays plain text for these
// fields, so we flatten the tree, preserving paragraph breaks.
func adfToText(node jiraADFNode) string {
	var b strings.Builder
	writeADFNode(&b, node)
	return strings.TrimSpace(b.String())
}

func writeADFNode(b *strings.Builder, node jiraADFNode) {
	if node.Text != "" {
		b.WriteString(node.Text)
	}

	for _, child := range node.Content {
		writeADFNode(b, child)
	}

	switch node.Type {
	case "paragraph", "heading", "listItem", "hardBreak":
		b.WriteString("\n")
	}
}
