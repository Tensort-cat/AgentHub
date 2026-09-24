import { useState } from "react";
import { useMutation, useQuery } from "@tanstack/react-query";
import { Check, CloudCog, ExternalLink } from "lucide-react";
import { getTool, listTools, setToolSubscription } from "../api/tools";
import { errorMessage } from "../api/client";
import { PageHeader } from "../components/layout/PageHeader";
import { Badge } from "../components/ui/Badge";
import { Dialog } from "../components/ui/Dialog";
import { EmptyState, ErrorState, PageSkeleton } from "../components/ui/States";
import { useToast } from "../components/ui/Toast";
import { formatDate } from "../lib/format";

export default function ToolsPage() {
  const { notify } = useToast();
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [sessionSubscriptions, setSessionSubscriptions] = useState<Set<string>>(() => new Set());
  const query = useQuery({ queryKey: ["tools"], queryFn: listTools });
  const detailQuery = useQuery({ queryKey: ["tool", selectedId], queryFn: () => getTool(selectedId!), enabled: Boolean(selectedId) });
  const subscriptionMutation = useMutation({
    mutationFn: ({ id, subscribed }: { id: string; subscribed: boolean }) => setToolSubscription(id, subscribed),
    onSuccess: (_, variables) => {
      setSessionSubscriptions((current) => {
        const next = new Set(current);
        if (variables.subscribed) next.add(variables.id); else next.delete(variables.id);
        return next;
      });
      notify(variables.subscribed ? "Tool subscribed for your account." : "Tool subscription removed.", "success");
    },
    onError: (error) => notify(errorMessage(error, "Unable to change the tool subscription."), "error"),
  });

  return (
    <div className="page">
      <PageHeader title="Tool center" description="Extend workflows with external capabilities." />
      <div className="inline-notice inline-notice--muted">The backend does not expose existing subscription state yet. Changes made here are shown for this page session only.</div>
      {query.isLoading ? <PageSkeleton /> : null}
      {query.isError ? <ErrorState description={errorMessage(query.error, "Unable to load the tool catalog.")} onRetry={() => void query.refetch()} /> : null}
      {query.data?.length === 0 ? <EmptyState title="No tools available" description="Enabled tools will appear here when the backend catalog is configured." /> : null}
      {query.data && query.data.length > 0 ? (
        <section className="tool-grid" aria-label="Available tools">
          {query.data.map((tool) => {
            const subscribed = sessionSubscriptions.has(tool.id);
            const isPending = subscriptionMutation.isPending && subscriptionMutation.variables?.id === tool.id;
            return (
              <article className="tool-tile" key={tool.id}>
                <button className="tool-tile__body" type="button" onClick={() => setSelectedId(tool.id)} aria-label={`View details for ${tool.name}`}>
                  <span className="tool-avatar"><CloudCog size={21} aria-hidden="true" /></span>
                  <span><strong>{tool.name}</strong><small>{tool.description || "No description"}</small></span>
                  <ExternalLink size={16} aria-hidden="true" />
                </button>
                <footer><Badge tone="blue">External tool</Badge><button className={`button ${subscribed ? "button--subscribed" : "button--secondary"}`} type="button" disabled={isPending} onClick={() => subscriptionMutation.mutate({ id: tool.id, subscribed: !subscribed })}>{subscribed ? <Check size={15} aria-hidden="true" /> : null}{isPending ? "Saving…" : subscribed ? "Subscribed" : "Subscribe"}</button></footer>
              </article>
            );
          })}
        </section>
      ) : null}
      <Dialog open={Boolean(selectedId)} title={detailQuery.data?.name ?? "Tool details"} description={detailQuery.data?.description} onClose={() => setSelectedId(null)} size="small">
        {detailQuery.isLoading ? <PageSkeleton rows={2} /> : null}
        {detailQuery.isError ? <ErrorState description={errorMessage(detailQuery.error, "Unable to load tool details.")} /> : null}
        {detailQuery.data ? <dl className="definition-list"><div><dt>Status</dt><dd><Badge tone={detailQuery.data.status === 2 ? "green" : "neutral"}>{detailQuery.data.status === 2 ? "Enabled" : "Disabled"}</Badge></dd></div><div><dt>Created</dt><dd>{formatDate(detailQuery.data.created_at)}</dd></div><div><dt>Integration</dt><dd>Connection details are not exposed by the API.</dd></div></dl> : null}
      </Dialog>
    </div>
  );
}

