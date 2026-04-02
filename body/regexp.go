package body

import "regexp"

var (
	rxEventAttrs       = regexp.MustCompile(`\s+on\w+\s*=\s*["'][^"']*["']`)
	rxHTMLComment      = regexp.MustCompile(`(?s)<!--.*?-->`)
	rxSpaceBetweenTags = regexp.MustCompile(`>\s+<`)
	rxMultipleSpaces   = regexp.MustCompile(`[^\S\n]+`)
	rxPositiveClasses  = regexp.MustCompile(`\b(content|article|post|entry|main|reader)\b`)
	rxNegativeClasses  = regexp.MustCompile(`\b(ad|ads|advert|banner|promo|sponsor|nav|icon|menu|breadcrumb|comment|comments|reply|discussion|share|social|facebook|twitter|linkedin|related|similar|recommended|footer|header|sidebar|widget|toolbox|pagination|pager|subscribe|newsletter|cookie|gdpr|feedback|footnotes)\b`)
)
