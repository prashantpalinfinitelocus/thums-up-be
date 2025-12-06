#!/bin/bash
set -e

echo "🛑 Stopping Strapi container..."
docker compose stop strapi || true

echo "🗑️  Removing Strapi container..."
docker compose rm -f strapi || true

echo "🗑️  Removing Strapi images..."
docker rmi thums-up-be-strapi 2>/dev/null || true
docker rmi thums-up-be_strapi 2>/dev/null || true

echo "🧹 Cleaning build cache..."
docker builder prune -f

echo "🧹 Cleaning local dist folders..."
rm -rf strapi/dist/
rm -rf strapi/.cache/
rm -rf strapi/.tmp/
rm -rf strapi/build/

echo "✅ Verifying cleanup..."
if [ -d "strapi/dist" ]; then
    echo "❌ ERROR: strapi/dist still exists!"
    exit 1
else
    echo "✅ strapi/dist removed successfully"
fi

echo "🔨 Rebuilding Strapi (this will take 2-3 minutes)..."
docker compose build --no-cache --pull strapi

echo "🚀 Starting Strapi..."
docker compose up -d strapi

echo "⏳ Waiting for Strapi to start..."
sleep 10

echo "🔍 Checking dist/config in container..."
docker exec thums_up_strapi sh -c "ls -la /srv/app/dist/config/ && file /srv/app/dist/config/plugins.js"

echo ""
echo "📋 Showing Strapi logs:"
docker logs thums_up_strapi --tail 30

echo ""
echo "✅ Done! Check if Strapi is running:"
echo "   docker logs -f thums_up_strapi"
echo "   http://localhost:1338/admin"

