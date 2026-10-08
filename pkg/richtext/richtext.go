package richtext

import (
	"html"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/microcosm-cc/bluemonday"
)

var policy = newPolicy()

func newPolicy() *bluemonday.Policy {
	p := bluemonday.NewPolicy()
	p.AllowElements("p", "br", "strong", "em", "ul", "ol", "li", "h2", "h3", "blockquote", "a")
	p.AllowAttrs("href").OnElements("a")
	p.AllowStandardURLs()
	p.RequireNoFollowOnLinks(true)
	p.AddTargetBlankToFullyQualifiedLinks(true)
	return p
}

var (
	blockEnd   = regexp.MustCompile(`(?i)</(p|li|h2|h3|blockquote)>|<br\s*/?>`)
	manySpaces = regexp.MustCompile(`\s+`)
	stripper   = bluemonday.StripTagsPolicy()
)

func PlainText(s string) string {
	s = blockEnd.ReplaceAllString(s, " ")
	s = stripper.Sanitize(s)
	s = html.UnescapeString(s)
	return strings.TrimSpace(manySpaces.ReplaceAllString(s, " "))
}

func Length(s string) int {
	return utf8.RuneCountInString(PlainText(s))
}

func Sanitize(s string) string {
	if PlainText(s) == "" {
		return ""
	}
	return strings.TrimSpace(policy.Sanitize(s))
}
