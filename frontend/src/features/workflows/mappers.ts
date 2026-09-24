import type {
  AgentHubFlowEdge,
  AgentHubFlowNode,
  BackendWorkflowEdge,
  BackendWorkflowNode,
} from "../../types/domain";

export function mapBackendNodeToFlowNode(node: BackendWorkflowNode): AgentHubFlowNode {
  return {
    id: node.id,
    type: "agentHub",
    position: { x: node.position_x, y: node.position_y },
    data: {
      backendId: node.id,
      name: node.name,
      nodeType: node.type,
      config: node.config,
    },
  };
}

export function mapBackendEdgeToFlowEdge(edge: BackendWorkflowEdge): AgentHubFlowEdge {
  return {
    id: edge.id,
    source: edge.source_node_id,
    target: edge.target_node_id,
    sourceHandle: edge.config.branch_rule_id,
    data: { backendId: edge.id },
    type: "smoothstep",
    animated: false,
  };
}

