# Post Feed Optimization - เปรียบเทียบแนวทาง

## 📊 เปรียบเทียบ 3 แนวทาง

### 1️⃣ แบบปัจจุบัน (GORM Preload)

```go
Preload("Author").
Preload("Media").
Preload("Tags").
Preload("SourcePost").
Preload("SourcePost.Author").
Preload("SourcePost.Media").
Preload("SourcePost.Tags")
```

**จำนวน Queries:** 8 queries
- 1 query สำหรับ posts
- 7 queries สำหรับ relationships

**ข้อดี:**
- ✅ เขียนง่าย
- ✅ Maintainable
- ✅ Type-safe

**ข้อเสีย:**
- ❌ N+1 problem
- ❌ Slow (8 database roundtrips)
- ❌ High database load

**Performance:**
- 20 posts = **8 queries**
- 1000 requests/sec = **8,000 queries/sec** 🔥

---

### 2️⃣ Raw SQL (เขียนเอง 100%)

```go
query := `
    SELECT
        p.*,
        u.id, u.username, u.display_name,
        m.id, m.url, m.type,
        t.id, t.name
    FROM posts p
    LEFT JOIN users u ON p.author_id = u.id
    LEFT JOIN post_media pm ON p.id = pm.post_id
    LEFT JOIN media m ON pm.media_id = m.id
    LEFT JOIN post_tags pt ON p.id = pt.post_id
    LEFT JOIN tags t ON pt.tag_id = t.id
    WHERE p.is_deleted = false
    ORDER BY p.created_at DESC
    LIMIT ? OFFSET ?
`
```

**จำนวน Queries:** 1 query ⚡

**ข้อดี:**
- ✅ เร็วที่สุด (1 query)
- ✅ Full control

**ข้อเสีย:**
- ❌ ต้องเขียน manual mapping
- ❌ เสี่ยง SQL injection
- ❌ ไม่ type-safe
- ❌ Maintenance ยาก
- ❌ Duplicate rows (Cartesian product ถ้า JOIN 1:N)

**Performance:**
- 20 posts = **1 query**
- แต่ต้อง manual mapping ทำให้ code ซับซ้อน

---

### 3️⃣ GORM Optimized (แนะนำ ⭐)

```go
// Step 1: Fetch posts
posts = db.Find(&posts).Limit(20)

// Step 2: Batch load authors (1 query)
authorIDs := extractAuthorIDs(posts)
authors = db.Where("id IN ?", authorIDs).Find(&authors)

// Step 3: Batch load media (1 query)
postIDs := extractPostIDs(posts)
media = db.Table("media").
    Joins("JOIN post_media ON ...").
    Where("post_id IN ?", postIDs).
    Find(&media)

// Step 4: Batch load tags (1 query)
tags = db.Table("tags").
    Joins("JOIN post_tags ON ...").
    Where("post_id IN ?", postIDs).
    Find(&tags)

// Step 5: Group data in memory
mapDataToPosts(posts, authors, media, tags)
```

**จำนวน Queries:** 4 queries
- 1 query สำหรับ posts
- 3 queries สำหรับ batch loading (authors, media, tags)

**ข้อดี:**
- ✅ เร็ว (4 queries vs 8 queries = **50% faster**)
- ✅ Type-safe (ใช้ GORM models)
- ✅ Maintainable
- ✅ ไม่มี Cartesian product
- ✅ Batch loading = efficient
- ✅ รองรับ caching ได้ง่าย

**ข้อเสีย:**
- ⚠️ ต้องเขียน grouping logic
- ⚠️ Code ยาวกว่า Preload

**Performance:**
- 20 posts = **4 queries**
- 1000 requests/sec = **4,000 queries/sec** (ลดลง 50%)

---

## 🎯 คำแนะนำ

### สำหรับระบบคุณ (Post Feed):

**ใช้แนวทาง 3: GORM Optimized + Redis Cache**

#### Phase 1: Optimize Queries (ทำเลย)
```go
// แทนที่ 8 Preloads ด้วย 4 batch queries
List() -> 4 queries
ListByAuthor() -> 4 queries
ListByTag() -> 4 queries
```

#### Phase 2: Add Caching (ต่อจากนั้น)
```go
// Cache hot posts in Redis
Cache Key: "feed:hot:page:1" -> TTL 5 minutes
Cache Key: "feed:new:page:1" -> TTL 1 minute
```

#### Phase 3: Add Database Indexes (สำคัญ!)
```sql
-- Index สำหรับ hot score
CREATE INDEX idx_posts_hot_score ON posts (
    (votes / POWER((EXTRACT(EPOCH FROM (NOW() - created_at)) / 3600.0) + 2, 1.5)) DESC
);

-- Index สำหรับ batch loading
CREATE INDEX idx_post_media_batch ON post_media (post_id, display_order);
CREATE INDEX idx_post_tags_batch ON post_tags (post_id);
```

---

## 📈 Performance Comparison

| Scenario | Current (Preload) | Raw SQL | GORM Optimized | Improvement |
|----------|-------------------|---------|----------------|-------------|
| Queries per request | 8 | 1 | 4 | **50% ↓** |
| Code complexity | Low | High | Medium | ✅ |
| Type safety | ✅ | ❌ | ✅ | ✅ |
| Maintenance | Easy | Hard | Medium | ✅ |
| Scalability | Poor | Good | Good | ✅ |
| Cache-friendly | ❌ | ⚠️ | ✅ | ✅ |

---

## 🚀 Implementation Plan

### Week 1: Core Optimization
- [ ] Implement `List()` with batch loading
- [ ] Implement `ListByAuthor()` with batch loading
- [ ] Add unit tests
- [ ] Benchmark tests

### Week 2: Extended Methods
- [ ] Implement `ListByTag()`
- [ ] Implement `Search()`
- [ ] Implement `GetCrossposts()`

### Week 3: Caching Layer
- [ ] Add Redis caching for hot/new feeds
- [ ] Implement cache invalidation
- [ ] Add cache warming

### Week 4: Database Optimization
- [ ] Add indexes
- [ ] Query performance analysis
- [ ] Load testing

---

## 💰 Cost Savings Estimate

**Current:**
- 1000 req/sec × 8 queries = 8,000 queries/sec
- Database: $200/month

**After Optimization:**
- 1000 req/sec × 4 queries = 4,000 queries/sec
- Database: $100/month
- Redis cache: $20/month
- **Total savings: $80/month**

**After Caching (90% cache hit rate):**
- 1000 req/sec × 10% × 4 queries = 400 queries/sec
- Database: $30/month
- Redis cache: $20/month
- **Total savings: $150/month**

---

## ✅ Recommendation

**ใช้ GORM Optimized (แนวทาง 3) เพราะ:**

1. **Balance ระหว่าง Performance & Maintainability**
2. **ลด queries ได้ 50% ทันทีโดยไม่ต้อง sacrifice type safety**
3. **รองรับ caching ได้ง่าย** (แค่ cache ที่ layer service)
4. **ไม่ต้องกลับมาแก้ใหม่** (scalable ถึง millions of users)
5. **Team สามารถ maintain ได้** (ไม่ยากเกินไป)

คุณต้องการให้ผมเริ่ม implement แนวทาง 3 ให้เลยไหมครับ? 🚀
