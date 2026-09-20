package downloader

import (
	"io"
	"math"
	"net/http"
	"time"
)

type Transport struct {
	inner  http.RoundTripper
	maxBps float64
	minBps float64
}

func NewTransport(inner http.RoundTripper, maxMbps, minMbps float64) *Transport {
	t := &Transport{
		inner:  inner,
		maxBps: maxMbps * 1024 * 1024 / 8,
		minBps: minMbps * 1024 * 1024 / 8,
	}
	if t.inner == nil {
		t.inner = http.DefaultTransport
	}
	return t
}

type slowReader struct {
	reader io.ReadCloser
	maxBps float64
	minBps float64
	start  time.Time
	read   int64
}

func (s *slowReader) Close() error {
	return s.reader.Close()
}

func (s *slowReader) Read(p []byte) (int, error) {
	n, err := s.reader.Read(p)
	if n > 0 {
		s.read += int64(n)
		elapsed := time.Since(s.start).Seconds()
		if elapsed > 0 {
			bps := float64(s.read) / elapsed
			var targetBps float64
			if s.maxBps > 0 && bps > s.maxBps {
				targetBps = s.maxBps
			} else if s.minBps > 0 && bps > s.minBps {
				targetBps = s.minBps
			}
			if targetBps > 0 && bps > targetBps {
				expected := time.Duration(float64(s.read)/targetBps*1000) * time.Millisecond
				wait := expected - time.Since(s.start)
				if wait > 0 {
					time.Sleep(wait)
				}
			}
		}
	}
	return n, err
}

func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := t.inner.RoundTrip(req)
	if err != nil {
		return resp, err
	}

	if t.maxBps > 0 || t.minBps > 0 {
		if resp.Body != nil {
			maxBps := math.Max(t.maxBps, t.minBps)
			minBps := math.Min(t.maxBps, t.minBps)
			if maxBps == 0 {
				maxBps = math.MaxFloat64
			}
			resp.Body = &slowReader{
				reader: resp.Body,
				maxBps: maxBps,
				minBps: minBps,
				start:  time.Now(),
			}
		}
	}

	return resp, nil
}
