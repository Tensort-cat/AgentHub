import { backendId, normalizeId } from "../../lib/json";
import type {
  AgentConfig,
  BranchConfig,
  ChatModelConfig,
  ChatTemplateConfig,
  EntityId,
  JsonId,
  RetrieverConfig,
  ToolConfig,
  WorkflowNodeConfig,
} from "../../types/domain";
import { WorkflowNodeType } from "../../types/domain";

type UnknownObject = Record<string, unknown>;

function objectValue(value: unknown): UnknownObject {
  return typeof value === "object" && value !== null ? (value as UnknownObject) : {};
}

function idValue(value: unknown): EntityId {
  return value === undefined || value === null ? "0" : normalizeId(value as JsonId);
}

function idArray(value: unknown): EntityId[] {
  return Array.isArray(value) ? value.map((item) => idValue(item)) : [];
}

export function normalizeNodeConfig(
  type: WorkflowNodeType,
  value: unknown,
): WorkflowNodeConfig {
  const config = objectValue(value);
  switch (type) {
    case WorkflowNodeType.ChatModel:
      return {
        model_id: idValue(config.model_id),
        temperature: Number(config.temperature ?? 0.7),
        max_tokens: Number(config.max_tokens ?? 1024),
        system_prompt: String(config.system_prompt ?? ""),
        tool_ids: idArray(config.tool_ids),
      } satisfies ChatModelConfig;
    case WorkflowNodeType.ChatTemplate:
      return {
        system_prompt: String(config.system_prompt ?? ""),
        user_prompt: String(config.user_prompt ?? "{input}"),
      } satisfies ChatTemplateConfig;
    case WorkflowNodeType.Branch: {
      const rawRules = Array.isArray(config.rules) ? config.rules : [];
      return {
        version: 1,
        trim_space: Boolean(config.trim_space),
        rules: rawRules.map((item) => {
          const rule = objectValue(item);
          return {
            id: String(rule.id ?? ""),
            label: String(rule.label ?? ""),
            operator: (rule.operator ?? "equals") as BranchConfig["rules"][number]["operator"],
            value: String(rule.value ?? ""),
            case_sensitive: Boolean(rule.case_sensitive),
          };
        }),
      } satisfies BranchConfig;
    }
    case WorkflowNodeType.Tool:
      return { tool_ids: idArray(config.tool_ids) } satisfies ToolConfig;
    case WorkflowNodeType.Retriever:
      return {
        kb_id: idValue(config.kb_id),
        top_k: Number(config.top_k ?? 5),
      } satisfies RetrieverConfig;
    case WorkflowNodeType.Agent:
      return {
        tool_calling_model: idValue(config.tool_calling_model),
      } satisfies AgentConfig;
    case WorkflowNodeType.Start:
    case WorkflowNodeType.End:
      return {};
  }
}

export function serializeNodeConfig(
  type: WorkflowNodeType,
  value: WorkflowNodeConfig,
): unknown {
  switch (type) {
    case WorkflowNodeType.ChatModel: {
      const config = value as ChatModelConfig;
      return {
        ...config,
        model_id: backendId(config.model_id),
        tool_ids: config.tool_ids.map(backendId),
      };
    }
    case WorkflowNodeType.Retriever: {
      const config = value as RetrieverConfig;
      return { ...config, kb_id: backendId(config.kb_id) };
    }
    case WorkflowNodeType.Tool: {
      const config = value as ToolConfig;
      return { tool_ids: config.tool_ids.map(backendId) };
    }
    case WorkflowNodeType.Agent: {
      const config = value as AgentConfig;
      return { tool_calling_model: backendId(config.tool_calling_model) };
    }
    default:
      return value;
  }
}

export function createRuleId(): string {
  return `rule_${crypto.randomUUID().replaceAll("-", "").slice(0, 10)}`;
}

export function defaultNodeConfig(
  type: WorkflowNodeType,
  defaults: { chatModelId?: string; knowledgeBaseId?: string },
): WorkflowNodeConfig {
  switch (type) {
    case WorkflowNodeType.ChatModel:
      return {
        model_id: defaults.chatModelId ?? "0",
        temperature: 0.7,
        max_tokens: 1024,
        system_prompt: "",
        tool_ids: [],
      };
    case WorkflowNodeType.ChatTemplate:
      return { system_prompt: "", user_prompt: "{input}" };
    case WorkflowNodeType.Branch:
      return {
        version: 1,
        trim_space: true,
        rules: [
          {
            id: createRuleId(),
            label: "Matched",
            operator: "equals",
            value: "matched",
            case_sensitive: false,
          },
        ],
      };
    case WorkflowNodeType.Tool:
      return { tool_ids: [] };
    case WorkflowNodeType.Retriever:
      return { kb_id: defaults.knowledgeBaseId ?? "0", top_k: 5 };
    case WorkflowNodeType.Agent:
      return { tool_calling_model: defaults.chatModelId ?? "0" };
    case WorkflowNodeType.Start:
    case WorkflowNodeType.End:
      return {};
  }
}

export function isConfigValid(type: WorkflowNodeType, config: WorkflowNodeConfig): boolean {
  switch (type) {
    case WorkflowNodeType.ChatModel: {
      const value = config as ChatModelConfig;
      return value.model_id !== "0" && value.temperature >= 0 && value.temperature <= 2 && value.max_tokens > 0;
    }
    case WorkflowNodeType.ChatTemplate:
      return (config as ChatTemplateConfig).user_prompt.length > 0;
    case WorkflowNodeType.Branch: {
      const rules = (config as BranchConfig).rules;
      const ids = new Set(rules.map((rule) => rule.id));
      return rules.length > 0 && ids.size === rules.length && rules.every((rule) => /^[A-Za-z0-9_-]{1,64}$/.test(rule.id) && rule.id !== "$default");
    }
    case WorkflowNodeType.Retriever: {
      const value = config as RetrieverConfig;
      return value.kb_id !== "0" && value.top_k >= 1;
    }
    case WorkflowNodeType.Agent:
      return false;
    default:
      return true;
  }
}

