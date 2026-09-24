import type { ReactNode } from "react";
import { AlertTriangle, Inbox, RefreshCw } from "lucide-react";

export function PageSkeleton({ rows = 5 }: { rows?: number }) {
  return (
    <div className="skeleton-list" aria-label="Loading" aria-busy="true">
      {Array.from({ length: rows }, (_, index) => (
        <div className="skeleton-row" key={index}>
          <span className="skeleton-block skeleton-block--title" />
          <span className="skeleton-block" />
          <span className="skeleton-block skeleton-block--short" />
        </div>
      ))}
    </div>
  );
}

export function EmptyState({
  title,
  description,
  action,
}: {
  title: string;
  description: string;
  action?: ReactNode;
}) {
  return (
    <div className="state-panel state-panel--empty">
      <Inbox size={24} aria-hidden="true" />
      <h2>{title}</h2>
      <p>{description}</p>
      {action}
    </div>
  );
}

export function ErrorState({
  title = "Unable to load this page",
  description,
  onRetry,
}: {
  title?: string;
  description: string;
  onRetry?: () => void;
}) {
  return (
    <div className="state-panel state-panel--error" role="alert">
      <AlertTriangle size={24} aria-hidden="true" />
      <h2>{title}</h2>
      <p>{description}</p>
      {onRetry ? (
        <button className="button button--secondary" type="button" onClick={onRetry}>
          <RefreshCw size={16} aria-hidden="true" /> Retry
        </button>
      ) : null}
    </div>
  );
}

