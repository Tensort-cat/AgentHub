import { useCallback, useEffect, useMemo, useRef, useState, type DragEvent } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Link, useParams } from "react-router-dom";
import {
  Background,
  BackgroundVariant,
  Controls,
  MiniMap,
  Panel,
  ReactFlow,
  ReactFlowProvider,
  useEdgesState,
  useNodesState,
  useNodesInitialized,
  useReactFlow,
  type Connection,
  type OnNodeDrag,
  type NodeTypes,
} from "@xyflow/react";
import { ArrowLeft, Clock3, PanelLeftClose, PanelRightClose, Play, Trash2 } from "lucide-react";
import { listKnowledgeBases } from "../api/knowledge";
import { listModels } from "../api/models";
import { listTools } from "../api/tools";
import {
  createEdge,
  createNode,
  deleteEdge,
  deleteNode,
  getWorkflow,
  listWorkflows,
  updateNode,
  updateWorkflow,
} from "../api/workflows";
import { errorMessage } from "../api/client";
import { Badge } from "../components/ui/Badge";
import { ErrorState } from "../components/ui/States";
import { useToast } from "../components/ui/Toast";
import { HistoryPanel } from "../features/workflows/HistoryPanel";
import { NodeInspector } from "../features/workflows/NodeInspector";
import { NodePalette } from "../features/workflows/NodePalette";
import { RunPanel } from "../features/workflows/RunPanel";
import { defaultNodeConfig } from "../features/workflows/config";
import { mapBackendEdgeToFlowEdge, mapBackendNodeToFlowNode } from "../features/workflows/mappers";
import { nodeDefinitionByType } from "../features/workflows/node-meta";
import { WorkflowNode } from "../features/workflows/nodes/WorkflowNode";
import type { AgentHubFlowEdge, AgentHubFlowNode, WorkflowNodeConfig } from "../types/domain";
import { ModelType, WorkflowNodeType, WorkflowStatus } from "../types/domain";

const nodeTypes: NodeTypes = { agentHub: WorkflowNode };

type SaveState = "saved" | "saving" | "failed";

function EditorWorkspace({ workflowId }: { workflowId: string }) {
  const queryClient = useQueryClient();
  const { notify } = useToast();
  const { screenToFlowPosition, fitView } = useReactFlow<AgentHubFlowNode, AgentHubFlowEdge>();
  const workflowQuery = useQuery({ queryKey: ["workflow", workflowId], queryFn: () => getWorkflow(workflowId) });
  const workflowsQuery = useQuery({ queryKey: ["workflows"], queryFn: () => listWorkflows() });
  const modelsQuery = useQuery({ queryKey: ["models"], queryFn: listModels });
  const knowledgeQuery = useQuery({ queryKey: ["knowledge-bases"], queryFn: listKnowledgeBases });
  const toolsQuery = useQuery({ queryKey: ["tools"], queryFn: listTools });
  const [nodes, setNodes, onNodesChange] = useNodesState<AgentHubFlowNode>([]);
  const [edges, setEdges, onEdgesChange] = useEdgesState<AgentHubFlowEdge>([]);
  const [selectedNodeId, setSelectedNodeId] = useState<string | null>(null);
  const [selectedEdgeId, setSelectedEdgeId] = useState<string | null>(null);
  const [saveState, setSaveState] = useState<SaveState>("saved");
  const [runOpen, setRunOpen] = useState(false);
  const [historyOpen, setHistoryOpen] = useState(false);
  const [paletteOpen, setPaletteOpen] = useState(true);
  const [inspectorOpen, setInspectorOpen] = useState(true);
  const nodesInitialized = useNodesInitialized();
  const hasFittedRef = useRef(false);

  useEffect(() => {
    if (!workflowQuery.data) return;
    setNodes(workflowQuery.data.nodes.map(mapBackendNodeToFlowNode));
    setEdges(workflowQuery.data.edges.map(mapBackendEdgeToFlowEdge));
  }, [workflowQuery.data, setEdges, setNodes]);

  useEffect(() => {
    if (!nodesInitialized || nodes.length === 0 || hasFittedRef.current) return;
    hasFittedRef.current = true;
    const frame = window.requestAnimationFrame(() => {
      void fitView({ padding: 0.22, duration: 0 });
    });
    return () => window.cancelAnimationFrame(frame);
  }, [fitView, nodes.length, nodesInitialized]);

  const refetchWorkflow = useCallback(async () => {
    await queryClient.invalidateQueries({ queryKey: ["workflow", workflowId] });
  }, [queryClient, workflowId]);

  const mutationError = useCallback((error: unknown, fallback: string) => {
    setSaveState("failed");
    notify(errorMessage(error, fallback), "error");
  }, [notify]);

  const createNodeMutation = useMutation({
    mutationFn: (input: Parameters<typeof createNode>[1]) => createNode(workflowId, input),
    onMutate: () => setSaveState("saving"),
    onSuccess: async () => { await refetchWorkflow(); setSaveState("saved"); notify("Node added.", "success"); },
    onError: (error) => mutationError(error, "Unable to add the node."),
  });
  const updateNodeMutation = useMutation({
    mutationFn: ({ id, type, input }: { id: string; type: WorkflowNodeType; input: { name?: string; position_x?: number; position_y?: number; config?: WorkflowNodeConfig } }) => updateNode(workflowId, id, type, input),
    onMutate: () => setSaveState("saving"),
    onSuccess: async (_, variables) => {
      if (variables.input.config !== undefined || variables.input.name !== undefined) await refetchWorkflow();
      setSaveState("saved");
    },
    onError: async (error) => { await refetchWorkflow(); mutationError(error, "Unable to save the node."); },
  });
  const deleteNodeMutation = useMutation({
    mutationFn: (id: string) => deleteNode(workflowId, id),
    onMutate: () => setSaveState("saving"),
    onSuccess: async () => { setSelectedNodeId(null); await refetchWorkflow(); setSaveState("saved"); notify("Node deleted.", "success"); },
    onError: (error) => mutationError(error, "Unable to delete the node."),
  });
  const createEdgeMutation = useMutation({
    mutationFn: (input: Parameters<typeof createEdge>[1]) => createEdge(workflowId, input),
    onMutate: () => setSaveState("saving"),
    onSuccess: async () => { await refetchWorkflow(); setSaveState("saved"); },
    onError: (error) => mutationError(error, "Unable to create the connection."),
  });
  const deleteEdgeMutation = useMutation({
    mutationFn: (id: string) => deleteEdge(workflowId, id),
    onMutate: () => setSaveState("saving"),
    onSuccess: async () => { setSelectedEdgeId(null); await refetchWorkflow(); setSaveState("saved"); notify("Connection deleted.", "success"); },
    onError: (error) => mutationError(error, "Unable to delete the connection."),
  });
  const statusMutation = useMutation({
    mutationFn: (status: WorkflowStatus) => updateWorkflow(workflowId, { status }),
    onSuccess: async () => { await queryClient.invalidateQueries({ queryKey: ["workflows"] }); notify("Workflow status updated.", "success"); },
    onError: (error) => notify(errorMessage(error, "Unable to update workflow status."), "error"),
  });

  const selectedNode = useMemo(() => nodes.find((node) => node.id === selectedNodeId) ?? null, [nodes, selectedNodeId]);
  const nodeTypeSet = useMemo(() => new Set(nodes.map((node) => node.data.nodeType)), [nodes]);
  const nodeById = useMemo(() => new Map(nodes.map((node) => [node.id, node])), [nodes]);
  const connectedRuleIds = useMemo(() => new Set(edges.filter((edge) => edge.source === selectedNodeId && edge.sourceHandle).map((edge) => edge.sourceHandle!)), [edges, selectedNodeId]);
  const summary = workflowsQuery.data?.find((workflow) => workflow.id === workflowId);
  const workflowStatus = summary?.status ?? WorkflowStatus.Draft;
  const chatModelId = modelsQuery.data?.find((model) => model.type === ModelType.Chat)?.id;
  const knowledgeBaseId = knowledgeQuery.data?.[0]?.id;

  const connectionError = useCallback((connection: Connection | AgentHubFlowEdge): string | null => {
    const source = connection.source ? nodeById.get(connection.source) : undefined;
    const target = connection.target ? nodeById.get(connection.target) : undefined;
    if (!source || !target) return "Choose a valid source and target.";
    if (source.id === target.id) return "A node cannot connect to itself.";
    if (source.data.nodeType === WorkflowNodeType.End) return "End nodes cannot have outgoing connections.";
    if (target.data.nodeType === WorkflowNodeType.Start) return "Start nodes cannot have incoming connections.";
    if (source.data.nodeType === WorkflowNodeType.Branch && target.data.nodeType === WorkflowNodeType.Branch) return "Branch nodes cannot connect directly to another Branch node.";
    if (target.data.nodeType === WorkflowNodeType.Branch && edges.some((edge) => edge.target === target.id)) return "A Branch node can have only one incoming connection.";
    if (target.data.nodeType === WorkflowNodeType.Branch && edges.some((edge) => edge.source === source.id)) return "The node before a Branch cannot have another outgoing connection.";
    const sourceAlreadyLeadsToBranch = edges.some((edge) => edge.source === source.id && nodeById.get(edge.target)?.data.nodeType === WorkflowNodeType.Branch);
    if (sourceAlreadyLeadsToBranch) return "A node feeding a Branch cannot have another outgoing connection.";
    if (source.data.nodeType === WorkflowNodeType.Branch) {
      if (!connection.sourceHandle) return "Choose a Branch rule outlet.";
      if (edges.some((edge) => edge.source === source.id && edge.sourceHandle === connection.sourceHandle)) return "Each Branch outlet can have only one connection.";
    } else if (edges.some((edge) => edge.source === source.id && edge.target === target.id)) {
      return "This connection already exists.";
    }
    return null;
  }, [edges, nodeById]);

  const onConnect = useCallback((connection: Connection) => {
    const problem = connectionError(connection);
    if (problem || !connection.source || !connection.target) {
      notify(problem ?? "Unable to create this connection.", "error");
      return;
    }
    const source = nodeById.get(connection.source)!;
    createEdgeMutation.mutate({
      source_node_id: connection.source,
      target_node_id: connection.target,
      config: source.data.nodeType === WorkflowNodeType.Branch ? { branch_rule_id: connection.sourceHandle ?? "$default" } : {},
    });
  }, [connectionError, createEdgeMutation, nodeById, notify]);

  const onDrop = useCallback((event: DragEvent<HTMLDivElement>) => {
    event.preventDefault();
    const type = Number(event.dataTransfer.getData("application/agenthub-node")) as WorkflowNodeType;
    const definition = nodeDefinitionByType.get(type);
    if (!definition || definition.unavailable) return;
    if ((type === WorkflowNodeType.Start || type === WorkflowNodeType.End) && nodeTypeSet.has(type)) {
      notify(`Only one ${definition.label} node is allowed.`, "error");
      return;
    }
    const position = screenToFlowPosition({ x: event.clientX, y: event.clientY });
    createNodeMutation.mutate({
      name: definition.label,
      type,
      position_x: Math.round(position.x),
      position_y: Math.round(position.y),
      config: defaultNodeConfig(type, { chatModelId, knowledgeBaseId }),
    });
  }, [chatModelId, createNodeMutation, knowledgeBaseId, nodeTypeSet, notify, screenToFlowPosition]);

  const onNodeDragStop: OnNodeDrag<AgentHubFlowNode> = useCallback((_, node) => {
    updateNodeMutation.mutate({ id: node.id, type: node.data.nodeType, input: { position_x: Math.round(node.position.x), position_y: Math.round(node.position.y) } });
  }, [updateNodeMutation]);

  if (workflowQuery.isLoading) return <div className="editor-loading" role="status"><span className="brand-mark" aria-hidden="true"><i /><i /><i /></span><p>Opening workflow canvas…</p></div>;
  if (workflowQuery.isError || !workflowQuery.data) return <div className="editor-error"><ErrorState description={errorMessage(workflowQuery.error, "Unable to load this workflow.")} onRetry={() => void workflowQuery.refetch()} /><Link className="button button--secondary" to="/workflows">Back to workflows</Link></div>;

  const workflow = workflowQuery.data;
  return (
    <main className={`workflow-editor ${paletteOpen ? "workflow-editor--palette-open" : ""} ${inspectorOpen ? "workflow-editor--inspector-open" : ""}`}>
      <header className="editor-header">
        <div className="editor-header__identity"><Link className="icon-button icon-button--quiet" to="/workflows" aria-label="Back to workflows"><ArrowLeft size={18} aria-hidden="true" /></Link><span className="editor-wordmark"><span className="brand-mark brand-mark--small" aria-hidden="true"><i /><i /><i /></span>AgentHub</span><span className="editor-divider" /><div><h1>{workflow.name}</h1><p>{workflow.description || "No description"}</p></div></div>
        <div className="editor-header__actions">
          <span className={`save-state save-state--${saveState}`}>{saveState === "saving" ? "Saving…" : saveState === "failed" ? "Save failed" : "Saved"}</span>
          <button className="status-button" type="button" disabled={statusMutation.isPending} onClick={() => statusMutation.mutate(workflowStatus === WorkflowStatus.Active ? WorkflowStatus.Draft : WorkflowStatus.Active)} title="Change workflow status"><Badge tone={workflowStatus === WorkflowStatus.Active ? "green" : "neutral"}>{workflowStatus === WorkflowStatus.Active ? "Enabled" : "Draft"}</Badge></button>
          <button className="button button--secondary editor-history-button" type="button" onClick={() => setHistoryOpen(true)}><Clock3 size={16} aria-hidden="true" />History</button>
          <button className="button button--primary" type="button" onClick={() => setRunOpen(true)}><Play size={16} fill="currentColor" aria-hidden="true" />Run</button>
        </div>
      </header>
      <div className="editor-body">
        {paletteOpen ? <NodePalette nodeTypes={nodeTypeSet} /> : null}
        <section className="flow-canvas" aria-label="Workflow canvas" onDrop={onDrop} onDragOver={(event) => { event.preventDefault(); event.dataTransfer.dropEffect = "move"; }}>
          <ReactFlow<AgentHubFlowNode, AgentHubFlowEdge>
            nodes={nodes}
            edges={edges}
            nodeTypes={nodeTypes}
            onNodesChange={onNodesChange}
            onEdgesChange={onEdgesChange}
            onConnect={onConnect}
            onNodeDragStop={onNodeDragStop}
            onNodeClick={(_, node) => { setSelectedNodeId(node.id); setSelectedEdgeId(null); setInspectorOpen(true); }}
            onEdgeClick={(_, edge) => { setSelectedEdgeId(edge.id); setSelectedNodeId(null); }}
            onPaneClick={() => { setSelectedNodeId(null); setSelectedEdgeId(null); }}
            isValidConnection={(connection) => !connectionError(connection)}
            deleteKeyCode={null}
            fitView
            fitViewOptions={{ padding: 0.22 }}
            minZoom={0.25}
            maxZoom={1.8}
          >
            <Background variant={BackgroundVariant.Dots} gap={20} size={1} color="#cbd3dd" />
            <Controls showInteractive={false} />
            <MiniMap pannable zoomable nodeStrokeWidth={3} nodeColor={(node) => node.selected ? "#2f6fed" : "#aeb9c6"} maskColor="rgba(241, 244, 247, .74)" />
            <Panel position="top-left" className="canvas-panel-buttons"><button className="icon-button canvas-panel-toggle" type="button" onClick={() => setPaletteOpen((value) => !value)} aria-label={paletteOpen ? "Hide node palette" : "Show node palette"} title={paletteOpen ? "Hide node palette" : "Show node palette"}><PanelLeftClose size={17} aria-hidden="true" /></button></Panel>
            <Panel position="top-right" className="canvas-panel-buttons"><button className="icon-button canvas-panel-toggle" type="button" onClick={() => setInspectorOpen((value) => !value)} aria-label={inspectorOpen ? "Hide inspector" : "Show inspector"} title={inspectorOpen ? "Hide inspector" : "Show inspector"}><PanelRightClose size={17} aria-hidden="true" /></button></Panel>
            {selectedEdgeId ? <Panel position="bottom-center" className="edge-action-panel"><span>Connection selected</span><button className="button button--danger-quiet" type="button" onClick={() => deleteEdgeMutation.mutate(selectedEdgeId)} disabled={deleteEdgeMutation.isPending}><Trash2 size={15} aria-hidden="true" />Delete connection</button></Panel> : null}
          </ReactFlow>
        </section>
        {inspectorOpen ? <NodeInspector key={selectedNode?.id ?? "workflow"} node={selectedNode} workflow={{ name: workflow.name, description: workflow.description, nodeCount: nodes.length, edgeCount: edges.length }} models={modelsQuery.data ?? []} knowledgeBases={knowledgeQuery.data ?? []} tools={toolsQuery.data ?? []} connectedRuleIds={connectedRuleIds} saving={updateNodeMutation.isPending || deleteNodeMutation.isPending} onSave={(input) => selectedNode && updateNodeMutation.mutate({ id: selectedNode.id, type: selectedNode.data.nodeType, input })} onDelete={() => selectedNode && deleteNodeMutation.mutate(selectedNode.id)} /> : null}
      </div>
      <RunPanel open={runOpen} workflowId={workflowId} onClose={() => setRunOpen(false)} onComplete={() => void queryClient.invalidateQueries({ queryKey: ["sessions", workflowId] })} />
      <HistoryPanel open={historyOpen} workflowId={workflowId} onClose={() => setHistoryOpen(false)} />
    </main>
  );
}

export default function WorkflowEditorPage() {
  const { workflowId = "" } = useParams();
  return <ReactFlowProvider><EditorWorkspace workflowId={workflowId} /></ReactFlowProvider>;
}
