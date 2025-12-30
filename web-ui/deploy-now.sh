#!/bin/bash

# Quick deploy script for Next-Board Web UI
# This uses Wrangler Pages to deploy manually

set -e

echo "🏗️  Building Next-Board Web UI..."
npm run build

echo ""
echo "✅ Build complete!"
echo ""
echo "📤 To deploy to Cloudflare Pages, run these commands:"
echo ""
echo "# First time only:"
echo "npm install -g wrangler"
echo "wrangler login"
echo ""
echo "# Deploy:"
echo "wrangler pages deploy dist --project-name=next-board-ui"
echo ""
echo "Or if you want to deploy RIGHT NOW, I can do it for you."
echo "Press ENTER to deploy, or Ctrl+C to cancel..."
read

echo ""
echo "🚀 Deploying to Cloudflare Pages..."
wrangler pages deploy dist --project-name=next-board-ui

echo ""
echo "✨ Deployment complete!"
echo ""
echo "⚙️  Don't forget to set environment variables in Cloudflare dashboard:"
echo "   VITE_API_BASE_URL = http://YOUR_BACKEND_IP:8080"
