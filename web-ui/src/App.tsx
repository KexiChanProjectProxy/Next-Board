import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { useEffect } from 'react';
import { useAuthStore } from '@/stores/authStore';
import { Toaster } from '@/components/ui/toaster';
import { ProtectedRoute } from '@/components/ProtectedRoute';
import { DashboardLayout } from '@/components/layout/DashboardLayout';
import { LoginPage } from '@/pages/auth/LoginPage';
import { DashboardPage } from '@/pages/user/DashboardPage';
import { NodesPage } from '@/pages/user/NodesPage';
import { UsagePage } from '@/pages/user/UsagePage';
import { SettingsPage } from '@/pages/user/SettingsPage';
import { AdminUsersPage } from '@/pages/admin/AdminUsersPage';
import { AdminNodesPage } from '@/pages/admin/AdminNodesPage';
import { AdminPlansPage } from '@/pages/admin/AdminPlansPage';
import { AdminLabelsPage } from '@/pages/admin/AdminLabelsPage';

function App() {
  const { isAuthenticated, fetchUser } = useAuthStore();

  useEffect(() => {
    if (isAuthenticated) {
      fetchUser();
    }
  }, [isAuthenticated]);

  return (
    <BrowserRouter>
      <Routes>
        {/* Public Routes */}
        <Route path="/login" element={<LoginPage />} />

        {/* Protected User Routes */}
        <Route
          path="/dashboard"
          element={
            <ProtectedRoute>
              <DashboardLayout>
                <DashboardPage />
              </DashboardLayout>
            </ProtectedRoute>
          }
        />
        <Route
          path="/nodes"
          element={
            <ProtectedRoute>
              <DashboardLayout>
                <NodesPage />
              </DashboardLayout>
            </ProtectedRoute>
          }
        />
        <Route
          path="/usage"
          element={
            <ProtectedRoute>
              <DashboardLayout>
                <UsagePage />
              </DashboardLayout>
            </ProtectedRoute>
          }
        />
        <Route
          path="/settings"
          element={
            <ProtectedRoute>
              <DashboardLayout>
                <SettingsPage />
              </DashboardLayout>
            </ProtectedRoute>
          }
        />

        {/* Protected Admin Routes */}
        <Route
          path="/admin/users"
          element={
            <ProtectedRoute requireAdmin>
              <DashboardLayout>
                <AdminUsersPage />
              </DashboardLayout>
            </ProtectedRoute>
          }
        />
        <Route
          path="/admin/nodes"
          element={
            <ProtectedRoute requireAdmin>
              <DashboardLayout>
                <AdminNodesPage />
              </DashboardLayout>
            </ProtectedRoute>
          }
        />
        <Route
          path="/admin/plans"
          element={
            <ProtectedRoute requireAdmin>
              <DashboardLayout>
                <AdminPlansPage />
              </DashboardLayout>
            </ProtectedRoute>
          }
        />
        <Route
          path="/admin/labels"
          element={
            <ProtectedRoute requireAdmin>
              <DashboardLayout>
                <AdminLabelsPage />
              </DashboardLayout>
            </ProtectedRoute>
          }
        />

        {/* Default Redirect */}
        <Route path="/" element={<Navigate to="/dashboard" replace />} />
        <Route path="*" element={<Navigate to="/dashboard" replace />} />
      </Routes>
      <Toaster />
    </BrowserRouter>
  );
}

export default App;
