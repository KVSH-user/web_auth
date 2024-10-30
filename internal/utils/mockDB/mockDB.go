package mockDB

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"web_auth/internal/adapters/db/postgres"
	"web_auth/internal/models"

	"github.com/go-faker/faker/v4"
)

func SeedDatabase(ctx context.Context, storage *postgres.Storage, numUsers, numMessages int) error {
	oneYearAgo := time.Now().AddDate(-1, 0, 0)

	for i := 0; i < numUsers; i++ {
		email := faker.Email()
		password := faker.Password()

		userID, err := storage.SaveUser(ctx, email, []byte(password))
		if err != nil {
			log.Printf("Error creating user: %v", err)
			continue
		}
		fmt.Printf("Created user %d: %s\n", userID, email)

		for j := 0; j < numMessages; j++ {
			createdAt := randomDateBetween(oneYearAgo, time.Now())
			message := models.Message{
				UserID:      userID,
				MessageText: faker.Sentence(),
				SenderType:  getRandomSenderType(),
				CreatedAt:   createdAt,
			}
			if err := storage.SaveMessage(ctx, message); err != nil {
				log.Printf("Error creating message for user %d: %v", userID, err)
			}
		}
	}

	fmt.Println("Database seeding completed successfully")
	return nil
}

func getRandomSenderType() string {
	types := []string{"user", "system"}
	return types[rand.Intn(len(types))] // Выбираем случайный индекс
}

func randomDateBetween(start, end time.Time) time.Time {
	delta := end.Sub(start)
	randomSeconds := rand.Int63n(int64(delta.Seconds()))
	return start.Add(time.Duration(randomSeconds) * time.Second)
}
