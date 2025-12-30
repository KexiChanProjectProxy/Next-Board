# Web UI Development Prompt

## Project Overview

Build a standalone web UI for Xboard Go (Next-Board), a high-performance proxy management system. The web UI should be a modern, responsive single-page application (SPA) that deploys to Cloudflare Pages and connects to the Go backend API.

## Backend API Context

The backend is a pure REST API server built with Go/Gin that provides:

- **Authentication**: JWT-based auth with access and refresh tokens
- **User Features**: Profile management, usage tracking, node access, plan details
- **Admin Features**: User/node/plan/label CRUD operations
- **Node Protocol**: Xboard-compatible protocol for proxy nodes

**API Documentation**: See `xboard-go/API.md` for complete endpoint reference

**Base API URL**: User-configurable (e.g., `https://api.example.com`)

## Technical Requirements

### Framework & Technology

**Recommended Stack:**
- **Framework**: React 18+ with TypeScript OR Vue 3 with TypeScript
- **Build Tool**: Vite (fast builds, optimized for Cloudflare Pages)
- **Styling**: Tailwind CSS for utility-first styling
- **State Management**:
  - React: Zustand or Redux Toolkit
  - Vue: Pinia
- **HTTP Client**: Axios with interceptors for auth
- **Routing**: React Router v6 OR Vue Router v4
- **UI Components**: shadcn/ui (React) OR Headless UI (Vue)
- **Charts**: Recharts or Chart.js for usage visualization
- **Date Handling**: date-fns or Day.js

**Alternative Lightweight Stack:**
- **Framework**: Solid.js OR Svelte for smaller bundle size
- All other requirements same as above

### Deployment Target

- **Platform**: Cloudflare Pages
- **Build Output**: Static files (HTML, JS, CSS, assets)
- **Environment Variables**: API base URL configurable via build-time env vars
- **SPA Routing**: Use `_redirects` or `_headers` file for client-side routing
- **Edge Optimization**: Leverage Cloudflare's global CDN

## Features to Implement

### 1. Authentication Pages

#### Login Page (`/login`)
- Email and password input fields
- "Remember me" checkbox (optional)
- Form validation with error messages
- Call `POST /api/v1/auth/login`
- Store access_token and refresh_token securely
- Redirect to dashboard on success
- Display API error messages

#### Auto Token Refresh
- Implement axios interceptor to catch 401 errors
- Automatically call `POST /api/v1/auth/refresh` with refresh_token
- Retry failed request with new access_token
- Logout user if refresh fails

### 2. User Dashboard (`/dashboard`)

**Layout:**
- Top navigation bar with logo, user email, logout button
- Sidebar navigation (Dashboard, Nodes, Usage, Settings)
- Main content area

**Dashboard Overview:**
- Display user profile info from `GET /api/v1/me`
- Show current plan details from `GET /api/v1/me/plan`
- Usage summary card:
  - Real vs billable traffic (upload/download)
  - Quota progress bar
  - Percentage used
  - Data from `GET /api/v1/me/usage`
- Quick stats cards (total nodes, plan name, etc.)

### 3. Nodes Page (`/nodes`)

**Features:**
- List all accessible nodes from `GET /api/v1/me/nodes`
- Display for each node:
  - Node name and type (vmess, vless, trojan, etc.)
  - Host and port
  - Status badge (active/inactive)
  - Traffic multiplier
  - Assigned labels with colored badges
- Filter by node type
- Search by node name
- Copy node info button
- Responsive grid/list layout

### 4. Usage Page (`/usage`)

**Current Usage Section:**
- Display from `GET /api/v1/me/usage`
- Show current billing period dates
- Real traffic: Upload, Download, Total (in GB/TB)
- Billable traffic: Upload, Download, Total (in GB/TB)
- Multiplier effect visualization
- Quota progress bar with color coding:
  - Green: < 50%
  - Yellow: 50-80%
  - Orange: 80-95%
  - Red: > 95%

**Usage History Section:**
- Chart showing traffic over time
- Note: Backend endpoint `GET /api/v1/me/usage/history` is not yet implemented
- Display placeholder message: "Historical data will be available soon"
- Prepare chart component for future integration

### 5. Settings Page (`/settings`)

**Profile Section:**
- Display user email
- Display user role
- Display created date
- Show plan assignment

**Telegram Integration:**
- Button to generate link token: `POST /api/v1/me/telegram/link`
- Display generated token in modal
- Show instructions to send `/link <token>` to bot
- Show current link status (linked/not linked)
- Display telegram_chat_id if linked

**Security Section:**
- Change password form (future)
- Active sessions list (future)

### 6. Admin Panel (`/admin/*`)

**Route Protection:**
- Check user role from token or API
- Redirect non-admin users to dashboard
- Show admin menu only for admin users

#### Admin - Users (`/admin/users`)
- List users with pagination: `GET /api/v1/admin/users?page=1&limit=20`
- Display table with columns:
  - ID, Email, Role, Plan, Banned status, Created date
- Search and filter functionality
- Actions per row:
  - Edit button (opens modal)
  - Delete button (with confirmation)
- Create user button: `POST /api/v1/admin/users`
- Edit user modal: `PUT /api/v1/admin/users/:id`
  - Update email, plan_id, banned status
- Delete confirmation: `DELETE /api/v1/admin/users/:id`
- Pagination controls

#### Admin - Nodes (`/admin/nodes`)
- List nodes: `GET /api/v1/admin/nodes`
- Display table with columns:
  - ID, Name, Type, Host, Port, Multiplier, Status
- Create node form: `POST /api/v1/admin/nodes`
  - Fields: name, node_type (dropdown), host, port, protocol_config (JSON), node_multiplier, label_ids (multi-select)
- Label assignment interface
- Pagination

#### Admin - Plans (`/admin/plans`)
- List plans: `GET /api/v1/admin/plans`
- Display cards or table:
  - Plan name, quota (in GB/TB), reset period, base multiplier
- Create plan form: `POST /api/v1/admin/plans`
  - Fields: name, quota_bytes (with GB/TB converter), reset_period (dropdown: none/daily/weekly/monthly/yearly), base_multiplier, label_ids
- Assign labels to plan
- Show which labels are included

#### Admin - Labels (`/admin/labels`)
- List labels: `GET /api/v1/admin/labels`
- Display cards with:
  - Label name, description, multiplier
- Create label form: `POST /api/v1/admin/labels`
  - Fields: name, description
- Color-coded label badges

### 7. Error Handling

**Global Error Handling:**
- Network errors: Display toast notification
- 401 Unauthorized: Trigger token refresh or logout
- 403 Forbidden: Redirect to access denied page
- 404 Not Found: Display not found page
- 500 Server Error: Display error message with retry button

**Form Validation:**
- Client-side validation before API calls
- Display backend error messages from API responses
- Format: `error.code` and `error.message`

## UI/UX Requirements

### Design System

**Color Palette:**
- Primary: Blue (#3B82F6) for main actions
- Success: Green (#10B981) for positive states
- Warning: Yellow (#F59E0B) for caution
- Danger: Red (#EF4444) for destructive actions
- Neutral: Gray shades for text and backgrounds

**Typography:**
- Font: Inter or System UI fonts
- Headings: Bold, larger sizes
- Body: Regular weight, readable line height

**Components:**
- Consistent button styles (primary, secondary, ghost, danger)
- Input fields with focus states
- Loading spinners for async operations
- Toast notifications for success/error messages
- Modal dialogs for confirmations and forms
- Dropdown menus for navigation

### Responsive Design

- **Mobile First**: Design for mobile, enhance for desktop
- **Breakpoints**:
  - Mobile: < 640px
  - Tablet: 640px - 1024px
  - Desktop: > 1024px
- **Navigation**: Hamburger menu on mobile, sidebar on desktop
- **Tables**: Horizontal scroll or card layout on mobile

### Accessibility

- Semantic HTML elements
- ARIA labels where needed
- Keyboard navigation support
- Focus indicators
- Color contrast compliance (WCAG AA)

## State Management

### Auth State

```typescript
interface AuthState {
  accessToken: string | null;
  refreshToken: string | null;
  user: User | null;
  isAuthenticated: boolean;
  login: (email: string, password: string) => Promise<void>;
  logout: () => void;
  refreshAccessToken: () => Promise<void>;
}
```

### User State

```typescript
interface UserState {
  profile: User | null;
  plan: Plan | null;
  nodes: Node[];
  usage: Usage | null;
  fetchProfile: () => Promise<void>;
  fetchPlan: () => Promise<void>;
  fetchNodes: () => Promise<void>;
  fetchUsage: () => Promise<void>;
}
```

### Admin State (if role === 'admin')

```typescript
interface AdminState {
  users: PaginatedResponse<User>;
  nodes: PaginatedResponse<Node>;
  plans: PaginatedResponse<Plan>;
  labels: PaginatedResponse<Label>;
  fetchUsers: (page: number, limit: number) => Promise<void>;
  createUser: (data: CreateUserRequest) => Promise<void>;
  // ... other admin actions
}
```

## API Integration

### Axios Setup

```typescript
// api/client.ts
import axios from 'axios';

const apiClient = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Request interceptor: Add auth token
apiClient.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('access_token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => Promise.reject(error)
);

// Response interceptor: Handle token refresh
apiClient.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config;

    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true;

      try {
        const refreshToken = localStorage.getItem('refresh_token');
        const response = await axios.post('/api/v1/auth/refresh', {
          refresh_token: refreshToken,
        });

        const { access_token } = response.data;
        localStorage.setItem('access_token', access_token);

        originalRequest.headers.Authorization = `Bearer ${access_token}`;
        return apiClient(originalRequest);
      } catch (refreshError) {
        // Refresh failed, logout user
        localStorage.clear();
        window.location.href = '/login';
        return Promise.reject(refreshError);
      }
    }

    return Promise.reject(error);
  }
);

export default apiClient;
```

### API Service Example

```typescript
// api/auth.ts
import apiClient from './client';

export const authApi = {
  login: async (email: string, password: string) => {
    const response = await apiClient.post('/api/v1/auth/login', {
      email,
      password,
    });
    return response.data;
  },

  refresh: async (refreshToken: string) => {
    const response = await apiClient.post('/api/v1/auth/refresh', {
      refresh_token: refreshToken,
    });
    return response.data;
  },
};

// api/user.ts
export const userApi = {
  getProfile: async () => {
    const response = await apiClient.get('/api/v1/me');
    return response.data.user;
  },

  getPlan: async () => {
    const response = await apiClient.get('/api/v1/me/plan');
    return response.data.plan;
  },

  getNodes: async () => {
    const response = await apiClient.get('/api/v1/me/nodes');
    return response.data.nodes;
  },

  getUsage: async () => {
    const response = await apiClient.get('/api/v1/me/usage');
    return response.data.usage;
  },
};
```

## Data Types

### Core Models

```typescript
interface User {
  id: number;
  email: string;
  role: 'admin' | 'user';
  plan_id: number | null;
  telegram_chat_id: number | null;
  telegram_linked_at: string | null;
  created_at: string;
}

interface Plan {
  id: number;
  name: string;
  quota_bytes: number;
  reset_period: 'none' | 'daily' | 'weekly' | 'monthly' | 'yearly';
  base_multiplier: number;
  labels: Label[];
  created_at: string;
}

interface Node {
  id: number;
  name: string;
  node_type: string;
  host: string;
  port: number;
  node_multiplier: number;
  status: 'active' | 'inactive';
  labels: Label[];
}

interface Label {
  id: number;
  name: string;
  description: string;
  multiplier: number;
}

interface Usage {
  real_bytes_up: number;
  real_bytes_down: number;
  billable_bytes_up: number;
  billable_bytes_down: number;
  period_start: string;
  period_end: string;
}

interface PaginatedResponse<T> {
  data: T[];
  pagination: {
    total: number;
    page: number;
    limit: number;
    pages: number;
  };
}
```

## Utility Functions

### Byte Formatting

```typescript
export function formatBytes(bytes: number, decimals = 2): string {
  if (bytes === 0) return '0 Bytes';

  const k = 1024;
  const dm = decimals < 0 ? 0 : decimals;
  const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB', 'PB'];

  const i = Math.floor(Math.log(bytes) / Math.log(k));

  return parseFloat((bytes / Math.pow(k, i)).toFixed(dm)) + ' ' + sizes[i];
}

// Example: formatBytes(1073741824) => "1.00 GB"
```

### Date Formatting

```typescript
import { format, formatDistanceToNow } from 'date-fns';

export function formatDate(dateString: string): string {
  return format(new Date(dateString), 'yyyy-MM-dd HH:mm:ss');
}

export function formatRelativeTime(dateString: string): string {
  return formatDistanceToNow(new Date(dateString), { addSuffix: true });
}
```

### Usage Percentage

```typescript
export function calculateUsagePercentage(
  billableUp: number,
  billableDown: number,
  quota: number
): number {
  const totalUsed = billableUp + billableDown;
  return (totalUsed / quota) * 100;
}

export function getUsageColor(percentage: number): string {
  if (percentage < 50) return 'green';
  if (percentage < 80) return 'yellow';
  if (percentage < 95) return 'orange';
  return 'red';
}
```

## Cloudflare Pages Deployment

### Project Structure

```
web-ui/
├── public/
│   ├── _redirects          # SPA routing redirect rules
│   └── favicon.ico
├── src/
│   ├── api/                # API client and services
│   ├── components/         # Reusable components
│   ├── pages/              # Page components
│   ├── stores/             # State management
│   ├── types/              # TypeScript types
│   ├── utils/              # Utility functions
│   ├── App.tsx             # Root component
│   └── main.tsx            # Entry point
├── .env.example
├── .gitignore
├── index.html
├── package.json
├── tsconfig.json
├── vite.config.ts
└── tailwind.config.js
```

### _redirects File

Create `public/_redirects` for SPA routing:

```
/* /index.html 200
```

### Environment Variables

**Development (.env.development):**
```
VITE_API_BASE_URL=http://localhost:8080
```

**Production (.env.production):**
```
VITE_API_BASE_URL=https://api.example.com
```

**Cloudflare Pages Configuration:**
- Set environment variable `VITE_API_BASE_URL` in Pages dashboard
- Build command: `npm run build`
- Build output directory: `dist`
- Node version: 18 or higher

### Build Configuration

**vite.config.ts:**
```typescript
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  build: {
    outDir: 'dist',
    sourcemap: false,
    minify: 'esbuild',
    rollupOptions: {
      output: {
        manualChunks: {
          vendor: ['react', 'react-dom', 'react-router-dom'],
          api: ['axios'],
        },
      },
    },
  },
});
```

### package.json Scripts

```json
{
  "scripts": {
    "dev": "vite",
    "build": "tsc && vite build",
    "preview": "vite preview",
    "lint": "eslint src --ext ts,tsx",
    "format": "prettier --write src"
  }
}
```

## Security Considerations

### Token Storage

- Store tokens in `localStorage` or `sessionStorage`
- Consider using `httpOnly` cookies if backend supports it
- Never log tokens to console in production

### XSS Protection

- Sanitize user input before rendering
- Use framework's built-in escaping (React/Vue auto-escapes)
- Validate data from API responses

### CORS Configuration

- Backend must include frontend domain in `cors_origins`
- Example: `["https://dashboard.example.com"]`
- For development: `["http://localhost:5173"]`

### API Key Protection

- Never commit API base URLs with credentials
- Use environment variables
- Cloudflare Pages automatically encrypts env vars

## Performance Optimization

### Code Splitting

- Lazy load admin pages (only load when needed)
- Use React.lazy() or Vue's defineAsyncComponent()
- Split vendor chunks from application code

### Caching

- Cache API responses where appropriate
- Use SWR or React Query for data fetching
- Implement stale-while-revalidate pattern

### Bundle Size

- Tree-shake unused code
- Use production builds
- Analyze bundle with `vite-bundle-visualizer`
- Target < 200KB initial bundle size

### Images & Assets

- Optimize images (WebP format)
- Use Cloudflare's image optimization
- Lazy load images below fold

## Testing Requirements

### Unit Tests

- Test utility functions (formatBytes, calculateUsagePercentage)
- Test state management logic
- Use Vitest for testing

### Integration Tests

- Test API service calls with mocked responses
- Test authentication flow
- Test form submissions

### E2E Tests (Optional)

- Use Playwright or Cypress
- Test critical user flows:
  - Login → Dashboard
  - View nodes
  - Admin create user

## Progressive Enhancement

### Offline Support (Optional)

- Service Worker for offline page
- Cache static assets
- Show "offline" indicator

### PWA Features (Optional)

- Add manifest.json
- Enable "Add to Home Screen"
- Push notifications for usage alerts (future)

## Deliverables

1. **Source Code**: Complete TypeScript project with all features
2. **README.md**: Setup instructions, development guide, deployment steps
3. **Environment Variables**: Document all required env vars
4. **Deployment Guide**: Step-by-step Cloudflare Pages deployment
5. **Screenshots**: Show key pages (login, dashboard, admin panel)
6. **Demo Credentials**: Provide test user/admin credentials for demo

## Success Criteria

- ✅ Successfully authenticates with backend API
- ✅ Displays user dashboard with real-time data
- ✅ All CRUD operations work for admin
- ✅ Responsive design works on mobile/tablet/desktop
- ✅ No console errors in production build
- ✅ Lighthouse score > 90 for Performance
- ✅ Deploys successfully to Cloudflare Pages
- ✅ CORS properly configured with backend

## Optional Enhancements

### Nice-to-Have Features

1. **Dark Mode**: Theme toggle with system preference detection
2. **Multi-language**: i18n support (English, Chinese)
3. **Export Data**: Export usage history as CSV
4. **QR Code**: Generate QR code for node configurations
5. **Notifications**: Toast notifications for quota warnings
6. **Search**: Global search across nodes and users (admin)
7. **Keyboard Shortcuts**: Quick navigation with hotkeys
8. **Activity Log**: Show recent actions (admin)

### Advanced Features

1. **Real-time Updates**: WebSocket connection for live stats
2. **Advanced Charts**: Interactive traffic visualization
3. **Bulk Operations**: Bulk user import/export (admin)
4. **API Playground**: Test API endpoints directly in UI (admin)
5. **Audit Trail**: Track all admin actions with timestamps

## Reference Materials

- **API Docs**: See `xboard-go/API.md` for all endpoints
- **Backend Config**: See `xboard-go/config.json` for CORS setup
- **Xboard Compatibility**: UniProxy V1/V2 protocol support

## Questions to Clarify with User

Before starting, confirm:

1. Framework preference (React vs Vue vs Svelte)?
2. UI component library preference (shadcn/ui vs Ant Design vs Material UI)?
3. State management preference (Zustand vs Redux vs Pinia)?
4. Dark mode required or optional?
5. Multi-language support needed?
6. Any branding guidelines (colors, logo, fonts)?
7. Target users: technical or non-technical?

## Next Steps

1. Set up Vite + React/Vue + TypeScript project
2. Configure Tailwind CSS and component library
3. Implement authentication flow with JWT
4. Build user dashboard and pages
5. Implement admin panel
6. Add responsive design and polish UI
7. Write tests
8. Deploy to Cloudflare Pages
9. Configure CORS on backend
10. Test end-to-end functionality

---

**Good luck building an amazing web UI! 🚀**

If you encounter any issues with the API, refer to `xboard-go/API.md` or check the backend logs for debugging.
