package utils

import (
	"regexp"
)

var (
	ReBrowserIE      = regexp.MustCompile("MSIE|Trident|InternetExplorer|Edge")
	ReBrowserSafari  = regexp.MustCompile("Safari")
	ReBrowserFirefox = regexp.MustCompile("Firefox")
)

func IsBrowserIE(userAgent string) bool {
	return ReBrowserIE.MatchString(userAgent)
}

func IsBrowserSafari(userAgent string) bool {
	return ReBrowserSafari.MatchString(userAgent)
}

func IsBrowserFireFox(userAgent string) bool {
	return ReBrowserFirefox.MatchString(userAgent)
}
