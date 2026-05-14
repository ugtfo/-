package service

import (
    "fmt"
    "log"
    "os"
    "path/filepath"
    "strings"
    "time"

    "log-parser/internal/model"
    "log-parser/internal/parser"
)

type Repository interface {
    CreateLog(fileName string) (int, error)
    SaveNodes(logID int, nodes []model.Node) (map[string]int, error)
    SavePorts(nodeMap map[string]int, ports []model.Port) error
    SaveNodesInfo(nodeMap map[string]int, nodesInfo []model.NodeInfo) error
    UpdateLogStatus(logID int, status string, nodeCount, portCount int) error
    GetLog(logID int) (*model.Log, error)
    GetTopology(logID int) (*model.Topology, error)
    GetNode(nodeID int) (*model.Node, error)
    GetNodePorts(nodeID int) ([]model.Port, error)
}

type Service struct {
    repo   Repository
    logger *log.Logger
}

func NewService(repo Repository, logger *log.Logger) *Service {
    return &Service{repo: repo, logger: logger}
}

func (s *Service) ParseLog(filePath string) (int, error) {
    start := time.Now()
    
    // Проверяем существование файла
    if _, err := os.Stat(filePath); os.IsNotExist(err) {
        return 0, fmt.Errorf("file not found: %s", filePath)
    }
    
    fileName := filepath.Base(filePath)
    
    // Создаем запись в логах
    logID, err := s.repo.CreateLog(fileName)
    if err != nil {
        return 0, fmt.Errorf("failed to create log entry: %w", err)
    }
    
    var data *parser.ParsedData
    
    // Определяем тип файла и парсим
    if strings.HasSuffix(fileName, ".zip") || strings.HasSuffix(fileName, ".tar.gz") {
        data, err = parser.ParseLogArchive(filePath, s.logger)
    } else {
        data, err = parser.ParseLogFile(filePath, s.logger)
    }
    
    if err != nil {
        s.repo.UpdateLogStatus(logID, "error", 0, 0)
        return 0, fmt.Errorf("parsing failed: %w", err)
    }
    
    // Сохраняем узлы
    nodeMap, err := s.repo.SaveNodes(logID, data.Nodes)
    if err != nil {
        s.repo.UpdateLogStatus(logID, "error", 0, 0)
        return 0, fmt.Errorf("failed to save nodes: %w", err)
    }
    
    // Сохраняем порты
    if err := s.repo.SavePorts(nodeMap, data.Ports); err != nil {
        s.repo.UpdateLogStatus(logID, "error", 0, 0)
        return 0, fmt.Errorf("failed to save ports: %w", err)
    }
    
    // Сохраняем информацию об узлах
    if err := s.repo.SaveNodesInfo(nodeMap, data.NodesInfo); err != nil {
        s.repo.UpdateLogStatus(logID, "error", 0, 0)
        return 0, fmt.Errorf("failed to save nodes info: %w", err)
    }
    
    // Обновляем статус
    nodeCount := len(data.Nodes)
    portCount := len(data.Ports)
    if err := s.repo.UpdateLogStatus(logID, "completed", nodeCount, portCount); err != nil {
        return 0, fmt.Errorf("failed to update log status: %w", err)
    }
    
    elapsed := time.Since(start)
    s.logger.Printf("Parsing completed for log_id=%d, nodes=%d, ports=%d, time=%v", logID, nodeCount, portCount, elapsed)
    
    return logID, nil
}

func (s *Service) GetLog(logID int) (*model.Log, error) {
    return s.repo.GetLog(logID)
}

func (s *Service) GetTopology(logID int) (*model.Topology, error) {
    return s.repo.GetTopology(logID)
}

func (s *Service) GetNode(nodeID int) (*model.Node, error) {
    return s.repo.GetNode(nodeID)
}

func (s *Service) GetNodePorts(nodeID int) ([]model.Port, error) {
    return s.repo.GetNodePorts(nodeID)
}
