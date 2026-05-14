package handler

import (
    "encoding/json"
    "log"
    "net/http"
    "strconv"
    "strings"
)

type Service interface {
    ParseLog(filePath string) (int, error)
    GetLog(logID int) (*model.Log, error)
    GetTopology(logID int) (*model.Topology, error)
    GetNode(nodeID int) (*model.Node, error)
    GetNodePorts(nodeID int) ([]model.Port, error)
}

type Handler struct {
    service Service
    logger  *log.Logger
}

func NewHandler(service Service, logger *log.Logger) *Handler {
    return &Handler{service: service, logger: logger}
}

func (h *Handler) HandleParse(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    var request struct {
        Path string `json:"path"`
    }
    
    if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
        h.logger.Printf("Error decoding request: %v", err)
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    logID, err := h.service.ParseLog(request.Path)
    if err != nil {
        h.logger.Printf("Error parsing log: %v", err)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    response := map[string]int{"log_id": logID}
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

func (h *Handler) HandleTopology(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    logID, err := extractIDFromURL(r.URL.Path, "/api/v1/topology/")
    if err != nil {
        http.Error(w, "Invalid log ID", http.StatusBadRequest)
        return
    }

    topology, err := h.service.GetTopology(logID)
    if err != nil {
        h.logger.Printf("Error getting topology: %v", err)
        http.Error(w, err.Error(), http.StatusNotFound)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(topology)
}

func (h *Handler) HandleNode(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    nodeID, err := extractIDFromURL(r.URL.Path, "/api/v1/node/")
    if err != nil {
        http.Error(w, "Invalid node ID", http.StatusBadRequest)
        return
    }

    node, err := h.service.GetNode(nodeID)
    if err != nil {
        h.logger.Printf("Error getting node: %v", err)
        http.Error(w, err.Error(), http.StatusNotFound)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(node)
}

func (h *Handler) HandlePort(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    nodeID, err := extractIDFromURL(r.URL.Path, "/api/v1/port/")
    if err != nil {
        http.Error(w, "Invalid node ID", http.StatusBadRequest)
        return
    }

    ports, err := h.service.GetNodePorts(nodeID)
    if err != nil {
        h.logger.Printf("Error getting ports: %v", err)
        http.Error(w, err.Error(), http.StatusNotFound)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(ports)
}

func (h *Handler) HandleLog(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    logID, err := extractIDFromURL(r.URL.Path, "/api/v1/log/")
    if err != nil {
        http.Error(w, "Invalid log ID", http.StatusBadRequest)
        return
    }

    log, err := h.service.GetLog(logID)
    if err != nil {
        h.logger.Printf("Error getting log: %v", err)
        http.Error(w, err.Error(), http.StatusNotFound)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(log)
}

func extractIDFromURL(path, prefix string) (int, error) {
    idStr := strings.TrimPrefix(path, prefix)
    idStr = strings.TrimSuffix(idStr, "/")
    return strconv.Atoi(idStr)
}
