#!/usr/bin/env python3
"""
Convert CSV topics file to JSON format for auto-post settings
Usage: python convert_topics_to_json.py topics.csv
"""

import csv
import json
import sys
from collections import defaultdict

def convert_csv_to_json(csv_file, output_file='topics.json'):
    """Convert CSV file to JSON format grouped by category and tone"""

    # Read CSV
    topics_by_category = defaultdict(lambda: defaultdict(list))

    with open(csv_file, 'r', encoding='utf-8') as f:
        reader = csv.DictReader(f)
        for row in reader:
            category = row.get('category', 'general')
            tone = row.get('tone', 'neutral')
            topic = row.get('topic', '').strip()

            if topic:
                topics_by_category[category][tone].append(topic)

    # Create settings for each category/tone combination
    settings = []

    for category, tones in topics_by_category.items():
        for tone, topics in tones.items():
            # Split into chunks of 50 (max topics per setting)
            chunk_size = 50
            for i in range(0, len(topics), chunk_size):
                chunk = topics[i:i+chunk_size]

                setting = {
                    "name": f"{category}_{tone}_{i//chunk_size + 1}",
                    "category": category,
                    "tone": tone,
                    "topics": chunk,
                    "topics_count": len(chunk)
                }
                settings.append(setting)

    # Write JSON
    output_data = {
        "total_topics": sum(len(topics) for tones in topics_by_category.values() for topics in tones.values()),
        "total_settings": len(settings),
        "settings": settings
    }

    with open(output_file, 'w', encoding='utf-8') as f:
        json.dump(output_data, f, ensure_ascii=False, indent=2)

    print(f"✅ Converted {output_data['total_topics']} topics into {output_data['total_settings']} settings")
    print(f"📄 Output: {output_file}")

    # Print summary
    print("\n📊 Summary by category:")
    for category, tones in topics_by_category.items():
        total = sum(len(topics) for topics in tones.values())
        print(f"  - {category}: {total} topics")
        for tone, topics in tones.items():
            print(f"    • {tone}: {len(topics)} topics")

def generate_sample_csv(output_file='sample_topics.csv'):
    """Generate a sample CSV file with 720 topics"""

    categories = {
        'platform_issues': [
            'ค่า fee แพงเกินไป - ร้านค้าลำบาก',
            'Delivery ช้า ลูกค้ารอนาน',
            'App crash บ่อย ใช้งานลำบาก',
            'ค่าบริการแพงกว่าที่อื่น',
            'ระบบการจ่ายเงินมีปัญหา',
        ],
        'rider_issues': [
            'Rider ได้เงินน้อย แต่เหนื่อยมาก',
            'ไม่มี insurance ไม่มีสวัสดิการ',
            'ต้องรอ order นาน ไม่มีรายได้',
            'ค่าน้ำมันแพง แต่ค่าส่งไม่พอ',
            'ถูกลูกค้าด่า ไม่มีใครช่วย',
        ],
        'restaurant_tips': [
            'วิธีเพิ่มยอดขาย 5 เทคนิค',
            'ทำ menu ให้ขายดี',
            'จัดการ cost ให้ได้กำไร',
            'Marketing ร้านอาหารยุคใหม่',
            'Customer service ที่ดี',
        ],
        'customer_tips': [
            'สั่งอาหารให้คุ้ม 10 วิธี',
            'เลือกร้านอาหารที่ดี',
            'ใช้ promotion อย่างไรให้คุ้ม',
            'ทิป rider ควรเท่าไหร่',
            'อาหารมาไม่ครบต้องทำไง',
        ],
        'industry_news': [
            'สถิติ food delivery ในไทย',
            'แนวโน้มธุรกิจอาหาร 2025',
            'เทคโนโลยีใหม่ในวงการ',
            'การแข่งขันในตลาด',
            'กฎหมายใหม่ที่ผู้ประกอบการต้องรู้',
        ]
    }

    tones = ['neutral', 'casual', 'professional', 'humorous', 'controversial']

    # Generate 720 topics (expand templates)
    topics = []
    target = 720

    while len(topics) < target:
        for category, templates in categories.items():
            for template in templates:
                for tone in tones:
                    if len(topics) >= target:
                        break

                    # Add variations
                    variations = [
                        template,
                        f"{template} - ประสบการณ์จริง",
                        f"{template} - สิ่งที่ต้องรู้",
                        f"{template} - มุมมองใหม่",
                        f"{template} - เคล็ดลับเด็ด",
                    ]

                    for var in variations:
                        if len(topics) >= target:
                            break
                        topics.append({
                            'category': category,
                            'topic': var,
                            'tone': tone
                        })

    # Write CSV
    with open(output_file, 'w', encoding='utf-8', newline='') as f:
        writer = csv.DictWriter(f, fieldnames=['category', 'topic', 'tone'])
        writer.writeheader()
        writer.writerows(topics[:target])

    print(f"✅ Generated {len(topics[:target])} sample topics")
    print(f"📄 Output: {output_file}")

if __name__ == '__main__':
    if len(sys.argv) > 1:
        if sys.argv[1] == '--generate-sample':
            generate_sample_csv()
        else:
            convert_csv_to_json(sys.argv[1])
    else:
        print("Usage:")
        print("  Generate sample CSV:")
        print("    python convert_topics_to_json.py --generate-sample")
        print("")
        print("  Convert CSV to JSON:")
        print("    python convert_topics_to_json.py topics.csv")
