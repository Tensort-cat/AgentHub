import { useQuery } from "@tanstack/react-query";
import { useNavigate } from "react-router-dom";
import { LogOut, ShieldCheck } from "lucide-react";
import { getCurrentUser } from "../api/auth";
import { errorMessage } from "../api/client";
import { PageHeader } from "../components/layout/PageHeader";
import { ErrorState, PageSkeleton } from "../components/ui/States";
import { clearAuth, readAuth } from "../lib/auth-storage";

export default function SettingsPage() {
  const navigate = useNavigate();
  const query = useQuery({ queryKey: ["current-user"], queryFn: getCurrentUser, initialData: readAuth()?.user });
  const signOut = () => { clearAuth(); navigate("/login", { replace: true }); };

  return (
    <div className="page page--narrow">
      <PageHeader title="Settings" description="Review your AgentHub account." />
      {query.isLoading ? <PageSkeleton rows={3} /> : null}
      {query.isError ? <ErrorState description={errorMessage(query.error, "Unable to load your account.")} onRetry={() => void query.refetch()} /> : null}
      {query.data ? (
        <section className="settings-panel">
          <header><div className="avatar avatar--large" aria-hidden="true">{query.data.name.slice(0, 2).toUpperCase()}</div><div><h2>{query.data.name}</h2><p>{query.data.email}</p></div></header>
          <dl className="settings-fields"><div><dt>Name</dt><dd>{query.data.name}</dd></div><div><dt>Email</dt><dd>{query.data.email}</dd></div><div><dt>Account ID</dt><dd className="mono">{query.data.id}</dd></div></dl>
          <div className="settings-note"><ShieldCheck size={18} aria-hidden="true" /><span>Profile editing is read-only because the current backend exposes no account update endpoint.</span></div>
          <footer><button className="button button--secondary" type="button" onClick={signOut}><LogOut size={17} aria-hidden="true" />Sign out</button></footer>
        </section>
      ) : null}
    </div>
  );
}

