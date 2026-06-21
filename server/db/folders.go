package db

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"schej.it/server/logger"
	"schej.it/server/models"
)

func CreateFolder(folder *models.Folder) (primitive.ObjectID, error) {
	result, err := FoldersCollection.InsertOne(context.Background(), folder)
	if err != nil {
		logger.StdErr.Panicln(err)
		return primitive.NilObjectID, err
	}
	return result.InsertedID.(primitive.ObjectID), nil
}

// GetWebhookFoldersForEvent returns the (non-deleted) folders that contain the given event for the
// given user AND have a Discord webhook configured. Used to broadcast event changes.
func GetWebhookFoldersForEvent(eventId primitive.ObjectID, userId primitive.ObjectID) []models.Folder {
	ctx := context.Background()

	// Find folder ids this event is mapped into for this user.
	cursor, err := FolderEventsCollection.Find(ctx, bson.M{"eventId": eventId, "userId": userId})
	if err != nil {
		logger.StdErr.Println(err)
		return nil
	}
	var mappings []models.FolderEvent
	if err := cursor.All(ctx, &mappings); err != nil {
		logger.StdErr.Println(err)
		return nil
	}
	if len(mappings) == 0 {
		return nil
	}
	folderIds := make([]primitive.ObjectID, 0, len(mappings))
	for _, m := range mappings {
		folderIds = append(folderIds, m.FolderId)
	}

	// Fetch the folders that have a webhook configured.
	fc, err := FoldersCollection.Find(ctx, bson.M{
		"_id":        bson.M{"$in": folderIds},
		"userId":     userId,
		"webhookUrl": bson.M{"$exists": true, "$nin": bson.A{"", nil}},
		"$or": bson.A{
			bson.M{"isDeleted": bson.M{"$exists": false}},
			bson.M{"isDeleted": false},
		},
	})
	if err != nil {
		logger.StdErr.Println(err)
		return nil
	}
	var folders []models.Folder
	if err := fc.All(ctx, &folders); err != nil {
		logger.StdErr.Println(err)
		return nil
	}
	return folders
}

func GetFolderById(folderId primitive.ObjectID, userId primitive.ObjectID) (*models.Folder, error) {
	var folder models.Folder
	err := FoldersCollection.FindOne(context.Background(), bson.M{
		"_id":    folderId,
		"userId": userId,
		"$or": bson.A{
			bson.M{"isDeleted": bson.M{"$exists": false}},
			bson.M{"isDeleted": false},
		},
	}).Decode(&folder)
	if err != nil {
		return nil, err
	}

	return &folder, nil
}

// GetReadableFolderById returns a folder the user is allowed to read: their own, or any public
// folder. Returns an error if not found / not readable.
func GetReadableFolderById(folderId primitive.ObjectID, userId primitive.ObjectID) (*models.Folder, error) {
	var folder models.Folder
	err := FoldersCollection.FindOne(context.Background(), bson.M{
		"_id": folderId,
		"$and": bson.A{
			bson.M{"$or": bson.A{
				bson.M{"isDeleted": bson.M{"$exists": false}},
				bson.M{"isDeleted": false},
			}},
			bson.M{"$or": bson.A{
				bson.M{"userId": userId},
				bson.M{"isPublic": true},
			}},
		},
	}).Decode(&folder)
	if err != nil {
		return nil, err
	}
	return &folder, nil
}

func GetAllFolders(userId primitive.ObjectID) ([]models.Folder, error) {
	notDeleted := bson.A{
		bson.M{"isDeleted": bson.M{"$exists": false}},
		bson.M{"isDeleted": false},
	}

	// The caller's own folders, plus any public folders owned by others (read-only).
	cursor, err := FoldersCollection.Find(context.Background(), bson.M{
		"$and": bson.A{
			bson.M{"$or": notDeleted},
			bson.M{"$or": bson.A{
				bson.M{"userId": userId},
				bson.M{"isPublic": true},
			}},
		},
	})
	if err != nil {
		return nil, err
	}

	var folders []models.Folder
	if err = cursor.All(context.Background(), &folders); err != nil {
		return nil, err
	}

	for i := range folders {
		// Events are mapped per owner, so fetch using the folder's owner (works for public folders
		// viewed by non-owners too).
		events, err := GetEventsInFolder(folders[i].Id, folders[i].UserId)
		if err != nil {
			return nil, err
		}
		if events != nil {
			folders[i].EventIds = events
		} else {
			folders[i].EventIds = []primitive.ObjectID{}
		}
		// Never expose another user's webhook URL (it embeds a secret token).
		if folders[i].UserId != userId {
			folders[i].WebhookUrl = nil
		}
	}

	return folders, nil
}

func GetEventsInFolder(folderId primitive.ObjectID, userId primitive.ObjectID) ([]primitive.ObjectID, error) {
	cursor, err := FolderEventsCollection.Find(context.Background(), bson.M{
		"folderId": folderId,
		"userId":   userId,
	}, options.Find().SetProjection(bson.M{"eventId": 1}))
	if err != nil {
		return nil, err
	}

	var eventIdsResponse []struct {
		EventId primitive.ObjectID `bson:"eventId"`
	}
	if err = cursor.All(context.Background(), &eventIdsResponse); err != nil {
		return nil, err
	}

	eventIds := make([]primitive.ObjectID, len(eventIdsResponse))
	for i, eventId := range eventIdsResponse {
		eventIds[i] = eventId.EventId
	}

	return eventIds, nil
}

func UpdateFolder(folderId primitive.ObjectID, userId primitive.ObjectID, updates bson.M) error {
	_, err := FoldersCollection.UpdateOne(context.Background(), bson.M{"_id": folderId, "userId": userId}, bson.M{"$set": updates})
	return err
}

func SetEventFolder(eventId primitive.ObjectID, folderId *primitive.ObjectID, userId primitive.ObjectID) error {
	ctx := context.Background()

	// Remove any existing mapping for this event
	_, err := FolderEventsCollection.DeleteMany(ctx, bson.M{"eventId": eventId, "userId": userId})
	if err != nil {
		return err
	}

	// If folderId is nil, we just un-assign it. Otherwise, create a new mapping
	if folderId != nil {
		_, err = FolderEventsCollection.InsertOne(ctx, models.FolderEvent{
			FolderId: *folderId,
			EventId:  eventId,
			UserId:   userId,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func DeleteFolder(folderId primitive.ObjectID, userId primitive.ObjectID) error {
	ctx := context.Background()
	// Mark this folder as deleted
	_, err := FoldersCollection.UpdateOne(ctx, bson.M{"_id": folderId, "userId": userId}, bson.M{"$set": bson.M{"isDeleted": true}})
	if err != nil {
		return err
	}

	// Find all event mappings for this folder
	eventIds, err := GetEventsInFolder(folderId, userId)
	if err != nil {
		return err
	}

	// Mark all events in this folder as deleted (if the user owns the event)
	if len(eventIds) > 0 {
		_, err = EventsCollection.UpdateMany(ctx, bson.M{"_id": bson.M{"$in": eventIds}, "ownerId": userId}, bson.M{"$set": bson.M{"isDeleted": true}})
		if err != nil {
			return err
		}
	}

	// Delete the mappings
	_, err = FolderEventsCollection.DeleteMany(ctx, bson.M{"folderId": folderId, "userId": userId})
	if err != nil {
		return err
	}

	return nil
}

// GetEventIdsInPublicFolders returns the IDs of all events mapped into any public (non-deleted)
// folder. Used so events shared via a public folder are visible (read-only) to every signed-in user.
func GetEventIdsInPublicFolders() []primitive.ObjectID {
	ctx := context.Background()

	// Public folder ids.
	fc, err := FoldersCollection.Find(ctx, bson.M{
		"isPublic": true,
		"$or": bson.A{
			bson.M{"isDeleted": bson.M{"$exists": false}},
			bson.M{"isDeleted": false},
		},
	}, options.Find().SetProjection(bson.M{"_id": 1, "userId": 1}))
	if err != nil {
		logger.StdErr.Println(err)
		return nil
	}
	var folders []models.Folder
	if err := fc.All(ctx, &folders); err != nil {
		logger.StdErr.Println(err)
		return nil
	}
	if len(folders) == 0 {
		return nil
	}

	// Event ids mapped into those folders (by the folder owner).
	ors := bson.A{}
	for _, f := range folders {
		ors = append(ors, bson.M{"folderId": f.Id, "userId": f.UserId})
	}
	ec, err := FolderEventsCollection.Find(ctx, bson.M{"$or": ors},
		options.Find().SetProjection(bson.M{"eventId": 1}))
	if err != nil {
		logger.StdErr.Println(err)
		return nil
	}
	var mappings []models.FolderEvent
	if err := ec.All(ctx, &mappings); err != nil {
		logger.StdErr.Println(err)
		return nil
	}
	ids := make([]primitive.ObjectID, 0, len(mappings))
	for _, m := range mappings {
		ids = append(ids, m.EventId)
	}
	return ids
}
