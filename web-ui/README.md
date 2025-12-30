# Next-Board Web UI

Modern, responsive web interface for Next-Board (Xboard Go), a high-performance proxy management system.

## Features

### User Features
- **Authentication**: Secure JWT-based login with automatic token refresh
- **Dashboard**: Overview of plan, usage, and available nodes
- **Nodes Management**: Browse, search, and filter proxy nodes with QR code generation
- **Usage Tracking**: Real-time data usage monitoring with interactive charts
- **Settings**: Profile management and Telegram bot integration
- **Responsive Design**: Fully responsive for mobile, tablet, and desktop

### Admin Features
- **User Management**: Create, update, and delete user accounts
- **Node Management**: CRUD operations for proxy nodes (coming soon)
- **Plan Management**: Manage subscription plans and quotas (coming soon)
- **Label Management**: Organize nodes with custom labels (coming soon)

## Tech Stack

- **Framework**: React 18 + TypeScript
- **Build Tool**: Vite
- **Styling**: Tailwind CSS
- **UI Components**: shadcn/ui (Radix UI primitives)
- **State Management**: Zustand
- **HTTP Client**: Axios with interceptors
- **Routing**: React Router v7
- **Charts**: Recharts
- **QR Codes**: qrcode library
- **Date Handling**: date-fns

## Prerequisites

- Node.js 18+ or higher
- npm or yarn
- Next-Board backend API running

## Installation

1. **Navigate to the web-ui directory**:
   ```bash
   cd /path/to/Next-Board/web-ui
   ```

2. **Install dependencies**:
   ```bash
   npm install
   ```

3. **Configure environment variables**:
   ```bash
   cp .env.example .env.development
   ```

   Edit `.env.development` and set your API base URL:
   ```env
   VITE_API_BASE_URL=http://localhost:8080
   ```

## Development

Start the development server:

```bash
npm run dev
```

The application will be available at `http://localhost:5173`

## Building for Production

Build the application:

```bash
npm run build
```

Preview the production build:

```bash
npm run preview
```

## Deployment to Cloudflare Pages

### Method 1: GitHub Integration (Recommended)

1. **Push your code to GitHub**
2. **Connect to Cloudflare Pages**:
   - Go to Cloudflare Pages Dashboard
   - Click "Create a project"
   - Connect your GitHub repository
3. **Configure build settings**:
   - Build command: `npm run build`
   - Build output directory: `dist`
   - Root directory: `web-ui` (if in subdirectory)
4. **Set environment variables**:
   - `VITE_API_BASE_URL`: Your production API URL
5. **Deploy**: Automatic on push

### CORS Configuration

Ensure your backend has the web UI domain in CORS configuration:

```json
{
  "server": {
    "cors_origins": [
      "http://localhost:5173",
      "https://your-domain.pages.dev"
    ]
  }
}
```

## Project Structure

```
web-ui/
├── public/
│   └── _redirects          # SPA routing for Cloudflare Pages
├── src/
│   ├── api/                # API client and services
│   ├── components/         # React components
│   ├── pages/              # Page components
│   ├── stores/             # Zustand stores
│   ├── types/              # TypeScript types
│   ├── utils/              # Utility functions
│   └── App.tsx             # Root component
├── package.json
├── vite.config.ts
└── tailwind.config.js
```

## License

Same as Next-Board project.
