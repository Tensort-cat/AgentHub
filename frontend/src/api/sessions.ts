import { normalizeId } from "../lib/json";
import type { JsonId, SessionMessage, WorkflowSession } from "../types/domain";
import { apiRequest } from "./client";

interface RawSession extends Omit<WorkflowSession, "id"> {
  id: JsonId;
}

interface RawMessage extends Omit<SessionMessage, "id"> {
  id: JsonId;
}

export async function listSessions(workflowId: string): Promise<WorkflowSession[]> {
  const data = await apiRequest<RawSession[] | null>(`/sessions?wf_id=${encodeURIComponent(workflowId)}&page=1&size=50`);
  return (data ?? []).map((session) => ({ ...session, id: normalizeId(session.id) }));
}

export async function listSessionMessages(sessionId: string): Promise<SessionMessage[]> {
  const data = await apiRequest<RawMessage[] | null>(`/sessions/${sessionId}?page=1&size=100`);
  return (data ?? [])
    .map((message) => ({ ...message, id: normalizeId(message.id) }))
    .reverse();
}

