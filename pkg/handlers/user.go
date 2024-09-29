package handlers

import (
	"encoding/json"

	"github.com/ayush3160/interview-bytes-backend/pkg/models"
	"github.com/ayush3160/interview-bytes-backend/utils"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

type userHandler struct {
	logger         *zap.Logger
	userCollection *mongo.Collection
}

func NewUserHandler(logger *zap.Logger, userCollection *mongo.Collection) *userHandler {
	return &userHandler{
		logger:         logger,
		userCollection: userCollection,
	}
}

func (u *userHandler) CreateUser(c *fiber.Ctx) error {

	ctx := c.Context()

	var user models.User

	if err := json.Unmarshal(c.Body(), &user); err != nil {
		u.logger.Error("Error parsing request body", zap.Error(err))
		return c.Status(fiber.StatusBadRequest).SendString("Invalid Request Body")
	}

	mongoResult := u.userCollection.FindOne(ctx, bson.M{"username": user.Username})

	if mongoResult.Err() == nil {
		return c.Status(fiber.StatusOK).SendString("User Already Exists")
	}

	hashedPassword, err := utils.HashPassword(user.Password)

	if err != nil {
		u.logger.Error("Error hashing password", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}

	user.Password = hashedPassword

	result, err := u.userCollection.InsertOne(ctx, user)

	if err != nil {
		u.logger.Error("Error inserting user", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).SendString("Error inserting user")
	}

	return c.Status(fiber.StatusCreated).JSON(result)
}

func (u *userHandler) Login(c *fiber.Ctx) error {
	ctx := c.Context()

	var user models.User

	if err := json.Unmarshal(c.Body(), &user); err != nil {
		u.logger.Error("Error parsing request body", zap.Error(err))
		return c.Status(fiber.StatusBadRequest).SendString("Invalid Request Body")
	}

	mongoResult := u.userCollection.FindOne(ctx, bson.M{"username": user.Username})

	if mongoResult.Err() != nil {
		return c.Status(fiber.StatusUnauthorized).SendString("Invalid Username")
	}

	var dbUser models.User

	if err := mongoResult.Decode(&dbUser); err != nil {
		u.logger.Error("Error decoding user", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}

	if !utils.ComparePassword(dbUser.Password, user.Password) {
		return c.Status(fiber.StatusUnauthorized).SendString("Invalid Password")
	}

	token, err := utils.CreateJWT(dbUser.Username, dbUser.ID)

	if err != nil {
		u.logger.Error("Error creating JWT", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}

	var response = struct {
		Token    string `json:"token"`
		UserName string `json:"username"`
	}{
		Token:    token,
		UserName: dbUser.Username,
	}

	return c.Status(fiber.StatusOK).JSON(response)
}
