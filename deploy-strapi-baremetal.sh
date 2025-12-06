#!/bin/bash
set -e

echo "======================================"
echo "🔧 Strapi Baremetal Fix Script"
echo "======================================"

# Step 1: Stop everything
echo ""
echo "Step 1/7: Stopping Strapi..."
docker compose stop strapi 2>/dev/null || true
docker compose rm -f strapi 2>/dev/null || true

# Step 2: Remove old images
echo ""
echo "Step 2/7: Removing old Docker images..."
docker rmi -f thums-up-be-strapi 2>/dev/null || true
docker rmi -f thums-up-be_strapi 2>/dev/null || true
docker images | grep strapi | awk '{print $3}' | xargs docker rmi -f 2>/dev/null || true

# Step 3: Clean build cache
echo ""
echo "Step 3/7: Cleaning Docker build cache..."
docker builder prune -af

# Step 4: Clean host directories
echo ""
echo "Step 4/7: Cleaning host directories..."
rm -rf strapi/dist/
rm -rf strapi/.cache/
rm -rf strapi/.tmp/
rm -rf strapi/build/
rm -rf strapi/node_modules/.cache/

# Step 5: Verify cleanup
echo ""
echo "Step 5/7: Verifying cleanup..."
if [ -d "strapi/dist" ]; then
    echo "❌ ERROR: strapi/dist still exists after cleanup!"
    ls -la strapi/dist/
    exit 1
else
    echo "✅ Host cleanup successful - no dist folder"
fi

# Step 6: Rebuild with complete no-cache
echo ""
echo "Step 6/7: Rebuilding Strapi (this takes 2-3 minutes)..."
docker compose build --no-cache --pull --progress=plain strapi 2>&1 | tail -50

# Step 7: Start and verify
echo ""
echo "Step 7/7: Starting Strapi and verifying..."
docker compose up -d strapi

sleep 5

# Check if container is running
if docker ps | grep -q thums_up_strapi; then
    echo "✅ Container is running"
    
    # Verify config structure inside container
    echo ""
    echo "Checking config files in container:"
    docker exec thums_up_strapi sh -c "ls -la /srv/app/dist/config/ && file /srv/app/dist/config/plugins.js"
    
    echo ""
    echo "Waiting for Strapi to initialize..."
    sleep 10
    
    echo ""
    echo "📋 Strapi logs:"
    docker logs thums_up_strapi --tail 50
else
    echo "❌ Container failed to start"
    docker logs thums_up_strapi
    exit 1
fi

echo ""
echo "======================================"
echo "✅ Deployment Complete!"
echo "======================================"
echo "Access Strapi at: http://YOUR_SERVER_IP:1338/admin"
echo "Monitor logs: docker logs -f thums_up_strapi"

