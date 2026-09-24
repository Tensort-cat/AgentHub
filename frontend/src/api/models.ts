import { normalizeId } from "../lib/json";
import type { JsonId, ModelConnection, ModelType } from "../types/domain";
import { apiRequest } from "./client";

interface RawModel extends Omit<ModelConnection, "id"> {
  id: JsonId;
}

export async function listModels(): Promise<ModelConnection[]> {
  const data = await apiRequest<RawModel[] | null>("/models");
  return (data ?? []).map((model) => ({ ...model, id: normalizeId(model.id) }));
}

export function createModel(input: {
  name: string;
  base_url: string;
  api_key: string;
  type: ModelType;
}): Promise<null> {
  // The backend does not accept provider yet, despite returning it in list DTOs.
  return apiRequest<null>("/models", { method: "POST", json: input });
}

export function updateModel(
  id: string,
  input: { name?: string; base_url?: string; api_key?: string },
): Promise<null> {
  return apiRequest<null>(`/models/${id}`, { method: "PUT", json: input });
}

export function deleteModel(id: string): Promise<null> {
  return apiRequest<null>(`/models/${id}`, { method: "DELETE" });
}

