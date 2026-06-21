package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Folder struct {
	Id     primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	UserId primitive.ObjectID `json:"userId" bson:"userId"`

	Name      string  `json:"name,omitempty" bson:"name,omitempty"`
	Color     *string `json:"color,omitempty" bson:"color,omitempty"`
	IsDeleted *bool   `json:"isDeleted,omitempty" bson:"isDeleted,omitempty"`

	// IsPublic, when true, makes the folder visible (read-only) to every signed-in user. Only users
	// from an admin domain (ADMIN_EMAIL_DOMAINS) may create/flag public folders.
	IsPublic *bool `json:"isPublic,omitempty" bson:"isPublic,omitempty"`

	// WebhookUrl, if set, is a Discord webhook the folder posts to when events are added to it or
	// scheduled. Validated to be a discord.com webhook URL (see services/discordwebhook).
	WebhookUrl *string `json:"webhookUrl,omitempty" bson:"webhookUrl,omitempty"`

	EventIds []primitive.ObjectID `json:"eventIds" bson:"-"`
}
