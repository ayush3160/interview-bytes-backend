package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Interview struct {
	ID           primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Title        string             `json:"title" bson:"title"`
	Description  string             `json:"description" bson:"description"`
	StartTime    string             `json:"startTime" bson:"startTime"`
	EndTime      string             `json:"endTime" bson:"endTime"`
	RoomId       string             `json:"roomId" bson:"roomId"`
	Host         string             `json:"host" bson:"host"`
	Status       string             `json:"status" bson:"status"`
	Participants []string           `json:"participants" bson:"participants"`
}

type InterviewStatus string

const (
	InterviewUpcoming InterviewStatus = "upcoming"
	InterviewFinished InterviewStatus = "finished"
)
