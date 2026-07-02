package main

import "strings"

// decodeHTMLEntities décode les entités HTML courantes
func decodeHTMLEntities(input string) string {
	// Remplacer les entités HTML courantes
	input = strings.ReplaceAll(input, "&amp;", "&")
	input = strings.ReplaceAll(input, "&#39;", "'")
	input = strings.ReplaceAll(input, "&quot;", "\"")
	return input
}