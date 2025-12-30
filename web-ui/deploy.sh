#!/bin/bash

# Next-Board Web UI - Manual Deploy Script
# Usage: ./deploy.sh

set -e

echo "🏗️  Building Next-Board Web UI..."
npm run build

echo ""
echo "📦 Build complete! Files in dist/"
echo ""
echo "To deploy to Cloudflare Pages using Wrangler:"
echo ""
echo "1. Install Wrangler:"
echo "   npm install -g wrangler"
echo ""
echo "2. Login:"
echo "   wrangler login"
echo ""
echo "3. Deploy:"
echo "   wrangler pages deploy dist --project-name=next-board-ui"
echo ""
echo "Or upload the dist/ folder directly in Cloudflare Pages dashboard"
