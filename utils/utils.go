package utils

import (
	"math/rand"
	"time"
)

func GenerateRandomRoomId() string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	roomID := make([]rune, 6)
	for i := range roomID {
		roomID[i] = letters[r.Intn(len(letters))]
	}
	return string(roomID)
}
