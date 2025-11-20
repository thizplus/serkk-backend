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
	Media []Media `json:"media"`
}

type Response struct {
	Data struct {
		Posts []Post `json:"posts"`
	} `json:"data"`
}

func main() {
	data, err := os.ReadFile("api_response.json")
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}

	var resp Response
	if err := json.Unmarshal(data, &resp); err != nil {
		log.Fatalf("Failed to parse JSON: %v", err)
	}

	fmt.Println("📊 Media count per post:")
	fmt.Println("=" + string(make([]byte, 60)) + "=\n")

	for i, post := range resp.Data.Posts {
		if i >= 5 {
			break
		}

		mediaIDs := make(map[string]int)
		for _, media := range post.Media {
			mediaIDs[media.ID]++
		}

		fmt.Printf("Post %d (ID: %s...):\n", i+1, post.ID[:8])
		fmt.Printf("  - Total media items in response: %d\n", len(post.Media))
		fmt.Printf("  - Unique media IDs: %d\n", len(mediaIDs))
		fmt.Printf("  - Duplicates: %d\n\n", len(post.Media)-len(mediaIDs))

		if len(mediaIDs) < 10 {
			fmt.Println("  Media ID duplicates:")
			for id, count := range mediaIDs {
				if count > 1 {
					fmt.Printf("    🔴 %s → %dx\n", id[:8], count)
				}
			}
			fmt.Println()
		}
	}
}
