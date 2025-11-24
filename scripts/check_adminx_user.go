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

type UserProfile struct {
	UserID      string `gorm:"column:user_id"`
	DisplayName string `gorm:"column:display_name"`
	Avatar      string `gorm:"column:avatar"`
}

type UsersCache struct {
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

	userID := "0a8218ff-b13c-4cce-9886-c791a72b49dd"

	fmt.Println("🔍 Checking adminx user data...")
	fmt.Println("=====================================")

	// Check users_identity
	var identity UsersIdentity
	err1 := db.Table("users_identity").Where("id = ?", userID).First(&identity).Error

	if err1 == nil {
		fmt.Println("✅ users_identity: FOUND")
		fmt.Printf("   Email: %s\n", identity.Email)
		fmt.Printf("   Username: %s\n", identity.Username)
	} else {
		fmt.Println("❌ users_identity: NOT FOUND")
		fmt.Printf("   Error: %v\n", err1)
	}

	// Check user_profiles
	var profile UserProfile
	err2 := db.Table("user_profiles").Where("user_id = ?", userID).First(&profile).Error

	if err2 == nil {
		fmt.Println("✅ user_profiles: FOUND")
		fmt.Printf("   DisplayName: %s\n", profile.DisplayName)
		fmt.Printf("   Avatar: %s\n", profile.Avatar)
	} else {
		fmt.Println("❌ user_profiles: NOT FOUND")
		fmt.Printf("   Error: %v\n", err2)
	}

	// Check users_cache
	var cache UsersCache
	err3 := db.Table("users_cache").Where("id = ?", userID).First(&cache).Error

	if err3 == nil {
		fmt.Println("✅ users_cache: FOUND")
		fmt.Printf("   Email: %s\n", cache.Email)
		fmt.Printf("   Username: %s\n", cache.Username)
	} else {
		fmt.Println("❌ users_cache: NOT FOUND")
		fmt.Printf("   Error: %v\n", err3)
	}

	fmt.Println("\n=====================================")
	fmt.Println("📝 Summary:")

	if err1 == nil && err2 == nil && err3 != nil {
		fmt.Println("⚠️  PROBLEM CONFIRMED!")
		fmt.Println("✅ V2 event created users_identity + user_profiles")
		fmt.Println("❌ BUT users_cache is missing (this causes 404 error)")
		fmt.Println("\n💡 Solution: Update handlers to use users_identity instead of users_cache")
	} else if err1 != nil && err2 != nil && err3 != nil {
		fmt.Println("❌ No data found in any table!")
		fmt.Println("💡 Event might not have been processed")
	}
}
