# Deployment Guide

## Quick Start

### Local Development

1. **Install dependencies**:
   ```bash
   npm install
   ```

2. **Configure environment**:
   ```bash
   cp .env.example .env.development
   # Edit .env.development to set VITE_API_BASE_URL
   ```

3. **Start development server**:
   ```bash
   npm run dev
   ```

4. **Build for production**:
   ```bash
   npm run build
   ```

## Cloudflare Pages Deployment

### Option 1: GitHub Integration (Recommended)

1. **Push to GitHub**:
   ```bash
   git add .
   git commit -m "Add Next-Board web UI"
   git push origin main
   ```

2. **Configure Cloudflare Pages**:
   - Go to: https://dash.cloudflare.com/
   - Select "Workers & Pages" → "Create application" → "Pages"
   - Connect to your GitHub repository
   - **Build settings**:
     - Framework preset: Vite
     - Build command: `npm run build`
     - Build output directory: `dist`
     - Root directory: `web-ui` (if not in root)
     - Node version: `18` or higher

3. **Environment Variables**:
   Add in Cloudflare Pages settings:
   ```
   VITE_API_BASE_URL=https://api.yourdomain.com
   ```

4. **Deploy**:
   - Click "Save and Deploy"
   - Cloudflare will build and deploy automatically
   - Future commits will auto-deploy

### Option 2: Direct Upload (Wrangler CLI)

1. **Install Wrangler**:
   ```bash
   npm install -g wrangler
   ```

2. **Login to Cloudflare**:
   ```bash
   wrangler login
   ```

3. **Build the project**:
   ```bash
   npm run build
   ```

4. **Deploy**:
   ```bash
   wrangler pages deploy dist --project-name=next-board
   ```

## Backend CORS Configuration

**CRITICAL**: Update your Next-Board backend `config.json`:

```json
{
  "server": {
    "listen_ip": "0.0.0.0",
    "listen_port": 8080,
    "cors_origins": [
      "http://localhost:5173",
      "https://your-app.pages.dev",
      "https://your-custom-domain.com"
    ]
  }
}
```

Restart your backend after updating CORS configuration.

## Custom Domain Setup

1. In Cloudflare Pages:
   - Go to your project settings
   - Click "Custom domains"
   - Add your domain (e.g., `dashboard.yourdomain.com`)

2. Update DNS:
   - Add CNAME record pointing to `your-app.pages.dev`
   - Or use Cloudflare nameservers

3. Update CORS:
   - Add your custom domain to backend `cors_origins`

## Build Output

Current build statistics:
- **Total size (gzipped)**: ~253 KB
- **Main bundle**: 111.63 KB
- **Charts chunk**: 104.99 KB
- **Vendor chunk**: 16.72 KB
- **API chunk**: 14.65 KB
- **CSS**: 5.33 KB

## Environment Variables

### Development (.env.development)
```env
VITE_API_BASE_URL=http://localhost:8080
```

### Production (Cloudflare Pages)
```env
VITE_API_BASE_URL=https://api.yourdomain.com
```

## Verification Checklist

- [ ] Dependencies installed successfully
- [ ] Build completes without errors
- [ ] Environment variables set correctly
- [ ] Backend CORS includes web UI domain
- [ ] Backend API is accessible from web UI
- [ ] Login works correctly
- [ ] Token refresh works
- [ ] All routes are accessible
- [ ] Mobile responsive design works
- [ ] Admin features work (for admin users)

## Troubleshooting

### Build Errors

**Module not found errors**:
```bash
rm -rf node_modules package-lock.json
npm install
```

**TypeScript errors**:
```bash
npm run build
# Check the error output and fix type issues
```

### Runtime Errors

**CORS errors**:
- Verify backend `cors_origins` includes your domain
- Check browser console for actual error
- Ensure no trailing slashes in URLs

**401 Unauthorized**:
- Check `VITE_API_BASE_URL` is correct
- Verify backend is running
- Check localStorage for tokens
- Try logging in again

**404 on refresh**:
- Ensure `public/_redirects` file exists
- Content should be: `/* /index.html 200`
- Cloudflare Pages handles this automatically

### Performance

**Slow load times**:
- Check bundle sizes: `npm run build`
- Verify Cloudflare CDN is serving assets
- Check for console errors
- Use browser dev tools Performance tab

## Post-Deployment

1. **Test login**: Verify authentication works
2. **Test user features**: Dashboard, nodes, usage, settings
3. **Test admin features**: User management (if admin)
4. **Test mobile**: Check responsive design on mobile devices
5. **Monitor errors**: Check browser console for errors
6. **Check performance**: Use Lighthouse for performance audit

## Support

For issues:
- Check browser console for errors
- Review backend logs
- Verify CORS configuration
- Check environment variables
- Test API endpoints directly

---

**Deployment completed successfully!** ✓
