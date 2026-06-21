// Package discordwebhook sends notifications to Discord webhook URLs configured on folders.
package discordwebhook

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"schej.it/server/logger"
)

// Brand colors for embeds (decimal RGB).
const (
	ColorGreen  = 0x00994C // scheduled / confirmed
	ColorBlue   = 0x3B82F6 // new event / call to action
	ColorOrange = 0xF59E0B // reminder / deadline approaching
)

// allowedHosts is the set of hosts a folder webhook URL may point at. Restricting to Discord's own
// hosts prevents this from being abused as a generic outbound poster / SSRF primitive (a folder owner
// can't point the "webhook" at an internal address).
var allowedHosts = map[string]bool{
	"discord.com":        true,
	"discordapp.com":     true,
	"ptb.discord.com":    true,
	"canary.discord.com": true,
}

// IsValidWebhookURL reports whether s is a well-formed Discord webhook URL (https, an allowed Discord
// host, and the /api/webhooks/ path). Used to validate user-supplied folder webhook URLs.
func IsValidWebhookURL(s string) bool {
	u, err := url.Parse(strings.TrimSpace(s))
	if err != nil {
		return false
	}
	if u.Scheme != "https" {
		return false
	}
	if !allowedHosts[strings.ToLower(u.Hostname())] {
		return false
	}
	return strings.HasPrefix(u.Path, "/api/webhooks/")
}

var client = &http.Client{Timeout: 10 * time.Second}

// Embed is a minimal Discord embed. Title is rendered as a clickable link when URL is set.
type Embed struct {
	Title       string
	Description string
	URL         string
	Color       int
}

// Button is a Discord link button (component style 5) — a non-interactive button that opens a URL.
// Link buttons are the only component type that works without an interaction backend.
type Button struct {
	Label string
	URL   string
}

// SendEmbed posts an embed (and optional link buttons) to the given Discord webhook URL. It is a
// no-op (logged) if the URL is not a valid Discord webhook. allowed_mentions suppresses
// @everyone/@here/role pings so a crafted event/folder name can never ping a channel.
//
// Link buttons: a plain (non-application) channel webhook can include non-interactive link buttons
// only when the request is sent with the `with_components=true` query parameter — without it Discord
// silently strips the components. We send embed + components together with that param, so the message
// has both a rich embed and clickable buttons. Intended to be called from a goroutine.
func SendEmbed(webhookURL string, embed Embed, buttons ...Button) {
	if !IsValidWebhookURL(webhookURL) {
		logger.StdErr.Printf("discordwebhook: refusing to send to non-Discord URL")
		return
	}

	// Discord limits: title 256, description 4096 (rune-safe).
	embed.Title = truncateRunes(embed.Title, 256)
	embed.Description = truncateRunes(embed.Description, 4096)

	e := map[string]any{}
	if embed.Title != "" {
		e["title"] = embed.Title
	}
	if embed.Description != "" {
		e["description"] = embed.Description
	}
	if embed.URL != "" {
		e["url"] = embed.URL
	}
	if embed.Color != 0 {
		e["color"] = embed.Color
	}

	payload := map[string]any{
		"embeds":           []any{e},
		"allowed_mentions": map[string]any{"parse": []string{}},
	}

	// Build the link-button action row (max 5 buttons), if any.
	row := []any{}
	for i, b := range buttons {
		if i >= 5 || strings.TrimSpace(b.URL) == "" {
			continue
		}
		row = append(row, map[string]any{
			"type":  2, // button
			"style": 5, // link
			"label": truncateRunes(b.Label, 80),
			"url":   b.URL,
		})
	}

	url := webhookURL
	if len(row) > 0 {
		payload["components"] = []any{map[string]any{"type": 1, "components": row}}
		// Required so a non-app channel webhook honors (rather than strips) the components.
		if strings.Contains(url, "?") {
			url += "&with_components=true"
		} else {
			url += "?with_components=true"
		}
	}

	post(url, payload)
}

// post marshals + sends the payload, returning true on a 2xx response.
func post(webhookURL string, payload map[string]any) bool {
	body, err := json.Marshal(payload)
	if err != nil {
		logger.StdErr.Printf("discordwebhook: marshal failed: %v", err)
		return false
	}
	resp, err := client.Post(webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		// Avoid logging the full URL (it contains the secret webhook token); url.Error embeds it.
		logger.StdErr.Printf("discordwebhook: post failed (transport error)")
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		// Log Discord's error body (it explains exactly what was rejected) so payload-structure
		// problems are debuggable. The body never contains the webhook token.
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		logger.StdErr.Printf("discordwebhook: status %d, response: %s", resp.StatusCode, string(respBody))
		return false
	}
	return true
}

func truncateRunes(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max])
}
