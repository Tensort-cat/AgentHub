import type { ComponentType } from "react";
import {
  Bot,
  Braces,
  GitBranch,
  LogIn,
  LogOut,
  MessageSquareText,
  Search,
  Wrench,
} from "lucide-react";
import { WorkflowNodeType } from "../../types/domain";

export interface NodeDefinition {
  type: WorkflowNodeType;
  label: string;
  description: string;
  category: "Input & control" | "AI" | "Knowledge" | "Tools" | "Output";
  icon: ComponentType<{ size?: number; className?: string; "aria-hidden"?: boolean | "true" | "false" }>;
  unavailable?: boolean;
  experimental?: boolean;
}

export const nodeDefinitions: NodeDefinition[] = [
  { type: WorkflowNodeType.Start, label: "Start", description: "Workflow input", category: "Input & control", icon: LogIn },
  { type: WorkflowNodeType.Branch, label: "Branch", description: "Route by ordered rules", category: "Input & control", icon: GitBranch },
  { type: WorkflowNodeType.ChatModel, label: "Chat model", description: "Generate with an LLM", category: "AI", icon: MessageSquareText },
  { type: WorkflowNodeType.ChatTemplate, label: "Chat template", description: "Shape the input prompt", category: "AI", icon: Braces },
  { type: WorkflowNodeType.Agent, label: "Agent", description: "Tool-calling reasoning loop", category: "AI", icon: Bot, unavailable: true },
  { type: WorkflowNodeType.Retriever, label: "Retriever", description: "Search a knowledge base", category: "Knowledge", icon: Search, experimental: true },
  { type: WorkflowNodeType.Tool, label: "Tools", description: "Expose subscribed tools", category: "Tools", icon: Wrench },
  { type: WorkflowNodeType.End, label: "End", description: "Workflow output", category: "Output", icon: LogOut },
];

export const nodeDefinitionByType = new Map(nodeDefinitions.map((definition) => [definition.type, definition]));

