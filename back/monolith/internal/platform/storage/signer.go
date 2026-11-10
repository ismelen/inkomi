package storage

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var ErrInvalidSignature = errors.New("invalid or expired file url")

// URLSigner builds and checks the urls of /files: an HMAC over the method, the key, the expiry
// and, for uploads, the most bytes allowed, so a url can't be reused for another file, a bigger
// body or after it expires.
type URLSigner struct {
	key     []byte
	baseURL string
	now     func() time.Time
}

func NewURLSigner(key []byte, baseURL string, now func() time.Time) *URLSigner {
	return &URLSigner{key: key, baseURL: strings.TrimRight(baseURL, "/"), now: now}
}

// URL returns the signed url of method on id/filename, valid for ttl; max < 0 means no limit.
func (s *URLSigner) URL(method, id, filename string, ttl time.Duration, max int64) (string, time.Time) {
	exp := s.now().Add(ttl).UTC()
	q := url.Values{}
	q.Set("exp", strconv.FormatInt(exp.Unix(), 10))
	if max >= 0 {
		q.Set("max", strconv.FormatInt(max, 10))
	}
	q.Set("sig", s.sign(method, id, filename, q.Get("exp"), q.Get("max")))
	path := "/files/" + url.PathEscape(id) + "/" + url.PathEscape(filename)
	return s.baseURL + path + "?" + q.Encode(), exp
}

// Verify checks the query of a request to method on id/filename and returns its byte limit,
// -1 when it has none.
func (s *URLSigner) Verify(method, id, filename string, q url.Values) (int64, error) {
	exp, err := strconv.ParseInt(q.Get("exp"), 10, 64)
	if err != nil || s.now().Unix() > exp {
		return 0, ErrInvalidSignature
	}
	want := s.sign(method, id, filename, q.Get("exp"), q.Get("max"))
	if !hmac.Equal([]byte(want), []byte(q.Get("sig"))) {
		return 0, ErrInvalidSignature
	}
	if q.Get("max") == "" {
		return -1, nil
	}
	max, err := strconv.ParseInt(q.Get("max"), 10, 64)
	if err != nil || max < 0 {
		return 0, ErrInvalidSignature
	}
	return max, nil
}

func (s *URLSigner) sign(method, id, filename, exp, max string) string {
	mac := hmac.New(sha256.New, s.key)
	fmt.Fprintf(mac, "%s\n%s\n%s\n%s\n%s", method, id, filename, exp, max)
	return hex.EncodeToString(mac.Sum(nil))
}
