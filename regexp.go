package parser

import "regexp"

var (
	rxJsonLdArticleTypes = regexp.MustCompile(`(?i)^Article|AdvertiserContentArticle|NewsArticle|AnalysisNewsArticle|AskPublicNewsArticle|BackgroundNewsArticle|OpinionNewsArticle|ReportageNewsArticle|ReviewNewsArticle|Report|SatiricalArticle|ScholarlyArticle|MedicalScholarlyArticle|SocialMediaPosting|BlogPosting|LiveBlogPosting|DiscussionForumPosting|TechArticle|APIReference$`)
	rxCDATA              = regexp.MustCompile(`^\s*<!\[CDATA\[|\]\]>\s*$`)
	rxSchemaOrg          = regexp.MustCompile(`(?i)^https?\:\/\/schema\.org\/?$`)
	rxEventAttrs         = regexp.MustCompile(`\s+on\w+\s*=\s*["'][^"']*["']`)
	rxURL                = regexp.MustCompile(`^https?:\/\/(www\.)?[-a-zA-Z0-9@:%._\+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}\b([-a-zA-Z0-9()@:%_\+.~#?&//=]*)$`)
	rxHTMLComment        = regexp.MustCompile(`(?s)<!--.*?-->`)
	rxSpaceBetweenTags   = regexp.MustCompile(`>\s+<`)
	rxMultipleSpaces     = regexp.MustCompile(`[^\S\n]+`)
	rxPositiveClasses    = regexp.MustCompile(`\b(content|article|post|entry|main|reader)\b`)
	rxNegativeClasses    = regexp.MustCompile(`\b(ad|ads|advert|banner|promo|sponsor|nav|icon|menu|breadcrumb|comment|comments|reply|discussion|share|social|facebook|twitter|linkedin|related|similar|recommended|footer|header|sidebar|widget|toolbox|pagination|pager|subscribe|newsletter|cookie|gdpr|feedback|footnotes)\b`)
)
