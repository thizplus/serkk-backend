# Performance Analysis - Post Feed Queries

## 📊 Current Status (หลังแก้ GORM Preload bug)

### ✅ ข่าวดี: GORM ทำ Batch Loading อัตโนมัติแล้ว!

หลังจากลบ `Joins()` ออกจาก Preload แล้ว, GORM จะใช้ batch loading อัตโนมัติ

**จำนวน Queries ปัจจุบัน: 6 queries** (ไม่ใช่ 8 อย่างที่คิด!)

```
Query #1: SELECT * FROM "users" WHERE id IN (...)           // Load authors
Query #2: SELECT * FROM "post_media" WHERE post_id IN (...) // Load junction table
Query #3: SELECT * FROM "media" WHERE id IN (...)           // Load media
Query #4: SELECT * FROM "post_tags" WHERE post_id IN (...)  // Load junction table
Query #5: SELECT * FROM "tags" WHERE id IN (...)            // Load tags
Query #6: SELECT * FROM "posts" WHERE ... LIMIT 20          // Load posts
```

### 📈 Performance Metrics

**Request สำหรับ 20 posts:**
- Total queries: **6 queries**
- Total time: **~52ms**
- Avg per query: **~8.7ms**

**Estimated Load:**
- 100 req/sec = **600 queries/sec** ✅ สบาย
- 1,000 req/sec = **6,000 queries/sec** ⚠️ เริ่มหนัก
- 10,000 req/sec = **60,000 queries/sec** ❌ ต้องใช้ cache

---

## 🎯 แผนการปรับปรุง

### Phase 1: Database Indexes (ทำก่อน) ⭐

สร้าง indexes ที่จำเป็น เพื่อ optimize queries ที่มีอยู่

```sql
-- 1. Index สำหรับ hot score (Query #6)
CREATE INDEX CONCURRENTLY idx_posts_hot_score
ON posts ((votes / POWER((EXTRACT(EPOCH FROM (NOW() - created_at)) / 3600.0) + 2, 1.5)))
WHERE is_deleted = false AND status = 'published';

-- 2. Index สำหรับ created_at sorting
CREATE INDEX CONCURRENTLY idx_posts_created_at
ON posts (created_at DESC)
WHERE is_deleted = false AND status = 'published';

-- 3. Index สำหรับ votes sorting
CREATE INDEX CONCURRENTLY idx_posts_votes
ON posts (votes DESC)
WHERE is_deleted = false AND status = 'published';

-- 4. Composite index สำหรับ batch loading media
CREATE INDEX CONCURRENTLY idx_post_media_composite
ON post_media (post_id, display_order);

-- 5. Index สำหรับ batch loading tags
CREATE INDEX CONCURRENTLY idx_post_tags_post_id
ON post_tags (post_id);

-- 6. Index สำหรับ author lookup
CREATE INDEX CONCURRENTLY idx_users_id_username
ON users (id) INCLUDE (username, display_name, karma);
```

**ผลลัพธ์ที่คาดหวัง:**
- Query time ลดลง 30-50%
- จาก ~52ms → ~25-35ms

---

### Phase 2: Redis Caching (ต่อมา)

Cache hot feeds เพื่อลด database load

```go
// Cache strategy
type CacheConfig struct {
    HotFeed:  5 * time.Minute,  // Cache hot posts 5 นาที
    NewFeed:  1 * time.Minute,  // Cache new posts 1 นาที
    TopFeed:  10 * time.Minute, // Cache top posts 10 นาที
}

// Implementation
func (s *PostService) GetFeed(sortBy string, page int) ([]*dto.PostResponse, error) {
    cacheKey := fmt.Sprintf("feed:%s:page:%d", sortBy, page)

    // Try cache first
    if cached, err := s.cache.Get(cacheKey); err == nil {
        return cached, nil
    }

    // Cache miss - query database
    posts, err := s.repo.List(ctx, offset, limit, sortBy)
    if err != nil {
        return nil, err
    }

    // Cache result
    s.cache.Set(cacheKey, posts, s.getCacheTTL(sortBy))

    return posts, nil
}
```

**ผลลัพธ์ที่คาดหวัง (90% cache hit rate):**
- Database queries ลดลง 90%
- จาก 6,000 queries/sec → 600 queries/sec

---

### Phase 3: Query Optimization (ถ้ายังไม่พอ)

ลด queries จาก 6 → 4 โดยรวม junction table กับ data table

```go
// ปัจจุบัน: 2 queries สำหรับ media
Query #2: post_media (junction)
Query #3: media (data)

// หลังปรับปรุง: 1 query เดียว
SELECT media.*, post_media.post_id, post_media.display_order
FROM media
INNER JOIN post_media ON post_media.media_id = media.id
WHERE post_media.post_id IN (...)
ORDER BY post_media.post_id, post_media.display_order
```

**ผลลัพธ์ที่คาดหวัง:**
- จาก 6 queries → 4 queries (ลด 33%)
- จาก ~52ms → ~35ms

---

## 💰 Cost Analysis

### ปัจจุบัน (6 queries, ไม่มี cache)

**Traffic: 1,000 req/sec**
- Queries/sec: 6,000
- Database: t3.medium ($100/month)
- **Total: $100/month**

### Phase 1 (Indexes)

**Traffic: 1,000 req/sec**
- Queries/sec: 6,000 (ยังเท่าเดิม)
- Response time: -40% (เร็วขึ้น)
- Database: t3.medium ($100/month)
- **Total: $100/month**
- **Benefit: User experience ดีขึ้น**

### Phase 2 (Indexes + Cache)

**Traffic: 1,000 req/sec (90% cache hit)**
- Queries/sec: 600 (ลด 90%)
- Database: t3.small ($50/month)
- Redis: cache.t3.micro ($20/month)
- **Total: $70/month**
- **Savings: $30/month (30%)**

### Phase 3 (Indexes + Cache + Optimized Queries)

**Traffic: 1,000 req/sec (90% cache hit)**
- Queries/sec: 400 (ลด 93%)
- Database: t3.micro ($30/month)
- Redis: cache.t3.micro ($20/month)
- **Total: $50/month**
- **Savings: $50/month (50%)**

---

## 🚀 Implementation Timeline

### Week 1: Database Indexes ⭐ (แนะนำเริ่มที่นี่)

**Day 1-2: สร้าง migration files**
- [ ] สร้าง migration สำหรับ indexes
- [ ] Test ใน development
- [ ] Verify query performance improvement

**Day 3-4: Deploy to production**
- [ ] Deploy migrations (CONCURRENTLY เพื่อไม่ lock table)
- [ ] Monitor performance metrics
- [ ] Verify no regression

**Day 5: Analyze results**
- [ ] Compare before/after metrics
- [ ] Document findings

**Expected Results:**
- ✅ Response time: -30-40%
- ✅ Zero downtime deployment
- ✅ No code changes needed

---

### Week 2-3: Redis Caching (ถ้าต้องการ scale เพิ่ม)

**Day 1-3: Setup Redis**
- [ ] Setup Redis instance
- [ ] Implement cache layer
- [ ] Add cache key strategy

**Day 4-5: Implement caching**
- [ ] Cache hot/new/top feeds
- [ ] Add cache invalidation
- [ ] Add monitoring

**Day 6-7: Deploy and monitor**
- [ ] Deploy to production
- [ ] Monitor cache hit rate
- [ ] Tune cache TTL

**Expected Results:**
- ✅ 90% cache hit rate
- ✅ Database load: -90%
- ✅ Cost savings: $30/month

---

### Week 4: Query Optimization (optional)

**Only if still need more optimization**

- [ ] Implement custom batch loading
- [ ] Reduce from 6 → 4 queries
- [ ] Test and deploy

---

## 📋 Recommendation

### ✅ ทำตอนนี้: Phase 1 (Database Indexes)

**เหตุผล:**
1. **Impact สูง, Effort ต่ำ** - แค่สร้าง migration
2. **Zero risk** - ไม่กระทบ code
3. **ไม่มี cost เพิ่ม** - แค่ใช้ disk space เพิ่ม
4. **Immediate benefit** - Response time ดีขึ้นทันที

### ⏳ ทำภายหลัง: Phase 2 (Redis Cache)

**เมื่อไหร่:**
- เมื่อ traffic > 500 req/sec
- เมื่อ database load > 70%
- เมื่อมี budget สำหรับ Redis

### 🤔 พิจารณา: Phase 3 (Query Optimization)

**เมื่อไหร่:**
- เมื่อ Phase 1 + 2 ยังไม่พอ
- เมื่อ traffic > 5,000 req/sec
- เมื่อต้องการ squeeze performance สุดๆ

---

## 🎯 Next Steps

**ผมแนะนำเริ่มจาก Phase 1 (Indexes) ก่อน:**

1. ✅ สร้าง migration files สำหรับ indexes
2. ✅ Test ใน development
3. ✅ Deploy to production
4. ✅ Monitor และ measure improvement

**คุณพร้อมให้ผมสร้าง migration files สำหรับ indexes ไหมครับ?** 🚀
