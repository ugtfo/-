package repository

import (
    "database/sql"
    "fmt"
    "log"
    "os"
    "path/filepath"
    "time"

    _ "github.com/lib/pq"
    "log-parser/internal/model"
)

type PostgresRepo struct {
    db     *sql.DB
    logger *log.Logger
}

func NewPostgresRepo(databaseURL string, logger *log.Logger) (*PostgresRepo, error) {
    db, err := sql.Open("postgres", databaseURL)
    if err != nil {
        return nil, fmt.Errorf("failed to open database: %w", err)
    }

    db.SetMaxOpenConns(25)
    db.SetMaxIdleConns(5)
    db.SetConnMaxLifetime(5 * time.Minute)

    if err := db.Ping(); err != nil {
        return nil, fmt.Errorf("failed to ping database: %w", err)
    }

    return &PostgresRepo{db: db, logger: logger}, nil
}

func (r *PostgresRepo) Close() error {
    return r.db.Close()
}

func (r *PostgresRepo) RunMigrations() error {
    migrationPath := "/migrations/001_init.sql"
    
    // Если запускаем локально, а не в контейнере
    if _, err := os.Stat(migrationPath); os.IsNotExist(err) {
        migrationPath = filepath.Join("..", migrationPath)
        if _, err := os.Stat(migrationPath); os.IsNotExist(err) {
            migrationPath = "migrations/001_init.sql"
        }
    }
    
    migration, err := os.ReadFile(migrationPath)
    if err != nil {
        return fmt.Errorf("failed to read migration file: %w", err)
    }

    _, err = r.db.Exec(string(migration))
    if err != nil {
        return fmt.Errorf("failed to execute migration: %w", err)
    }

    r.logger.Println("Migrations completed successfully")
    return nil
}

func (r *PostgresRepo) CreateLog(fileName string) (int, error) {
    var logID int
    err := r.db.QueryRow(
        "INSERT INTO logs (file_name, status, created_at) VALUES ($1, 'processing', NOW()) RETURNING id",
        fileName,
    ).Scan(&logID)
    return logID, err
}

func (r *PostgresRepo) SaveNodes(logID int, nodes []model.Node) (map[string]int, error) {
    nodeMap := make(map[string]int)
    
    for _, node := range nodes {
        var nodeID int
        err := r.db.QueryRow(
            "INSERT INTO nodes (log_id, node_id, node_type, hostname, model) VALUES ($1, $2, $3, $4, $5) RETURNING id",
            logID, node.NodeID, node.NodeType, node.Hostname, node.Model,
        ).Scan(&nodeID)
        if err != nil {
            return nil, fmt.Errorf("failed to insert node: %w", err)
        }
        nodeMap[node.NodeID] = nodeID
    }
    
    return nodeMap, nil
}

func (r *PostgresRepo) SavePorts(nodeMap map[string]int, ports []model.Port) error {
    for i := range ports {
        // Находим ID узла по имени порта (предполагаем, что имя порта содержит node_id)
        for nodeID, dbNodeID := range nodeMap {
            if ports[i].PortName != "" {
                _, err := r.db.Exec(
                    "INSERT INTO ports (node_id, port_name, port_type, speed, status) VALUES ($1, $2, $3, $4, $5)",
                    dbNodeID, ports[i].PortName, ports[i].PortType, ports[i].Speed, ports[i].Status,
                )
                if err != nil {
                    return fmt.Errorf("failed to insert port: %w", err)
                }
                break
            }
            _ = nodeID // Заглушка для неиспользуемой переменной
        }
    }
    return nil
}

func (r *PostgresRepo) SaveNodesInfo(nodeMap map[string]int, nodesInfo []model.NodeInfo) error {
    for _, info := range nodesInfo {
        for _, dbNodeID := range nodeMap {
            _, err := r.db.Exec(
                "INSERT INTO nodes_info (node_id, cpu, memory, disk, os) VALUES ($1, $2, $3, $4, $5)",
                dbNodeID, info.CPU, info.Memory, info.Disk, info.OS,
            )
            if err != nil {
                return fmt.Errorf("failed to insert node info: %w", err)
            }
            break
        }
    }
    return nil
}

func (r *PostgresRepo) UpdateLogStatus(logID int, status string, nodeCount, portCount int) error {
    _, err := r.db.Exec(
        "UPDATE logs SET status = $1, node_count = $2, port_count = $3 WHERE id = $4",
        status, nodeCount, portCount, logID,
    )
    return err
}

func (r *PostgresRepo) GetLog(logID int) (*model.Log, error) {
    log := &model.Log{}
    err := r.db.QueryRow(
        "SELECT id, file_name, status, node_count, port_count, created_at FROM logs WHERE id = $1",
        logID,
    ).Scan(&log.ID, &log.FileName, &log.Status, &log.NodeCount, &log.PortCount, &log.CreatedAt)
    
    if err == sql.ErrNoRows {
        return nil, fmt.Errorf("log not found")
    }
    return log, err
}

func (r *PostgresRepo) GetTopology(logID int) (*model.Topology, error) {
    topology := &model.Topology{}
    
    // Получаем узлы
    rows, err := r.db.Query("SELECT id, log_id, node_id, node_type, hostname, model FROM nodes WHERE log_id = $1", logID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    for rows.Next() {
        var node model.Node
        if err := rows.Scan(&node.ID, &node.LogID, &node.NodeID, &node.NodeType, &node.Hostname, &node.Model); err != nil {
            return nil, err
        }
        topology.Nodes = append(topology.Nodes, node)
    }
    
    // Получаем порты
    portRows, err := r.db.Query(
        "SELECT p.id, p.node_id, p.port_name, p.port_type, p.speed, p.status FROM ports p JOIN nodes n ON p.node_id = n.id WHERE n.log_id = $1",
        logID,
    )
    if err != nil {
        return nil, err
    }
    defer portRows.Close()
    
    for portRows.Next() {
        var port model.Port
        if err := portRows.Scan(&port.ID, &port.NodeID, &port.PortName, &port.PortType, &port.Speed, &port.Status); err != nil {
            return nil, err
        }
        topology.Ports = append(topology.Ports, port)
    }
    
    return topology, nil
}

func (r *PostgresRepo) GetNode(nodeID int) (*model.Node, error) {
    node := &model.Node{}
    err := r.db.QueryRow(
        "SELECT id, log_id, node_id, node_type, hostname, model FROM nodes WHERE id = $1",
        nodeID,
    ).Scan(&node.ID, &node.LogID, &node.NodeID, &node.NodeType, &node.Hostname, &node.Model)
    
    if err == sql.ErrNoRows {
        return nil, fmt.Errorf("node not found")
    }
    return node, err
}

func (r *PostgresRepo) GetNodePorts(nodeID int) ([]model.Port, error) {
    rows, err := r.db.Query("SELECT id, node_id, port_name, port_type, speed, status FROM ports WHERE node_id = $1", nodeID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var ports []model.Port
    for rows.Next() {
        var port model.Port
        if err := rows.Scan(&port.ID, &port.NodeID, &port.PortName, &port.PortType, &port.Speed, &port.Status); err != nil {
            return nil, err
        }
        ports = append(ports, port)
    }
    
    return ports, nil
}
