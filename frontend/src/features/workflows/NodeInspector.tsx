import { useEffect, useState, type FormEvent } from "react";
import { Link } from "react-router-dom";
import { ChevronDown, ChevronUp, Plus, Trash2 } from "lucide-react";
import { ConfirmDialog } from "../../components/ui/Dialog";
import type {
  AgentHubFlowNode,
  BranchConfig,
  BranchOperator,
  ChatModelConfig,
  ChatTemplateConfig,
  KnowledgeBaseSummary,
  ModelConnection,
  RetrieverConfig,
  ToolConfig,
  ToolSummary,
  WorkflowNodeConfig,
} from "../../types/domain";
import { ModelType, WorkflowNodeType } from "../../types/domain";
import { createRuleId, isConfigValid } from "./config";
import { nodeDefinitionByType } from "./node-meta";

const operators: { value: BranchOperator; label: string }[] = [
  { value: "equals", label: "Equals" },
  { value: "contains", label: "Contains" },
  { value: "starts_with", label: "Starts with" },
  { value: "ends_with", label: "Ends with" },
  { value: "regex", label: "Regular expression" },
];

function toggleId(ids: string[], id: string): string[] {
  return ids.includes(id) ? ids.filter((item) => item !== id) : [...ids, id];
}

function ToolChecklist({ ids, tools, onChange }: { ids: string[]; tools: ToolSummary[]; onChange: (ids: string[]) => void }) {
  if (tools.length === 0) return <div className="inspector-empty"><p>No tools are available.</p><Link to="/tools">Open Tool center</Link></div>;
  return (
    <div className="check-list">
      {tools.map((tool) => <label key={tool.id}><input type="checkbox" checked={ids.includes(tool.id)} onChange={() => onChange(toggleId(ids, tool.id))} /><span><strong>{tool.name}</strong><small>{tool.description}</small></span></label>)}
    </div>
  );
}

function BranchFields({ config, connectedRuleIds, onChange }: { config: BranchConfig; connectedRuleIds: Set<string>; onChange: (config: BranchConfig) => void }) {
  const updateRule = (index: number, field: keyof BranchConfig["rules"][number], value: string | boolean) => {
    onChange({ ...config, rules: config.rules.map((rule, ruleIndex) => ruleIndex === index ? { ...rule, [field]: value } : rule) });
  };
  const move = (index: number, direction: -1 | 1) => {
    const destination = index + direction;
    if (destination < 0 || destination >= config.rules.length) return;
    const rules = [...config.rules];
    [rules[index], rules[destination]] = [rules[destination], rules[index]];
    onChange({ ...config, rules });
  };
  const add = () => onChange({ ...config, rules: [...config.rules, { id: createRuleId(), label: `Rule ${config.rules.length + 1}`, operator: "equals", value: "", case_sensitive: false }] });
  const remove = (index: number) => onChange({ ...config, rules: config.rules.filter((_, ruleIndex) => ruleIndex !== index) });

  return (
    <>
      <label className="switch-field"><input type="checkbox" checked={config.trim_space} onChange={(event) => onChange({ ...config, trim_space: event.target.checked })} /><span><strong>Trim whitespace</strong><small>Trim input and string values before matching.</small></span></label>
      <div className="branch-rule-list">
        {config.rules.map((rule, index) => (
          <fieldset className="branch-rule" key={rule.id}>
            <legend>Rule {index + 1}</legend>
            <div className="branch-rule__actions"><button className="icon-button icon-button--quiet" type="button" onClick={() => move(index, -1)} disabled={index === 0} aria-label={`Move rule ${index + 1} up`}><ChevronUp size={15} aria-hidden="true" /></button><button className="icon-button icon-button--quiet" type="button" onClick={() => move(index, 1)} disabled={index === config.rules.length - 1} aria-label={`Move rule ${index + 1} down`}><ChevronDown size={15} aria-hidden="true" /></button><button className="icon-button icon-button--quiet text-danger" type="button" onClick={() => remove(index)} disabled={config.rules.length === 1 || connectedRuleIds.has(rule.id)} aria-label={`Delete rule ${index + 1}`} title={connectedRuleIds.has(rule.id) ? "Delete the connected edge first" : "Delete rule"}><Trash2 size={15} aria-hidden="true" /></button></div>
            <label className="field"><span>Label</span><input value={rule.label} onChange={(event) => updateRule(index, "label", event.target.value)} /></label>
            <label className="field"><span>Operator</span><select value={rule.operator} onChange={(event) => updateRule(index, "operator", event.target.value as BranchOperator)}>{operators.map((operator) => <option key={operator.value} value={operator.value}>{operator.label}</option>)}</select></label>
            <label className="field"><span>{rule.operator === "regex" ? "Pattern" : "Value"}</span><input className={rule.operator === "regex" ? "mono" : ""} value={rule.value} onChange={(event) => updateRule(index, "value", event.target.value)} /><small>An empty value can match every input.</small></label>
            <label className="switch-field switch-field--compact"><input type="checkbox" checked={rule.case_sensitive} onChange={(event) => updateRule(index, "case_sensitive", event.target.checked)} /><span>Case sensitive</span></label>
            <code className="rule-id">Handle: {rule.id}</code>
          </fieldset>
        ))}
      </div>
      <button className="button button--secondary button--full" type="button" onClick={add}><Plus size={15} aria-hidden="true" />Add rule</button>
      <div className="default-outlet"><span>Default outlet</span><code>$default</code></div>
    </>
  );
}

function ConfigFields({
  type,
  config,
  models,
  knowledgeBases,
  tools,
  connectedRuleIds,
  onChange,
}: {
  type: WorkflowNodeType;
  config: WorkflowNodeConfig;
  models: ModelConnection[];
  knowledgeBases: KnowledgeBaseSummary[];
  tools: ToolSummary[];
  connectedRuleIds: Set<string>;
  onChange: (config: WorkflowNodeConfig) => void;
}) {
  if (type === WorkflowNodeType.ChatModel) {
    const value = config as ChatModelConfig;
    const chatModels = models.filter((model) => model.type === ModelType.Chat);
    return <><label className="field"><span>Model <b aria-hidden="true">*</b></span><select value={value.model_id} onChange={(event) => onChange({ ...value, model_id: event.target.value })}><option value="0">Select a chat model</option>{chatModels.map((model) => <option key={model.id} value={model.id}>{model.name}</option>)}</select>{chatModels.length === 0 ? <small>No chat models available. <Link to="/models">Create a model</Link>.</small> : null}</label><label className="field"><span>System prompt</span><textarea className="nodrag" value={value.system_prompt} onChange={(event) => onChange({ ...value, system_prompt: event.target.value })} rows={6} placeholder="Define how the model should respond." /></label><div className="field-pair"><label className="field"><span>Temperature</span><input type="number" min="0" max="2" step="0.1" value={value.temperature} onChange={(event) => onChange({ ...value, temperature: Number(event.target.value) })} /></label><label className="field"><span>Max tokens</span><input type="number" min="1" step="1" value={value.max_tokens} onChange={(event) => onChange({ ...value, max_tokens: Number(event.target.value) })} /></label></div><div className="field"><span>Tools</span><small className="field-help">Subscription state is not readable from the current API. Select only tools you have subscribed to.</small><ToolChecklist ids={value.tool_ids} tools={tools} onChange={(tool_ids) => onChange({ ...value, tool_ids })} /></div></>;
  }
  if (type === WorkflowNodeType.ChatTemplate) {
    const value = config as ChatTemplateConfig;
    return <><label className="field"><span>System prompt</span><textarea className="nodrag" value={value.system_prompt} onChange={(event) => onChange({ ...value, system_prompt: event.target.value })} rows={5} /></label><label className="field"><span>User prompt <b aria-hidden="true">*</b></span><textarea className="nodrag mono" value={value.user_prompt} onChange={(event) => onChange({ ...value, user_prompt: event.target.value })} rows={7} required /><small>Use <code>{"{input}"}</code> to insert the upstream value.</small></label></>;
  }
  if (type === WorkflowNodeType.Branch) return <BranchFields config={config as BranchConfig} connectedRuleIds={connectedRuleIds} onChange={onChange} />;
  if (type === WorkflowNodeType.Tool) {
    const value = config as ToolConfig;
    return <div className="field"><span>Tools</span><small className="field-help">The backend validates that every selected tool is subscribed.</small><ToolChecklist ids={value.tool_ids} tools={tools} onChange={(tool_ids) => onChange({ tool_ids })} /></div>;
  }
  if (type === WorkflowNodeType.Retriever) {
    const value = config as RetrieverConfig;
    return <><div className="form-warning">Retriever configuration is supported, but the current backend node whitelist prevents workflows containing it from running.</div><label className="field"><span>Knowledge base <b aria-hidden="true">*</b></span><select value={value.kb_id} onChange={(event) => onChange({ ...value, kb_id: event.target.value })}><option value="0">Select knowledge</option>{knowledgeBases.map((base) => <option key={base.id} value={base.id}>{base.name}</option>)}</select>{knowledgeBases.length === 0 ? <small>No knowledge bases available. <Link to="/knowledge">Create knowledge</Link>.</small> : null}</label><label className="field"><span>Top K</span><input type="number" min="1" step="1" value={value.top_k} onChange={(event) => onChange({ ...value, top_k: Number(event.target.value) })} /></label></>;
  }
  if (type === WorkflowNodeType.Agent) return <div className="form-warning">Agent execution is not implemented by the backend. This node cannot be saved as a runnable configuration.</div>;
  return <p className="inspector-copy">This node has no type-specific configuration.</p>;
}

export function NodeInspector({
  node,
  workflow,
  models,
  knowledgeBases,
  tools,
  connectedRuleIds,
  saving,
  onSave,
  onDelete,
}: {
  node: AgentHubFlowNode | null;
  workflow: { name: string; description: string; nodeCount: number; edgeCount: number };
  models: ModelConnection[];
  knowledgeBases: KnowledgeBaseSummary[];
  tools: ToolSummary[];
  connectedRuleIds: Set<string>;
  saving: boolean;
  onSave: (input: { name: string; config: WorkflowNodeConfig }) => void;
  onDelete: () => void;
}) {
  const [name, setName] = useState(() => node?.data.name ?? "");
  const [config, setConfig] = useState<WorkflowNodeConfig>(() => structuredClone(node?.data.config ?? {}));
  const [error, setError] = useState("");
  const [confirmDelete, setConfirmDelete] = useState(false);

  useEffect(() => {
    if (!node) return;
    setName(node.data.name);
    setConfig(structuredClone(node.data.config));
    setError("");
  }, [node?.data.config, node?.data.name, node?.id]);

  if (!node) {
    return (
      <aside className="node-inspector node-inspector--empty">
        <header><h2>Workflow</h2><p>Select a node to configure it.</p></header>
        <div className="workflow-overview"><strong>{workflow.name}</strong><p>{workflow.description || "No description"}</p><dl><div><dt>Nodes</dt><dd>{workflow.nodeCount}</dd></div><div><dt>Connections</dt><dd>{workflow.edgeCount}</dd></div></dl></div>
      </aside>
    );
  }

  const definition = nodeDefinitionByType.get(node.data.nodeType)!;
  const submit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!name.trim()) { setError("Node name is required."); return; }
    if (!isConfigValid(node.data.nodeType, config)) { setError("Complete the required configuration before saving."); return; }
    setError("");
    onSave({ name: name.trim(), config });
  };

  return (
    <aside className="node-inspector">
      <header><div><h2>Node settings</h2><p>{definition.label}</p></div><span className={`inspector-type inspector-type--${node.data.nodeType}`}>{definition.label}</span></header>
      <form className="inspector-form" onSubmit={submit}>
        {error ? <div className="form-alert" role="alert">{error}</div> : null}
        <section className="inspector-section"><h3>General</h3><label className="field"><span>Node name <b aria-hidden="true">*</b></span><input className="nodrag" value={name} onChange={(event) => setName(event.target.value)} required /></label></section>
        <section className="inspector-section"><h3>Configuration</h3><ConfigFields type={node.data.nodeType} config={config} models={models} knowledgeBases={knowledgeBases} tools={tools} connectedRuleIds={connectedRuleIds} onChange={setConfig} /></section>
        <div className="inspector-save"><button className="button button--primary button--full" type="submit" disabled={saving || node.data.nodeType === WorkflowNodeType.Agent}>{saving ? "Saving…" : "Save changes"}</button></div>
      </form>
      <section className="inspector-danger"><h3>Danger zone</h3><button className="button button--danger-quiet button--full" type="button" onClick={() => setConfirmDelete(true)}><Trash2 size={16} aria-hidden="true" />Delete node</button></section>
      <ConfirmDialog open={confirmDelete} title="Delete node?" description={`“${node.data.name}” and all of its connections will be removed.`} confirmLabel="Delete node" busy={saving} onClose={() => setConfirmDelete(false)} onConfirm={() => { setConfirmDelete(false); onDelete(); }} />
    </aside>
  );
}
