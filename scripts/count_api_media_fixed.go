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
	data, err := os.ReadFile("api_response_fixed.json")
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}

	var resp Response
	if err := json.Unmarshal(data, &resp); err != nil {
		log.Fatalf("Failed to parse JSON: %v", err)
	}

	fmt.Println("✅ API Response Check (After Fix):")
	fmt.Println("=" + string(make([]byte, 60)) + "=\n")

	allGood := true
	for i, post := range resp.Data.Posts {
		if i >= 5 {
			break
		}

		mediaIDs := make(map[string]int)
		for _, media := range post.Media {
			mediaIDs[media.ID]++
		}

		status := "✅"
		if len(post.Media) > 10 {
			status = "⚠️ "
			allGood = false
		}
		if len(post.Media) != len(mediaIDs) {
			status = "🔴"
			allGood = false
		}

		fmt.Printf("%s Post %d: %s\n", status, i+1, post.Title)
		fmt.Printf("   ID: %s...\n", post.ID[:8])
		fmt.Printf("   Total media: %d\n", len(post.Media))
		fmt.Printf("   Unique IDs: %d\n", len(mediaIDs))
		if len(post.Media) != len(mediaIDs) {
			fmt.Printf("   🔴 DUPLICATES: %d\n", len(post.Media)-len(mediaIDs))
		}
		fmt.Println()
	}

	if allGood {
		fmt.Println("🎉 SUCCESS! All posts have correct media count (no duplicates)")
	} else {
		fmt.Println("❌ FAILED! Some posts still have issues")
	}
}
