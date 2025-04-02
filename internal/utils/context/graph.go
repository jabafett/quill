package context

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"gonum.org/v1/gonum/graph/simple"

	"github.com/jabafett/quill/internal/utils/debug"
)

// NodeType defines the type of a node in the graph.
type NodeType string

const (
	FileNode   NodeType = "file"
	SymbolNode NodeType = "symbol"
)

// GraphNode implements graph.Node and stores our specific data.
type GraphNode struct {
	NodeID int64    `json:"id"`   // Unique int64 ID required by gonum
	Type   NodeType `json:"type"` // FileNode or SymbolNode

	// File attributes (only relevant if Type == FileNode)
	FilePath string `json:"filePath,omitempty"`

	// Symbol attributes (only relevant if Type == SymbolNode)
	SymbolName      string `json:"symbolName,omitempty"`
	SymbolKind      string `json:"symbolKind,omitempty"` // From SymbolContext.Type
	SymbolStartLine int    `json:"symbolStartLine,omitempty"`
	SymbolEndLine   int    `json:"symbolEndLine,omitempty"`
	SymbolFilePath  string `json:"symbolFilePath,omitempty"` // File containing the symbol
}

// ID returns the unique int64 ID for the node.
func (n *GraphNode) ID() int64 {
	return n.NodeID
}

// String returns a string representation (optional but helpful for debugging).
func (n *GraphNode) String() string {
	if n.Type == FileNode {
		return fmt.Sprintf("File:%d(%s)", n.NodeID, n.FilePath)
	}
	if n.Type == SymbolNode {
		return fmt.Sprintf("Symbol:%d(%s#%s@%d)", n.NodeID, n.SymbolFilePath, n.SymbolName, n.SymbolStartLine)
	}
	return fmt.Sprintf("Node:%d(UnknownType)", n.NodeID)
}

// DependencyGraph manages the gonum graph and mappings.
type DependencyGraph struct {
	Graph *simple.DirectedGraph // The core gonum graph

	// Mappings for easy lookup
	nodeMap   map[string]*GraphNode // Maps string ID ("file:/path", "symbol:/path#Name@Line") -> GraphNode
	idCounter int64                 // For generating unique int64 node IDs
	mu        sync.RWMutex          // Protects nodeMap and idCounter during concurrent access
}

// NewDependencyGraph initializes an empty graph structure.
func NewDependencyGraph() *DependencyGraph {
	return &DependencyGraph{
		Graph:     simple.NewDirectedGraph(),
		nodeMap:   make(map[string]*GraphNode),
		idCounter: 0,
		// mu is initialized implicitly
	}
}

// createFileStringID generates a unique string identifier for a file node.
func createFileStringID(path string) string {
	return "file:" + path
}

// createSymbolStringID generates a unique string identifier for a symbol node.
func createSymbolStringID(symbol SymbolContext) string {
	// Using StartLine helps differentiate symbols with the same name in the same file
	return fmt.Sprintf("symbol:%s#%s@%d", symbol.FilePath, symbol.Name, symbol.StartLine)
}

// getNode retrieves or creates a node in the graph, ensuring uniqueness.
// It handles locking for concurrent safety.
func (g *DependencyGraph) getNode(stringID string, nodeType NodeType, data map[string]interface{}) *GraphNode {
	g.mu.RLock()
	node, exists := g.nodeMap[stringID]
	g.mu.RUnlock()

	if exists {
		return node
	}

	// Node doesn't exist, need write lock to create it
	g.mu.Lock()
	defer g.mu.Unlock()

	// Double-check after acquiring write lock
	node, exists = g.nodeMap[stringID]
	if exists {
		return node
	}

	// Create the new node
	newNodeID := g.idCounter
	g.idCounter++

	newNode := &GraphNode{
		NodeID: newNodeID,
		Type:   nodeType,
	}

	// Populate data based on type
	switch nodeType {
	case FileNode:
		if path, ok := data["path"].(string); ok {
			newNode.FilePath = path
		}
	case SymbolNode:
		if name, ok := data["name"].(string); ok {
			newNode.SymbolName = name
		}
		if kind, ok := data["kind"].(string); ok {
			newNode.SymbolKind = kind
		}
		if start, ok := data["startLine"].(int); ok {
			newNode.SymbolStartLine = start
		}
		if end, ok := data["endLine"].(int); ok {
			newNode.SymbolEndLine = end
		}
		if path, ok := data["filePath"].(string); ok {
			newNode.SymbolFilePath = path
		}
	}

	g.nodeMap[stringID] = newNode
	g.Graph.AddNode(newNode) // Add to the actual gonum graph

	debug.Log("Added Node: %s (ID: %d)", stringID, newNodeID)
	return newNode
}

// addEdge adds a directed edge between two nodes.
func (g *DependencyGraph) addEdge(fromNode, toNode *GraphNode) {
	if fromNode == nil || toNode == nil {
		debug.Log("Warning: Attempted to add edge with nil node(s)")
		return
	}
	// Check if edge already exists (optional, simple.DirectedGraph handles duplicates)
	// if g.Graph.HasEdgeFromTo(fromNode.ID(), toNode.ID()) {
	//     return
	// }
	edge := simple.Edge{F: fromNode, T: toNode}
	g.Graph.SetEdge(edge) // Creates or updates the edge
}

// BuildDependencyGraph constructs the graph from the repository context using gonum.
func BuildDependencyGraph(repoCtx *RepositoryContext) (*DependencyGraph, error) {
	if repoCtx == nil || repoCtx.Files == nil {
		return nil, fmt.Errorf("cannot build graph from nil or empty repository context")
	}

	graph := NewDependencyGraph()
	allFiles := repoCtx.Files // For import resolution lookup

	// --- Pass 1: Add all file and symbol nodes & Define edges ---
	debug.Log("Graph Build - Pass 1: Adding Nodes and Define Edges...")
	nodeBuildCount := 0
	defineEdgeCount := 0
	for path, fileCtx := range allFiles {
		if fileCtx == nil {
			continue
		}

		// Get/Create file node
		fileStringID := createFileStringID(path)
		fileNode := graph.getNode(fileStringID, FileNode, map[string]interface{}{"path": path})
		nodeBuildCount++

		// Get/Create symbol nodes and 'Defines' edges for this file
		for _, symbol := range fileCtx.Symbols {
			// Ensure symbol has file path associated (should be set during analysis)
			symbol.FilePath = path // Make sure this is set correctly in analyzer

			symbolStringID := createSymbolStringID(symbol)
			symbolNode := graph.getNode(symbolStringID, SymbolNode, map[string]interface{}{
				"name":      symbol.Name,
				"kind":      symbol.Type,
				"startLine": symbol.StartLine,
				"endLine":   symbol.EndLine,
				"filePath":  path,
			})
			nodeBuildCount++

			// Add edge: File -> Defines -> Symbol
			graph.addEdge(fileNode, symbolNode)
			defineEdgeCount++
		}
	}
	debug.Log("Graph Build - Pass 1 Complete: %d nodes created/retrieved, %d 'Define' edges added.", nodeBuildCount, defineEdgeCount)

	// --- Pass 2: Add import edges ---
	debug.Log("Graph Build - Pass 2: Adding Import Edges...")
	importEdgeCount := 0
	// This requires resolving import paths, which can be complex.
	for path, fileCtx := range allFiles {
		if fileCtx == nil {
			continue
		}
		fileStringID := createFileStringID(path)
		importingNode := graph.nodeMap[fileStringID] // Should exist from Pass 1

		if importingNode == nil {
			debug.Log("Warning: Importing node %s not found during Pass 2", fileStringID)
			continue
		}

		for _, imp := range fileCtx.Imports {
			// TODO: Implement robust import path resolution based on language and project structure.
			resolvedPath := resolveImportPath(imp, path, allFiles) // Needs implementation (same as previous plan)

			if resolvedPath != "" {
				targetFileStringID := createFileStringID(resolvedPath)
				// Lookup the target node (don't create here, it should exist from Pass 1 if tracked)
				graph.mu.RLock()
				targetNode, exists := graph.nodeMap[targetFileStringID]
				graph.mu.RUnlock()

				if exists {
					graph.addEdge(importingNode, targetNode)
					importEdgeCount++
				} else {
					debug.Log("Skipping import edge: Target file %s (from import '%s' in %s) not found in graph.", resolvedPath, imp, path)
				}
			}
		}
	}
	debug.Log("Graph Build - Pass 2 Complete: %d 'Import' edges added.", importEdgeCount)

	// --- Pass 3: Add Calls/References edges (Future Enhancement) ---
	// Requires analyzer changes. Logic would be similar:
	// 1. Get symbol node A.
	// 2. Find symbol node B that A calls/references (needs lookup logic).
	// 3. Add edge A -> B.

	debug.Log("Graph build complete. Total Nodes: %d", graph.Graph.Nodes().Len())
	return graph, nil
}

// resolveImportPath (Placeholder - Needs language-specific logic)
// Tries to find a file in the context that matches the import string.
func resolveImportPath(importPath string, importingFilePath string, allFiles map[string]*FileContext) string {
	// Extremely basic example: Check for exact match or relative path
	// This needs to handle Go packages, Python modules, JS/TS paths, etc.
	potentialPath := importPath // Assume direct match first

	// Simplistic relative path check (assuming Unix-like paths)
	if strings.HasPrefix(importPath, ".") {
		// Resolve relative to the *directory* of the importing file
		absPath := filepath.Join(filepath.Dir(importingFilePath), importPath)
		// Clean the path (removes ../, ./) and ensures consistent separators
		potentialPath = filepath.Clean(absPath)
		// Ensure it uses '/' separators if needed for map keys, though filepath should handle OS specifics
		potentialPath = filepath.ToSlash(potentialPath)
	}

	// Check if the exact resolved path exists
	if _, exists := allFiles[potentialPath]; exists {
		return potentialPath
	}

	// Add more sophisticated checks based on language conventions...
	// e.g., check GOPATH/GOROOT for Go, node_modules for JS/TS, sys.path for Python etc.
	// This part is highly language-dependent.

	// Final check: maybe the import *is* the filename key? (e.g., Python module name maps to file)
	if _, exists := allFiles[importPath]; exists {
		return importPath
	}

	return "" // Could not resolve
}

// findNodes performs a breadth-first search from start nodes up to maxDepth.
// It collects nodes that match the filter criteria.
func (g *DependencyGraph) findNodes(startStringIDs []string, maxDepth int, filter func(*GraphNode) bool) (map[int64]*GraphNode, error) {
	if g.Graph == nil {
		return nil, fmt.Errorf("graph is nil")
	}

	foundNodes := make(map[int64]*GraphNode)
	visited := make(map[int64]int) // Store depth visited at

	queue := make([]*GraphNode, 0)

	// Initialize queue with starting nodes
	g.mu.RLock()
	for _, startID := range startStringIDs {
		if node, exists := g.nodeMap[startID]; exists {
			if _, seen := visited[node.ID()]; !seen {
				queue = append(queue, node)
				visited[node.ID()] = 0 // Depth 0
				if filter(node) {
					foundNodes[node.ID()] = node
				}
			}
		} else {
			debug.Log("Warning: Start node %s not found in graph", startID)
		}
	}
	g.mu.RUnlock()

	head := 0
	for head < len(queue) {
		current := queue[head]
		head++

		currentDepth := visited[current.ID()]
		if currentDepth >= maxDepth {
			continue
		}

		// Use gonum's graph traversal methods
		neighbors := g.Graph.From(current.ID()) // Get nodes reachable *from* current
		for neighbors.Next() {
			neighbor := neighbors.Node().(*GraphNode) // Type assertion needed
			if _, seen := visited[neighbor.ID()]; !seen {
				visited[neighbor.ID()] = currentDepth + 1
				queue = append(queue, neighbor)
				if filter(neighbor) {
					foundNodes[neighbor.ID()] = neighbor
				}
			}
		}
	}

	return foundNodes, nil
}

// FindRelatedFiles finds files related to the initial set of files within a certain depth.
// Currently follows outgoing edges (e.g., files imported *by* startFiles).
// To find files *importing* startFiles, graph traversal needs adjustment (e.g., using g.Graph.To()).
func (g *DependencyGraph) FindRelatedFiles(startFiles []string, maxDepth int) ([]string, error) {
	startNodeIDs := make([]string, len(startFiles))
	for i, f := range startFiles {
		startNodeIDs[i] = createFileStringID(f)
	}

	// Filter for FileNode types
	fileFilter := func(node *GraphNode) bool {
		return node.Type == FileNode
	}

	relatedNodesMap, err := g.findNodes(startNodeIDs, maxDepth, fileFilter)
	if err != nil {
		return nil, fmt.Errorf("error finding related nodes: %w", err)
	}

	// Extract file paths
	relatedFilePaths := make([]string, 0, len(relatedNodesMap))
	for _, node := range relatedNodesMap {
		relatedFilePaths = append(relatedFilePaths, node.FilePath)
	}

	// TODO: Consider adding traversal in the reverse direction (files that import startFiles)
	// This would involve using g.Graph.To(nodeID) or a separate BFS.

	return relatedFilePaths, nil
}

// FindRelatedSymbols finds symbols related to a set of starting files or symbols.
// Example: Finds symbols defined within the start files or files reachable from them.
func (g *DependencyGraph) FindRelatedSymbols(startFiles []string, maxDepth int) ([]SymbolContext, error) {
	startNodeIDs := make([]string, len(startFiles))
	for i, f := range startFiles {
		startNodeIDs[i] = createFileStringID(f)
	}

	// Find *all* nodes reachable first (files and symbols)
	// We need files to find symbols defined *by* them.
	reachableNodesMap, err := g.findNodes(startNodeIDs, maxDepth, func(node *GraphNode) bool { return true }) // Find all reachable nodes
	if err != nil {
		return nil, fmt.Errorf("error finding reachable nodes: %w", err)
	}

	relatedSymbols := make([]SymbolContext, 0)
	processedSymbols := make(map[int64]bool) // Avoid duplicates if symbols are reached directly

	// Iterate through reachable nodes. If it's a file, find symbols it defines.
	// If it's a symbol directly, add it.
	for _, node := range reachableNodesMap {
		if node.Type == FileNode {
			// Find symbols defined by this file node (follow outgoing edges)
			defines := g.Graph.From(node.ID())
			for defines.Next() {
				definedNode := defines.Node().(*GraphNode)
				if definedNode.Type == SymbolNode {
					if !processedSymbols[definedNode.ID()] {
						relatedSymbols = append(relatedSymbols, SymbolContext{
							Name:      definedNode.SymbolName,
							Type:      definedNode.SymbolKind,
							StartLine: definedNode.SymbolStartLine,
							EndLine:   definedNode.SymbolEndLine,
							FilePath:  definedNode.SymbolFilePath,
						})
						processedSymbols[definedNode.ID()] = true
					}
				}
			}
		} else if node.Type == SymbolNode {
			// If a symbol node was reached directly (e.g., via future Calls/References edges)
			if !processedSymbols[node.ID()] {
				relatedSymbols = append(relatedSymbols, SymbolContext{
					Name:      node.SymbolName,
					Type:      node.SymbolKind,
					StartLine: node.SymbolStartLine,
					EndLine:   node.SymbolEndLine,
					FilePath:  node.SymbolFilePath,
				})
				processedSymbols[node.ID()] = true
			}
		}
	}

	return relatedSymbols, nil
}

// GetContextForFiles generates a string summary of related files/symbols for the AI prompt.
func (g *DependencyGraph) GetContextForFiles(files []string, maxDepth int) (string, error) {
	// --- Find Related Files ---
	relatedFiles, err := g.FindRelatedFiles(files, maxDepth)
	if err != nil {
		// Don't fail entirely, maybe just log the error and proceed without file context
		debug.Log("Warning: Failed to find related files for context: %v", err)
		relatedFiles = []string{} // Ensure it's not nil
	}

	// Filter out the input files themselves from the related list for brevity
	inputFilesMap := make(map[string]struct{})
	for _, f := range files {
		inputFilesMap[f] = struct{}{}
	}
	filteredRelatedFiles := make([]string, 0, len(relatedFiles))
	for _, rf := range relatedFiles {
		if _, isInput := inputFilesMap[rf]; !isInput {
			filteredRelatedFiles = append(filteredRelatedFiles, rf)
		}
	}

	// --- Find Related Symbols ---
	// Find symbols within the *original* changed files + 1 level of related files
	searchFiles := make([]string, len(files))
	copy(searchFiles, files)
	if len(filteredRelatedFiles) > 0 {
		searchFiles = append(searchFiles, filteredRelatedFiles...)
	}

	relatedSymbols, err := g.FindRelatedSymbols(searchFiles, 2) // Search depth 1 within relevant files
	if err != nil {
		// Log error, proceed without symbol context
		debug.Log("Warning: Failed to find related symbols for context: %v", err)
		relatedSymbols = []SymbolContext{}
	}

	// --- Format the output string ---
	var sb strings.Builder
	hasContext := false

	if len(filteredRelatedFiles) > 0 {
		hasContext = true
		sb.WriteString("Related files (via imports):\n")
		limit := 10
		for i, path := range filteredRelatedFiles {
			if i >= limit {
				sb.WriteString(fmt.Sprintf("... and %d more files\n", len(filteredRelatedFiles)-i))
				break
			}
			sb.WriteString(fmt.Sprintf("- %s\n", path))
		}
	}

	if len(relatedSymbols) > 0 {
		if hasContext {
			sb.WriteString("\n")
		} // Add separator if files were listed
		hasContext = true
		sb.WriteString("Key symbols found in changed/related files:\n")
		limit := 15
		for i, symbol := range relatedSymbols {
			if i >= limit {
				sb.WriteString(fmt.Sprintf("... and %d more symbols\n", len(relatedSymbols)-i))
				break
			}
			// Provide concise symbol info
			sb.WriteString(fmt.Sprintf("- %s %s (%s:%d)\n", symbol.Type, symbol.Name, filepath.Base(symbol.FilePath), symbol.StartLine))
		}
	}

	if !hasContext {
		return "", nil // No relevant context found
	}

	return sb.String(), nil
}
