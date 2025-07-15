# New Files Summary from Ultrathink Recommendations

## Files That Would Need to Be Created

### 1. RAGHandler (Not Recommended Initially)
**Path**: `internal/logging/raghandler.go`

**Purpose**: Embed reflection logs into a vector database for self-improvement

**Dependencies Required**:
- FAISS or alternative vector database
- Embedding model (BERT/SentenceTransformers)

**Status**: Not recommended for initial implementation due to:
- High complexity
- Missing infrastructure
- Unclear immediate value
- Placeholder embedding function

### 2. Vector Database Support (If RAG is Pursued)
**Path**: `internal/db/faiss.go` (or similar)

**Purpose**: Provide vector database functionality

**Note**: This would require:
- New database schema for vectors
- Integration with embedding models
- Significant new dependencies

## Configuration Changes Needed

### 1. Logging Configuration Structure
**File to Modify**: `internal/config/config.go`

**Add**:
```go
// LoggingOutput defines a log output type and its options.
type LoggingOutput struct {
    Type    string                 `json:"type"`
    Options map[string]interface{} `json:"options"`
}

// LoggingModule defines a module's logging level.
type LoggingModule struct {
    Level string `json:"level"`
}

// LoggingConfig holds logging configuration.
type LoggingConfig struct {
    Outputs []LoggingOutput          `json:"outputs"`
    Modules map[string]LoggingModule `json:"modules"`
}
```

**Update Config struct to include**:
```go
Logging LoggingConfig `json:"logging"`
```

### 2. SetupLogging Function
**File to Modify**: `internal/config/config.go`

**Add**: A function to initialize logging based on configuration

## Files That Should NOT Be Created/Modified

### 1. Enhanced Logs Table (internal/tui/components/logs/table.go)
**Reason**: Current implementation is more sophisticated than the proposal
- Uses pubsub for real-time updates
- Has better architecture than proposed fsnotify approach

### 2. Reflection Sidebar (internal/tui/components/chat/sidebar.go)
**Reason**: Would replace valuable file tracking functionality
- Current sidebar shows modified files with diffs
- Reflection display should be added elsewhere, not replace this

## Recommended Implementation Order

1. **First Phase** (Configuration only):
   - Add LoggingConfig types to config.go
   - Implement SetupLogging function (without RAG)
   - Test with existing handlers

2. **Second Phase** (If needed):
   - Evaluate need for additional log outputs
   - Consider simpler alternatives to RAG (full-text search)
   
3. **Future Consideration** (If proven necessary):
   - RAG implementation with proper vector database
   - Embedding model integration

## Dependencies to Add

### Required for Basic Implementation:
- None (uses existing infrastructure)

### Required for Full Proposal (Not Recommended):
- Vector database (FAISS/Chroma/Weaviate)
- Embedding model library
- fsnotify (not recommended - current pubsub is better)

## Configuration File Updates

### .opencode.json Example
```json
{
  "logging": {
    "outputs": [
      {"type": "console"},
      {"type": "file", "options": {"path": "opencode.log"}},
      {"type": "session", "options": {"path": "session.log"}}
    ],
    "modules": {
      "main": {"level": "info"},
      "llm": {"level": "debug"}
    }
  }
}
```

Note: RAG output type excluded until infrastructure is in place.