import React from 'react';
import { Routes, Route, Navigate, useLocation } from 'react-router-dom';

import { MainLayout } from '@/layouts/MainLayout';
import { Dashboard } from '@/pages/Dashboard';
import { EnvironmentDetails } from '@/pages/EnvironmentDetails';
import { CreateEnvironment } from '@/pages/CreateEnvironment';
import { EditEnvironment } from '@/pages/EditEnvironment';
import { Logs } from '@/pages/Logs';
import { Login } from '@/pages/Login';
import { NotFound } from '@/pages/NotFound';
import Users from '@/pages/Users';

// Protected Route component
const ProtectedRoute: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const location = useLocation();
  const isAuthenticated = localStorage.getItem('authToken');
  
  if (!isAuthenticated) {
    return <Navigate to="/login" state={{ from: location }} replace />;
  }
  
  return <>{children}</>;
};

export const AppRoutes: React.FC = () => {
  return (
    <Routes>
      <Route path="/login" element={<Login />} />
      <Route
        path="/"
        element={
          <ProtectedRoute>
            <MainLayout />
          </ProtectedRoute>
        }
      >
        <Route index element={<Navigate to="/dashboard" replace />} />
        <Route path="dashboard" element={<Dashboard />} />
        <Route path="environments/create" element={<CreateEnvironment />} />
        <Route path="environments/:id" element={<EnvironmentDetails />} />
        <Route path="environments/:id/edit" element={<EditEnvironment />} />
        <Route path="logs" element={<Logs />} />
        <Route path="users" element={<Users />} />
        <Route path="*" element={<NotFound />} />
      </Route>
    </Routes>
  );
};
