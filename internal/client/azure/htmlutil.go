package azure

import (
	"html"
	"regexp"
	"strings"
)

var (
	azureBlockTagRe = regexp.MustCompile(`(?i)</(p|div|li|br)\s*>`)
	azureBrTagRe    = regexp.MustCompile(`(?i)<br\s*/?>`)
	azureAnyTagRe   = regexp.MustCompile(`<[^>]*>`)
	azureBlankLines = regexp.MustCompile(`\n{3,}`)
)

// stripHTML converts Azure DevOps's HTML-formatted rich text fields
// (description, comments) into plain text, since sneak only ever displays
// plain text for these fields.
func stripHTML(s string) string {
	if s == "" {
		return ""
	}

	// Turn common block-level boundaries into line breaks before removing
	// tags entirely, so paragraphs/list items don't get smashed together.
	out := azureBrTagRe.ReplaceAllString(s, "\n")
	out = azureBlockTagRe.ReplaceAllString(out, "\n")
	out = azureAnyTagRe.ReplaceAllString(out, "")

	out = html.UnescapeString(out)
	out = azureBlankLines.ReplaceAllString(out, "\n\n")

	return strings.TrimSpace(out)
}
