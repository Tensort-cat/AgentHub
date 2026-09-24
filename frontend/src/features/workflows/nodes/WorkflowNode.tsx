import { memo, useEffect } from "react";
import { Handle, Position, useUpdateNodeInternals, type NodeProps } from "@xyflow/react";
import { AlertCircle } from "lucide-react";
import type {
  AgentHubFlowNode,
  BranchConfig,
  ChatModelConfig,
  ChatTemplateConfig,
  RetrieverConfig,
  ToolConfig,
} from "../../../types/domain";
import { WorkflowNodeType } from "../../../types/domain";
import { isConfigValid } from "../config";
import { nodeDefinitionByType } from "../node-meta";

function nodeSummary(data: AgentHubFlowNode["data"]): string {
  switch (data.nodeType) {
    case WorkflowNodeType.ChatModel: {
      const config = data.config as ChatModelConfig;
      return config.model_id === "0" ? "Choose a model" : `Temp ${config.temperature} · ${config.max_tokens} tokens`;
    }
    case WorkflowNodeType.ChatTemplate: {
      const config = data.config as ChatTemplateConfig;
      return config.user_prompt ? "Prompt template ready" : "User prompt required";
    }
    case WorkflowNodeType.Branch:
      return `${(data.config as BranchConfig).rules.length} ordered ${(data.config as BranchConfig).rules.length === 1 ? "rule" : "rules"}`;
    case WorkflowNodeType.Tool:
      return `${(data.config as ToolConfig).tool_ids.length} ${(data.config as ToolConfig).tool_ids.length === 1 ? "tool" : "tools"} selected`;
    case WorkflowNodeType.Retriever: {
      const config = data.config as RetrieverConfig;
      return config.kb_id === "0" ? "Choose knowledge" : `Top ${config.top_k} matches`;
    }
    case WorkflowNodeType.Agent:
      return "Backend support pending";
    case WorkflowNodeType.Start:
      return "Workflow input";
    case WorkflowNodeType.End:
      return "Workflow output";
  }
}

function WorkflowNodeComponent({ id, data, selected }: NodeProps<AgentHubFlowNode>) {
  const definition = nodeDefinitionByType.get(data.nodeType)!;
  const Icon = definition.icon;
  const valid = isConfigValid(data.nodeType, data.config);
  const branchConfig = data.nodeType === WorkflowNodeType.Branch ? data.config as BranchConfig : null;
  const updateNodeInternals = useUpdateNodeInternals();
  const ruleKey = branchConfig?.rules.map((rule) => rule.id).join("|") ?? "";

  useEffect(() => {
    if (branchConfig) updateNodeInternals(id);
  }, [branchConfig, id, ruleKey, updateNodeInternals]);

  return (
    <div className={`workflow-node workflow-node--${data.nodeType} ${selected ? "workflow-node--selected" : ""} ${valid ? "" : "workflow-node--invalid"}`}>
      {data.nodeType !== WorkflowNodeType.Start ? <Handle type="target" position={Position.Left} /> : null}
      <header><span className="workflow-node__icon"><Icon size={15} aria-hidden="true" /></span><strong>{data.name}</strong>{!valid ? <AlertCircle className="workflow-node__warning" size={15} aria-label="Configuration required" /> : null}</header>
      <div className="workflow-node__summary">{nodeSummary(data)}</div>
      {branchConfig ? (
        <div className="branch-outputs">
          {branchConfig.rules.map((rule) => (
            <div className="branch-output" key={rule.id}>
              <span>{rule.label || rule.id}</span>
              <Handle className="branch-handle" type="source" position={Position.Right} id={rule.id} />
            </div>
          ))}
          <div className="branch-output branch-output--default"><span>Default</span><Handle className="branch-handle" type="source" position={Position.Right} id="$default" /></div>
        </div>
      ) : data.nodeType !== WorkflowNodeType.End ? <Handle type="source" position={Position.Right} /> : null}
    </div>
  );
}

export const WorkflowNode = memo(WorkflowNodeComponent);
