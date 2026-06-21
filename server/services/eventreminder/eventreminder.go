// Package eventreminder runs an in-process poller that posts a "starting now" Discord notification
// when a scheduled event's start time arrives. This is a single-instance scheduler (the deployment
// runs one server); it needs no external job queue.
package eventreminder

import (
	"fmt"
	"strings"
	"time"

	"schej.it/server/db"
	"schej.it/server/logger"
	"schej.it/server/services/discordwebhook"
	"schej.it/server/utils"
)

const (
	pollInterval = 1 * time.Minute
	// graceWindow bounds how late a "starting now" ping may be sent, so a server restart doesn't
	// emit notifications for events whose start was missed long ago.
	graceWindow = 15 * time.Minute
	// deadlineReminderLead is how far before the response deadline the "submit your availability"
	// reminder fires. The reminder is sent on the first poll where the deadline is within this window.
	deadlineReminderLead = 24 * time.Hour
)

// StartPoller launches the background poller. Call once at startup.
func StartPoller() {
	go func() {
		ticker := time.NewTicker(pollInterval)
		defer ticker.Stop()
		for range ticker.C {
			runOnce()
		}
	}()
	logger.StdOut.Println("Event start-notification poller started")
}

func runOnce() {
	defer func() {
		if r := recover(); r != nil {
			logger.StdErr.Printf("eventreminder poller panic: %v", r)
		}
	}()

	now := time.Now().UTC()
	for _, event := range db.GetEventsDueForStartNotification(now, graceWindow) {
		if event.ScheduledEvent == nil {
			continue
		}
		startDate := event.ScheduledEvent.StartDate

		// Post to any of the owner's webhook folders containing this event.
		folders := db.GetWebhookFoldersForEvent(event.Id, event.OwnerId)
		if len(folders) > 0 {
			eventUrl := fmt.Sprintf("%s/e/%s", utils.GetBaseUrl(), utils.Coalesce(event.ShortId))
			desc := "⏰ It's time — the meeting is starting **right now**!"
			if link := strings.TrimSpace(event.MeetingLink); link != "" {
				desc += fmt.Sprintf("\n\n📹 **[Join the meeting now](%s)**", link)
			}
			desc += fmt.Sprintf("\n\n🔗 [View the event](%s)", eventUrl)
			embed := discordwebhook.Embed{
				Title:       fmt.Sprintf("🔴 %s is starting now!", sanitizeTitle(event.Name)),
				Description: desc,
				URL:         eventUrl,
				Color:       discordwebhook.ColorGreen,
			}
			buttons := []discordwebhook.Button{}
			if link := strings.TrimSpace(event.MeetingLink); link != "" {
				buttons = append(buttons, discordwebhook.Button{Label: "📹 Join now", URL: link})
			}
			buttons = append(buttons, discordwebhook.Button{Label: "🔗 View the event", URL: eventUrl})
			for _, folder := range folders {
				discordwebhook.SendEmbed(utils.Coalesce(folder.WebhookUrl), embed, buttons...)
			}
		}

		// Mark as notified regardless (nothing to send if no webhook folder) so we don't re-process.
		db.MarkStartNotified(event.Id, startDate)
	}

	notifyDeadlinesApproaching(now)
}

// notifyDeadlinesApproaching posts a "submit your availability" reminder with a live Discord
// countdown to each webhook folder when an event's response deadline comes within the lead window.
func notifyDeadlinesApproaching(now time.Time) {
	for _, event := range db.GetEventsDueForDeadlineReminder(now, deadlineReminderLead) {
		if event.ResponseDeadline == nil {
			continue
		}
		deadline := *event.ResponseDeadline

		folders := db.GetWebhookFoldersForEvent(event.Id, event.OwnerId)
		if len(folders) > 0 {
			eventUrl := fmt.Sprintf("%s/e/%s", utils.GetBaseUrl(), utils.Coalesce(event.ShortId))
			unix := deadline.Time().Unix()
			// <t:UNIX:R> renders a live relative countdown ("in 23 hours") in Discord.
			desc := fmt.Sprintf(
				"⏳ Voting closes <t:%d:R> — submit your availability before then!\n\n🔗 [Open the event](%s)",
				unix, eventUrl,
			)
			embed := discordwebhook.Embed{
				Title:       fmt.Sprintf("⏰ Last call: %s", sanitizeTitle(event.Name)),
				Description: desc,
				URL:         eventUrl,
				Color:       discordwebhook.ColorOrange,
			}
			button := discordwebhook.Button{Label: "🗳️ Submit availability", URL: eventUrl}
			for _, folder := range folders {
				discordwebhook.SendEmbed(utils.Coalesce(folder.WebhookUrl), embed, button)
			}
		}

		// Mark regardless (nothing to send if no webhook folder) so we don't re-process this deadline.
		db.MarkDeadlineReminderSent(event.Id, deadline)
	}
}

// sanitizeTitle strips control chars and defangs mass-mention tokens from a user-controlled event
// name before it goes into a Discord embed (allowed_mentions also suppresses pings).
func sanitizeTitle(s string) string {
	s = strings.Map(func(r rune) rune {
		if r == '\t' {
			return ' '
		}
		if r < 0x20 || r == 0x7f {
			return -1
		}
		return r
	}, strings.TrimSpace(s))
	s = strings.ReplaceAll(s, "@everyone", "@​everyone")
	s = strings.ReplaceAll(s, "@here", "@​here")
	return s
}
