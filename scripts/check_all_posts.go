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

	fmt.Println("📊 Checking all posts for media duplicates...\n")
	fmt.Println(string(make([]byte, 70)))

	totalDuplicates := 0
	postsWithDuplicates := 0

	for i, post := range resp.Data.Posts {
		if len(post.Media) == 0 {
			continue
		}

		mediaIDs := make(map[string]int)
		for _, media := range post.Media {
			mediaIDs[media.ID]++
		}

		duplicates := len(post.Media) - len(mediaIDs)

		fmt.Printf("\n%d. Post %s\n", i+1, post.ID[:8])
		fmt.Printf("   Title: %s\n", post.Title)
		fmt.Printf("   Media: %d total, %d unique", len(post.Media), len(mediaIDs))

		if duplicates > 0 {
			fmt.Printf(" ❌ (%d duplicates)\n", duplicates)
			totalDuplicates += duplicates
			postsWithDuplicates++
		} else {
			fmt.Printf(" ✅\n")
		}
	}

	fmt.Println("\n" + string(make([]byte, 70)))
	fmt.Printf("\n📈 Summary:\n")
	fmt.Printf("   Total posts checked: %d\n", len(resp.Data.Posts))
	fmt.Printf("   Posts with duplicates: %d\n", postsWithDuplicates)
	fmt.Printf("   Total duplicate entries: %d\n", totalDuplicates)

	if postsWithDuplicates == 0 {
		fmt.Println("\n🎉 SUCCESS! All posts have correct media counts!")
	} else {
		fmt.Println("\n⚠️  Some posts still have duplicates")
	}
}
