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

type Post struct {
	ID    string  `json:"id"`
	Title string  `json:"title"`
	Media []Media `json:"media"`
}

type Response struct {
	Data struct {
		Posts []Post `json:"posts"`
	} `json:"data"`
}

func main() {
	data, err := os.ReadFile("api_debug.json")
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}

	var resp Response
	if err := json.Unmarshal(data, &resp); err != nil {
		log.Fatalf("Failed to parse JSON: %v", err)
	}

	targetID := "207af412-c890-4560-83b2-a5642933367e"

	fmt.Println("🔍 Searching for post:", targetID[:8]+"...")
	fmt.Println("=" + string(make([]byte, 60)) + "=\n")

	for i, post := range resp.Data.Posts {
		if post.ID == targetID {
			fmt.Printf("✅ Found at position %d\n", i+1)
			fmt.Printf("Title: %s\n", post.Title)
			fmt.Printf("Total media items: %d\n\n", len(post.Media))

			// Count unique IDs
			mediaIDs := make(map[string]int)
			for _, media := range post.Media {
				mediaIDs[media.ID]++
			}

			fmt.Printf("Unique media IDs: %d\n", len(mediaIDs))
			fmt.Printf("Duplicates: %d\n\n", len(post.Media)-len(mediaIDs))

			if len(mediaIDs) <= 20 {
				fmt.Println("Media ID duplicates:")
				for id, count := range mediaIDs {
					if count > 1 {
						fmt.Printf("  🔴 %s → %dx\n", id[:8], count)
					} else {
						fmt.Printf("  ✅ %s → 1x\n", id[:8])
					}
				}
			}

			return
		}
	}

	fmt.Printf("❌ Post %s not found in response\n", targetID[:8])
}
