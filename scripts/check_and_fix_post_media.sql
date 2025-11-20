-- ============================================
-- Script: ตรวจสอบและแก้ไขปัญหา post_media ที่มี media มากเกินไป
-- Created: 2025-01-16
-- ============================================

-- ============================================
-- STEP 1: ตรวจสอบปัญหา
-- ============================================

-- 1.1 ดูจำนวน media ในแต่ละ post (Top 20 posts ที่มี media เยอะที่สุด)
SELECT
    post_id,
    COUNT(*) as media_count,
    ARRAY_AGG(media_id ORDER BY display_order) as media_ids
FROM post_media
GROUP BY post_id
HAVING COUNT(*) > 10
ORDER BY media_count DESC
LIMIT 20;

-- 1.2 สถิติรวม
SELECT
    'Total posts' as metric,
    COUNT(DISTINCT post_id) as count
FROM post_media
UNION ALL
SELECT
    'Posts with >10 media' as metric,
    COUNT(*) as count
FROM (
    SELECT post_id, COUNT(*) as cnt
    FROM post_media
    GROUP BY post_id
    HAVING COUNT(*) > 10
) subq
UNION ALL
SELECT
    'Total post_media records' as metric,
    COUNT(*) as count
FROM post_media;

-- ============================================
-- STEP 2: สำรองข้อมูลก่อนลบ (สร้างตารางสำรอง)
-- ============================================

-- สร้างตารางสำรองข้อมูล
CREATE TABLE IF NOT EXISTS post_media_backup AS
SELECT * FROM post_media WHERE 1=0;

-- สำรองข้อมูลที่จะถูกลบ (media ที่เกิน 10 ตัว)
INSERT INTO post_media_backup
SELECT pm.*
FROM post_media pm
INNER JOIN (
    SELECT
        post_id,
        media_id,
        ROW_NUMBER() OVER (PARTITION BY post_id ORDER BY display_order, media_id) as rn
    FROM post_media
) ranked
ON pm.post_id = ranked.post_id
AND pm.media_id = ranked.media_id
WHERE ranked.rn > 10;

-- ดูจำนวนที่จะถูกลบ
SELECT COUNT(*) as will_be_deleted FROM post_media_backup;

-- ============================================
-- STEP 3: ลบ media ที่เกิน 10 ตัว (เก็บแค่ 10 ตัวแรกตาม display_order)
-- ============================================

-- ⚠️ คำเตือน: ขั้นตอนนี้จะลบข้อมูลจริง กรุณาตรวจสอบก่อน!
-- ลบ media ที่เกิน 10 ตัว
DELETE FROM post_media
WHERE (post_id, media_id) IN (
    SELECT post_id, media_id
    FROM (
        SELECT
            post_id,
            media_id,
            ROW_NUMBER() OVER (PARTITION BY post_id ORDER BY display_order, media_id) as rn
        FROM post_media
    ) ranked
    WHERE rn > 10
);

-- ============================================
-- STEP 4: ตรวจสอบผลลัพธ์หลังลบ
-- ============================================

-- ดูจำนวน media ในแต่ละ post อีกครั้ง
SELECT
    post_id,
    COUNT(*) as media_count
FROM post_media
GROUP BY post_id
HAVING COUNT(*) > 10
ORDER BY media_count DESC;

-- ควรได้ผลลัพธ์ว่าไม่มี post ไหนที่มี media เกิน 10 แล้ว

-- ============================================
-- STEP 5: จัดเรียง display_order ใหม่
-- ============================================

-- จัดเรียง display_order ให้ต่อเนื่อง (0, 1, 2, 3, ...)
UPDATE post_media
SET display_order = subq.new_order - 1
FROM (
    SELECT
        post_id,
        media_id,
        ROW_NUMBER() OVER (PARTITION BY post_id ORDER BY display_order, media_id) as new_order
    FROM post_media
) subq
WHERE post_media.post_id = subq.post_id
AND post_media.media_id = subq.media_id;

-- ============================================
-- STEP 6: สรุปผลลัพธ์
-- ============================================

SELECT
    'Total posts' as metric,
    COUNT(DISTINCT post_id) as count
FROM post_media
UNION ALL
SELECT
    'Max media per post' as metric,
    MAX(cnt) as count
FROM (
    SELECT COUNT(*) as cnt
    FROM post_media
    GROUP BY post_id
) subq
UNION ALL
SELECT
    'Avg media per post' as metric,
    ROUND(AVG(cnt)::numeric, 2) as count
FROM (
    SELECT COUNT(*) as cnt
    FROM post_media
    GROUP BY post_id
) subq;
