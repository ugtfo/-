package parser

import (
    "bufio"
    "fmt"
    "log"
    "os"
    "path/filepath"
    "strings"

    "log-parser/internal/model"
)

type ParsedData struct {
    Nodes     []model.Node
    Ports     []model.Port
    NodesInfo []model.NodeInfo
}

func ParseLogFile(filePath string, logger *log.Logger) (*ParsedData, error) {
    file, err := os.Open(filePath)
    if err != nil {
        return nil, fmt.Errorf("failed to open file: %w", err)
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
    var currentSection string
    var currentNodeID string
    var currentNode model.Node
    var currentPorts []model.Port
    var currentInfo model.NodeInfo
    
    data := &ParsedData{}
    nodeMap := make(map[string]bool)

    for scanner.Scan() {
        line := strings.TrimSpace(scanner.Text())
        if line == "" {
            continue
        }

        // Определяем секции
        if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
            // Сохраняем предыдущий узел при смене секции
            if currentNodeID != "" && !nodeMap[currentNodeID] {
                data.Nodes = append(data.Nodes, currentNode)
                if len(currentPorts) > 0 {
                    data.Ports = append(data.Ports, currentPorts...)
                }
                if currentInfo != (model.NodeInfo{}) {
                    data.NodesInfo = append(data.NodesInfo, currentInfo)
                }
                nodeMap[currentNodeID] = true
            }
            
            currentSection = line
            continue
        }

        parts := strings.SplitN(line, "=", 2)
        if len(parts) != 2 {
            return nil, fmt.Errorf("invalid line format: %s", line)
        }

        key := strings.TrimSpace(parts[0])
        value := strings.TrimSpace(parts[1])

        switch currentSection {
        case "[NODE]":
            switch key {
            case "node_id":
                currentNodeID = value
                currentNode = model.Node{NodeID: value}
                currentPorts = nil
                currentInfo = model.NodeInfo{}
            case "type":
                currentNode.NodeType = value
            case "hostname":
                currentNode.Hostname = value
            case "model":
                currentNode.Model = value
            }
            
        case "[PORTS]":
            port := model.Port{PortName: key}
            portFields := strings.Split(value, ",")
            for _, field := range portFields {
                fieldParts := strings.SplitN(field, ":", 2)
                if len(fieldParts) == 2 {
                    switch strings.TrimSpace(fieldParts[0]) {
                    case "type":
                        port.PortType = strings.TrimSpace(fieldParts[1])
                    case "speed":
                        port.Speed = strings.TrimSpace(fieldParts[1])
                    case "status":
                        port.Status = strings.TrimSpace(fieldParts[1])
                    }
                }
            }
            currentPorts = append(currentPorts, port)
            
        case "[INFO]":
            switch key {
            case "cpu":
                currentInfo.CPU = value
            case "memory":
                currentInfo.Memory = value
            case "disk":
                currentInfo.Disk = value
            case "os":
                currentInfo.OS = value
            }
        }
    }

    // Сохраняем последний узел
    if currentNodeID != "" && !nodeMap[currentNodeID] {
        data.Nodes = append(data.Nodes, currentNode)
        if len(currentPorts) > 0 {
            data.Ports = append(data.Ports, currentPorts...)
        }
        if currentInfo != (model.NodeInfo{}) {
            data.NodesInfo = append(data.NodesInfo, currentInfo)
        }
    }

    if err := scanner.Err(); err != nil {
        return nil, fmt.Errorf("error reading file: %w", err)
    }

    if len(data.Nodes) == 0 {
        return nil, fmt.Errorf("no valid nodes found in file")
    }

    return data, nil
}

func ParseLogArchive(archivePath string, logger *log.Logger) (*ParsedData, error) {
    // Предполагаем, что архив уже распакован в папке data/
    logger.Printf("Parsing log from archive: %s", archivePath)
    
    var allData ParsedData
    
    err := filepath.Walk(archivePath, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        
        if !info.IsDir() && strings.HasSuffix(info.Name(), ".log") {
            logger.Printf("Parsing file: %s", path)
            data, err := ParseLogFile(path, logger)
            if err != nil {
                return fmt.Errorf("error parsing %s: %w", path, err)
            }
            
            allData.Nodes = append(allData.Nodes, data.Nodes...)
            allData.Ports = append(allData.Ports, data.Ports...)
            allData.NodesInfo = append(allData.NodesInfo, data.NodesInfo...)
        }
        return nil
    })
    
    if err != nil {
        return nil, err
    }
    
    if len(allData.Nodes) == 0 {
        return nil, fmt.Errorf("no log files found in archive")
    }
    
    return &allData, nil
}
