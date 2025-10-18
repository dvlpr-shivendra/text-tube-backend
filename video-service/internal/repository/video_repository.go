package repository

import (
	"context"
	"time"

	"videoservice/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type VideoRepository struct {
	channelCollection *mongo.Collection
	videoCollection   *mongo.Collection
}

func NewVideoRepository(db *mongo.Database) *VideoRepository {
	return &VideoRepository{
		channelCollection: db.Collection("channels"),
		videoCollection:   db.Collection("videos"),
	}
}

// Channel operations
func (r *VideoRepository) GetCachedChannel(ctx context.Context, channelID string) (*models.Channel, error) {
	var channel models.Channel
	err := r.channelCollection.FindOne(ctx, bson.M{"channel_id": channelID}).Decode(&channel)
	if err != nil {
		return nil, err
	}
	return &channel, nil
}

func (r *VideoRepository) CacheChannel(ctx context.Context, channel *models.Channel) error {
	channel.CachedAt = time.Now()
	filter := bson.M{"channel_id": channel.ChannelID}
	update := bson.M{"$set": channel}
	opts := options.Update().SetUpsert(true)
	_, err := r.channelCollection.UpdateOne(ctx, filter, update, opts)
	return err
}

// Video operations
func (r *VideoRepository) GetCachedVideos(ctx context.Context, channelID string, maxAge time.Duration) ([]models.Video, error) {
	cutoff := time.Now().Add(-maxAge)
	filter := bson.M{
		"channel_id": channelID,
		"cached_at":  bson.M{"$gte": cutoff},
	}
	
	cursor, err := r.videoCollection.Find(ctx, filter, options.Find().SetSort(bson.M{"published_at": -1}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var videos []models.Video
	if err := cursor.All(ctx, &videos); err != nil {
		return nil, err
	}
	return videos, nil
}

func (r *VideoRepository) CacheVideos(ctx context.Context, videos []models.Video) error {
	if len(videos) == 0 {
		return nil
	}

	var operations []mongo.WriteModel
	for _, video := range videos {
		video.CachedAt = time.Now()
		filter := bson.M{"video_id": video.VideoID}
		update := bson.M{"$set": video}
		operation := mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(update).SetUpsert(true)
		operations = append(operations, operation)
	}

	_, err := r.videoCollection.BulkWrite(ctx, operations)
	return err
}

func (r *VideoRepository) GetCachedVideo(ctx context.Context, videoID string, maxAge time.Duration) (*models.Video, error) {
	cutoff := time.Now().Add(-maxAge)
	filter := bson.M{
		"video_id":  videoID,
		"cached_at": bson.M{"$gte": cutoff},
	}

	var video models.Video
	err := r.videoCollection.FindOne(ctx, filter).Decode(&video)
	if err != nil {
		return nil, err
	}
	return &video, nil
}
