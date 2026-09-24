import { backendId, normalizeId } from "../lib/json";
import type {
  DocumentMeta,
  JsonId,
  KnowledgeBaseDetail,
  KnowledgeBaseSummary,
} from "../types/domain";
import { apiRequest } from "./client";

interface RawKnowledgeBaseSummary extends Omit<KnowledgeBaseSummary, "id"> {
  id: JsonId;
}

interface RawDocument extends Omit<DocumentMeta, "id"> {
  id: JsonId;
}

interface RawKnowledgeBaseDetail extends Omit<KnowledgeBaseDetail, "id" | "embedder_id" | "docs"> {
  id: JsonId;
  embedder_id: JsonId;
  docs: RawDocument[] | null;
}

export async function listKnowledgeBases(): Promise<KnowledgeBaseSummary[]> {
  const data = await apiRequest<RawKnowledgeBaseSummary[] | null>("/kb");
  return (data ?? []).map((item) => ({ ...item, id: normalizeId(item.id) }));
}

export function createKnowledgeBase(input: {
  name: string;
  description: string;
  embedder_id: string;
}): Promise<null> {
  return apiRequest<null>("/kb", {
    method: "POST",
    json: { ...input, embedder_id: backendId(input.embedder_id) },
  });
}

export async function getKnowledgeBase(id: string): Promise<KnowledgeBaseDetail> {
  const data = await apiRequest<RawKnowledgeBaseDetail>(`/kb/${id}`);
  return {
    ...data,
    id: normalizeId(data.id),
    embedder_id: normalizeId(data.embedder_id),
    docs: (data.docs ?? []).map((doc) => ({ ...doc, id: normalizeId(doc.id) })),
  };
}

export async function uploadDocument(knowledgeBaseId: string, file: File): Promise<DocumentMeta> {
  const form = new FormData();
  form.set("knowledge_base_id", knowledgeBaseId);
  form.set("file", file);
  const data = await apiRequest<Omit<RawDocument, "created_at">>("/docs", {
    method: "POST",
    body: form,
  });
  return { ...data, id: normalizeId(data.id), created_at: new Date().toISOString() };
}

