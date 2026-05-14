package model

import (
    "time"
)

type Log struct {
    ID        int       `json:"id"`
    FileName  string    `json:"file_name"`
    Status    string    `json:"status"`
    NodeCount int       `json:"node_count"`
    PortCount int       `json:"port_count"`
    CreatedAt time.Time `json:"created_at"`
}

type Node struct {
    ID        int    `json:"id"`
    LogID     int    `json:"log_id"`
    NodeID    string `json:"node_id"`
    NodeType  string `json:"node_type"` // host или switch
    Hostname  string `json:"hostname,omitempty"`
    Model     string `json:"model,omitempty"`
}

type Port struct {
    ID          int    `json:"id"`
    NodeID      int    `json:"node_id"`
    PortName    string `json:"port_name"`
    PortType    string `json:"port_type"` // ethernet, fiber и т.д.
    Speed       string `json:"speed,omitempty"`
    Status      string `json:"status"`    // up/down
}

type NodeInfo struct {
    ID          int    `json:"id"`
    NodeID      int    `json:"node_id"`
    CPU         string `json:"cpu,omitempty"`
    Memory      string `json:"memory,omitempty"`
    Disk        string `json:"disk,omitempty"`
    OS          string `json:"os,omitempty"`
}

type Topology struct {
    Nodes []Node `json:"nodes"`
    Ports []Port `json:"ports"`
    Links []Link `json:"links,omitempty"` // связи между портами
}

type Link struct {
    SourceNodeID int    `json:"source_node_id"`
    SourcePort   string `json:"source_port"`
    TargetNodeID int    `json:"target_node_id"`
    TargetPort   string `json:"target_port"`
}
