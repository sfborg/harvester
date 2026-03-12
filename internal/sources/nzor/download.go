package nzor

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gnames/gn"
	"github.com/sfborg/harvester/internal/sysio"
)

const apiURL = "https://data.nzor.org.nz/v1/names"

// Download fetches all NZOR names pages into a local JSONL file.
// It is resumable: if nzor.jsonl already exists from a previous partial run,
// download continues from the next page rather than starting over.
// If nzor.jsonl is absent (fresh run or wiped by another dataset), the cache
// is reset first so no stale files from other datasets linger.
func (n *nzor) Download() (string, error) {
	if n.cfg.SkipDownload {
		return "", nil
	}

	// nzor.jsonl absent → fresh start; wipe any leftover cache from other datasets.
	// nzor.jsonl present → resuming; preserve it.
	if _, err := os.Stat(n.jsonlPath); os.IsNotExist(err) {
		if err := sysio.ResetCache(n.cfg); err != nil {
			return "", err
		}
	}

	// Done sentinel present → download already complete.
	if _, err := os.Stat(n.donePath); err == nil {
		slog.Info("NZOR data already downloaded, skipping")
		gn.Info("NZOR data already downloaded, skipping")
		return "", nil
	}

	startPage := countValidLines(n.jsonlPath) + 1

	f, err := os.OpenFile(n.jsonlPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return "", fmt.Errorf("opening nzor.jsonl: %w", err)
	}
	defer f.Close()

	ctx := context.Background()
	w := bufio.NewWriter(f)

	slog.Info("downloading NZOR", "startPage", startPage)
	gn.Info(fmt.Sprintf("Downloading NZOR starting from page %d", startPage))

	for page := startPage; ; page++ {
		if page%100 == 0 {
			slog.Info("downloading NZOR", "page", page)
			gn.Info(fmt.Sprintf("Downloading NZOR page %d", page))
		}

		body, err := n.fetchPage(ctx, page)
		if err != nil {
			return "", err
		}

		// Parse before writing — guarantees no partial lines land in the file.
		var resp nzorResponse
		if err := json.Unmarshal(body, &resp); err != nil {
			return "", fmt.Errorf("parsing NZOR page %d: %w", page, err)
		}

		if _, err := fmt.Fprintln(w, string(body)); err != nil {
			return "", fmt.Errorf("writing page %d: %w", page, err)
		}
		if err := w.Flush(); err != nil {
			return "", fmt.Errorf("flushing page %d: %w", page, err)
		}

		if len(resp.Names) == 0 {
			break
		}

		time.Sleep(100 * time.Millisecond)
	}

	if err := os.WriteFile(n.donePath, []byte("done"), 0644); err != nil {
		return "", fmt.Errorf("writing done sentinel: %w", err)
	}

	slog.Info("NZOR download complete")
	gn.Info("NZOR download complete")
	return "", nil
}

func (n *nzor) Extract(_ string) error {
	return nil
}

// fetchPage retrieves a single API page with exponential-backoff retries.
// Transient errors (network failures, 5xx, 429) are retried up to 5 times.
// Client errors (4xx except 429) are returned immediately.
func (n *nzor) fetchPage(ctx context.Context, pageNum int) ([]byte, error) {
	url := fmt.Sprintf("%s?page=%d", apiURL, pageNum)
	const maxRetries = 5
	var lastErr error

	for attempt := range maxRetries {
		if attempt > 0 {
			wait := time.Duration(1<<uint(attempt)) * time.Second
			slog.Warn("NZOR fetch error, retrying",
				"page", pageNum, "attempt", attempt, "wait", wait, "err", lastErr)
			time.Sleep(wait)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Accept", "application/json")

		resp, err := n.http.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			resp.Body.Close()
			wait := time.Duration(1<<uint(attempt)) * time.Second
			if ra := resp.Header.Get("Retry-After"); ra != "" {
				if secs, err := strconv.Atoi(ra); err == nil {
					wait = time.Duration(secs) * time.Second
				}
			}
			lastErr = fmt.Errorf("rate limited (HTTP 429)")
			time.Sleep(wait)
			continue
		}

		if resp.StatusCode >= 500 {
			resp.Body.Close()
			lastErr = fmt.Errorf("server error: HTTP %d", resp.StatusCode)
			continue
		}

		if resp.StatusCode >= 400 {
			resp.Body.Close()
			return nil, fmt.Errorf("client error: HTTP %d", resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("reading response body: %w", err)
			continue
		}

		return body, nil
	}

	return nil, fmt.Errorf("page %d failed after %d attempts: %w", pageNum, maxRetries, lastErr)
}

// countValidLines returns the number of leading lines in path that are valid JSON.
// Stops at the first invalid or empty line. Returns 0 if the file does not exist.
func countValidLines(path string) int {
	f, err := os.Open(path)
	if err != nil {
		return 0
	}
	defer f.Close()

	count := 0
	scanner := bufio.NewScanner(f)
	// 10 MB buffer — NZOR pages can be large.
	scanner.Buffer(make([]byte, 10*1024*1024), 10*1024*1024)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		if json.Valid(line) {
			count++
		} else {
			break
		}
	}
	return count
}
