import { useEffect, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { Clock3, MessageSquare } from "lucide-react";
import { listSessionMessages, listSessions } from "../../api/sessions";
import { errorMessage } from "../../api/client";
import { Dialog } from "../../components/ui/Dialog";
import { EmptyState, ErrorState, PageSkeleton } from "../../components/ui/States";
import { formatDateTime } from "../../lib/format";
import { MessageType } from "../../types/domain";

const messageLabels: Record<MessageType, string> = {
  [MessageType.User]: "User",
  [MessageType.Assistant]: "Assistant",
  [MessageType.Tool]: "Tool",
  [MessageType.System]: "System",
};

export function HistoryPanel({ open, workflowId, onClose }: { open: boolean; workflowId: string; onClose: () => void }) {
  const [selectedSessionId, setSelectedSessionId] = useState<string | null>(null);
  const sessionsQuery = useQuery({ queryKey: ["sessions", workflowId], queryFn: () => listSessions(workflowId), enabled: open });
  const messagesQuery = useQuery({ queryKey: ["session-messages", selectedSessionId], queryFn: () => listSessionMessages(selectedSessionId!), enabled: open && Boolean(selectedSessionId) });

  useEffect(() => {
    if (!open) setSelectedSessionId(null);
  }, [open]);

  return (
    <Dialog open={open} title="Run history" description="Inspect workflow sessions and the messages produced during each run." onClose={onClose} size="large">
      <div className="history-panel">
        <section className="session-list" aria-label="Workflow sessions">
          {sessionsQuery.isLoading ? <PageSkeleton rows={4} /> : null}
          {sessionsQuery.isError ? <ErrorState description={errorMessage(sessionsQuery.error, "Unable to load run history.")} /> : null}
          {sessionsQuery.data?.length === 0 ? <EmptyState title="No runs yet" description="Completed workflow runs will appear here." /> : null}
          {sessionsQuery.data?.map((session) => <button className={selectedSessionId === session.id ? "session-row session-row--selected" : "session-row"} type="button" key={session.id} onClick={() => setSelectedSessionId(session.id)}><Clock3 size={16} aria-hidden="true" /><span><strong>{session.title || "Untitled run"}</strong><small>{formatDateTime(session.created_at)}</small></span></button>)}
        </section>
        <section className="message-history" aria-label="Session messages">
          {!selectedSessionId ? <div className="history-placeholder"><MessageSquare size={24} aria-hidden="true" /><p>Select a run to inspect its messages.</p></div> : null}
          {messagesQuery.isLoading ? <PageSkeleton rows={4} /> : null}
          {messagesQuery.isError ? <ErrorState description={errorMessage(messagesQuery.error, "Unable to load session messages.")} /> : null}
          {selectedSessionId && messagesQuery.data?.length === 0 ? <EmptyState title="No messages" description="This session has no stored messages." /> : null}
          {messagesQuery.data?.map((message) => <article className={`history-message history-message--${message.type}`} key={message.id}><header><strong>{messageLabels[message.type] ?? "Message"}</strong><time dateTime={message.created_at}>{formatDateTime(message.created_at)}</time></header><pre>{message.content}</pre></article>)}
        </section>
      </div>
    </Dialog>
  );
}

