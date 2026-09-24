import http from "node:http";

const ok = (data) => JSON.stringify({ code: 0, msg: "ok", data });
const nodeId = "2080000000000000101";
const modelId = "2080000000000000201";
const kbId = "2080000000000000301";
const toolId = "2080000000000000401";

const server = http.createServer((request, response) => {
  const url = new URL(request.url, "http://127.0.0.1:8000");
  response.setHeader("Content-Type", "application/json");
  let body = ok(null);
  if (url.pathname === "/api/v1/auth/login") body = `{"code":0,"msg":"ok","data":{"token":"Bearer qa-token","id":2080000000000000001,"name":"Yunpeng","email":"demo@example.com","avatar":""}}`;
  else if (url.pathname === "/api/v1/users/me") body = `{"code":0,"msg":"ok","data":{"id":2080000000000000001,"name":"Yunpeng","email":"demo@example.com","avatar":""}}`;
  else if (url.pathname === "/api/v1/workflows" && request.method === "GET") body = `{"code":0,"msg":"ok","data":[{"id":2080000000000000002,"name":"Support triage","description":"Classify incoming issues and draft a grounded response.","status":1,"created_at":"2026-09-24T10:00:00Z"},{"id":2080000000000000003,"name":"Release notes","description":"Turn merged changes into a concise changelog.","status":2,"created_at":"2026-09-20T10:00:00Z"}]}`;
  else if (url.pathname === "/api/v1/workflows/2080000000000000002") body = `{"code":0,"msg":"ok","data":{"id":2080000000000000002,"name":"Support triage","description":"Classify incoming issues and draft a grounded response.","nodes":[{"id":${nodeId},"name":"Start","type":7,"position_x":40,"position_y":180,"config":{}},{"id":2080000000000000102,"name":"Prepare prompt","type":2,"position_x":350,"position_y":170,"config":{"system_prompt":"Be concise and factual.","user_prompt":"Triage this request: {input}"}},{"id":2080000000000000103,"name":"Priority route","type":3,"position_x":670,"position_y":145,"config":{"version":1,"trim_space":true,"rules":[{"id":"urgent","label":"Urgent","operator":"contains","value":"urgent","case_sensitive":false},{"id":"billing","label":"Billing","operator":"contains","value":"invoice","case_sensitive":false}]}},{"id":2080000000000000104,"name":"Draft response","type":1,"position_x":1010,"position_y":45,"config":{"model_id":${modelId},"temperature":0.4,"max_tokens":1200,"system_prompt":"Draft a helpful support reply.","tool_ids":[]}},{"id":2080000000000000105,"name":"End","type":8,"position_x":1340,"position_y":180,"config":{}}],"edges":[{"id":2080000000000000111,"source_node_id":${nodeId},"target_node_id":2080000000000000102,"config":{}},{"id":2080000000000000112,"source_node_id":2080000000000000102,"target_node_id":2080000000000000103,"config":{}},{"id":2080000000000000113,"source_node_id":2080000000000000103,"target_node_id":2080000000000000104,"config":{"branch_rule_id":"urgent"}},{"id":2080000000000000114,"source_node_id":2080000000000000103,"target_node_id":2080000000000000105,"config":{"branch_rule_id":"billing"}},{"id":2080000000000000115,"source_node_id":2080000000000000103,"target_node_id":2080000000000000105,"config":{"branch_rule_id":"$default"}},{"id":2080000000000000116,"source_node_id":2080000000000000104,"target_node_id":2080000000000000105,"config":{}}]}}`;
  else if (url.pathname === "/api/v1/models") body = `{"code":0,"msg":"ok","data":[{"id":${modelId},"name":"GPT-4o mini","provider":1,"base_url":"https://api.openai.com/v1","type":1,"created_at":"2026-09-10T10:00:00Z"},{"id":2080000000000000202,"name":"Text embedding","provider":1,"base_url":"https://api.openai.com/v1","type":2,"created_at":"2026-09-10T10:00:00Z"}]}`;
  else if (url.pathname === "/api/v1/kb") body = `{"code":0,"msg":"ok","data":[{"id":${kbId},"name":"Product handbook","description":"Policies, product behavior, and support playbooks.","created_at":"2026-09-12T10:00:00Z"}]}`;
  else if (url.pathname === `/api/v1/kb/${kbId}`) body = `{"code":0,"msg":"ok","data":{"id":${kbId},"name":"Product handbook","description":"Policies, product behavior, and support playbooks.","embedder_id":2080000000000000202,"created_at":"2026-09-12T10:00:00Z","docs":[{"id":2080000000000000311,"name":"support-policy.md","size":18240,"status":4,"type":".md","created_at":"2026-09-18T10:00:00Z"}]}}`;
  else if (url.pathname === "/api/v1/tools") body = `{"code":0,"msg":"ok","data":[{"id":${toolId},"name":"Weather","description":"Look up current conditions by city."},{"id":2080000000000000402,"name":"Calculator","description":"Evaluate numeric expressions safely."}]}`;
  else if (url.pathname === `/api/v1/tools/${toolId}`) body = `{"code":0,"msg":"ok","data":{"id":${toolId},"name":"Weather","description":"Look up current conditions by city.","status":2,"avatar":"","created_at":"2026-09-01T10:00:00Z"}}`;
  else if (url.pathname === "/api/v1/sessions") body = `{"code":0,"msg":"ok","data":[{"id":2080000000000000501,"title":"Urgent billing issue","created_at":"2026-09-24T10:00:00Z"}]}`;
  else if (url.pathname === "/api/v1/sessions/2080000000000000501") body = `{"code":0,"msg":"ok","data":[{"id":2080000000000000511,"content":"Here is the drafted reply.","type":2,"created_at":"2026-09-24T10:00:03Z"},{"id":2080000000000000510,"content":"Urgent billing issue","type":1,"created_at":"2026-09-24T10:00:00Z"}]}`;
  response.end(body);
});

server.listen(8000, "127.0.0.1");
