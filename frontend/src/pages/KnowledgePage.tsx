import { useEffect, useMemo, useState, type FormEvent } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Link } from "react-router-dom";
import { ArrowUpRight, BookOpen, Plus } from "lucide-react";
import { createKnowledgeBase, listKnowledgeBases } from "../api/knowledge";
import { listModels } from "../api/models";
import { errorMessage } from "../api/client";
import { PageHeader } from "../components/layout/PageHeader";
import { Dialog } from "../components/ui/Dialog";
import { EmptyState, ErrorState, PageSkeleton } from "../components/ui/States";
import { useToast } from "../components/ui/Toast";
import { formatDate } from "../lib/format";
import { ModelType } from "../types/domain";

function KnowledgeDialog({
  open,
  busy,
  models,
  onClose,
  onSubmit,
}: {
  open: boolean;
  busy: boolean;
  models: { id: string; name: string }[];
  onClose: () => void;
  onSubmit: (input: { name: string; description: string; embedder_id: string }) => void;
}) {
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [embedderId, setEmbedderId] = useState("");

  useEffect(() => {
    if (!open) return;
    setName("");
    setDescription("");
    setEmbedderId(models[0]?.id ?? "");
  }, [open, models]);

  const submit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    onSubmit({ name: name.trim(), description: description.trim(), embedder_id: embedderId });
  };

  return (
    <Dialog open={open} title="New knowledge base" description="Choose the embedding model used to index and retrieve documents." onClose={onClose} footer={<><button className="button button--secondary" type="button" onClick={onClose} disabled={busy}>Cancel</button><button className="button button--primary" type="submit" form="knowledge-form" disabled={busy || !name.trim() || !embedderId}>{busy ? "Creating…" : "Create knowledge base"}</button></>}>
      <form id="knowledge-form" className="form-stack" onSubmit={submit}>
        <label className="field"><span>Name <b aria-hidden="true">*</b></span><input value={name} onChange={(event) => setName(event.target.value)} required autoFocus /></label>
        <label className="field"><span>Description</span><textarea value={description} onChange={(event) => setDescription(event.target.value)} rows={3} /></label>
        <label className="field"><span>Embedding model <b aria-hidden="true">*</b></span><select value={embedderId} onChange={(event) => setEmbedderId(event.target.value)} required><option value="" disabled>Select a model</option>{models.map((model) => <option key={model.id} value={model.id}>{model.name}</option>)}</select></label>
      </form>
    </Dialog>
  );
}

export default function KnowledgePage() {
  const queryClient = useQueryClient();
  const { notify } = useToast();
  const [dialogOpen, setDialogOpen] = useState(false);
  const knowledgeQuery = useQuery({ queryKey: ["knowledge-bases"], queryFn: listKnowledgeBases });
  const modelsQuery = useQuery({ queryKey: ["models"], queryFn: listModels });
  const embeddingModels = useMemo(() => (modelsQuery.data ?? []).filter((model) => model.type === ModelType.Embedding), [modelsQuery.data]);
  const mutation = useMutation({
    mutationFn: createKnowledgeBase,
    onSuccess: async () => { await queryClient.invalidateQueries({ queryKey: ["knowledge-bases"] }); setDialogOpen(false); notify("Knowledge base created.", "success"); },
    onError: (error) => notify(errorMessage(error, "Unable to create the knowledge base."), "error"),
  });

  const openCreate = () => {
    if (embeddingModels.length === 0) {
      notify("Create an embedding model before creating a knowledge base.", "info");
      return;
    }
    setDialogOpen(true);
  };

  return (
    <div className="page">
      <PageHeader title="Knowledge" description="Manage the source material used for retrieval." actions={<button className="button button--primary" type="button" onClick={openCreate} disabled={modelsQuery.isLoading}><Plus size={17} aria-hidden="true" />New knowledge base</button>} />
      {!modelsQuery.isLoading && embeddingModels.length === 0 ? <div className="inline-notice">An embedding model is required before you can create a knowledge base. <Link to="/models">Go to models</Link></div> : null}
      {knowledgeQuery.isLoading ? <PageSkeleton /> : null}
      {knowledgeQuery.isError ? <ErrorState description={errorMessage(knowledgeQuery.error, "Unable to load knowledge bases.")} onRetry={() => void knowledgeQuery.refetch()} /> : null}
      {knowledgeQuery.data?.length === 0 ? <EmptyState title="No knowledge bases" description="Create a knowledge base, then add Markdown documents for retrieval." action={embeddingModels.length > 0 ? <button className="button button--primary" type="button" onClick={openCreate}>Create knowledge base</button> : <Link className="button button--primary" to="/models">Add embedding model</Link>} /> : null}
      {knowledgeQuery.data && knowledgeQuery.data.length > 0 ? (
        <section className="resource-table" aria-label="Knowledge bases">
          <div className="resource-table__header resource-table__row--knowledge"><span>Knowledge base</span><span>Created</span><span><span className="sr-only">Open</span></span></div>
          {knowledgeQuery.data.map((base) => (
            <article className="resource-table__row resource-table__row--knowledge" key={base.id}>
              <Link className="resource-identity" to={`/knowledge/${base.id}`}><span className="resource-icon resource-icon--knowledge"><BookOpen size={17} aria-hidden="true" /></span><span><strong>{base.name}</strong><small>{base.description || "No description"}</small></span></Link>
              <time dateTime={base.created_at}>{formatDate(base.created_at)}</time>
              <Link className="icon-button icon-button--quiet" to={`/knowledge/${base.id}`} aria-label={`Open ${base.name}`}><ArrowUpRight size={17} aria-hidden="true" /></Link>
            </article>
          ))}
        </section>
      ) : null}
      <KnowledgeDialog open={dialogOpen} busy={mutation.isPending} models={embeddingModels} onClose={() => setDialogOpen(false)} onSubmit={(input) => mutation.mutate(input)} />
    </div>
  );
}

