package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/distributed-louvain/pkg/actor"
	"github.com/distributed-louvain/pkg/actors"
	"github.com/distributed-louvain/pkg/cluster"
	"github.com/distributed-louvain/pkg/config"
	"github.com/distributed-louvain/pkg/graph"
)

const (
	DefaultPort           = 8080
	DefaultMaxIterations  = 20
	DefaultTimeout        = 60 * time.Second
	DefaultGracePeriod    = 2 * time.Second
	DefaultDataPath       = "data/karate_club.csv"
)

func main() {
	var (
		configPath = flag.String("config", "", "Path to configuration file (YAML)")
	)
	flag.Parse()

	var cfg *config.Config
	var err error

	if *configPath != "" {
		cfg, err = config.LoadConfig(*configPath)
		if err != nil {
			log.Printf("Failed to load config: %v", err)
			os.Exit(1)
		}
		log.Printf("Loaded configuration from %s", *configPath)
	} else {
		log.Printf("No configuration file provided")
		os.Exit(1)
	}

	log.Printf("Starting node %s on port %d", cfg.MachineID, cfg.Port)

	provider := cluster.NewSimpleProvider(cfg.MachineID, true)

	for _, peer := range cfg.Network.Peers {
		provider.RegisterMachine(peer.ID, peer.Address)
		log.Printf("Registered peer: %s -> %s", peer.ID, peer.Address)
	}

	if err := registerPeerActors(provider, cfg.MachineID); err != nil {
		log.Printf("Warning: Failed to register peer actors: %v", err)
		os.Exit(1)
	}

	system := actor.NewActorSystem(cfg.MachineID, provider)
	provider.SetActorSystem(system)

	var coordinatorPID actor.PID
	if cfg.IsCoordinator {
		coordinatorPID = actor.NewPID(cfg.MachineID, "coordinator")
		coordinatorActor := actors.NewCoordinatorActor(coordinatorPID, system, cfg.Algorithm.MaxIterations)
		if err := system.Register(coordinatorActor); err != nil {
			log.Fatalf("Failed to register coordinator: %v", err)
		}
		provider.SetCoordinator(coordinatorPID)
		log.Printf("Registered coordinator actor")
	} else {
		coordinatorPID = actor.NewPID("coordinator", "coordinator")
		log.Printf("Using remote coordinator: %s", coordinatorPID)
	}

	aggregators := make([]*actors.AggregatorActor, cfg.Actors.Aggregators)
	for i := 0; i < cfg.Actors.Aggregators; i++ {
		aggregatorPID := actor.NewPID(cfg.MachineID, fmt.Sprintf("aggregator-%d", i))
		aggregator := actors.NewAggregatorActor(aggregatorPID, system, coordinatorPID, cfg.Actors.Partitions)
		if err := system.Register(aggregator); err != nil {
			log.Fatalf("Failed to register aggregator %d: %v", i, err)
		}
		if err := provider.RegisterActor(actor.AggregatorType, aggregatorPID); err != nil {
			log.Fatalf("Failed to register aggregator %d in provider: %v", i, err)
		}
		aggregators[i] = aggregator
	}

	partitions := make([]*actors.PartitionActor, cfg.Actors.Partitions)
	for i := 0; i < cfg.Actors.Partitions; i++ {
		partitionPID := actor.NewPID(cfg.MachineID, fmt.Sprintf("partition-%d", i))
		partition := actors.NewPartitionActor(partitionPID, system, coordinatorPID)
		if err := system.Register(partition); err != nil {
			log.Fatalf("Failed to register partition %d: %v", i, err)
		}
		if err := provider.RegisterActor(actor.PartitionType, partitionPID); err != nil {
			log.Fatalf("Failed to register partition %d in provider: %v", i, err)
		}
		partitions[i] = partition
	}

	// if err := system.Start(); err != nil {
	// 	log.Fatalf("Failed to start actor system: %v", err)
	// }

	// sigChan := make(chan os.Signal, 1)
	// signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// select {
	// case <-sigChan:
	// 	log.Println("Received shutdown signal")
	// case <-time.After(cfg.Algorithm.Timeout):
	// 	log.Println("Algorithm execution timeout")
	// }

	// log.Println("Shutting down...")
	// system.Shutdown()
	// log.Println("Shutdown complete")
}

func loadGraphData(dataPath string) ([]graph.Edge, int, error) {
	edges, err := graph.ReadEdgesFromCSV(dataPath)
	if err != nil {
		return nil, 0, err
	}

	totalWeight := 0
	for _, edge := range edges {
		totalWeight += edge.W
	}
	totalWeight = totalWeight / 2
	log.Printf("Total graph weight: %d", totalWeight)

	return edges, totalWeight, nil
}

func registerPeerActors(provider *cluster.SimpleProvider, selfMachineID string) error {
	configsDir := "configs"

	files, err := ioutil.ReadDir(configsDir)
	if err != nil {
		return fmt.Errorf("failed to read configs directory: %v", err)
	}

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".yaml") && !strings.HasSuffix(file.Name(), ".yml") {
			continue
		}

		if strings.Contains(file.Name(), "coordinator") {
			log.Printf("Skipping coordinator configuration: %s", file.Name())
			continue
		}

		configPath := filepath.Join(configsDir, file.Name())

		peerCfg, err := config.LoadConfig(configPath)
		if err != nil {
			return fmt.Errorf("failed to load peer config %s: %v", configPath, err)
		}

		if peerCfg.MachineID == selfMachineID {
			log.Printf("Skipping self configuration: %s", file.Name())
			continue
		}

		log.Printf("Registering actors from peer node: %s", peerCfg.MachineID)

		for i := 0; i < peerCfg.Actors.Aggregators; i++ {
			aggregatorPID := actor.NewPID(peerCfg.MachineID, fmt.Sprintf("aggregator-%d", i))
			if err := provider.RegisterActor(actor.AggregatorType, aggregatorPID); err != nil {
				return fmt.Errorf("failed to register peer aggregator %s: %v", aggregatorPID, err)
			} else {
				log.Printf("Registered peer aggregator: %s", aggregatorPID)
			}
		}

		for i := 0; i < peerCfg.Actors.Partitions; i++ {
			partitionPID := actor.NewPID(peerCfg.MachineID, fmt.Sprintf("partition-%d", i))
			if err := provider.RegisterActor(actor.PartitionType, partitionPID); err != nil {
				return fmt.Errorf("failed to register peer partition %s: %v", partitionPID, err)
			} else {
				log.Printf("Registered peer partition: %s", partitionPID)
			}
		}
	}

	return nil
}
