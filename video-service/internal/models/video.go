package models

import "time"

type Channel struct {
	ID          string    `bson:"_id,omitempty"`
	ChannelID   string    `bson:"channel_id"`
	Title       string    `bson:"title"`
	Description string    `bson:"description"`
	Thumbnail   string    `bson:"thumbnail"`
	CachedAt    time.Time `bson:"cached_at"`
}

type Video struct {
	ID           string    `bson:"_id,omitempty"`
	VideoID      string    `bson:"video_id"`
	Title        string    `bson:"title"`
	Description  string    `bson:"description"`
	Thumbnail    string    `bson:"thumbnail"`
	PublishedAt  string    `bson:"published_at"`
	ChannelID    string    `bson:"channel_id"`
	ChannelTitle string    `bson:"channel_title"`
	ViewCount    int64     `bson:"view_count"`
	LikeCount    int64     `bson:"like_count"`
	CachedAt     time.Time `bson:"cached_at"`
}
