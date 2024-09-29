package handlers

import (
	"encoding/json"
	"sync"

	"github.com/ayush3160/interview-bytes-backend/pkg/models"
	"github.com/ayush3160/interview-bytes-backend/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

type roomHandler struct {
	logger      *zap.Logger
	interviewDb *mongo.Collection
}

var (
	rooms     = make(map[string]map[*websocket.Conn]bool)
	roomLimit = 2
	mu        sync.Mutex
)

func NewRoomHandler(logger *zap.Logger) *roomHandler {
	return &roomHandler{
		logger: logger,
	}
}

func (r *roomHandler) CreateRoom(c *fiber.Ctx) error {

	ctx := c.Context()

	var interview models.Interview

	if err := json.Unmarshal(c.Body(), &interview); err != nil {
		r.logger.Error("Error parsing request body", zap.Error(err))
		return c.Status(fiber.StatusBadRequest).SendString("Invalid Request Body")
	}

	interview.Host = c.Locals("id").(string)

	mu.Lock()

	roomId := utils.GenerateRandomRoomId()
	rooms[roomId] = make(map[*websocket.Conn]bool)

	mu.Unlock()

	interview.RoomId = roomId
	interview.Status = string(models.InterviewUpcoming)

	interviewID, err := r.interviewDb.InsertOne(ctx, interview)

	if err != nil {
		r.logger.Error("Error inserting interview", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).SendString("Error inserting interview")
	}

	interview.ID = interviewID.InsertedID.(primitive.ObjectID)

	return c.Status(fiber.StatusOK).JSON(interview)
}

func (r *roomHandler) CanAccessRoom(c *fiber.Ctx) error {
	roomId := c.Params("roomId")

	var mailID string

	if err := json.Unmarshal(c.Body(), &mailID); err != nil {
		r.logger.Error("Error parsing request body", zap.Error(err))
		return c.Status(fiber.StatusBadRequest).SendString("Invalid Request Body")
	}

	if roomId == "" {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid Room ID")
	}

	ctx := c.Context()

	var interview models.Interview

	queryResult := r.interviewDb.FindOne(ctx, bson.M{"roomId": roomId})

	if queryResult.Err() != nil {

		if queryResult.Err() == mongo.ErrNoDocuments {
			r.logger.Error("Error fetching interview", zap.Error(queryResult.Err()))
			return c.Status(fiber.StatusOK).SendString("No Such Interview Exists")
		}
	}

	if err := queryResult.Decode(&interview); err != nil {
		r.logger.Error("Error decoding interview", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}

	if interview.Participants[0] != mailID && len(rooms[roomId]) >= roomLimit {
		return c.Status(fiber.StatusUnauthorized).SendString("You are not authorized to join this interview")
	} else {
		return c.Status(fiber.StatusOK).SendString("You are authorized to join this interview")
	}

}

func (r *roomHandler) JoinRoom(conn *websocket.Conn) {
	roomId := conn.Params("roomId")

	mu.Lock()

	rooms[roomId][conn] = true

	mu.Unlock()

	for {
		mt, msg, err := conn.ReadMessage()
		if err != nil {
			r.logger.Error("Error reading message", zap.Error(err))
			break
		}

		for client := range rooms[roomId] {
			if client != conn {
				if err = client.WriteMessage(mt, msg); err != nil {
					r.logger.Error("Error writing message", zap.Error(err))
					break
				}
			}
		}
	}

	delete(rooms[roomId], conn)
	if len(rooms[roomId]) == 0 {
		delete(rooms, roomId)
	}

}
