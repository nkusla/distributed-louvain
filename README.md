# Distributed Louvain Algorithm

A distributed implementation of the Louvain community detection algorithm in Go using an actor-based architecture with message-passing communication. The system partitions graphs across multiple nodes for parallel processing and can run in both standalone mode (single machine) and distributed mode (multiple nodes via Docker).

## Features

- **Distributed Processing**: Partitions graphs across multiple nodes for parallel processing
- **Actor-Based Architecture**: Uses an actor system for message-passing between components
- **CRDT Support**: Conflict-free replicated data types for distributed state management
- **Flexible Deployment**: Supports both standalone and distributed Docker-based deployment
- **Configurable**: YAML-based configuration for coordinator and worker nodes

## Architecture

The system consists of three main actor types:

- **Coordinator**: Orchestrates the algorithm phases and aggregates results
- **Partition**: Manages a subset of the graph and performs local optimizations
- **Aggregator**: Collects and combines intermediate results from partitions

## Getting Started

### Prerequisites

- Go 1.23.4 or later
- Docker 28.1.1 or later (for distributed mode)
- Docker Compose 2.35 or later (for distributed mode)

### Running in Standalone Mode

```bash
# Build the standalone binary
make build

# Run the algorithm on a single machine
make run
```

### Running in Distributed Mode

```bash
# Deploy the system with Docker Compose
make deploy

# Start the algorithm on the coordinator
make start

# Stop the system
make stop
```

## Project Structure

```
├── cmd/                   # Application entry points
│   ├── standalone/        # Standalone mode
│   └── node/              # Distributed node
├── pkg/
│   ├── actor/             # Actor system implementation
│   ├── actors/            # Coordinator, Partition, and Aggregator actors
│   ├── cluster/           # Cluster management and transport
│   ├── graph/             # Graph data structures
│   ├── graphio/           # Graph I/O utilities
│   ├── messages/          # Message types for actor communication
│   └── crdt/              # CRDT implementations
├── data/                  # Sample graph datasets
└── deploy/                # Docker deployment files
```

## Configuration

Configuration files are located in `deploy/configs/`. Each node requires:

- Machine ID
- Listen address and port
- Coordinator address (for worker nodes)
- Number of partitions and aggregators
- Graph data file path

## Testing

```bash
make test
```

This will run the unit tests for the project. The tests cover serialization of messages and CRDT functionality to ensure correctness in distributed communication.