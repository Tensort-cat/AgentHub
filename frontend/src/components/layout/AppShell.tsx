import { useState } from "react";
import { NavLink, Outlet, useNavigate } from "react-router-dom";
import {
  BookOpen,
  Boxes,
  ChevronLeft,
  ChevronRight,
  Database,
  LogOut,
  Menu,
  Settings,
  Workflow,
} from "lucide-react";
import { clearAuth, readAuth } from "../../lib/auth-storage";

const navigation = [
  { to: "/workflows", label: "Workflows", icon: Workflow },
  { to: "/knowledge", label: "Knowledge", icon: BookOpen },
  { to: "/tools", label: "Tools", icon: Boxes },
  { to: "/models", label: "Models", icon: Database },
];

export function AppShell() {
  const [collapsed, setCollapsed] = useState(false);
  const [mobileOpen, setMobileOpen] = useState(false);
  const navigate = useNavigate();
  const user = readAuth()?.user;
  const initials = user?.name
    .split(/\s+/)
    .map((part) => part[0])
    .join("")
    .slice(0, 2)
    .toUpperCase() || "AH";

  const signOut = () => {
    clearAuth();
    navigate("/login", { replace: true });
  };

  return (
    <div className={`app-shell ${collapsed ? "app-shell--collapsed" : ""}`}>
      <button className="mobile-nav-toggle icon-button" type="button" onClick={() => setMobileOpen(true)} aria-label="Open navigation">
        <Menu size={20} aria-hidden="true" />
      </button>
      {mobileOpen ? <button className="sidebar-scrim" type="button" aria-label="Close navigation" onClick={() => setMobileOpen(false)} /> : null}
      <aside className={`sidebar ${mobileOpen ? "sidebar--mobile-open" : ""}`}>
        <div className="brand-lockup">
          <span className="brand-mark" aria-hidden="true"><i /><i /><i /></span>
          <span className="brand-name">AgentHub</span>
        </div>
        <nav className="primary-nav" aria-label="Primary navigation">
          {navigation.map(({ to, label, icon: Icon }) => (
            <NavLink key={to} to={to} onClick={() => setMobileOpen(false)} className={({ isActive }) => `nav-link ${isActive ? "nav-link--active" : ""}`} title={collapsed ? label : undefined}>
              <Icon size={18} aria-hidden="true" />
              <span>{label}</span>
            </NavLink>
          ))}
        </nav>
        <div className="sidebar__footer">
          <NavLink to="/settings" onClick={() => setMobileOpen(false)} className={({ isActive }) => `nav-link ${isActive ? "nav-link--active" : ""}`} title={collapsed ? "Settings" : undefined}>
            <Settings size={18} aria-hidden="true" />
            <span>Settings</span>
          </NavLink>
          <div className="account-block">
            <div className="avatar" aria-hidden="true">{initials}</div>
            <div className="account-block__copy">
              <strong>{user?.name ?? "Account"}</strong>
              <span>{user?.email ?? ""}</span>
            </div>
            <button className="icon-button icon-button--quiet" type="button" onClick={signOut} aria-label="Sign out" title="Sign out">
              <LogOut size={17} aria-hidden="true" />
            </button>
          </div>
          <button className="collapse-button" type="button" onClick={() => setCollapsed((value) => !value)} aria-label={collapsed ? "Expand sidebar" : "Collapse sidebar"}>
            {collapsed ? <ChevronRight size={16} aria-hidden="true" /> : <ChevronLeft size={16} aria-hidden="true" />}
            <span>{collapsed ? "Expand" : "Collapse"}</span>
          </button>
        </div>
      </aside>
      <main className="app-main"><Outlet /></main>
    </div>
  );
}

