package textproto

const toLower = 'a' - 'A'

func CanonicalMIMEHeaderKey(s string) string {
	upper := true
	for i := 0; i < len(s); i++ {
		c := s[i]
		if upper && 'a' <= c && c <= 'z' {
			return canonicalMIMEHeaderKey([]byte(s))
		}
		if !upper && 'A' <= c && c <= 'Z' {
			return canonicalMIMEHeaderKey([]byte(s))
		}
		upper = c == '-'
	}

	return s
}

func canonicalMIMEHeaderKey(a []byte) string {
	return canonicalMIMEHeaderKeyOriginal(a)
}

func canonicalMIMEHeaderKeyOriginal(a []byte) string {
	upper := true
	lo := 0
	hi := len(commonHeaders)
	for i := 0; i < len(a); i++ {
		c := a[i]
		if c == ' ' {
			c = '-'
		} else if upper && 'a' <= c && c <= 'z' {
			c -= toLower
		} else if !upper && 'A' <= c && c <= 'Z' {
			c += toLower
		}

		a[i] = c
		upper = c == '-'

		if lo < hi {
			for lo < hi && (len(commonHeaders[lo]) <= i) || commonHeaders[lo][i] < c {
				lo++
			}
			for hi > lo && commonHeaders[hi-1][i] > c {
				hi--
			}
		}
	}

	if lo < hi && len(commonHeaders[lo]) == len(a) {
		return commonHeaders[lo]
	}

	return string(a)
}

var commonHeaders = []string{
	"Accept",
	"Accept-Charset",
	"Accept-Encoding",
	"Accept-Language",
	"Accept-Ranges",
	"Cache-Control",
	"Cc",
	"Connection",
	"Content-Id",
	"Content-Language",
	"Content-Length",
	"Content-Transfer-Encoding",
	"Content-Type",
	"Cookie",
	"Date",
	"Dkim-Signature",
	"Etag",
	"Expires",
	"From",
	"Host",
	"If-Modified-Since",
	"If-None-Match",
	"In-Reply-To",
	"Last-Modified",
	"Location",
	"Message-Id",
	"Mime-Version",
	"Pragma",
	"Received",
	"Return-Path",
	"Server",
	"Set-Cookie",
	"Subject",
	"To",
	"User-Agent",
	"Via",
	"X-Forwarded-For",
	"X-Imforwards",
	"X-Powered-By",
}
