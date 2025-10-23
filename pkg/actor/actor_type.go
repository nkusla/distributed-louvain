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
		return "coordinator"
	case AggregatorType:
		return "aggregator"
	case PartitionType:
		return "partition"
	default:
		return "Unknown"
	}
}
