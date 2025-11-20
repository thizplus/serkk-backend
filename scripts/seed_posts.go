package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"time"
)

const (
	API_URL = "http://localhost:8080/api/v1/posts"
	TOKEN   = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6InRoZXB0aGFpLmptQGdtYWlsLmNvbSIsImV4cCI6MTc2Mzc0NDI1MywiaWF0IjoxNzYzMTM5NDUzLCJyb2xlIjoidXNlciIsInVzZXJfaWQiOiI0YWExMGUxYi0wNmM0LTRiMDktOGJkOS01Y2VhOTRjZDM3MjMiLCJ1c2VybmFtZSI6InRoZXB0aGFpIn0.UgZeYGOU7JdsShLnAfunzpnmo8H7XfkelIlfUNq5pgY"
)

type CreatePostRequest struct {
	Title      string   `json:"title"`
	Content    string   `json:"content"`
	Type       string   `json:"type"`
	Tags       []string `json:"tags"`
	IsNSFW     bool     `json:"isNsfw"`
	IsSpoiler  bool     `json:"isSpoiler"`
	IsOriginal bool     `json:"isOriginal"`
}

var (
	titles = []string{
		"วิธีการเขียน Go ให้เร็วขึ้น 10 เท่า",
		"React vs Vue: เลือกอะไรดีในปี 2025?",
		"10 VSCode Extensions ที่ Developer ต้องมี",
		"Docker คืออะไร? เริ่มต้นอย่างไร",
		"Kubernetes สำหรับมือใหม่",
		"PostgreSQL Performance Tuning Tips",
		"เทคนิคการออกแบบ Database ที่มืออาชีพใช้",
		"Clean Code: หลักการที่ทุกคนควรรู้",
		"Git Commands ที่ใช้บ่อยที่สุด",
		"Microservices Architecture คืออะไร?",
		"REST API vs GraphQL: ข้อดีข้อเสียของแต่ละตัว",
		"JWT Authentication แบบ Secure",
		"SQL vs NoSQL: เลือกใช้ยังไง",
		"Redis Caching Strategies",
		"Nginx Configuration สำหรับ Production",
		"CI/CD Pipeline ด้วย GitHub Actions",
		"TypeScript Tips & Tricks",
		"Next.js 14 มีอะไรใหม่บ้าง",
		"Tailwind CSS: Utility-First Framework",
		"Figma to Code: Best Practices",
		"Responsive Design ในยุค 2025",
		"Web Performance Optimization",
		"SEO สำหรับ Single Page Applications",
		"Progressive Web Apps (PWA) คืออะไร",
		"WebSockets vs Server-Sent Events",
		"OAuth 2.0 Authentication Flow",
		"CORS: Cross-Origin Resource Sharing",
		"Content Security Policy (CSP)",
		"XSS และวิธีป้องกัน",
		"SQL Injection: เข้าใจและป้องกัน",
		"Rate Limiting สำหรับ API",
		"API Versioning Strategies",
		"Pagination: Offset vs Cursor",
		"Full-Text Search ด้วย Elasticsearch",
		"Message Queue ด้วย RabbitMQ",
		"Event-Driven Architecture",
		"SOLID Principles ที่ทุกคนควรรู้",
		"Design Patterns ในการพัฒนาซอฟต์แวร์",
		"Test-Driven Development (TDD)",
		"Unit Testing Best Practices",
		"Integration Testing vs E2E Testing",
		"Debugging Tips สำหรับ Developer",
		"VS Code Keyboard Shortcuts",
		"Terminal Commands ที่ใช้บ่อย",
		"Vim: Editor ที่ทรงพลัง",
		"Tmux สำหรับ Terminal Multiplexing",
		"Shell Scripting Basics",
		"Python vs Go: เปรียบเทียบ",
		"Rust Programming Language",
		"WebAssembly คืออะไร?",
	}

	contents = []string{
		"ในบทความนี้จะแชร์เทคนิคและ best practices ที่จะช่วยให้คุณเขียนโค้ดได้ดีขึ้น รวดเร็วขึ้น และ maintainable มากขึ้น\n\nเริ่มจากการเข้าใจ fundamentals ให้ดีก่อน แล้วค่อยไปต่อที่ advanced topics ทีละขั้น อย่าเพิ่งรีบ!\n\nสิ่งสำคัญคือการฝึกฝนอย่างสม่ำเสมอ และเรียนรู้จากโค้ดของคนอื่นด้วย",
		"มาดูข้อดีข้อเสียของแต่ละตัวกันครับ\n\n**ข้อดี:**\n- Performance ดี\n- Community ใหญ่\n- Documentation ครบถ้วน\n\n**ข้อเสีย:**\n- Learning curve สูง\n- Bundle size ใหญ่\n\nโดยรวมแล้วแต่ละตัวก็มีจุดเด่นของตัวเอง ขึ้นอยู่กับ use case",
		"แชร์ extensions ที่ใช้ประจำทุกวัน ช่วยให้การเขียนโค้ดเร็วขึ้นมาก!\n\n1. ESLint - ตรวจสอบ code quality\n2. Prettier - Format code อัตโนมัติ\n3. GitLens - ดู git history\n4. Auto Import - Import modules อัตโนมัติ\n5. Bracket Pair Colorizer - ดู brackets ง่ายขึ้น",
		"Step by step guide สำหรับมือใหม่ที่อยากเริ่มต้น\n\nเริ่มจากการติดตั้ง Docker Desktop แล้วลองรัน hello world container ก่อน\n\nพอคุ้นเคยแล้วค่อยไป learn เรื่อง Dockerfile, docker-compose, และ best practices ต่างๆ",
		"อธิบายแนวคิดและวิธีการใช้งาน พร้อมตัวอย่างจริง\n\nKubernetes ช่วยจัดการ containers ในระดับ production ให้คุณได้\n\nแม้จะดูซับซ้อนตอนแรก แต่พอเข้าใจแล้วจะเห็นว่ามันทรงพลังมาก",
		"Tips and tricks สำหรับการ optimize database performance\n\n**สิ่งที่ควรทำ:**\n- สร้าง indexes ที่เหมาะสม\n- ใช้ EXPLAIN ANALYZE เพื่อดู query plan\n- Optimize slow queries\n- Connection pooling\n\n**สิ่งที่ไม่ควรทำ:**\n- N+1 queries\n- SELECT * without WHERE\n- Missing indexes",
		"หลักการออกแบบ database ที่ดี จะช่วยให้ระบบ scalable และ maintainable\n\nเริ่มจาก normalization ให้ดีก่อน แล้วค่อย denormalize ตามความจำเป็น\n\nอย่าลืม plan สำหรับการ scale ในอนาคตด้วย",
		"Clean Code principles ที่ทุกคนควรปฏิบัติตาม\n\n- ตั้งชื่อ variables/functions ให้มีความหมาย\n- Functions ควรทำแค่สิ่งเดียว (Single Responsibility)\n- Comment เฉพาะส่วนที่จำเป็น\n- DRY (Don't Repeat Yourself)\n\nโค้ดที่ดีคือโค้ดที่อ่านเข้าใจง่าย",
		"Git commands ที่ใช้บ่อยในการทำงานจริง\n\n```bash\ngit status\ngit add .\ngit commit -m \"message\"\ngit push origin main\ngit pull\ngit branch\ngit checkout -b new-branch\ngit merge\n```\n\nควร commit บ่อยๆ และเขียน commit message ให้ดี",
		"อธิบายสถาปัตยกรรมแบบ microservices พร้อมข้อดีข้อเสีย\n\n**ข้อดี:**\n- Scale แต่ละ service ได้อิสระ\n- Deploy แยกกันได้\n- Technology stack ที่แตกต่างกันได้\n\n**ข้อเสีย:**\n- ซับซ้อนขึ้น\n- Debugging ยากขึ้น\n- Network latency",
	}

	tags = [][]string{
		{"golang", "programming", "tutorial"},
		{"javascript", "react", "vue"},
		{"vscode", "productivity", "tools"},
		{"docker", "devops", "containers"},
		{"kubernetes", "devops", "cloud"},
		{"postgresql", "database", "performance"},
		{"database", "design", "architecture"},
		{"clean-code", "best-practices", "programming"},
		{"git", "version-control", "tutorial"},
		{"microservices", "architecture", "backend"},
		{"api", "rest", "graphql"},
		{"security", "authentication", "jwt"},
		{"database", "sql", "nosql"},
		{"redis", "caching", "performance"},
		{"nginx", "devops", "production"},
		{"cicd", "github-actions", "automation"},
		{"typescript", "javascript", "programming"},
		{"nextjs", "react", "web"},
		{"tailwindcss", "css", "frontend"},
		{"figma", "design", "frontend"},
		{"responsive", "css", "web"},
		{"performance", "optimization", "web"},
		{"seo", "spa", "web"},
		{"pwa", "web", "mobile"},
		{"websockets", "realtime", "web"},
		{"oauth", "security", "authentication"},
		{"cors", "security", "web"},
		{"csp", "security", "web"},
		{"xss", "security", "web"},
		{"security", "sql-injection", "database"},
		{"api", "rate-limiting", "backend"},
		{"api", "versioning", "backend"},
		{"pagination", "api", "performance"},
		{"elasticsearch", "search", "database"},
		{"rabbitmq", "message-queue", "backend"},
		{"architecture", "event-driven", "backend"},
		{"solid", "principles", "programming"},
		{"design-patterns", "programming", "architecture"},
		{"tdd", "testing", "programming"},
		{"testing", "unit-testing", "programming"},
		{"testing", "integration", "e2e"},
		{"debugging", "programming", "tips"},
		{"vscode", "shortcuts", "productivity"},
		{"terminal", "cli", "productivity"},
		{"vim", "editor", "productivity"},
		{"tmux", "terminal", "productivity"},
		{"shell", "scripting", "automation"},
		{"python", "golang", "programming"},
		{"rust", "programming", "systems"},
		{"webassembly", "web", "performance"},
	}
)

func createPost(post CreatePostRequest, index int) error {
	jsonData, err := json.Marshal(post)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	req, err := http.NewRequest("POST", API_URL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+TOKEN)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to create post (status %d): %s", resp.StatusCode, string(body))
	}

	log.Printf("✅ [%d/50] Created: %s", index+1, post.Title)
	return nil
}

func main() {
	rand.Seed(time.Now().UnixNano())
	log.Println("🚀 Starting to create 50 test posts...")
	log.Println("📝 API URL:", API_URL)
	log.Println("")

	successCount := 0
	failCount := 0

	for i := 0; i < 50; i++ {
		// Random content and tags
		titleIdx := i % len(titles)
		contentIdx := rand.Intn(len(contents))
		tagsIdx := i % len(tags)

		post := CreatePostRequest{
			Title:      titles[titleIdx],
			Content:    contents[contentIdx],
			Type:       "text",
			Tags:       tags[tagsIdx],
			IsNSFW:     false,
			IsSpoiler:  false,
			IsOriginal: rand.Intn(2) == 1,
		}

		err := createPost(post, i)
		if err != nil {
			log.Printf("❌ [%d/50] Error: %v", i+1, err)
			failCount++
		} else {
			successCount++
		}

		// Random delay between 100-500ms to ensure different created_at timestamps
		delay := time.Duration(100+rand.Intn(400)) * time.Millisecond
		time.Sleep(delay)
	}

	log.Println("")
	log.Println("==================================================")
	log.Printf("✅ Success: %d posts", successCount)
	log.Printf("❌ Failed: %d posts", failCount)
	log.Printf("📊 Total: %d posts", successCount+failCount)
	log.Println("==================================================")
	log.Println("")
	log.Println("🎉 Done! You can now test cursor-based pagination with:")
	log.Println("   GET http://localhost:8080/api/v1/posts?limit=20&sort=new")
	log.Println("   GET http://localhost:8080/api/v1/posts?limit=20&sort=hot")
	log.Println("   GET http://localhost:8080/api/v1/posts?limit=20&sort=top")
}
