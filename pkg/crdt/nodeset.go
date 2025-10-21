package crdt

type NodeTransition struct {
	NodeID          int
	CommunityID     int
	ModularityDelta float64
}

// NodeSet is modified Max-Set CRDT for conflict resolution
type NodeSet struct {
	transitions map[int]*NodeTransition // nodeID -> best transition with highest modularity delta
}

func NewNodeSet() *NodeSet {
	return &NodeSet{
		transitions: make(map[int]*NodeTransition),
	}
}

func (n *NodeSet) Add(nodeID, communityID int, delta float64) {
	// Symmetry breaking: only allow moves where nodeID > communityID
	if nodeID <= communityID {
		communityID, nodeID = nodeID, communityID
	}

	existing, exists := n.transitions[nodeID]

	if !exists || delta > existing.ModularityDelta {
		n.transitions[nodeID] = &NodeTransition{
			NodeID:          nodeID,
			CommunityID:     communityID,
			ModularityDelta: delta,
		}
	}
}

func (n *NodeSet) Get(nodeID int) (*NodeTransition, bool) {
	transition, exists := n.transitions[nodeID]
	return transition, exists
}

func (n *NodeSet) GetAll() []*NodeTransition {
	transitions := make([]*NodeTransition, 0, len(n.transitions))
	for _, transition := range n.transitions {
		transitions = append(transitions, transition)
	}
	return transitions
}

func (n *NodeSet) Merge(other *NodeSet) {
	for nodeID, transition := range other.transitions {
		existing, exists := n.transitions[nodeID]

		if !exists || transition.ModularityDelta > existing.ModularityDelta {
			n.transitions[nodeID] = &NodeTransition{
				NodeID:          transition.NodeID,
				CommunityID:     transition.CommunityID,
				ModularityDelta: transition.ModularityDelta,
			}
		}
	}
}

func (n *NodeSet) Clear() {
	n.transitions = make(map[int]*NodeTransition)
}

func (n *NodeSet) Size() int {
	return len(n.transitions)
}

func (n *NodeSet) Clone() *NodeSet {
	clone := NewNodeSet()
	for nodeID, transition := range n.transitions {
		clone.transitions[nodeID] = &NodeTransition{
			NodeID:          transition.NodeID,
			CommunityID:     transition.CommunityID,
			ModularityDelta: transition.ModularityDelta,
		}
	}
	return clone
}
