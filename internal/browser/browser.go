package browser

import (
	"context"
	"fmt"
	"time"

	"github.com/playwright-community/playwright-go"
)

// DefaultTimeout is the default navigation timeout for screenshots.
const DefaultTimeout = 30 * time.Second

// Browser wraps a Playwright browser instance.
type Browser struct {
	pw      *playwright.Playwright
	browser playwright.Browser
	timeout time.Duration
}

// Start initializes Playwright and launches a WebKit instance.
func Start(timeout time.Duration) (*Browser, error) {
	pw, err := playwright.Run()
	if err != nil {
		return nil, fmt.Errorf("starting playwright: %w", err)
	}
	b, err := pw.WebKit.Launch()
	if err != nil {
		_ = pw.Stop()
		return nil, fmt.Errorf("launching webkit: %w", err)
	}
	return &Browser{pw: pw, browser: b, timeout: timeout}, nil
}

// Stop closes the browser and stops Playwright.
func (b *Browser) Stop() {
	if b.browser != nil {
		_ = b.browser.Close()
	}
	if b.pw != nil {
		_ = b.pw.Stop()
	}
}

// TakeScreenshot navigates to url and returns a PNG screenshot.
func (b *Browser) TakeScreenshot(reqCtx context.Context, url string, width, height, delayMs int) ([]byte, error) {
	ctx, err := b.browser.NewContext(playwright.BrowserNewContextOptions{
		Viewport: &playwright.Size{
			Width:  width,
			Height: height,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("creating browser context: %w", err)
	}
	defer func() { _ = ctx.Close() }()

	page, err := ctx.NewPage()
	if err != nil {
		return nil, fmt.Errorf("creating page: %w", err)
	}
	defer func() { _ = page.Close() }()

	if err := reqCtx.Err(); err != nil {
		return nil, err
	}

	_, err = page.Goto(url, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
		Timeout:   playwright.Float(float64(b.timeout.Milliseconds())),
	})
	if err != nil {
		return nil, fmt.Errorf("navigating to %s: %w", url, err)
	}

	if delayMs > 0 {
		select {
		case <-time.After(time.Duration(delayMs) * time.Millisecond):
		case <-reqCtx.Done():
			return nil, reqCtx.Err()
		}
	}

	if err := reqCtx.Err(); err != nil {
		return nil, err
	}

	png, err := page.Screenshot()
	if err != nil {
		return nil, fmt.Errorf("taking screenshot: %w", err)
	}
	return png, nil
}
