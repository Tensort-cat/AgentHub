import { lazy, Suspense, useSyncExternalStore, type ReactNode } from "react";
import { Navigate, Route, Routes } from "react-router-dom";
import { AppShell } from "./components/layout/AppShell";
import { readAuth } from "./lib/auth-storage";

const LoginPage = lazy(() => import("./pages/auth/LoginPage"));
const RegisterPage = lazy(() => import("./pages/auth/RegisterPage"));
const WorkflowsPage = lazy(() => import("./pages/WorkflowsPage"));
const WorkflowEditorPage = lazy(() => import("./pages/WorkflowEditorPage"));
const KnowledgePage = lazy(() => import("./pages/KnowledgePage"));
const KnowledgeDetailPage = lazy(() => import("./pages/KnowledgeDetailPage"));
const ToolsPage = lazy(() => import("./pages/ToolsPage"));
const ModelsPage = lazy(() => import("./pages/ModelsPage"));
const SettingsPage = lazy(() => import("./pages/SettingsPage"));
const NotFoundPage = lazy(() => import("./pages/NotFoundPage"));

function subscribeAuth(callback: () => void) {
  window.addEventListener("agenthub:auth-change", callback);
  window.addEventListener("storage", callback);
  return () => {
    window.removeEventListener("agenthub:auth-change", callback);
    window.removeEventListener("storage", callback);
  };
}

function useIsAuthenticated() {
  return useSyncExternalStore(subscribeAuth, () => Boolean(readAuth()?.token), () => false);
}

function Protected({ children }: { children: ReactNode }) {
  return useIsAuthenticated() ? children : <Navigate to="/login" replace />;
}

function PublicOnly({ children }: { children: ReactNode }) {
  return useIsAuthenticated() ? <Navigate to="/workflows" replace /> : children;
}

export default function App() {
  return (
    <Suspense fallback={<div className="route-loader" role="status">Loading workspace…</div>}>
      <Routes>
        <Route path="/login" element={<PublicOnly><LoginPage /></PublicOnly>} />
        <Route path="/register" element={<PublicOnly><RegisterPage /></PublicOnly>} />
        <Route element={<Protected><AppShell /></Protected>}>
          <Route index element={<Navigate to="/workflows" replace />} />
          <Route path="/workflows" element={<WorkflowsPage />} />
          <Route path="/knowledge" element={<KnowledgePage />} />
          <Route path="/knowledge/:knowledgeBaseId" element={<KnowledgeDetailPage />} />
          <Route path="/tools" element={<ToolsPage />} />
          <Route path="/models" element={<ModelsPage />} />
          <Route path="/settings" element={<SettingsPage />} />
        </Route>
        <Route path="/workflows/:workflowId" element={<Protected><WorkflowEditorPage /></Protected>} />
        <Route path="*" element={<NotFoundPage />} />
      </Routes>
    </Suspense>
  );
}

