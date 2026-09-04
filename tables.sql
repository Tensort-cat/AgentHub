CREATE TABLE users (
    id BIGINT PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    email VARCHAR(100) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    avatar VARCHAR(255),
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    deleted_at DATETIME(3) DEFAULT NULL
);

CREATE TABLE workflows (
    id BIGINT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    description TEXT,
    name VARCHAR(100) NOT NULL,
    status TINYINT NOT NULL DEFAULT 1,
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    deleted_at DATETIME(3),
    INDEX idx_user(user_id)
);

CREATE TABLE sessions (
    id BIGINT PRIMARY KEY,
    workflow_id BIGINT NOT NULL,
    title VARCHAR(100),
    created_at DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3),
    deleted_at DATETIME(3),
    INDEX idx_workflow(workflow_id)
);

CREATE TABLE messages (
    id BIGINT PRIMARY KEY,
    session_id BIGINT NOT NULL,
    type TINYINT NOT NULL,
    content LONGTEXT NOT NULL,
    created_at DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3),
    deleted_at DATETIME(3),
    INDEX idx_session(session_id)
);

CREATE TABLE tools (
    id BIGINT NOT NULL,

    name VARCHAR(100) NOT NULL,
    description TEXT,
    avatar VARCHAR(255),

    type TINYINT NOT NULL DEFAULT 1,

    mcp_server_url VARCHAR(500),
    mcp_tool_name VARCHAR(100),

    status TINYINT NOT NULL DEFAULT 1,

    created_at DATETIME(3) NOT NULL,
    updated_at DATETIME(3) NOT NULL,

    PRIMARY KEY (id),

    UNIQUE KEY uk_mcp_tool (
        mcp_server_url,
        mcp_tool_name
    ),

    KEY idx_status (status)
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci;

CREATE TABLE models (
    id BIGINT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    name VARCHAR(100) NOT NULL,
    base_url VARCHAR(255) NOT NULL,
    api_key VARCHAR(255) NOT NULL,
    type TINYINT NOT NULL,
    created_at DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3),
    deleted_at DATETIME(3),
    INDEX idx_user(user_id)
);

CREATE TABLE workflow_nodes (
    id BIGINT PRIMARY KEY,
    workflow_id BIGINT NOT NULL,
    name VARCHAR(100) NOT NULL,
    type TINYINT NOT NULL,
    position_x INT NOT NULL,
    position_y INT NOT NULL,
    config JSON NOT NULL,
    created_at DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3),
    deleted_at DATETIME(3),
    INDEX idx_workflow(workflow_id)
);

CREATE TABLE workflow_edges (
    id BIGINT PRIMARY KEY,
    workflow_id BIGINT NOT NULL,
    source_node_id BIGINT NOT NULL,
    target_node_id BIGINT NOT NULL,
    config JSON,
    created_at DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3),
    deleted_at DATETIME(3),
    INDEX idx_workflow(workflow_id),
    INDEX idx_source(source_node_id),
    INDEX idx_target(target_node_id)
);

CREATE TABLE knowledge_base (
    id BIGINT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    description TEXT,
    name VARCHAR(100) NOT NULL,
    embedder_id BIGINT NOT NULL,
    created_at DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3),
    deleted_at DATETIME(3),
    INDEX idx_user(user_id),
    INDEX idx_embedder(embedder_id)
);

CREATE TABLE documents (
    id BIGINT PRIMARY KEY,
    knowledge_base_id BIGINT NOT NULL,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(20) NOT NULL,
    size BIGINT NOT NULL,
    status TINYINT NOT NULL DEFAULT 0,
    chunk_num INT DEFAULT 0,
    created_at DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3),
    deleted_at DATETIME(3),
    INDEX idx_kb(knowledge_base_id)
);

CREATE TABLE user_tools (
    id BIGINT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    tool_id BIGINT NOT NULL,
    status TINYINT NOT NULL DEFAULT 1,
    created_at DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3),
    deleted_at DATETIME(3),
    UNIQUE KEY uk_user_tool(user_id, tool_id),
    INDEX idx_user(user_id),
    INDEX idx_tool(tool_id)
);
