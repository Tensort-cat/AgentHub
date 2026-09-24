import { Info } from "lucide-react";
import type { DragEvent } from "react";
import { nodeDefinitions } from "./node-meta";
import { WorkflowNodeType } from "../../types/domain";

const categories = ["Input & control", "AI", "Knowledge", "Tools", "Output"] as const;

export function NodePalette({ nodeTypes }: { nodeTypes: Set<WorkflowNodeType> }) {
  const drag = (event: DragEvent<HTMLButtonElement>, type: WorkflowNodeType) => {
    event.dataTransfer.setData("application/agenthub-node", String(type));
    event.dataTransfer.effectAllowed = "move";
  };

  return (
    <aside className="node-palette" aria-label="Node palette">
      <header><h2>Add nodes</h2><p>Drag a node onto the canvas.</p></header>
      <div className="node-palette__scroll">
        {categories.map((category) => (
          <section className="palette-group" key={category}>
            <h3>{category}</h3>
            {nodeDefinitions.filter((item) => item.category === category).map((item) => {
              const Icon = item.icon;
              const singletonExists = (item.type === WorkflowNodeType.Start || item.type === WorkflowNodeType.End) && nodeTypes.has(item.type);
              const disabled = item.unavailable || singletonExists;
              return (
                <button key={item.type} className="palette-item" type="button" draggable={!disabled} disabled={disabled} onDragStart={(event) => drag(event, item.type)} title={item.unavailable ? "Agent execution is not implemented by the backend" : singletonExists ? `Only one ${item.label} node is allowed` : `Drag ${item.label} to canvas`}>
                  <span className={`palette-item__icon palette-item__icon--${item.type}`}><Icon size={16} aria-hidden="true" /></span>
                  <span><strong>{item.label}</strong><small>{item.description}</small></span>
                  {item.unavailable ? <em>Coming soon</em> : item.experimental ? <em>Backend fix needed</em> : null}
                </button>
              );
            })}
          </section>
        ))}
      </div>
      <div className="palette-footnote"><Info size={14} aria-hidden="true" /><span>Node creation is saved immediately.</span></div>
    </aside>
  );
}

