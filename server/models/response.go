package models

import "go.mongodb.org/mongo-driver/bson/primitive"

// EventTopic is a discussion topic / agenda item suggested by a respondent on an event.
type EventTopic struct {
	Id        primitive.ObjectID `json:"id" bson:"_id"`
	Name      string             `json:"name" bson:"name"`
	Text      string             `json:"text" bson:"text"`
	CreatedAt primitive.DateTime `json:"createdAt" bson:"createdAt"`
}

// CurrentPrivacyPolicyVersion is stamped onto guest responses when they consent to the privacy
// policy. Bump this when the policy materially changes so old consents can be distinguished.
const CurrentPrivacyPolicyVersion = "2026-06-21"

type EventResponse struct {
	Id      primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	EventId primitive.ObjectID `json:"eventId" bson:"eventId"`

	UserId   string    `json:"userId" bson:"userId"`
	Response *Response `json:"response" bson:"response"`
}

// A response object containing an array of times that the given user is available
type Response struct {
	// Guest information
	Name  string `json:"name" bson:"name,omitempty"`
	Email string `json:"email" bson:"email,omitempty"`

	// Recorded privacy-policy consent for guest submissions (not set for signed-in users, who
	// consent at sign-in). Provides provable, versioned consent for GDPR purposes.
	ConsentedToPrivacyPolicyAt *primitive.DateTime `json:"consentedToPrivacyPolicyAt" bson:"consentedToPrivacyPolicyAt,omitempty"`
	PrivacyPolicyVersion       string              `json:"privacyPolicyVersion" bson:"privacyPolicyVersion,omitempty"`

	// User information
	UserId primitive.ObjectID `json:"userId" bson:"userId,omitempty"`
	User   *User              `json:"user" bson:",omitempty"`

	// TimezoneOffset is the respondent's JS getTimezoneOffset() (minutes behind UTC, positive =
	// behind) captured AT SUBMISSION TIME. Stored per-response because users travel — the value on
	// the User document reflects only their last sign-in. Nil for responses submitted before this
	// was captured. Surfaced onto response.User.TimezoneOffset when responses are returned.
	TimezoneOffset *int `json:"timezoneOffset" bson:"timezoneOffset,omitempty"`

	// Availability
	Availability []primitive.DateTime `json:"availability" bson:"availability"`
	IfNeeded     []primitive.DateTime `json:"ifNeeded" bson:"ifNeeded"`

	// Mapping from the start date of a day to the available times for that day
	ManualAvailability *map[primitive.DateTime][]primitive.DateTime `json:"manualAvailability" bson:"manualAvailability,omitempty"`

	// Calendar availability variables for Availability Groups feature
	UseCalendarAvailability *bool                `json:"useCalendarAvailability" bson:"useCalendarAvailability,omitempty"`
	EnabledCalendars        *map[string][]string `json:"enabledCalendars" bson:"enabledCalendars,omitempty"` // Maps email to an array of sub calendar ids
	CalendarOptions         *CalendarOptions     `json:"calendarOptions" bson:"calendarOptions,omitempty"`
}
