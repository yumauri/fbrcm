package harness

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

type HoverflyOptions struct {
	Binary          string
	Directory       string
	CertificatePath string
	KeyPath         string
	GuardPath       string
	AllowedRequests []HTTPExpectation
	CassettePath    string
	Capture         bool
	SkipUpstreamTLS bool
}

type Hoverfly struct {
	proxyURL string
	adminURL string
	command  *exec.Cmd
	logs     lockedBuffer
	wait     chan error
	stopOnce sync.Once
	stopErr  error
}

type Journal struct {
	Entries []JournalEntry `json:"journal"`
	Total   int            `json:"total"`
}

type JournalEntry struct {
	Mode    string `json:"mode"`
	Request struct {
		Method      string `json:"method"`
		Destination string `json:"destination"`
		Path        string `json:"path"`
		Query       string `json:"query"`
	} `json:"request"`
	Response struct {
		Status int `json:"status"`
	} `json:"response"`
}

type lockedBuffer struct {
	mu     sync.Mutex
	buffer bytes.Buffer
}

func (b *lockedBuffer) Write(raw []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buffer.Write(raw)
}

func (b *lockedBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buffer.String()
}

func GenerateCA(ctx context.Context, hoverflyBinary, directory string) (certificatePath, keyPath string, returnErr error) {
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return "", "", fmt.Errorf("create Hoverfly CA directory: %w", err)
	}
	proxyPort, err := reservePort()
	if err != nil {
		return "", "", err
	}
	adminPort, err := reservePort()
	if err != nil {
		return "", "", err
	}
	cmd := exec.Command(
		hoverflyBinary,
		"-generate-ca-cert", "-cert-name", "fbrcm-e2e.proxy", "-cert-org", "fbrcm E2E",
		"-pp", strconv.Itoa(proxyPort), "-ap", strconv.Itoa(adminPort), "-listen-on-host", "127.0.0.1",
		"-log-level", "error", "-logs-file", filepath.Join(directory, "hoverfly-ca.log"),
	)
	cmd.Dir = directory
	var combined bytes.Buffer
	cmd.Stdout = &combined
	cmd.Stderr = &combined
	if err := cmd.Start(); err != nil {
		return "", "", fmt.Errorf("start Hoverfly CA generation: %w", err)
	}
	certificatePath = filepath.Join(directory, "cert.pem")
	keyPath = filepath.Join(directory, "key.pem")
	wait := make(chan error, 1)
	go func() { wait <- cmd.Wait() }()
	deadline := time.NewTimer(10 * time.Second)
	ticker := time.NewTicker(25 * time.Millisecond)
	defer deadline.Stop()
	defer ticker.Stop()
	generated := false
	for !generated {
		select {
		case <-ctx.Done():
			_ = cmd.Process.Kill()
			<-wait
			return "", "", fmt.Errorf("generate Hoverfly CA: %w", ctx.Err())
		case err := <-wait:
			return "", "", processExitError("Hoverfly exited while generating CA", err, combined.String())
		case <-deadline.C:
			_ = cmd.Process.Kill()
			<-wait
			return "", "", fmt.Errorf("timed out generating Hoverfly CA\n%s", combined.String())
		case <-ticker.C:
			_, certErr := os.Stat(certificatePath)
			_, keyErr := os.Stat(keyPath)
			if certErr == nil && keyErr == nil {
				_, pairErr := tls.LoadX509KeyPair(certificatePath, keyPath)
				generated = pairErr == nil
			}
		}
	}
	_ = cmd.Process.Signal(os.Interrupt)
	select {
	case <-wait:
	case <-time.After(2 * time.Second):
		_ = cmd.Process.Kill()
		<-wait
	}
	return certificatePath, keyPath, nil
}

func StartHoverfly(ctx context.Context, options HoverflyOptions) (*Hoverfly, error) {
	if err := os.MkdirAll(options.Directory, 0o700); err != nil {
		return nil, fmt.Errorf("create Hoverfly directory: %w", err)
	}
	proxyPort, err := reservePort()
	if err != nil {
		return nil, err
	}
	adminPort, err := reservePort()
	if err != nil {
		return nil, err
	}
	args := []string{
		"-pp", strconv.Itoa(proxyPort),
		"-ap", strconv.Itoa(adminPort),
		"-listen-on-host", "127.0.0.1",
		"-cert", options.CertificatePath,
		"-key", options.KeyPath,
		"-destination", ".",
		"-middleware", options.GuardPath,
		"-log-level", "error",
		"-log-no-color",
		"-logs-file", filepath.Join(options.Directory, "hoverfly.log"),
	}
	if options.Capture {
		args = append(args, "-capture")
	} else {
		args = append(args, "-import", options.CassettePath)
	}
	if options.SkipUpstreamTLS {
		args = append(args, "-tls-verification=false")
	}
	hoverfly := &Hoverfly{
		proxyURL: fmt.Sprintf("http://127.0.0.1:%d", proxyPort),
		adminURL: fmt.Sprintf("http://127.0.0.1:%d", adminPort),
		wait:     make(chan error, 1),
	}
	cmd := exec.Command(options.Binary, args...)
	cmd.Dir = options.Directory
	cmd.Env = hoverflyEnvironment(options.AllowedRequests)
	cmd.Stdout = &hoverfly.logs
	cmd.Stderr = &hoverfly.logs
	hoverfly.command = cmd
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start Hoverfly: %w", err)
	}
	go func() { hoverfly.wait <- cmd.Wait() }()
	if err := hoverfly.waitUntilReady(ctx); err != nil {
		_ = hoverfly.Stop()
		return nil, err
	}
	return hoverfly, nil
}

func (h *Hoverfly) ProxyURL() string {
	return h.proxyURL
}

func (h *Hoverfly) Stop() error {
	if h == nil || h.command == nil || h.command.Process == nil {
		return nil
	}
	h.stopOnce.Do(func() {
		_ = h.command.Process.Signal(os.Interrupt)
		select {
		case err := <-h.wait:
			if err != nil && !strings.Contains(err.Error(), "signal: interrupt") {
				h.stopErr = fmt.Errorf("Hoverfly exited: %w\n%s", err, h.logs.String())
			}
		case <-time.After(2 * time.Second):
			if err := h.command.Process.Kill(); err != nil {
				h.stopErr = fmt.Errorf("kill Hoverfly: %w", err)
				return
			}
			<-h.wait
		}
	})
	return h.stopErr
}

func (h *Hoverfly) Journal(ctx context.Context) (Journal, error) {
	var journal Journal
	if err := h.adminJSON(ctx, http.MethodGet, "/api/v2/journal?sort=timestarted:asc", nil, &journal); err != nil {
		return Journal{}, err
	}
	return journal, nil
}

func (h *Hoverfly) ExportSimulation(ctx context.Context, secrets ...string) ([]byte, error) {
	raw, err := h.adminBytes(ctx, http.MethodGet, "/api/v2/simulation", nil)
	if err != nil {
		return nil, err
	}
	return SanitizeSimulation(raw, secrets...)
}

func (h *Hoverfly) waitUntilReady(ctx context.Context) error {
	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		select {
		case err := <-h.wait:
			return processExitError("Hoverfly exited before becoming ready", err, h.logs.String())
		default:
		}
		requestContext, cancel := context.WithTimeout(ctx, 250*time.Millisecond)
		_, err := h.adminBytes(requestContext, http.MethodGet, "/api/v2/hoverfly", nil)
		cancel()
		if err == nil {
			return nil
		}
		time.Sleep(25 * time.Millisecond)
	}
	return fmt.Errorf("Hoverfly did not become ready\n%s", h.logs.String())
}

func processExitError(message string, err error, logs string) error {
	if err == nil {
		return fmt.Errorf("%s without an error\n%s", message, logs)
	}
	return fmt.Errorf("%s: %w\n%s", message, err, logs)
}

func (h *Hoverfly) adminJSON(ctx context.Context, method, path string, body []byte, target any) error {
	raw, err := h.adminBytes(ctx, method, path, body)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(raw, target); err != nil {
		return fmt.Errorf("decode Hoverfly response from %s: %w", path, err)
	}
	return nil
}

func (h *Hoverfly) adminBytes(ctx context.Context, method, path string, body []byte) ([]byte, error) {
	request, err := http.NewRequestWithContext(ctx, method, h.adminURL+path, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create Hoverfly admin request: %w", err)
	}
	client := &http.Client{Timeout: 2 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("call Hoverfly admin API: %w", err)
	}
	defer func() { _ = response.Body.Close() }()
	raw, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("read Hoverfly admin response: %w", err)
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("Hoverfly admin API returned %s: %s", response.Status, raw)
	}
	return raw, nil
}

func reservePort() (int, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, fmt.Errorf("reserve loopback port: %w", err)
	}
	defer func() { _ = listener.Close() }()
	return listener.Addr().(*net.TCPAddr).Port, nil
}

func hoverflyEnvironment(allowedRequests []HTTPExpectation) []string {
	values := environmentMap(os.Environ())
	for _, key := range []string{"ALL_PROXY", "all_proxy", "HTTP_PROXY", "http_proxy", "HTTPS_PROXY", "https_proxy", "NO_PROXY", "no_proxy"} {
		delete(values, key)
	}
	raw, err := json.Marshal(allowedRequests)
	if err != nil {
		panic(fmt.Sprintf("encode allowed E2E requests: %v", err))
	}
	values["FBRCM_E2E_ALLOWED_REQUESTS"] = string(raw)
	return flattenEnvironment(values)
}

func SanitizeSimulation(raw []byte, secrets ...string) ([]byte, error) {
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, fmt.Errorf("decode captured simulation: %w", err)
	}
	scrubSensitiveKeys(value)
	if root, ok := value.(map[string]any); ok {
		if meta, ok := root["meta"].(map[string]any); ok {
			meta["timeExported"] = fixedTimestamp
		}
	}
	formatted, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("encode sanitized simulation: %w", err)
	}
	formatted = append(formatted, '\n')
	lower := strings.ToLower(string(formatted))
	for _, forbidden := range []string{"private_key", "refresh_token"} {
		if strings.Contains(lower, forbidden) {
			return nil, fmt.Errorf("captured simulation contains forbidden credential field %q", forbidden)
		}
	}
	for _, secret := range secrets {
		if secret != "" && bytes.Contains(formatted, []byte(secret)) {
			return nil, fmt.Errorf("captured simulation contains a supplied secret")
		}
	}
	return formatted, nil
}

func scrubSensitiveKeys(value any) {
	switch typed := value.(type) {
	case map[string]any:
		for key, child := range typed {
			switch strings.ToLower(key) {
			case "authorization", "proxy-authorization", "set-cookie", "access_token", "refresh_token", "private_key":
				delete(typed, key)
			default:
				scrubSensitiveKeys(child)
			}
		}
	case []any:
		for _, child := range typed {
			scrubSensitiveKeys(child)
		}
	}
}
