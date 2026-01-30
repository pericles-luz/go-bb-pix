package bbpix

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/pericles-luz/go-bb-pix/internal/auth"
	"github.com/pericles-luz/go-bb-pix/internal/transport"
	"github.com/pericles-luz/go-bb-pix/pix"
	"github.com/pericles-luz/go-bb-pix/pixauto"
)

// Client is the main client for the Banco do Brasil PIX API
type Client struct {
	config     Config
	httpClient *http.Client
	apiURL     string
	oauthURL   string

	// Lazy-initialized clients
	pixClient     *pix.Client
	pixAutoClient *pixauto.Client
	mu            sync.Mutex
}

// New creates a new Banco do Brasil PIX client
func New(config Config, opts ...Option) (*Client, error) {
	// Validate config
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Apply options
	options := defaultClientOptions()
	for _, opt := range opts {
		opt(options)
	}

	// Get environment URLs
	oauthURL, apiURL := config.Environment.URLs()

	// Create client
	client := &Client{
		config:   config,
		apiURL:   apiURL,
		oauthURL: oauthURL,
	}

	// Build HTTP client with transport chain
	client.httpClient = client.buildHTTPClient(options)

	return client, nil
}

// buildHTTPClient builds an HTTP client with the transport chain
func (c *Client) buildHTTPClient(opts *clientOptions) *http.Client {
	// Start with base transport or custom HTTP client
	var baseTransport http.RoundTripper
	if opts.httpClient != nil {
		baseTransport = opts.httpClient.Transport
		if baseTransport == nil {
			baseTransport = http.DefaultTransport
		}
	} else {
		baseTransport = http.DefaultTransport
	}

	// Create OAuth2 token provider
	tokenProvider := auth.NewOAuth2Provider(c.oauthURL, c.config.ClientID, c.config.ClientSecret)

	// Build transport chain (innermost to outermost):
	// 1. Base transport
	// 2. Circuit breaker (fail-fast protection)
	// 3. Retry (exponential backoff)
	// 4. Auth (inject OAuth2 token)
	// 5. Logging (log requests/responses)

	// Apply circuit breaker
	var currentTransport http.RoundTripper = transport.NewCircuitBreakerTransport(
		baseTransport,
		opts.circuitBreakerMaxFailures,
		opts.circuitBreakerResetTimeout,
	)

	// Apply retry
	currentTransport = transport.NewRetryTransport(
		currentTransport,
		opts.maxRetries,
		opts.initialBackoff,
	)

	// Apply auth
	currentTransport = transport.NewAuthTransport(
		currentTransport,
		tokenProvider,
		c.config.DeveloperAppKey,
	)

	// Apply logging
	currentTransport = transport.NewLoggingTransport(
		currentTransport,
		opts.logger,
	)

	// Create HTTP client with configured transport and timeout
	return &http.Client{
		Transport: currentTransport,
		Timeout:   opts.timeout,
	}
}

// PIX returns the PIX client
// The client is lazily initialized and cached
func (c *Client) PIX() *pix.Client {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.pixClient == nil {
		c.pixClient = pix.NewClient(c.httpClient, c.apiURL)
	}

	return c.pixClient
}

// PIXAuto returns the PIX Automático client
// The client is lazily initialized and cached
func (c *Client) PIXAuto() *pixauto.Client {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.pixAutoClient == nil {
		c.pixAutoClient = pixauto.NewClient(c.httpClient, c.apiURL)
	}

	return c.pixAutoClient
}
