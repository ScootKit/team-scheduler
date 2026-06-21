package calendar

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/emersion/go-ical"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"schej.it/server/models"
	"schej.it/server/utils"
)

// maxICSFeedBytes caps how much of a user-supplied ICS feed we will read. Without this, a malicious
// feed can drive unbounded memory use, or deeply-nested BEGIN blocks that make the go-ical decoder
// recurse until the process hits a fatal (unrecoverable) stack overflow. A few MB is plenty for a
// real calendar feed; capping the input bounds both DoS vectors.
const maxICSFeedBytes = 8 << 20 // 8 MiB

type ICSCalendar struct {
	models.ICSCalendarAuth
}

func (cal *ICSCalendar) GetCalendarList() (map[string]models.SubCalendar, error) {
	return map[string]models.SubCalendar{
		"default": {
			Name:    cal.Label,
			Enabled: utils.TruePtr(),
		},
	}, nil
}

func (cal *ICSCalendar) GetCalendarEvents(calendarId string, timeMin time.Time, timeMax time.Time) ([]models.CalendarEvent, error) {
	// Fetch the data and ensure the fetch was successful. The feed URL is
	// user-supplied, so we use the SSRF-safe client (see safe_http.go) to keep
	// the request from reaching internal/metadata addresses.
	resp, err := safeGet(cal.FeedURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch ICS feed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ICS feed returned status %d", resp.StatusCode)
	}

	// Parse, but bound how much we read so a hostile feed can't exhaust memory or recurse the
	// decoder into a fatal stack overflow.
	decoder := ical.NewDecoder(io.LimitReader(resp.Body, maxICSFeedBytes))
	parsedCal, err := decoder.Decode()
	if err != nil {
		return nil, fmt.Errorf("failed to parse ICS data: %v", err)
	}

	var events []models.CalendarEvent

	// Loop through parsed data and append each to events array
	for _, component := range parsedCal.Children {
		if component.Name != ical.CompEvent {
			continue
		}

		summary := component.Props.Get(ical.PropSummary)
		uid := component.Props.Get(ical.PropUID)
		dtStart := component.Props.Get(ical.PropDateTimeStart)
		dtEnd := component.Props.Get(ical.PropDateTimeEnd)

		if dtStart == nil || dtEnd == nil {
			continue
		}

		allDay := false
		var startTime, endTime time.Time

		// Check that event is not all day
		if !strings.Contains(dtStart.Value, "T") {
			startTime, err = time.Parse("20060102", dtStart.Value)
			if err != nil {
				continue
			}
			endTime, err = time.Parse("20060102", dtEnd.Value)
			if err != nil {
				continue
			}
			allDay = true
		} else {
			startTime, err = parseTimeWithTZ(dtStart)
			if err != nil {
				continue
			}

			endTime, err = parseTimeWithTZ(dtEnd)
			if err != nil {
				continue
			}
		}

		if endTime.Before(timeMin) || startTime.After(timeMax) {
			continue
		}

		free := false
		if transp := component.Props.Get(ical.PropTransparency); transp != nil {
			// TRANSPARENT means free, OPAQUE means busy
			free = strings.EqualFold(transp.Value, "TRANSPARENT")
		}

		summaryStr := ""
		if summary != nil {
			summaryStr = summary.Value
		}

		uidStr := ""
		if uid != nil {
			uidStr = uid.Value
		}

		events = append(events, models.CalendarEvent{
			Id:         uidStr,
			CalendarId: calendarId,
			Summary:    summaryStr,
			StartDate:  primitive.NewDateTimeFromTime(startTime),
			EndDate:    primitive.NewDateTimeFromTime(endTime),
			Free:       free,
			AllDay:     allDay,
		})
	}

	return events, nil
}

// parseTimeWithTZ parses an ICS datetime property, honoring a TZID parameter
// when present and otherwise treating the value as UTC.
func parseTimeWithTZ(prop *ical.Prop) (time.Time, error) {
	if prop == nil {
		return time.Time{}, fmt.Errorf("missing datetime property")
	}

	timeStr := prop.Value
	tzID := prop.Params.Get("TZID")

	var t time.Time
	var err error

	if tzID != "" {
		loc, err := time.LoadLocation(tzID)
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid timezone: %v", err)
		}
		//lint:ignore SA4006 err is in fact used later in the code
		t, err = time.ParseInLocation("20060102T150405", timeStr, loc)
	} else {
		t, err = time.Parse("20060102T150405Z", timeStr)
	}

	if err != nil {
		return time.Time{}, fmt.Errorf("unable to parse time: %v", err)
	}

	return t, nil
}
