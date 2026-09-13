package libgen

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
)

type ResilientReader struct {
	client     *http.Client
	req        *http.Request
	body       io.ReadCloser
	downloaded int64
	total      int64
	retries    int
	maxRetries int
}

func NewResilientReader(client *http.Client, req *http.Request, initialResp *http.Response) *ResilientReader {
	return &ResilientReader{
		client:     client,
		req:        req,
		body:       initialResp.Body,
		total:      initialResp.ContentLength,
		maxRetries: 5,
	}
}

func (r *ResilientReader) Read(p []byte) (int, error) {
	for {
		n, err := r.body.Read(p)
		r.downloaded += int64(n)

		if err == nil {
			return n, nil
		}

		if err == io.EOF {
			// Check if we actually reached the total expected length, if known
			if r.total > 0 && r.downloaded < r.total {
				err = io.ErrUnexpectedEOF
			} else {
				return n, io.EOF
			}
		}

		// If it's an unexpected EOF or network error, and we have retries left, try to resume
		r.body.Close()
		r.retries++
		if r.retries > r.maxRetries {
			return n, fmt.Errorf("download failed after %d retries: %w", r.maxRetries, err)
		}

		// Reconnect with Range header
		r.req.Header.Set("Range", "bytes="+strconv.FormatInt(r.downloaded, 10)+"-")

		// Optional: add a small delay before retry
		// time.Sleep(1 * time.Second)

		resp, reqErr := r.client.Do(r.req)
		if reqErr != nil {
			err = reqErr
			continue // try again on next loop iteration
		}

		if resp.StatusCode != http.StatusPartialContent && resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			err = fmt.Errorf("unexpected status code on resume: %d", resp.StatusCode)
			continue
		}

		// Success! Swap the body and continue reading
		r.body = resp.Body
		if n > 0 {
			return n, nil
		}
		// If n == 0, we loop again to read from the new body
	}
}

func (r *ResilientReader) Close() error {
	if r.body != nil {
		return r.body.Close()
	}
	return nil
}
