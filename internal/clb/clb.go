// Package clb provides a client for the ChecklistBank API.
// It handles authentication, triggering dataset exports, polling
// for completion, and downloading the resulting archives. It also
// provides helpers for converting downloaded COLDP archives to SFGA.
package clb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gnames/gn"
)

// Client communicates with the ChecklistBank API.
type Client struct {
	api      string
	user     string
	password string
	token    string
	http     *http.Client
}

// ExportRequest describes parameters for a CLB dataset export.
type ExportRequest struct {
	Format         string `json:"format"`
	Synonyms       bool   `json:"synonyms"`
	BareNames      bool   `json:"bareNames"`
	Extended       bool   `json:"extended"`
	Extinct        *bool  `json:"extinct"`
	Classification bool   `json:"classification"`
	TaxGroups      bool   `json:"taxGroups"`
	Root           *Root  `json:"root,omitempty"`
	MinRank        string `json:"minRank,omitempty"`
	TabFormat      string `json:"tabFormat,omitempty"`
}

// Root specifies the root taxon for a filtered export.
type Root struct {
	ID string `json:"id"`
}

// ExportStatus represents the status of a CLB export job.
type ExportStatus struct {
	Key      string `json:"key"`
	Status   string `json:"status"`
	Error    string `json:"error"`
	Download string `json:"download"`
	Request  any    `json:"request"`
}

// New creates a ChecklistBank API client.
func New(api, user, password string) *Client {
	return &Client{
		api:      strings.TrimRight(api, "/"),
		user:     user,
		password: password,
		http: &http.Client{
			Timeout: 5 * time.Minute,
		},
	}
}

// Login authenticates with the ChecklistBank API. It first tries to
// load a cached token from the config directory. If no valid cached
// token is found, it performs a fresh login using HTTP Basic Auth
// and caches the new token.
func (c *Client) Login() error {
	// Try cached token first.
	if token, ok := loadCachedToken(c.api, c.user); ok {
		c.token = token
		gn.Info("Using cached ChecklistBank token")
		return nil
	}

	url := c.api + "/user/login"
	slog.Info("logging in to ChecklistBank", "url", url)
	gn.Info("Logging in to ChecklistBank")

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("creating login request: %w", err)
	}
	req.SetBasicAuth(c.user, c.password)

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("login request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading login response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf(
			"login failed (HTTP %d): %s", resp.StatusCode, string(body),
		)
	}

	c.token = strings.Trim(string(body), " \t\n\r\"")
	if c.token == "" {
		return fmt.Errorf("login returned empty token")
	}

	saveCachedToken(c.api, c.user, c.token)
	slog.Info("ChecklistBank login successful")
	return nil
}

// freshLogin forces a new login, bypassing the cache. Used when a
// cached token turns out to be rejected by the server.
func (c *Client) freshLogin() error {
	slog.Info("cached token rejected, performing fresh login")
	gn.Info("Token expired, logging in again")

	// Delete stale cache.
	if path, err := tokenPath(); err == nil {
		os.Remove(path)
	}

	c.token = ""
	url := c.api + "/user/login"

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("creating login request: %w", err)
	}
	req.SetBasicAuth(c.user, c.password)

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("login request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading login response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf(
			"login failed (HTTP %d): %s", resp.StatusCode, string(body),
		)
	}

	c.token = strings.Trim(string(body), " \t\n\r\"")
	if c.token == "" {
		return fmt.Errorf("login returned empty token")
	}

	saveCachedToken(c.api, c.user, c.token)
	return nil
}

// TriggerExport starts an asynchronous export for the given dataset
// and returns the export UUID. If the server returns 401, it
// performs a fresh login and retries once.
func (c *Client) TriggerExport(
	datasetID int,
	er ExportRequest,
) (string, error) {
	key, err := c.triggerExport(datasetID, er)
	if err != nil && strings.Contains(err.Error(), "HTTP 401") {
		if loginErr := c.freshLogin(); loginErr != nil {
			return "", loginErr
		}
		return c.triggerExport(datasetID, er)
	}
	return key, err
}

func (c *Client) triggerExport(
	datasetID int,
	er ExportRequest,
) (string, error) {
	url := fmt.Sprintf("%s/dataset/%d/export", c.api, datasetID)
	slog.Info("triggering CLB export", "dataset", datasetID, "url", url)
	gn.Info("Triggering ChecklistBank export for dataset %d", datasetID)

	body, err := json.Marshal(er)
	if err != nil {
		return "", fmt.Errorf("marshaling export request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("creating export request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("export request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading export response: %w", err)
	}

	if resp.StatusCode != http.StatusOK &&
		resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf(
			"export trigger failed (HTTP %d): %s",
			resp.StatusCode, string(respBody),
		)
	}

	exportKey := strings.Trim(string(respBody), " \t\n\r\"")
	if exportKey == "" {
		return "", fmt.Errorf("export returned empty key")
	}

	slog.Info("CLB export triggered", "exportKey", exportKey)
	gn.Info("Export triggered, key: %s", exportKey)
	return exportKey, nil
}

// WaitForExport polls the export status until the export finishes,
// fails, or is canceled. It returns the download URL on success.
func (c *Client) WaitForExport(exportKey string) (string, error) {
	url := fmt.Sprintf("%s/export/%s", c.api, exportKey)
	slog.Info("waiting for CLB export", "exportKey", exportKey)
	gn.Info("Waiting for export to complete...")

	pollInterval := 10 * time.Second
	maxWait := 30 * time.Minute
	start := time.Now()

	for {
		if time.Since(start) > maxWait {
			return "", fmt.Errorf(
				"export %s timed out after %v", exportKey, maxWait,
			)
		}

		time.Sleep(pollInterval)

		es, err := c.getExportStatus(url)
		if err != nil {
			slog.Warn("error polling export status", "error", err)
			continue
		}

		st := strings.ToLower(es.Status)
		slog.Info("export status", "key", exportKey, "status", st)

		switch st {
		case "finished":
			gn.Info("Export finished")
			return es.Download, nil
		case "failed":
			return "", fmt.Errorf(
				"export %s failed: %s", exportKey, es.Error,
			)
		case "canceled":
			return "", fmt.Errorf("export %s was canceled", exportKey)
		case "waiting", "blocked", "running":
			gn.Info("Export status: %s", es.Status)
		default:
			slog.Warn("unknown export status", "status", es.Status)
		}
	}
}

// DownloadExport downloads a completed export archive from downloadURL
// to destDir and returns the path to the downloaded file.
func (c *Client) DownloadExport(
	downloadURL, destDir string,
) (string, error) {
	slog.Info("downloading CLB export", "url", downloadURL)
	gn.Info("Downloading export from %s", downloadURL)

	req, err := http.NewRequest(http.MethodGet, downloadURL, nil)
	if err != nil {
		return "", fmt.Errorf("creating download request: %w", err)
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("download request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf(
			"download failed (HTTP %d): %s",
			resp.StatusCode, string(body),
		)
	}

	fname := filepath.Base(downloadURL)
	if fname == "" || fname == "." || fname == "/" {
		fname = "export.zip"
	}
	outPath := filepath.Join(destDir, fname)
	f, err := os.Create(outPath)
	if err != nil {
		return "", fmt.Errorf("creating output file: %w", err)
	}
	defer f.Close()

	written, err := io.Copy(f, resp.Body)
	if err != nil {
		return "", fmt.Errorf("writing export file: %w", err)
	}

	slog.Info("export downloaded", "path", outPath, "bytes", written)
	gn.Info("Export downloaded: %s (%d bytes)", outPath, written)
	return outPath, nil
}

// FetchDatasetAlias returns a short name for the dataset suitable for
// use as an output file name. It first checks the dataset's alias
// field; if empty, it derives an acronym from the first letter of
// each word in the title.
func (c *Client) FetchDatasetAlias(datasetID int) (string, error) {
	url := fmt.Sprintf("%s/dataset/%d.json", c.api, datasetID)
	slog.Info("fetching dataset metadata", "dataset", datasetID)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("creating dataset request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("dataset request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf(
			"dataset fetch failed (HTTP %d): %s",
			resp.StatusCode, string(body),
		)
	}

	var ds datasetMeta
	if err := json.NewDecoder(resp.Body).Decode(&ds); err != nil {
		return "", fmt.Errorf("decoding dataset metadata: %w", err)
	}

	if ds.Alias != "" {
		slog.Info("using dataset alias", "alias", ds.Alias)
		return sanitizeAlias(ds.Alias), nil
	}

	// Derive acronym from first letter of each word in the title.
	alias := acronymFromTitle(ds.Title)
	slog.Info("derived alias from title", "title", ds.Title, "alias", alias)
	return alias, nil
}

type datasetMeta struct {
	Alias string `json:"alias"`
	Title string `json:"title"`
}

// sanitizeAlias lowercases an alias and replaces spaces with
// underscores so it is safe to use as a file name.
func sanitizeAlias(alias string) string {
	alias = strings.ToLower(alias)
	alias = strings.ReplaceAll(alias, " ", "_")
	return alias
}

// acronymFromTitle takes the first letter of each word in a title
// and returns it lowercased as a short identifier.
func acronymFromTitle(title string) string {
	words := strings.Fields(title)
	var b strings.Builder
	for _, w := range words {
		if len(w) > 0 {
			b.WriteByte(w[0])
		}
	}
	return strings.ToLower(b.String())
}

// SearchTaxon searches for a taxon by name within a dataset and
// returns the taxon ID of the first accepted match. The rank and
// status parameters narrow the search. This resolves taxon names
// to their current IDs, which may change over time.
func (c *Client) SearchTaxon(
	datasetID int,
	name, rank, status string,
) (string, error) {
	url := fmt.Sprintf(
		"%s/dataset/%d/nameusage/search?q=%s",
		c.api, datasetID, name,
	)
	if rank != "" {
		url += "&rank=" + rank
	}
	if status != "" {
		url += "&status=" + status
	}
	slog.Info("searching CLB taxon", "dataset", datasetID, "name", name)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("creating search request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("search request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf(
			"taxon search failed (HTTP %d): %s",
			resp.StatusCode, string(body),
		)
	}

	var result searchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decoding search response: %w", err)
	}

	// Look for the first exact name match.
	for _, r := range result.Result {
		if strings.EqualFold(r.Usage.Name.ScientificName, name) {
			slog.Info(
				"found taxon",
				"name", name,
				"id", r.Usage.ID,
				"rank", r.Usage.Name.Rank,
			)
			gn.Info(
				"Resolved taxon '%s' to ID '%s'", name, r.Usage.ID,
			)
			return r.Usage.ID, nil
		}
	}

	// Fall back to first result if no exact name match.
	if len(result.Result) > 0 {
		id := result.Result[0].Usage.ID
		slog.Info(
			"using first search result for taxon",
			"name", name, "id", id,
		)
		gn.Info("Resolved taxon '%s' to ID '%s' (best match)", name, id)
		return id, nil
	}

	return "", fmt.Errorf(
		"taxon '%s' not found in dataset %d", name, datasetID,
	)
}

type searchResponse struct {
	Result []searchResult `json:"result"`
	Total  int            `json:"total"`
}

type searchResult struct {
	Usage searchUsage `json:"usage"`
}

type searchUsage struct {
	ID     string     `json:"id"`
	Name   searchName `json:"name"`
	Status string     `json:"status"`
}

type searchName struct {
	ScientificName string `json:"scientificName"`
	Rank           string `json:"rank"`
}

// getExportStatus fetches the current status of an export job.
func (c *Client) getExportStatus(url string) (*ExportStatus, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	var status ExportStatus
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, fmt.Errorf("decoding export status: %w", err)
	}

	return &status, nil
}

// FindColdpZip looks for a COLDP zip archive in the given directory.
func FindColdpZip(dir string) (string, error) {
	matches, _ := filepath.Glob(filepath.Join(dir, "*.zip"))
	if len(matches) == 0 {
		return "", fmt.Errorf("no COLDP archive found in %s", dir)
	}
	return matches[0], nil
}
