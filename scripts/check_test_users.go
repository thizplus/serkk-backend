package main

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type UsersIdentity struct {
	ID       string `gorm:"column:id"`
	Email    string `gorm:"column:email"`
	Username string `gorm:"column:username"`
}

func main() {
	dsn := "host=localhost user=postgres password=n147369 dbname=gofiber_social port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	var users []UsersIdentity
	result := db.Table("users_identity").
		Where("deleted_at IS NULL").
		Order("created_at DESC").
		Limit(10).
		Find(&users)

	if result.Error != nil {
		log.Fatal("Failed to query users:", result.Error)
	}

	fmt.Println("📋 Recent users in users_identity:")
	fmt.Println("=====================================")
	for i, user := range users {
		fmt.Printf("%d. ID: %s\n", i+1, user.ID)
		fmt.Printf("   Email: %s\n", user.Email)
		fmt.Printf("   Username: %s\n", user.Username)
		fmt.Println()
	}

	if len(users) > 0 {
		fmt.Printf("\n✅ Found %d users\n", len(users))
		fmt.Println("\nTo test profile consolidation, you can:")
		fmt.Printf("1. Login with one of these users via Auth Service (8088)\n")
		fmt.Printf("2. Or create a new test user\n")
	} else {
		fmt.Println("❌ No users found. Please create a test user first.")
	}
}
