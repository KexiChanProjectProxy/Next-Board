# Manual Deployment with Wrangler Pages

If Cloudflare Pages automatic deployment isn't working, you can deploy manually.

## Prerequisites

```bash
npm install -g wrangler
wrangler login
```

## Build and Deploy

```bash
cd /home/kexi/Next-Board/web-ui

# Build the project
npm run build

# Deploy to Cloudflare Pages (NOT Workers)
wrangler pages deploy dist --project-name=next-board-ui
```

## First Time Deployment

The first time you run this, Wrangler will:
1. Ask you to confirm the project name
2. Create the project automatically
3. Deploy the files

## Subsequent Deployments

After the first deployment, just run:

```bash
npm run build
wrangler pages deploy dist --project-name=next-board-ui
```

## Set Environment Variables

After first deployment:
1. Go to Cloudflare Dashboard
2. Find your project under Workers & Pages  
3. Settings → Environment variables
4. Add: VITE_API_BASE_URL = http://YOUR_BACKEND:8080
5. Redeploy:
   ```bash
   npm run build
   wrangler pages deploy dist --project-name=next-board-ui
   ```

## Notes

- Use `wrangler pages deploy` (NOT `wrangler deploy`)
- `wrangler deploy` is for Workers, not Pages
- Pages requires Node.js 20+ for Wrangler

## Success Message

You should see:
```
✨ Success! Uploaded 15 files (2.53 sec)
✨ Deployment complete! Your site is available at:
   https://next-board-ui.pages.dev
```
