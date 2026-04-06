package body

import "regexp"

var (
	rxEventAttrs      = regexp.MustCompile(`\s+on\w+\s*=\s*["'][^"']*["']`)
	rxHTMLComment     = regexp.MustCompile(`(?s)<!--.*?-->`)
	rxPositiveClasses = regexp.MustCompile(`\b(content|article|post|entry|main|reader)\b`)
	rxNegativeClasses = regexp.MustCompile(`\b(ad|ads|advert|banner|promo|sponsor|nav|page-nav|post-navigation|icon|menu|breadcrumb|comment|comments|reply|discussion|share|social|facebook|twitter|linkedin|related|similar|recommended|footer|header|sidebar|widget|toolbox|pagination|pager|subscribe|newsletter|cookie|gdpr|feedback|footnotes)\b`)
	rxDoantionText    = regexp.MustCompile(`\b(subscribe|donation|patreon|support|paypal)\b`)
)
