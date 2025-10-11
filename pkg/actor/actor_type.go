package actor

type ActorType int

const (
	CoordinatorType ActorType = iota
	AggregatorType
	PartitionType
)

func (at ActorType) String() string {
	switch at {
	case CoordinatorType:
		return "Coordinator"
	case AggregatorType:
		return "Aggregator"
	case PartitionType:
		return "Partition"
	default:
		return "Unknown"
	}
}
