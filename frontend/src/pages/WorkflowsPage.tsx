import { useEffect, useState, type FormEvent } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Link } from "react-router-dom";
import { ArrowUpRight, MoreHorizontal, Plus, Workflow as WorkflowIcon } from "lucide-react";
import { createWorkflow, deleteWorkflow, listWorkflows, updateWorkflow } from "../api/workflows";
import { errorMessage } from "../api/client";
import { PageHeader } from "../components/layout/PageHeader";
import { Badge } from "../components/ui/Badge";
import { ConfirmDialog, Dialog } from "../components/ui/Dialog";
import { EmptyState, ErrorState, PageSkeleton } from "../components/ui/States";
import { useToast } from "../components/ui/Toast";
import { formatDate } from "../lib/format";
import type { WorkflowSummary } from "../types/domain";
import { WorkflowStatus } from "../types/domain";

function WorkflowFormDialog({
  open,
  workflow,
  busy,
  onClose,
  onSubmit,
}: {
  open: boolean;
  workflow: WorkflowSummary | null;
  busy: boolean;
  onClose: () => void;
  onSubmit: (input: { name: string; description: string }) => void;
}) {
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");

  useEffect(() => {
    if (!open) return;
    setName(workflow?.name ?? "");
    setDescription(workflow?.description ?? "");
  }, [open, workflow]);

  const submit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    onSubmit({ name: name.trim(), description: description.trim() });
  };

  return (
    <Dialog
      open={open}
      title={workflow ? "Edit workflow" : "New workflow"}
      description={workflow ? "Update the details shown in your workflow list." : "Create a blank workspace, then add nodes on the canvas."}
      onClose={onClose}
      footer={
        <>
          <button className="button button--secondary" type="button" onClick={onClose} disabled={busy}>Cancel</button>
          <button className="button button--primary" type="submit" form="workflow-form" disabled={busy || !name.trim()}>{busy ? "Saving…" : workflow ? "Save changes" : "Create workflow"}</button>
        </>
      }
    >
      <form id="workflow-form" className="form-stack" onSubmit={submit}>
        <label className="field"><span>Name <b aria-hidden="true">*</b></span><input value={name} onChange={(event) => setName(event.target.value)} maxLength={100} required autoFocus /></label>
        <label className="field"><span>Description</span><textarea value={description} onChange={(event) => setDescription(event.target.value)} rows={4} placeholder="What does this workflow accomplish?" /></label>
      </form>
    </Dialog>
  );
}

export default function WorkflowsPage() {
  const queryClient = useQueryClient();
  const { notify } = useToast();
  const [formOpen, setFormOpen] = useState(false);
  const [editing, setEditing] = useState<WorkflowSummary | null>(null);
  const [deleting, setDeleting] = useState<WorkflowSummary | null>(null);
  const query = useQuery({ queryKey: ["workflows"], queryFn: () => listWorkflows() });

  const createMutation = useMutation({
    mutationFn: createWorkflow,
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["workflows"] });
      setFormOpen(false);
      notify("Workflow created.", "success");
    },
    onError: (error) => notify(errorMessage(error, "Unable to create the workflow."), "error"),
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, input }: { id: string; input: { name: string; description: string } }) => updateWorkflow(id, input),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["workflows"] });
      setFormOpen(false);
      setEditing(null);
      notify("Workflow details saved.", "success");
    },
    onError: (error) => notify(errorMessage(error, "Unable to update the workflow."), "error"),
  });

  const deleteMutation = useMutation({
    mutationFn: deleteWorkflow,
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["workflows"] });
      setDeleting(null);
      notify("Workflow deleted.", "success");
    },
    onError: (error) => notify(errorMessage(error, "Unable to delete the workflow."), "error"),
  });

  const openCreate = () => {
    setEditing(null);
    setFormOpen(true);
  };

  return (
    <div className="page">
      <PageHeader title="Workflows" description="Build and manage your AI workflows." actions={<button className="button button--primary" type="button" onClick={openCreate}><Plus size={17} aria-hidden="true" />New workflow</button>} />
      {query.isLoading ? <PageSkeleton /> : null}
      {query.isError ? <ErrorState description={errorMessage(query.error, "Check that the AgentHub server is running.")} onRetry={() => void query.refetch()} /> : null}
      {query.data?.length === 0 ? <EmptyState title="No workflows yet" description="Create your first workflow to build an AI application." action={<button className="button button--primary" type="button" onClick={openCreate}>Create workflow</button>} /> : null}
      {query.data && query.data.length > 0 ? (
        <section className="resource-table" aria-label="Workflows">
          <div className="resource-table__header resource-table__row--workflow"><span>Workflow</span><span>Status</span><span>Created</span><span><span className="sr-only">Actions</span></span></div>
          {query.data.map((workflow) => (
            <article className="resource-table__row resource-table__row--workflow" key={workflow.id}>
              <Link className="resource-identity" to={`/workflows/${workflow.id}`}>
                <span className="resource-icon"><WorkflowIcon size={17} aria-hidden="true" /></span>
                <span><strong>{workflow.name}</strong><small>{workflow.description || "No description"}</small></span>
              </Link>
              <span><Badge tone={workflow.status === WorkflowStatus.Active ? "green" : "neutral"}>{workflow.status === WorkflowStatus.Active ? "Enabled" : "Draft"}</Badge></span>
              <time dateTime={workflow.created_at}>{formatDate(workflow.created_at)}</time>
              <div className="row-actions">
                <Link className="icon-button icon-button--quiet" to={`/workflows/${workflow.id}`} aria-label={`Open ${workflow.name}`} title="Open workflow"><ArrowUpRight size={17} aria-hidden="true" /></Link>
                <details className="action-menu">
                  <summary className="icon-button icon-button--quiet" aria-label={`More actions for ${workflow.name}`}><MoreHorizontal size={18} aria-hidden="true" /></summary>
                  <div className="action-menu__popover">
                    <button type="button" onClick={() => { setEditing(workflow); setFormOpen(true); }}>Edit details</button>
                    <button className="text-danger" type="button" onClick={() => setDeleting(workflow)}>Delete workflow</button>
                  </div>
                </details>
              </div>
            </article>
          ))}
        </section>
      ) : null}
      <WorkflowFormDialog
        open={formOpen}
        workflow={editing}
        busy={createMutation.isPending || updateMutation.isPending}
        onClose={() => { setFormOpen(false); setEditing(null); }}
        onSubmit={(input) => editing ? updateMutation.mutate({ id: editing.id, input }) : createMutation.mutate(input)}
      />
      <ConfirmDialog open={Boolean(deleting)} title="Delete workflow?" description={`“${deleting?.name ?? "This workflow"}” and its nodes and edges will be removed.`} confirmLabel="Delete workflow" busy={deleteMutation.isPending} onClose={() => setDeleting(null)} onConfirm={() => deleting && deleteMutation.mutate(deleting.id)} />
    </div>
  );
}

