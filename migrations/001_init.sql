CREATE TABLE IF NOT EXISTS logs (
    id SERIAL PRIMARY KEY,
    file_name VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'processing',
    node_count INTEGER DEFAULT 0,
    port_count INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS nodes (
    id SERIAL PRIMARY KEY,
    log_id INTEGER REFERENCES logs(id) ON DELETE CASCADE,
    node_id VARCHAR(255) NOT NULL,
    node_type VARCHAR(50) NOT NULL,
    hostname VARCHAR(255),
    model VARCHAR(255),
    UNIQUE(log_id, node_id)
);

CREATE TABLE IF NOT EXISTS ports (
    id SERIAL PRIMARY KEY,
    node_id INTEGER REFERENCES nodes(id) ON DELETE CASCADE,
    port_name VARCHAR(255) NOT NULL,
    port_type VARCHAR(50),
    speed VARCHAR(50),
    status VARCHAR(20) DEFAULT 'down'
);

CREATE TABLE IF NOT EXISTS nodes_info (
    id SERIAL PRIMARY KEY,
    node_id INTEGER REFERENCES nodes(id) ON DELETE CASCADE,
    cpu VARCHAR(255),
    memory VARCHAR(255),
    disk VARCHAR(255),
    os VARCHAR(255)
);

CREATE INDEX IF NOT EXISTS idx_nodes_log_id ON nodes(log_id);
CREATE INDEX IF NOT EXISTS idx_ports_node_id ON ports(node_id);
CREATE INDEX IF NOT EXISTS idx_nodes_info_node_id ON nodes_info(node_id);
