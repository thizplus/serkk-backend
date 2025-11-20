#!/usr/bin/env python3
"""
Import topics from JSON to create auto-post settings via API
Usage: python import_topics.py topics.json --api-url http://localhost:3000 --token YOUR_JWT_TOKEN
"""

import json
import sys
import argparse
import requests
from time import sleep

def import_topics(json_file, api_url, jwt_token, bot_user_id):
    """Import topics from JSON file and create settings via API"""

    # Read JSON
    with open(json_file, 'r', encoding='utf-8') as f:
        data = json.load(f)

    settings = data.get('settings', [])
    total = len(settings)

    print(f"📦 Found {total} settings to import")
    print(f"🎯 Total topics: {data.get('total_topics', 0)}")
    print("")

    headers = {
        'Authorization': f'Bearer {jwt_token}',
        'Content-Type': 'application/json'
    }

    success_count = 0
    failed_count = 0

    for i, setting in enumerate(settings, 1):
        print(f"[{i}/{total}] Creating setting: {setting['name']}...")

        # Prepare request body
        body = {
            'botUserId': bot_user_id,
            'isEnabled': False,  # เริ่มต้นปิดไว้ก่อน
            'cronSchedule': '0 * * * *',  # ทุกชั่วโมง
            'model': 'gpt-4o-mini',
            'tone': setting['tone'],
            'topics': setting['topics'],
            'maxTokens': 1500,
            'enableVariations': True,
            'requireApproval': False,
            'useBatchMode': False,
            'batchSize': 1
        }

        try:
            response = requests.post(
                f'{api_url}/api/v1/auto-post/settings',
                headers=headers,
                json=body,
                timeout=10
            )

            if response.status_code in [200, 201]:
                result = response.json()
                print(f"  ✅ Success! ID: {result.get('id')}")
                print(f"     Topics: {len(setting['topics'])} | Tone: {setting['tone']}")
                success_count += 1
            else:
                print(f"  ❌ Failed! Status: {response.status_code}")
                print(f"     Error: {response.text}")
                failed_count += 1

        except Exception as e:
            print(f"  ❌ Error: {str(e)}")
            failed_count += 1

        # Rate limiting
        sleep(0.5)
        print("")

    print("="*60)
    print(f"📊 Import Summary:")
    print(f"  ✅ Success: {success_count}")
    print(f"  ❌ Failed: {failed_count}")
    print(f"  📝 Total: {total}")
    print("")

    if success_count > 0:
        print("🎯 Next steps:")
        print("  1. Review settings: GET /api/v1/auto-post/settings")
        print("  2. Test manual trigger for each setting")
        print("  3. Enable settings: POST /api/v1/auto-post/settings/{id}/enable")

def enable_all_settings(api_url, jwt_token):
    """Enable all auto-post settings"""

    headers = {
        'Authorization': f'Bearer {jwt_token}',
        'Content-Type': 'application/json'
    }

    # Get all settings
    response = requests.get(
        f'{api_url}/api/v1/auto-post/settings',
        headers=headers,
        params={'limit': 100}
    )

    if response.status_code != 200:
        print(f"❌ Failed to fetch settings: {response.status_code}")
        return

    settings = response.json().get('settings', [])
    print(f"📦 Found {len(settings)} settings")
    print("")

    enabled_count = 0
    for setting in settings:
        setting_id = setting['id']
        name = f"{setting.get('tone', 'unknown')} ({len(setting.get('topics', []))} topics)"

        print(f"Enabling: {name}...")

        try:
            response = requests.post(
                f'{api_url}/api/v1/auto-post/settings/{setting_id}/enable',
                headers=headers
            )

            if response.status_code == 200:
                print(f"  ✅ Enabled!")
                enabled_count += 1
            else:
                print(f"  ❌ Failed: {response.status_code}")

        except Exception as e:
            print(f"  ❌ Error: {str(e)}")

        sleep(0.3)

    print("")
    print(f"✅ Enabled {enabled_count}/{len(settings)} settings")

if __name__ == '__main__':
    parser = argparse.ArgumentParser(description='Import auto-post topics')
    parser.add_argument('json_file', help='Topics JSON file')
    parser.add_argument('--api-url', default='http://localhost:3000', help='API base URL')
    parser.add_argument('--token', required=True, help='JWT authentication token')
    parser.add_argument('--bot-user-id', required=True, help='Bot user UUID')
    parser.add_argument('--enable-all', action='store_true', help='Enable all settings after import')

    args = parser.parse_args()

    print("🚀 Auto-Post Topics Importer")
    print("="*60)
    print(f"📁 File: {args.json_file}")
    print(f"🌐 API: {args.api_url}")
    print(f"🤖 Bot User: {args.bot_user_id}")
    print("="*60)
    print("")

    import_topics(args.json_file, args.api_url, args.token, args.bot_user_id)

    if args.enable_all:
        print("")
        print("🔄 Enabling all settings...")
        print("")
        enable_all_settings(args.api_url, args.token)
