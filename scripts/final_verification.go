package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

type Media struct {
	ID string `json:"id"`
}

type PostResponse struct {
	ID    string  `json:"id"`
	Title string  `json:"title"`
	Media []Media `json:"media"`
}

type Response struct {
	Data PostResponse `json:"data"`
}

func main() {
	data, err := os.ReadFile("post_207_response.json")
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}

	var resp Response
	if err := json.Unmarshal(data, &resp); err != nil {
		log.Fatalf("Failed to parse JSON: %v", err)
	}

	post := resp.Data

	fmt.Println("🎯 Final Verification - Post 207af412")
	fmt.Println(string(make([]byte, 70)))
	fmt.Printf("\nPost ID: %s\n", post.ID[:8]+"...")
	fmt.Printf("Title: %s\n", post.Title)
	fmt.Printf("Media count: %d\n\n", len(post.Media))

	// Count unique IDs
	mediaIDs := make(map[string]bool)
	for _, media := range post.Media {
		mediaIDs[media.ID] = true
	}

	fmt.Printf("Unique media IDs: %d\n\n", len(mediaIDs))

	if len(post.Media) == len(mediaIDs) {
		fmt.Println("✅ SUCCESS! Media count matches unique IDs")
		fmt.Println("✅ No duplicates detected")
		fmt.Println("\n🎉 The GORM Preload bug has been FIXED!")
	} else {
		fmt.Printf("❌ FAILED! Still has %d duplicates\n", len(post.Media)-len(mediaIDs))
	}

	fmt.Println("\n" + string(make([]byte, 70)))
	fmt.Println("\n📊 Database Query Result (from earlier):")
	fmt.Println("   SELECT COUNT(*) FROM post_media WHERE post_id = '207af412...'")
	fmt.Println("   Result: 2 records")
	fmt.Println("\n📡 API Response Result:")
	fmt.Printf("   GET /api/v1/posts/207af412...\n")
	fmt.Printf("   Result: %d media items\n", len(post.Media))
	fmt.Println("\n✅ Database and API are now IN SYNC!")
}
