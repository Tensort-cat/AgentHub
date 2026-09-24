import { useEffect, useMemo, useState, type FormEvent } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { KeyRound, MoreHorizontal, Plus } from "lucide-react";
import { createModel, deleteModel, listModels, updateModel } from "../api/models";
import { errorMessage } from "../api/client";
import { PageHeader } from "../components/layout/PageHeader";
import { Badge } from "../components/ui/Badge";
import { ConfirmDialog, Dialog } from "../components/ui/Dialog";
import { EmptyState, ErrorState, PageSkeleton } from "../components/ui/States";
import { useToast } from "../components/ui/Toast";
import { formatDate } from "../lib/format";
import type { ModelConnection } from "../types/domain";
import { ModelType } from "../types/domain";

function providerName(provider: number): string {
  if (provider === 1) return "OpenAI";
  if (provider === 2) return "Ark";
  return "Not reported";
}

function ModelDialog({
  open,
  model,
  busy,
  onClose,
  onSubmit,
}: {
  open: boolean;
  model: ModelConnection | null;
  busy: boolean;
  onClose: () => void;
  onSubmit: (input: { name: string; base_url: string; api_key: string; type: ModelType }) => void;
}) {
  const [name, setName] = useState("");
  const [baseUrl, setBaseUrl] = useState("");
  const [apiKey, setApiKey] = useState("");
  const [type, setType] = useState<ModelType>(ModelType.Chat);

  useEffect(() => {
    if (!open) return;
    setName(model?.name ?? "");
    setBaseUrl(model?.base_url ?? "");
    setApiKey("");
    setType(model?.type ?? ModelType.Chat);
  }, [open, model]);

  const submit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    onSubmit({ name: name.trim(), base_url: baseUrl.trim(), api_key: apiKey, type });
  };

  return (
    <Dialog
      open={open}
      title={model ? "Edit model" : "Add model"}
      description={model ? "Update connection details. The model type cannot be changed." : "Connect a chat or embedding model to AgentHub."}
      onClose={onClose}
      footer={<><button className="button button--secondary" type="button" onClick={onClose} disabled={busy}>Cancel</button><button className="button button--primary" type="submit" form="model-form" disabled={busy || !name.trim() || !baseUrl.trim() || (!model && !apiKey)}>{busy ? "Saving…" : model ? "Save changes" : "Add model"}</button></>}
    >
      <form id="model-form" className="form-stack" onSubmit={submit}>
        <label className="field"><span>Name <b aria-hidden="true">*</b></span><input value={name} onChange={(event) => setName(event.target.value)} required autoFocus /></label>
        <label className="field"><span>Base URL <b aria-hidden="true">*</b></span><input type="url" value={baseUrl} onChange={(event) => setBaseUrl(event.target.value)} placeholder="https://api.example.com/v1" required /></label>
        <label className="field"><span>API key {!model ? <b aria-hidden="true">*</b> : null}</span><input type="password" value={apiKey} onChange={(event) => setApiKey(event.target.value)} autoComplete="off" required={!model} /><small>{model ? "Leave empty to keep the existing key." : "The key is sent only when this connection is saved."}</small></label>
        <label className="field"><span>Type <b aria-hidden="true">*</b></span><select value={type} onChange={(event) => setType(Number(event.target.value) as ModelType)} disabled={Boolean(model)}><option value={ModelType.Chat}>Chat model</option><option value={ModelType.Embedding}>Embedding model</option></select></label>
        <div className="form-note">Provider selection is unavailable because the current create API does not accept a provider field.</div>
      </form>
    </Dialog>
  );
}

export default function ModelsPage() {
  const queryClient = useQueryClient();
  const { notify } = useToast();
  const [tab, setTab] = useState<ModelType>(ModelType.Chat);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editing, setEditing] = useState<ModelConnection | null>(null);
  const [deleting, setDeleting] = useState<ModelConnection | null>(null);
  const query = useQuery({ queryKey: ["models"], queryFn: listModels });
  const filtered = useMemo(() => (query.data ?? []).filter((model) => model.type === tab), [query.data, tab]);

  const createMutation = useMutation({
    mutationFn: createModel,
    onSuccess: async () => { await queryClient.invalidateQueries({ queryKey: ["models"] }); setDialogOpen(false); notify("Model connection added.", "success"); },
    onError: (error) => notify(errorMessage(error, "Unable to add the model."), "error"),
  });
  const updateMutation = useMutation({
    mutationFn: ({ id, input }: { id: string; input: { name: string; base_url: string; api_key?: string } }) => updateModel(id, input),
    onSuccess: async () => { await queryClient.invalidateQueries({ queryKey: ["models"] }); setDialogOpen(false); setEditing(null); notify("Model connection updated.", "success"); },
    onError: (error) => notify(errorMessage(error, "Unable to update the model."), "error"),
  });
  const deleteMutation = useMutation({
    mutationFn: deleteModel,
    onSuccess: async () => { await queryClient.invalidateQueries({ queryKey: ["models"] }); setDeleting(null); notify("Model connection deleted.", "success"); },
    onError: (error) => notify(errorMessage(error, "Unable to delete the model."), "error"),
  });

  const openCreate = () => { setEditing(null); setDialogOpen(true); };

  return (
    <div className="page">
      <PageHeader title="Models" description="Configure models used by workflows and knowledge bases." actions={<button className="button button--primary" type="button" onClick={openCreate}><Plus size={17} aria-hidden="true" />Add model</button>} />
      <div className="tabs" role="tablist" aria-label="Model types">
        <button role="tab" aria-selected={tab === ModelType.Chat} className={tab === ModelType.Chat ? "tabs__active" : ""} type="button" onClick={() => setTab(ModelType.Chat)}>Chat models <span>{(query.data ?? []).filter((item) => item.type === ModelType.Chat).length}</span></button>
        <button role="tab" aria-selected={tab === ModelType.Embedding} className={tab === ModelType.Embedding ? "tabs__active" : ""} type="button" onClick={() => setTab(ModelType.Embedding)}>Embedding models <span>{(query.data ?? []).filter((item) => item.type === ModelType.Embedding).length}</span></button>
      </div>
      {query.isLoading ? <PageSkeleton /> : null}
      {query.isError ? <ErrorState description={errorMessage(query.error, "Unable to load model connections.")} onRetry={() => void query.refetch()} /> : null}
      {!query.isLoading && !query.isError && filtered.length === 0 ? <EmptyState title={tab === ModelType.Chat ? "No chat models" : "No embedding models"} description={tab === ModelType.Chat ? "Add a chat model before configuring model nodes." : "Add an embedding model before creating a knowledge base."} action={<button className="button button--primary" type="button" onClick={openCreate}>Add model</button>} /> : null}
      {filtered.length > 0 ? (
        <section className="resource-table" aria-label={tab === ModelType.Chat ? "Chat models" : "Embedding models"}>
          <div className="resource-table__header resource-table__row--model"><span>Model</span><span>Provider</span><span>Created</span><span><span className="sr-only">Actions</span></span></div>
          {filtered.map((model) => (
            <article className="resource-table__row resource-table__row--model" key={model.id}>
              <div className="resource-identity"><span className="resource-icon"><KeyRound size={17} aria-hidden="true" /></span><span><strong>{model.name}</strong><small>{model.base_url}</small></span></div>
              <span><Badge tone={model.provider ? "blue" : "neutral"}>{providerName(model.provider)}</Badge></span>
              <time dateTime={model.created_at}>{formatDate(model.created_at)}</time>
              <details className="action-menu row-action-single"><summary className="icon-button icon-button--quiet" aria-label={`More actions for ${model.name}`}><MoreHorizontal size={18} aria-hidden="true" /></summary><div className="action-menu__popover"><button type="button" onClick={() => { setEditing(model); setDialogOpen(true); }}>Edit connection</button><button className="text-danger" type="button" onClick={() => setDeleting(model)}>Delete model</button></div></details>
            </article>
          ))}
        </section>
      ) : null}
      <ModelDialog open={dialogOpen} model={editing} busy={createMutation.isPending || updateMutation.isPending} onClose={() => { setDialogOpen(false); setEditing(null); }} onSubmit={(input) => {
        if (!editing) { createMutation.mutate(input); return; }
        const update = { name: input.name, base_url: input.base_url, ...(input.api_key ? { api_key: input.api_key } : {}) };
        updateMutation.mutate({ id: editing.id, input: update });
      }} />
      <ConfirmDialog open={Boolean(deleting)} title="Delete model?" description={`“${deleting?.name ?? "This model"}” will no longer be available to workflows or knowledge bases.`} confirmLabel="Delete model" busy={deleteMutation.isPending} onClose={() => setDeleting(null)} onConfirm={() => deleting && deleteMutation.mutate(deleting.id)} />
    </div>
  );
}

