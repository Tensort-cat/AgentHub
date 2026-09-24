import { backendId, normalizeId, parseJson } from "../lib/json";
import type {
  BackendWorkflowEdge,
  BackendWorkflowNode,
  JsonId,
  WorkflowDetail,
  WorkflowNodeConfig,
  WorkflowRunState,
  WorkflowStatus,
  WorkflowSummary,
} from "../types/domain";
import { normalizeNodeConfig, serializeNodeConfig } from "../features/workflows/config";
import { apiRequest, apiUrl, ApiError } from "./client";
import { readAuth } from "../lib/auth-storage";

interface RawWorkflowSummary extends Omit<WorkflowSummary, "id"> {
  id: JsonId;
}

interface RawWorkflowNode extends Omit<BackendWorkflowNode, "id" | "config"> {
  id: JsonId;
  config: unknown;
}

interface RawWorkflowEdge extends Omit<BackendWorkflowEdge, "id" | "source_node_id" | "target_node_id"> {
  id: JsonId;
  source_node_id: JsonId;
  target_node_id: JsonId;
}

interface RawWorkflowDetail extends Omit<WorkflowDetail, "id" | "nodes" | "edges"> {
  id: JsonId;
  nodes: RawWorkflowNode[] | null;
  edges: RawWorkflowEdge[] | null;
}

export async function listWorkflows(page = 1, size = 50): Promise<WorkflowSummary[]> {
  const data = await apiRequest<RawWorkflowSummary[] | null>(`/workflows?page=${page}&size=${size}`);
  return (data ?? []).map((workflow) => ({ ...workflow, id: normalizeId(workflow.id) }));
}

export function createWorkflow(input: { name: string; description: string }): Promise<null> {
  return apiRequest<null>("/workflows", { method: "POST", json: input });
}

export function updateWorkflow(
  id: string,
  input: { name?: string; description?: string; status?: WorkflowStatus },
): Promise<null> {
  return apiRequest<null>(`/workflows/${id}`, { method: "PUT", json: input });
}

export function deleteWorkflow(id: string): Promise<null> {
  return apiRequest<null>(`/workflows/${id}`, { method: "DELETE" });
}

export async function getWorkflow(id: string): Promise<WorkflowDetail> {
  const data = await apiRequest<RawWorkflowDetail>(`/workflows/${id}`);
  return {
    ...data,
    id: normalizeId(data.id),
    nodes: (data.nodes ?? []).map((node) => ({
      ...node,
      id: normalizeId(node.id),
      config: normalizeNodeConfig(node.type, node.config),
    })),
    edges: (data.edges ?? []).map((edge) => ({
      ...edge,
      id: normalizeId(edge.id),
      source_node_id: normalizeId(edge.source_node_id),
      target_node_id: normalizeId(edge.target_node_id),
    })),
  };
}

export function createNode(
  workflowId: string,
  input: {
    name: string;
    type: BackendWorkflowNode["type"];
    position_x: number;
    position_y: number;
    config: WorkflowNodeConfig;
  },
): Promise<null> {
  return apiRequest<null>(`/workflows/${workflowId}/nodes`, {
    method: "POST",
    json: { ...input, config: serializeNodeConfig(input.type, input.config) },
  });
}

export function updateNode(
  workflowId: string,
  nodeId: string,
  type: BackendWorkflowNode["type"],
  input: { name?: string; position_x?: number; position_y?: number; config?: WorkflowNodeConfig },
): Promise<null> {
  return apiRequest<null>(`/workflows/${workflowId}/nodes/${nodeId}`, {
    method: "PUT",
    json: {
      ...input,
      ...(input.config === undefined ? {} : { config: serializeNodeConfig(type, input.config) }),
    },
  });
}

export function deleteNode(workflowId: string, nodeId: string): Promise<null> {
  return apiRequest<null>(`/workflows/${workflowId}/nodes/${nodeId}`, { method: "DELETE" });
}

export function createEdge(
  workflowId: string,
  input: { source_node_id: string; target_node_id: string; config: { branch_rule_id?: string } },
): Promise<null> {
  return apiRequest<null>(`/workflows/${workflowId}/edges`, {
    method: "POST",
    json: {
      ...input,
      source_node_id: backendId(input.source_node_id),
      target_node_id: backendId(input.target_node_id),
    },
  });
}

export function deleteEdge(workflowId: string, edgeId: string): Promise<null> {
  return apiRequest<null>(`/workflows/${workflowId}/edges/${edgeId}`, { method: "DELETE" });
}

export async function runWorkflow(workflowId: string, input: string): Promise<WorkflowRunState> {
  const data = await apiRequest<{ task_id: string; status: "queued" }>("/workflows/run", {
    method: "POST",
    json: { id: backendId(workflowId), input },
  });
  return { task_id: data.task_id, status: data.status };
}

function normalizeRunState(raw: Omit<WorkflowRunState, "session_id"> & { session_id?: JsonId }): WorkflowRunState {
  const { session_id: sessionId, ...state } = raw;
  return {
    ...state,
    ...(sessionId === undefined ? {} : { session_id: normalizeId(sessionId) }),
  };
}

export async function streamWorkflowRun(
  taskId: string,
  signal: AbortSignal,
  onState: (state: WorkflowRunState) => void,
): Promise<void> {
  const token = readAuth()?.token;
  const response = await fetch(apiUrl(`/workflows/run/${encodeURIComponent(taskId)}/events`), {
    headers: {
      Accept: "text/event-stream",
      ...(token ? { Authorization: token } : {}),
    },
    signal,
  });
  if (!response.ok || !response.body) {
    const text = await response.text();
    let message = "Unable to connect to the run stream.";
    try {
      message = parseJson<{ msg?: string }>(text).msg ?? message;
    } catch {
      // Keep the actionable fallback when the response is not JSON.
    }
    throw new ApiError(message, response.status === 403 ? 1002 : 4001, response.status);
  }

  const reader = response.body.getReader();
  const decoder = new TextDecoder();
  let buffer = "";

  while (true) {
    const { value, done } = await reader.read();
    if (done) break;
    buffer += decoder.decode(value, { stream: true }).replaceAll("\r\n", "\n");
    const events = buffer.split("\n\n");
    buffer = events.pop() ?? "";
    for (const eventBlock of events) {
      if (!eventBlock || eventBlock.startsWith(":")) continue;
      let eventName = "message";
      const dataLines: string[] = [];
      for (const line of eventBlock.split("\n")) {
        if (line.startsWith("event:")) eventName = line.slice(6).trim();
        if (line.startsWith("data:")) dataLines.push(line.slice(5).trimStart());
      }
      if (eventName === "workflow_run" && dataLines.length > 0) {
        const state = parseJson<Omit<WorkflowRunState, "session_id"> & { session_id?: JsonId }>(dataLines.join("\n"));
        onState(normalizeRunState(state));
      }
      if (eventName === "error" && dataLines.length > 0) {
        const error = parseJson<{ msg?: string }>(dataLines.join("\n"));
        throw new ApiError(error.msg ?? "The run stream closed with an error.", 4001, 200);
      }
    }
  }
}
