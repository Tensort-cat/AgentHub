import type { Edge, Node } from "@xyflow/react";

export type EntityId = string;
export type JsonId = string | number | bigint;

export enum ModelType {
  Chat = 1,
  Embedding = 2,
}

export enum WorkflowStatus {
  Draft = 1,
  Active = 2,
}

export enum WorkflowNodeType {
  ChatModel = 1,
  ChatTemplate = 2,
  Branch = 3,
  Tool = 4,
  Retriever = 5,
  Agent = 6,
  Start = 7,
  End = 8,
}

export enum DocumentStatus {
  Pending = 1,
  Parsing = 2,
  Indexing = 3,
  Success = 4,
  Error = 5,
}

export enum MessageType {
  User = 1,
  Assistant = 2,
  Tool = 3,
  System = 4,
}

export type BranchOperator =
  | "equals"
  | "contains"
  | "starts_with"
  | "ends_with"
  | "regex";

export interface EmptyConfig {
  readonly [key: string]: never;
}

export interface ChatModelConfig {
  model_id: EntityId;
  temperature: number;
  max_tokens: number;
  system_prompt: string;
  tool_ids: EntityId[];
}

export interface ChatTemplateConfig {
  system_prompt: string;
  user_prompt: string;
}

export interface BranchRule {
  id: string;
  label: string;
  operator: BranchOperator;
  value: string;
  case_sensitive: boolean;
}

export interface BranchConfig {
  version: 1;
  trim_space: boolean;
  rules: BranchRule[];
}

export interface ToolConfig {
  tool_ids: EntityId[];
}

export interface RetrieverConfig {
  kb_id: EntityId;
  top_k: number;
}

export interface AgentConfig {
  tool_calling_model: EntityId;
}

export type WorkflowNodeConfig =
  | EmptyConfig
  | ChatModelConfig
  | ChatTemplateConfig
  | BranchConfig
  | ToolConfig
  | RetrieverConfig
  | AgentConfig;

export interface User {
  id: EntityId;
  name: string;
  email: string;
  avatar: string;
}

export interface ModelConnection {
  id: EntityId;
  name: string;
  provider: number;
  base_url: string;
  type: ModelType;
  created_at: string;
}

export interface WorkflowSummary {
  id: EntityId;
  name: string;
  description: string;
  status: WorkflowStatus;
  created_at: string;
}

export interface BackendWorkflowNode {
  id: EntityId;
  name: string;
  type: WorkflowNodeType;
  position_x: number;
  position_y: number;
  config: WorkflowNodeConfig;
}

export interface BackendWorkflowEdge {
  id: EntityId;
  source_node_id: EntityId;
  target_node_id: EntityId;
  config: { branch_rule_id?: string };
}

export interface WorkflowDetail {
  id: EntityId;
  name: string;
  description: string;
  nodes: BackendWorkflowNode[];
  edges: BackendWorkflowEdge[];
}

export interface KnowledgeBaseSummary {
  id: EntityId;
  name: string;
  description: string;
  created_at: string;
}

export interface DocumentMeta {
  id: EntityId;
  name: string;
  size: number;
  status: DocumentStatus;
  type: ".md" | ".txt";
  created_at: string;
}

export interface KnowledgeBaseDetail extends KnowledgeBaseSummary {
  embedder_id: EntityId;
  docs: DocumentMeta[];
}

export interface ToolSummary {
  id: EntityId;
  name: string;
  description: string;
}

export interface ToolDetail extends ToolSummary {
  status: number;
  avatar: string;
  created_at: string;
}

export interface WorkflowSession {
  id: EntityId;
  title: string;
  created_at: string;
}

export interface SessionMessage {
  id: EntityId;
  content: string;
  type: MessageType;
  created_at: string;
}

export type RunStatus = "idle" | "queued" | "running" | "success" | "failed";

export interface WorkflowRunState {
  task_id: string;
  status: Exclude<RunStatus, "idle">;
  session_id?: EntityId;
  result?: string;
  error?: string;
  updated_at?: string;
}

export interface AgentHubNodeData extends Record<string, unknown> {
  backendId: EntityId;
  name: string;
  nodeType: WorkflowNodeType;
  config: WorkflowNodeConfig;
}

export type AgentHubFlowNode = Node<AgentHubNodeData, "agentHub">;
export type AgentHubFlowEdge = Edge<{ backendId: EntityId }>;

