package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/distributed-louvain/pkg/actor"
	"github.com/distributed-louvain/pkg/actors"
	"github.com/distributed-louvain/pkg/cluster"
	"github.com/distributed-louvain/pkg/graph"
)

const (
	NumPartitions       = 4
	NumAggregators      = 2
	DataPath            = "data/karate_club.csv"
	MachineID           = "machine-0"
	MaxIterations       = 20
	AlgorithmTimeout    = 60 * time.Second
	ShutdownGracePeriod = 2 * time.Second
)

func main() {
	fmt.Println("Starting standalone mode")

	edges, totalWeight, err := graph.LoadGraphData(DataPath)
	if err != nil {
		log.Fatalf("Failed to load graph data: %v", err)
	}

	provider := cluster.NewSimpleProvider(MachineID, false)
	system := actor.NewActorSystem(MachineID, provider)

	coordinatorPID := actor.NewPID(MachineID, "coordinator")
	coordinator := actors.NewCoordinatorActor(coordinatorPID, system, MaxIterations)
	if err := system.Register(coordinator); err != nil {
		log.Fatalf("Failed to register coordinator: %v", err)
	}
	provider.SetCoordinator(coordinatorPID)

	for i := 0; i < NumAggregators; i++ {
		aggregatorPID := actor.NewPID(MachineID, fmt.Sprintf("aggregator-%d", i))
		aggregator := actors.NewAggregatorActor(aggregatorPID, system, coordinatorPID, NumPartitions)
		if err := system.Register(aggregator); err != nil {
			log.Fatalf("Failed to register aggregator %d: %v", i, err)
		}
		if err := provider.RegisterActor(actor.AggregatorType, aggregatorPID); err != nil {
			log.Fatalf("Failed to register aggregator %d in provider: %v", i, err)
		}
	}

	for i := 0; i < NumPartitions; i++ {
		partitionPID := actor.NewPID(MachineID, fmt.Sprintf("partition-%d", i))
		partition := actors.NewPartitionActor(partitionPID, system, coordinatorPID)
		if err := system.Register(partition); err != nil {
			log.Fatalf("Failed to register partition %d: %v", i, err)
		}
		if err := provider.RegisterActor(actor.PartitionType, partitionPID); err != nil {
			log.Fatalf("Failed to register partition %d in provider: %v", i, err)
		}
	}

	if err := system.Start(); err != nil {
		log.Fatalf("Failed to start actor system: %v", err)
	}

	coordinator.StartAlgorithm(edges, totalWeight)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	select {
	case <-sigChan:
		log.Println("Received shutdown signal")
	case <-time.After(AlgorithmTimeout):
		log.Println("Algorithm execution timeout")
	}

	log.Println("Shutting down...")
	system.Shutdown()
	log.Println("Shutdown complete")
}
