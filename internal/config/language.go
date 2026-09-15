package config

import "strings"

var languageNames = map[string]string{
	"en": "English",
	"ru": "Russian",
	"uk": "Ukrainian",
	"be": "Belarusian",
	"de": "German",
	"fr": "French",
	"es": "Spanish",
	"it": "Italian",
	"pt": "Portuguese",
	"pl": "Polish",
	"nl": "Dutch",
	"sv": "Swedish",
	"no": "Norwegian",
	"nb": "Norwegian Bokmål",
	"da": "Danish",
	"fi": "Finnish",
	"cs": "Czech",
	"sk": "Slovak",
	"hu": "Hungarian",
	"ro": "Romanian",
	"bg": "Bulgarian",
	"el": "Greek",
	"tr": "Turkish",
	"ar": "Arabic",
	"he": "Hebrew",
	"fa": "Persian",
	"hi": "Hindi",
	"zh": "Chinese",
	"ja": "Japanese",
	"ko": "Korean",
	"th": "Thai",
	"vi": "Vietnamese",
	"id": "Indonesian",
	"kk": "Kazakh",
}

func LanguageLabel(target string) string {
	trimmed := strings.TrimSpace(target)
	if trimmed == "" {
		return trimmed
	}
	key := strings.ToLower(trimmed)
	if name, ok := languageNames[key]; ok {
		return name + " (" + trimmed + ")"
	}
	base := key
	if i := strings.IndexAny(base, "-_"); i >= 0 {
		base = base[:i]
	}
	if name, ok := languageNames[base]; ok {
		return name + " (" + trimmed + ")"
	}
	return trimmed
}
