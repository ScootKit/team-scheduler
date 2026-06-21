// Package discordwebhook sends notifications to Discord webhook URLs configured on folders.
package discordwebhook

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"

	"schej.it/server/logger"
)

// Brand colors for embeds (decimal RGB).
const (
	ColorGreen = 0x00994C // scheduled / confirmed
	ColorBlue  = 0x3B82F6 // new event / call to action
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

// Embed is a minimal Discord embed. Title is rendered as a clickable link when URL is set, which
// serves as the call-to-action (standard channel webhooks cannot send interactive buttons).
type Embed struct {
	Title       string
	Description string
	URL         string
	Color       int
}

// SendEmbed posts a single embed to the given Discord webhook URL. It is a no-op (logged) if the URL
// is not a valid Discord webhook. allowed_mentions is set to suppress @everyone/@here/role pings so a
// crafted event or folder name can never ping a channel. Intended to be called from a goroutine.
func SendEmbed(webhookURL string, embed Embed) {
	if !IsValidWebhookURL(webhookURL) {
		logger.StdErr.Printf("discordwebhook: refusing to send to non-Discord URL")
		return
	}

	// Discord limits: embed title 256, description 4096.
	if len(embed.Title) > 256 {
		embed.Title = embed.Title[:256]
	}
	if len(embed.Description) > 4096 {
		embed.Description = embed.Description[:4096]
	}

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
		"embeds": []any{e},
		// Never allow mentions to be parsed from webhook content.
		"allowed_mentions": map[string]any{"parse": []string{}},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		logger.StdErr.Printf("discordwebhook: marshal failed: %v", err)
		return
	}

	resp, err := client.Post(webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		logger.StdErr.Printf("discordwebhook: post failed: %v", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		logger.StdErr.Printf("discordwebhook: webhook returned status %d", resp.StatusCode)
	}
}
