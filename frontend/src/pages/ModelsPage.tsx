import { useEffect, useMemo, useState, type FormEvent, type KeyboardEvent } from "react";
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
  onSubmit: (input: { name: string; base_url: string; api_key: string; type: ModelType; provider: number }) => void;
}) {
  const [name, setName] = useState("");
  const [baseUrl, setBaseUrl] = useState("");
  const [apiKey, setApiKey] = useState("");
  const [type, setType] = useState<ModelType>(ModelType.Chat);
  const [provider, setProvider] = useState(0);

  useEffect(() => {
    if (!open) return;
    setName(model?.name ?? "");
    setBaseUrl(model?.base_url ?? "");
    setApiKey("");
    setType(model?.type ?? ModelType.Chat);
    setProvider(model?.provider ?? 0);
  }, [open, model]);

  const submit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    onSubmit({ name: name.trim(), base_url: baseUrl.trim(), api_key: apiKey, type, provider });
  };

  return (
    <Dialog
      open={open}
      title={model ? "Edit model" : "Add model"}
      description={model ? "Update connection details. The model type cannot be changed." : "Connect a chat or embedding model to AgentHub."}
      onClose={onClose}
      footer={<><button className="button button--secondary" type="button" onClick={onClose} disabled={busy}>Cancel</button><button className="button button--primary" type="submit" form="model-form" disabled={busy || !provider || !name.trim() || !baseUrl.trim() || (!model && !apiKey)}>{busy ? "Saving…" : model ? "Save changes" : "Add model"}</button></>}
    >
      <form id="model-form" className="form-stack" onSubmit={submit}>
        <label className="field"><span>Name <b aria-hidden="true">*</b></span><input name="model_name" autoComplete="off" value={name} onChange={(event) => setName(event.target.value)} required /></label>
        <label className="field"><span>Base URL <b aria-hidden="true">*</b></span><input name="base_url" type="url" autoComplete="off" value={baseUrl} onChange={(event) => setBaseUrl(event.target.value)} placeholder="e.g. https://api.example.com/v1…" required /></label>
        <label className="field"><span>API key {!model ? <b aria-hidden="true">*</b> : null}</span><input name="api_key" type="password" value={apiKey} onChange={(event) => setApiKey(event.target.value)} autoComplete="off" required={!model} /><small>{model ? "Leave empty to keep the existing key." : "The key is sent only when this connection is saved."}</small></label>
        <label className="field"><span>Type <b aria-hidden="true">*</b></span><select name="model_type" autoComplete="off" value={type} onChange={(event) => setType(Number(event.target.value) as ModelType)} disabled={Boolean(model)}><option value={ModelType.Chat}>Chat model</option><option value={ModelType.Embedding}>Embedding model</option></select></label>
        <label className="field"><span>Provider <b aria-hidden="true">*</b></span><select name="provider" value={provider} onChange={(event) => setProvider(Number(event.target.value))} required><option value={0} disabled>Select provider</option><option value={1}>OpenAI compatible</option><option value={2}>Ark</option></select></label>
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
    mutationFn: ({ id, input }: { id: string; input: { name: string; base_url: string; api_key?: string; provider: number } }) => updateModel(id, input),
    onSuccess: async () => { await queryClient.invalidateQueries({ queryKey: ["models"] }); setDialogOpen(false); setEditing(null); notify("Model connection updated.", "success"); },
    onError: (error) => notify(errorMessage(error, "Unable to update the model."), "error"),
  });
  const deleteMutation = useMutation({
    mutationFn: deleteModel,
    onSuccess: async () => { await queryClient.invalidateQueries({ queryKey: ["models"] }); setDeleting(null); notify("Model connection deleted.", "success"); },
    onError: (error) => notify(errorMessage(error, "Unable to delete the model."), "error"),
  });

  const openCreate = () => { setEditing(null); setDialogOpen(true); };
  const handleTabKeyDown = (event: KeyboardEvent<HTMLButtonElement>) => {
    let next: ModelType | null = null;
    if (event.key === "ArrowLeft" || event.key === "ArrowRight") next = tab === ModelType.Chat ? ModelType.Embedding : ModelType.Chat;
    if (event.key === "Home") next = ModelType.Chat;
    if (event.key === "End") next = ModelType.Embedding;
    if (next === null) return;
    event.preventDefault();
    setTab(next);
    window.requestAnimationFrame(() => document.getElementById(`model-tab-${next}`)?.focus());
  };

  return (
    <div className="page">
      <PageHeader title="Models" description="Configure models used by workflows and knowledge bases." actions={<button className="button button--primary" type="button" onClick={openCreate}><Plus size={17} aria-hidden="true" />Add model</button>} />
      <div className="tabs" role="tablist" aria-label="Model types">
        <button id={`model-tab-${ModelType.Chat}`} role="tab" aria-controls="model-tabpanel" aria-selected={tab === ModelType.Chat} tabIndex={tab === ModelType.Chat ? 0 : -1} className={tab === ModelType.Chat ? "tabs__active" : ""} type="button" onClick={() => setTab(ModelType.Chat)} onKeyDown={handleTabKeyDown}>Chat models <span>{(query.data ?? []).filter((item) => item.type === ModelType.Chat).length}</span></button>
        <button id={`model-tab-${ModelType.Embedding}`} role="tab" aria-controls="model-tabpanel" aria-selected={tab === ModelType.Embedding} tabIndex={tab === ModelType.Embedding ? 0 : -1} className={tab === ModelType.Embedding ? "tabs__active" : ""} type="button" onClick={() => setTab(ModelType.Embedding)} onKeyDown={handleTabKeyDown}>Embedding models <span>{(query.data ?? []).filter((item) => item.type === ModelType.Embedding).length}</span></button>
      </div>
      <div id="model-tabpanel" role="tabpanel" aria-labelledby={`model-tab-${tab}`}>
        {query.isLoading ? <PageSkeleton /> : null}
        {query.isError ? <ErrorState description={errorMessage(query.error, "Unable to load model connections.")} onRetry={() => void query.refetch()} /> : null}
        {!query.isLoading && !query.isError && filtered.length === 0 ? <EmptyState title={tab === ModelType.Chat ? "No chat models" : "No embedding models"} description={tab === ModelType.Chat ? "Add a chat model before configuring model nodes." : "Add an embedding model before creating a knowledge base."} action={<button className="button button--primary" type="button" onClick={openCreate}>Add model</button>} /> : null}
        {filtered.length > 0 ? (
          <table className="resource-table" aria-label={tab === ModelType.Chat ? "Chat models" : "Embedding models"}>
            <thead><tr className="resource-table__header resource-table__row--model"><th scope="col">Model</th><th scope="col">Provider</th><th scope="col">Created</th><th scope="col"><span className="sr-only">Actions</span></th></tr></thead>
            <tbody>
            {filtered.map((model) => (
              <tr className="resource-table__row resource-table__row--model" key={model.id}>
                <td><div className="resource-identity"><span className="resource-icon"><KeyRound size={17} aria-hidden="true" /></span><span><strong>{model.name}</strong><small>{model.base_url}</small></span></div></td>
                <td><Badge tone={model.provider ? "blue" : "neutral"}>{providerName(model.provider)}</Badge></td>
                <td><time dateTime={model.created_at}>{formatDate(model.created_at)}</time></td>
                <td className="table-action-cell"><details className="action-menu row-action-single"><summary className="icon-button icon-button--quiet" aria-label={`More actions for ${model.name}`}><MoreHorizontal size={18} aria-hidden="true" /></summary><div className="action-menu__popover"><button type="button" onClick={() => { setEditing(model); setDialogOpen(true); }}>Edit connection</button><button className="text-danger" type="button" onClick={() => setDeleting(model)}>Delete model</button></div></details></td>
              </tr>
            ))}
            </tbody>
          </table>
        ) : null}
      </div>
      <ModelDialog open={dialogOpen} model={editing} busy={createMutation.isPending || updateMutation.isPending} onClose={() => { setDialogOpen(false); setEditing(null); }} onSubmit={(input) => {
        if (!editing) { createMutation.mutate(input); return; }
        const update = { name: input.name, base_url: input.base_url, provider: input.provider, ...(input.api_key ? { api_key: input.api_key } : {}) };
        updateMutation.mutate({ id: editing.id, input: update });
      }} />
      <ConfirmDialog open={Boolean(deleting)} title="Delete model?" description={`“${deleting?.name ?? "This model"}” will no longer be available to workflows or knowledge bases.`} confirmLabel="Delete model" busy={deleteMutation.isPending} onClose={() => setDeleting(null)} onConfirm={() => deleting && deleteMutation.mutate(deleting.id)} />
    </div>
  );
}
