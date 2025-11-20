package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	API_URL      = "http://localhost:8080/api/v1/posts"
	TOKEN        = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6InRoZXB0aGFpLmptQGdtYWlsLmNvbSIsImV4cCI6MTc2Mzc0NDI1MywiaWF0IjoxNzYzMTM5NDUzLCJyb2xlIjoidXNlciIsInVzZXJfaWQiOiI0YWExMGUxYi0wNmM0LTRiMDktOGJkOS01Y2VhOTRjZDM3MjMiLCJ1c2VybmFtZSI6InRoZXB0aGFpIn0.UgZeYGOU7JdsShLnAfunzpnmo8H7XfkelIlfUNq5pgY"
	TOTAL_POSTS  = 5000
	BATCH_SIZE   = 5    // จำนวน concurrent requests (ลดลงเพื่อหลีก rate limit)
	RETRY_COUNT  = 3    // จำนวนครั้งที่ retry
	DELAY_MS     = 100  // delay ระหว่าง requests (ms)
)

type CreatePostRequest struct {
	Title    string      `json:"title"`
	Content  string      `json:"content"`
	MediaIDs []uuid.UUID `json:"mediaIds"`
	Tags     []string    `json:"tags"`
	IsDraft  bool        `json:"isDraft"`
}

type Media struct {
	ID  uuid.UUID `gorm:"type:uuid;primaryKey"`
	URL string    `gorm:"type:text"`
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
		"GraphQL vs REST API",
		"สอน JavaScript สำหรับมือใหม่",
		"Python สำหรับ Data Science",
		"Machine Learning เบื้องต้น",
		"Deep Learning with TensorFlow",
		"Natural Language Processing",
		"Computer Vision Basics",
		"Blockchain Technology",
		"Smart Contracts with Solidity",
		"Web3 Development",
		"NFT Development Guide",
		"DeFi Protocols explained",
		"Cybersecurity Best Practices",
		"Ethical Hacking Basics",
		"Penetration Testing Guide",
		"Network Security Fundamentals",
		"Cloud Security AWS",
		"DevOps Best Practices",
		"Infrastructure as Code",
		"Terraform Tutorial",
		"Ansible Automation",
		"Jenkins CI/CD Pipeline",
		"Monitoring with Prometheus",
		"Logging with ELK Stack",
		"Container Orchestration",
		"Service Mesh with Istio",
		"API Gateway Patterns",
		"Load Balancing Strategies",
		"Caching Strategies",
		"Database Sharding",
		"Database Replication",
		"Backup and Recovery",
		"Disaster Recovery Planning",
		"High Availability Systems",
		"Scalability Patterns",
		"Performance Optimization",
		"Code Review Best Practices",
		"Agile Development",
		"Scrum Framework",
		"Kanban Methodology",
		"Project Management",
		"Technical Documentation",
		"API Documentation",
		"Software Architecture",
		"System Design Interview",
		"Algorithm Design",
		"Data Structures",
		"Time Complexity",
		"Space Complexity",
		"Dynamic Programming",
		"Graph Algorithms",
		"Tree Algorithms",
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
		"Tutorial แบบ step-by-step พร้อมตัวอย่างโค้ดและภาพประกอบ\n\nเหมาะสำหรับคนที่เพิ่งเริ่มต้น จะมีการอธิบายแนวคิดและหลักการพื้นฐาน\n\nรวมถึง best practices และ common pitfalls ที่ควรหลีกเลี่ยง",
		"Deep dive ลงในรายละเอียดเชิงเทคนิค พร้อมแนวทางการแก้ปัญหา\n\nวิเคราะห์ edge cases ต่างๆ และวิธีการจัดการ\n\nเหมาะสำหรับคนที่มี experience แล้วอยากเจาะลึกมากขึ้น",
		"เปรียบเทียบแต่ละแนวทาง pros and cons อย่างละเอียด\n\nช่วยให้คุณตัดสินใจเลือกเครื่องมือหรือวิธีการที่เหมาะสม\n\nพร้อมตัวอย่าง use cases จริงในแต่ละสถานการณ์",
		"รวม tips และ tricks ที่นำไปใช้ได้จริง ช่วยประหยัดเวลา\n\nเรียนรู้จากประสบการณ์ของคนที่ผ่านมาแล้ว\n\nหลีกเลี่ยงข้อผิดพลาดที่คนอื่นเคยทำ",
		"Architecture patterns และการออกแบบระบบที่ scalable\n\nอธิบายการออกแบบ component ต่างๆ และความสัมพันธ์\n\nพร้อมแนวทางการ implement ในโปรเจค production",
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
	}
)

func connectDB() (*gorm.DB, error) {
	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️  Warning: .env file not found, using environment variables")
	}

	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbSSLMode := os.Getenv("DB_SSL_MODE")

	if dbSSLMode == "" {
		dbSSLMode = "disable"
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dbHost, dbPort, dbUser, dbPassword, dbName, dbSSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	return db, nil
}

func getMediaIDs(db *gorm.DB) ([]uuid.UUID, error) {
	var media []Media
	err := db.Table("media").Find(&media).Error
	if err != nil {
		return nil, err
	}

	if len(media) == 0 {
		return nil, fmt.Errorf("no media found in database")
	}

	mediaIDs := make([]uuid.UUID, len(media))
	for i, m := range media {
		mediaIDs[i] = m.ID
	}

	return mediaIDs, nil
}

func randomMediaIDs(allMediaIDs []uuid.UUID) []uuid.UUID {
	if len(allMediaIDs) == 0 {
		return []uuid.UUID{}
	}

	// Random 1-10 images per post (max limit)
	count := rand.Intn(10) + 1
	if count > len(allMediaIDs) {
		count = len(allMediaIDs)
	}

	// Shuffle and pick first N
	shuffled := make([]uuid.UUID, len(allMediaIDs))
	copy(shuffled, allMediaIDs)
	rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	return shuffled[:count]
}

func createPost(post CreatePostRequest, index int) error {
	var lastErr error

	for attempt := 1; attempt <= RETRY_COUNT; attempt++ {
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

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("failed to send request: %w", err)
			if attempt < RETRY_COUNT {
				time.Sleep(time.Duration(attempt*500) * time.Millisecond)
				continue
			}
			return lastErr
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusOK {
			return nil
		}

		lastErr = fmt.Errorf("status %d: %s", resp.StatusCode, string(body))

		// Retry on server errors (5xx) or rate limit (429)
		if resp.StatusCode >= 500 || resp.StatusCode == 429 {
			if attempt < RETRY_COUNT {
				time.Sleep(time.Duration(attempt*1000) * time.Millisecond)
				continue
			}
		}

		// Don't retry on client errors (4xx except 429)
		if resp.StatusCode >= 400 && resp.StatusCode < 500 && resp.StatusCode != 429 {
			return lastErr
		}
	}

	return lastErr
}

func worker(id int, jobs <-chan CreatePostRequest, results chan<- error, wg *sync.WaitGroup, successCount *int, failCount *int, mu *sync.Mutex, errorLog *[]string) {
	defer wg.Done()

	for post := range jobs {
		err := createPost(post, id)
		if err == nil {
			mu.Lock()
			*successCount++
			count := *successCount
			mu.Unlock()

			if count%100 == 0 {
				log.Printf("✅ Progress: %d/%d posts created", count, TOTAL_POSTS)
			}
		} else {
			mu.Lock()
			*failCount++
			// Only log first 10 errors
			if len(*errorLog) < 10 {
				*errorLog = append(*errorLog, fmt.Sprintf("Post '%s': %v", post.Title, err))
			}
			mu.Unlock()
		}
		results <- err

		// Delay between requests
		time.Sleep(time.Duration(DELAY_MS) * time.Millisecond)
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())

	log.Println("🚀 Starting to create 5000 posts with media...")
	log.Println("📝 API URL:", API_URL)
	log.Println("")

	// Connect to database
	log.Println("🔌 Connecting to database...")
	db, err := connectDB()
	if err != nil {
		log.Fatalf("❌ Failed to connect database: %v", err)
	}
	log.Println("✅ Database connected!")

	// Get all media IDs
	log.Println("🖼️  Fetching media from database...")
	mediaIDs, err := getMediaIDs(db)
	if err != nil {
		log.Fatalf("❌ Failed to get media: %v", err)
	}
	log.Printf("✅ Found %d media items in database", len(mediaIDs))
	log.Println("")

	// Prepare jobs
	jobs := make(chan CreatePostRequest, TOTAL_POSTS)
	results := make(chan error, TOTAL_POSTS)

	var wg sync.WaitGroup
	var mu sync.Mutex
	successCount := 0
	failCount := 0
	errorLog := []string{}

	// Start workers
	log.Printf("🔧 Starting %d concurrent workers...", BATCH_SIZE)
	for w := 1; w <= BATCH_SIZE; w++ {
		wg.Add(1)
		go worker(w, jobs, results, &wg, &successCount, &failCount, &mu, &errorLog)
	}

	// Send jobs
	log.Println("📤 Sending jobs to workers...")
	startTime := time.Now()

	for i := 0; i < TOTAL_POSTS; i++ {
		titleIdx := i % len(titles)
		contentIdx := rand.Intn(len(contents))
		tagsIdx := rand.Intn(len(tags))

		post := CreatePostRequest{
			Title:    titles[titleIdx],
			Content:  contents[contentIdx],
			MediaIDs: randomMediaIDs(mediaIDs),
			Tags:     tags[tagsIdx],
			IsDraft:  false,
		}

		jobs <- post
	}
	close(jobs)

	// Wait for all workers to finish
	log.Println("⏳ Waiting for all workers to finish...")
	wg.Wait()
	close(results)

	// Drain results
	for range results {
	}

	elapsed := time.Since(startTime)

	log.Println("")
	log.Println("==================================================")
	log.Printf("✅ Success: %d posts", successCount)
	log.Printf("❌ Failed: %d posts", failCount)
	log.Printf("📊 Total: %d posts", successCount+failCount)
	log.Printf("⏱️  Time: %.2f seconds", elapsed.Seconds())
	if successCount > 0 {
		log.Printf("🚀 Speed: %.2f posts/second", float64(successCount)/elapsed.Seconds())
	}
	log.Println("==================================================")

	// Show error samples
	if len(errorLog) > 0 {
		log.Println("")
		log.Println("❌ Sample errors (first 10):")
		for i, errMsg := range errorLog {
			log.Printf("   %d. %s", i+1, errMsg)
		}
	}
	log.Println("")
	log.Println("🎉 Done! You can now test cursor-based pagination with:")
	log.Println("   GET http://localhost:8080/api/v1/posts?limit=20&sort=new")
	log.Println("   GET http://localhost:8080/api/v1/posts?limit=20&sort=hot")
	log.Println("   GET http://localhost:8080/api/v1/posts?limit=20&sort=top")
	log.Println("")
	log.Println("🔍 Test deep pagination:")
	log.Println("   Load 20 posts per page and keep loading with cursor")
	log.Println("   Check that posts don't duplicate and performance stays fast!")
}
