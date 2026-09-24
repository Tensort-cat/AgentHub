import { backendId, normalizeId } from "../lib/json";
import type { JsonId, ToolDetail, ToolSummary } from "../types/domain";
import { apiRequest } from "./client";

interface RawToolSummary extends Omit<ToolSummary, "id"> {
  id: JsonId;
}

interface RawToolDetail extends Omit<ToolDetail, "id"> {
  id: JsonId;
}

export async function listTools(): Promise<ToolSummary[]> {
  const data = await apiRequest<RawToolSummary[] | null>("/tools");
  return (data ?? []).map((tool) => ({ ...tool, id: normalizeId(tool.id) }));
}

export async function getTool(id: string): Promise<ToolDetail> {
  const data = await apiRequest<RawToolDetail>(`/tools/${id}`);
  return { ...data, id: normalizeId(data.id) };
}

export function setToolSubscription(id: string, subscribed: boolean): Promise<null> {
  return apiRequest<null>("/user_tools", {
    method: "POST",
    json: { tool_id: backendId(id), status: subscribed ? 2 : 1 },
  });
}

