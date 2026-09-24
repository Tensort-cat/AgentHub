import { useRef, useState, type DragEvent } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Link, useParams } from "react-router-dom";
import { ArrowLeft, Download, FileText, UploadCloud } from "lucide-react";
import { getKnowledgeBase, uploadDocument } from "../api/knowledge";
import { listModels } from "../api/models";
import { errorMessage } from "../api/client";
import { Badge } from "../components/ui/Badge";
import { EmptyState, ErrorState, PageSkeleton } from "../components/ui/States";
import { useToast } from "../components/ui/Toast";
import { formatBytes, formatDateTime } from "../lib/format";
import { DocumentStatus } from "../types/domain";

const documentStatus = {
  [DocumentStatus.Pending]: { label: "Waiting", tone: "neutral" },
  [DocumentStatus.Parsing]: { label: "Parsing", tone: "blue" },
  [DocumentStatus.Indexing]: { label: "Indexing", tone: "amber" },
  [DocumentStatus.Success]: { label: "Ready", tone: "green" },
  [DocumentStatus.Error]: { label: "Failed", tone: "red" },
} as const;

export default function KnowledgeDetailPage() {
  const { knowledgeBaseId = "" } = useParams();
  const queryClient = useQueryClient();
  const { notify } = useToast();
  const inputRef = useRef<HTMLInputElement>(null);
  const [dragging, setDragging] = useState(false);
  const query = useQuery({
    queryKey: ["knowledge-base", knowledgeBaseId],
    queryFn: () => getKnowledgeBase(knowledgeBaseId),
    enabled: Boolean(knowledgeBaseId),
    refetchInterval: (current) => {
      const data = current.state.data;
      return data?.docs.some((doc) => [DocumentStatus.Pending, DocumentStatus.Parsing, DocumentStatus.Indexing].includes(doc.status)) ? 2500 : false;
    },
  });
  const modelsQuery = useQuery({ queryKey: ["models"], queryFn: listModels });
  const uploadMutation = useMutation({
    mutationFn: (file: File) => uploadDocument(knowledgeBaseId, file),
    onSuccess: async () => { await queryClient.invalidateQueries({ queryKey: ["knowledge-base", knowledgeBaseId] }); notify("Document uploaded and indexed.", "success"); },
    onError: (error) => notify(errorMessage(error, "Unable to upload this document."), "error"),
  });

  const upload = (file?: File) => {
    if (!file) return;
    if (!file.name.toLowerCase().endsWith(".md")) {
      notify("Only Markdown (.md) files can be uploaded.", "error");
      return;
    }
    uploadMutation.mutate(file);
  };

  const drop = (event: DragEvent<HTMLDivElement>) => {
    event.preventDefault();
    setDragging(false);
    upload(event.dataTransfer.files[0]);
  };

  if (query.isLoading) return <div className="page"><PageSkeleton /></div>;
  if (query.isError || !query.data) return <div className="page"><ErrorState description={errorMessage(query.error, "Unable to load this knowledge base.")} onRetry={() => void query.refetch()} /></div>;

  const base = query.data;
  const embedder = modelsQuery.data?.find((model) => model.id === base.embedder_id);

  return (
    <div className="page">
      <Link className="back-link" to="/knowledge"><ArrowLeft size={16} aria-hidden="true" />Knowledge</Link>
      <header className="detail-header">
        <div><h1>{base.name}</h1><p>{base.description || "No description"}</p></div>
        <div className="detail-meta"><span>Embedding model</span><strong>{embedder?.name ?? `Model ${base.embedder_id}`}</strong></div>
      </header>
      <section className="content-section">
        <div className="section-heading"><div><h2>Documents</h2><p>Markdown files available to retrieval nodes.</p></div><button className="button button--primary" type="button" onClick={() => inputRef.current?.click()} disabled={uploadMutation.isPending}><UploadCloud size={17} aria-hidden="true" />{uploadMutation.isPending ? "Uploading…" : "Upload Markdown"}</button></div>
        <input ref={inputRef} className="sr-only" type="file" accept=".md,text/markdown" onChange={(event) => { upload(event.target.files?.[0]); event.target.value = ""; }} />
        <div className={`drop-zone ${dragging ? "drop-zone--active" : ""}`} onDragEnter={(event) => { event.preventDefault(); setDragging(true); }} onDragOver={(event) => event.preventDefault()} onDragLeave={() => setDragging(false)} onDrop={drop}>
          <UploadCloud size={22} aria-hidden="true" /><span>Drop a <strong>.md</strong> file here</span><small>One file at a time</small>
        </div>
        {base.docs.length === 0 ? <EmptyState title="No documents" description="Upload a Markdown file to add source material to this knowledge base." /> : (
          <div className="resource-table document-table" aria-label="Documents">
            <div className="resource-table__header resource-table__row--document"><span>Name</span><span>Status</span><span>Size</span><span>Created</span><span><span className="sr-only">Download</span></span></div>
            {base.docs.map((doc) => {
              const status = documentStatus[doc.status] ?? documentStatus[DocumentStatus.Error];
              return <article className="resource-table__row resource-table__row--document" key={doc.id}><div className="resource-identity"><span className="resource-icon"><FileText size={17} aria-hidden="true" /></span><span><strong>{doc.name}</strong><small>{doc.type}</small></span></div><span><Badge tone={status.tone}>{status.label}</Badge></span><span>{formatBytes(doc.size)}</span><time dateTime={doc.created_at}>{formatDateTime(doc.created_at)}</time><button className="icon-button icon-button--quiet" type="button" disabled title="Download is unavailable until the backend route parameter is fixed" aria-label={`Download ${doc.name} unavailable`}><Download size={16} aria-hidden="true" /></button></article>;
            })}
          </div>
        )}
      </section>
    </div>
  );
}

