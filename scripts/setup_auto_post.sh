#!/bin/bash

# Auto-Post Setup Script
# ช่วยตั้งค่า Auto-Post System แบบอัตโนมัติ

set -e

echo "🚀 Auto-Post Setup Script"
echo "=========================================="
echo ""

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Check if .env exists
if [ ! -f ".env" ]; then
    echo -e "${RED}❌ Error: .env file not found${NC}"
    echo "Please create .env file first"
    exit 1
fi

echo "Step 1: Checking configuration..."
echo ""

# Check CSV file
CSV_FILE="${1:-suekk_720_posts.csv}"
if [ ! -f "$CSV_FILE" ]; then
    echo -e "${RED}❌ Error: CSV file not found: $CSV_FILE${NC}"
    echo "Usage: ./setup_auto_post.sh [csv_file]"
    echo "Example: ./setup_auto_post.sh suekk_720_posts.csv"
    exit 1
fi

echo -e "${GREEN}✅ Found CSV file: $CSV_FILE${NC}"

# Check BOT_USER_ID in .env
if grep -q "^AUTO_POST_BOT_USER_ID=" .env; then
    BOT_USER_ID=$(grep "^AUTO_POST_BOT_USER_ID=" .env | cut -d '=' -f2)
    if [ -z "$BOT_USER_ID" ] || [ "$BOT_USER_ID" = "your-bot-user-uuid-here" ]; then
        echo -e "${RED}❌ Error: AUTO_POST_BOT_USER_ID not configured in .env${NC}"
        echo ""
        echo "Please set your bot user ID in .env:"
        echo "AUTO_POST_BOT_USER_ID=your-uuid-here"
        exit 1
    fi
    echo -e "${GREEN}✅ Bot User ID configured: $BOT_USER_ID${NC}"
else
    echo -e "${RED}❌ Error: AUTO_POST_BOT_USER_ID not found in .env${NC}"
    echo "Please add to .env:"
    echo "AUTO_POST_BOT_USER_ID=your-uuid-here"
    exit 1
fi

# Check OpenAI API Key
if grep -q "^OPENAI_API_KEY=" .env; then
    OPENAI_KEY=$(grep "^OPENAI_API_KEY=" .env | cut -d '=' -f2)
    if [ -z "$OPENAI_KEY" ] || [ "$OPENAI_KEY" = "sk-your-openai-api-key-here" ]; then
        echo -e "${YELLOW}⚠️  Warning: OPENAI_API_KEY not configured${NC}"
        echo "Auto-post will not work without OpenAI API key"
    else
        echo -e "${GREEN}✅ OpenAI API Key configured${NC}"
    fi
fi

echo ""
echo "=========================================="
echo "Step 2: Importing topics from CSV..."
echo "=========================================="
echo ""

# Run import script
go run scripts/import_csv_to_db.go "$CSV_FILE"

IMPORT_STATUS=$?

echo ""
echo "=========================================="

if [ $IMPORT_STATUS -eq 0 ]; then
    echo -e "${GREEN}✅ Import completed successfully!${NC}"
    echo ""
    echo "📊 Summary:"
    echo "  - CSV file imported"
    echo "  - Settings created in database"
    echo "  - Topics ready to use"
    echo ""
    echo "=========================================="
    echo "Next Steps:"
    echo "=========================================="
    echo ""
    echo "1. Review imported settings:"
    echo "   psql -U postgres -d gofiber_template -c \"SELECT id, tone, jsonb_array_length(topics) as topics_count FROM auto_post_settings;\""
    echo ""
    echo "2. Test manual trigger (optional):"
    echo "   curl -X POST http://localhost:3000/api/v1/auto-post/settings/{SETTING_ID}/trigger \\"
    echo "     -H \"Authorization: Bearer YOUR_JWT_TOKEN\""
    echo ""
    echo "3. Enable settings:"
    echo "   # Via API:"
    echo "   curl -X POST http://localhost:3000/api/v1/auto-post/settings/{SETTING_ID}/enable \\"
    echo "     -H \"Authorization: Bearer YOUR_JWT_TOKEN\""
    echo ""
    echo "   # Via SQL (enable all):"
    echo "   psql -U postgres -d gofiber_template -c \"UPDATE auto_post_settings SET is_enabled = true WHERE bot_user_id = '$BOT_USER_ID';\""
    echo ""
    echo "4. Restart server to activate scheduler:"
    echo "   # Docker:"
    echo "   docker-compose restart app"
    echo ""
    echo "   # Direct:"
    echo "   ./bin/api"
    echo ""
    echo "=========================================="
    echo -e "${GREEN}🎉 Setup complete!${NC}"
    echo "=========================================="
else
    echo -e "${RED}❌ Import failed!${NC}"
    echo "Please check the error messages above"
    exit 1
fi
