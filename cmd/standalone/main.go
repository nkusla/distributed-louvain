package main

import (
	"context"
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
	MachineID           = "machine-0"
	MaxIterations       = 20
	AlgorithmTimeout    = 60 * time.Second
	ShutdownGracePeriod = 2 * time.Second
)

func main() {
	fmt.Println("Starting standalone mode")

	edges, err := graph.ReadEdgesFromCSV("data/karate_club.csv")
	if err != nil {
		log.Fatalf("Failed to read CSV: %v", err)
	}
	log.Printf("Loaded %d edges from karate_club.csv", len(edges))

	totalWeight := 0
	for _, edge := range edges {
		totalWeight += edge.W
	}
	totalWeight = totalWeight / 2
	log.Printf("Total graph weight: %d", totalWeight)

	provider := cluster.NewSimpleProvider(MachineID)
	system := actor.NewActorSystem(MachineID, provider)

	coordinatorPID := actor.NewPID(MachineID, "coordinator")
	coordinator := actors.NewCoordinatorActor(coordinatorPID, system, MaxIterations)
	if err := system.Register(coordinator); err != nil {
		log.Fatalf("Failed to register coordinator: %v", err)
	}
	provider.SetCoordinator(coordinatorPID)

	aggregators := make([]*actors.AggregatorActor, NumAggregators)
	for i := 0; i < NumAggregators; i++ {
		aggregatorPID := actor.NewPID(MachineID, fmt.Sprintf("aggregator-%d", i))
		aggregator := actors.NewAggregatorActor(aggregatorPID, system, coordinatorPID, NumPartitions)
		if err := system.Register(aggregator); err != nil {
			log.Fatalf("Failed to register aggregator %d: %v", i, err)
		}
		if err := provider.RegisterActor(actor.AggregatorType, aggregatorPID); err != nil {
			log.Fatalf("Failed to register aggregator %d in provider: %v", i, err)
		}
		aggregators[i] = aggregator
	}

	partitions := make([]*actors.PartitionActor, NumPartitions)
	for i := 0; i < NumPartitions; i++ {
		partitionPID := actor.NewPID(MachineID, fmt.Sprintf("partition-%d", i))
		partition := actors.NewPartitionActor(partitionPID, system, coordinatorPID)
		if err := system.Register(partition); err != nil {
			log.Fatalf("Failed to register partition %d: %v", i, err)
		}
		if err := provider.RegisterActor(actor.PartitionType, partitionPID); err != nil {
			log.Fatalf("Failed to register partition %d in provider: %v", i, err)
		}
		partitions[i] = partition
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := system.Start(); err != nil {
		log.Fatalf("Failed to start actor system: %v", err)
	}

	coordinator.Start(ctx)
	for _, aggregator := range aggregators {
		aggregator.Start(ctx)
	}
	for _, partition := range partitions {
		partition.Start(ctx)
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
	cancel()
	coordinator.Stop()
	time.Sleep(ShutdownGracePeriod)
	log.Println("Shutdown complete")
}
