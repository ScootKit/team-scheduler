package discordwebhook

import "testing"

func TestIsValidWebhookURL(t *testing.T) {
	valid := []string{
		"https://discord.com/api/webhooks/123/abcDEF",
		"https://discordapp.com/api/webhooks/123/abcDEF",
		"https://ptb.discord.com/api/webhooks/1/2",
		"https://canary.discord.com/api/webhooks/1/2",
	}
	for _, u := range valid {
		if !IsValidWebhookURL(u) {
			t.Errorf("expected valid: %s", u)
		}
	}

	invalid := []string{
		"",
		"http://discord.com/api/webhooks/1/2", // not https
		"https://evil.com/api/webhooks/1/2",   // wrong host (SSRF/abuse)
		"https://discord.com.evil.com/api/webhooks/1/2", // host suffix trick
		"https://discord.com/webhooks/1/2",              // wrong path
		"https://127.0.0.1/api/webhooks/1/2",            // internal address (SSRF)
		"https://169.254.169.254/api/webhooks/1/2",      // cloud metadata (SSRF)
		"https://discord.com@evil.com/api/webhooks/1/2", // userinfo trick → host is evil.com
		"not a url",
	}
	for _, u := range invalid {
		if IsValidWebhookURL(u) {
			t.Errorf("expected invalid: %s", u)
		}
	}
}
