package crdt

import (
	"encoding/json"
)

type NodeTransitionJSON struct {
	NodeID          int     `json:"node_id"`
	CommunityID     int     `json:"community_id"`
	ModularityDelta float64 `json:"modularity_delta"`
}

type NodeSetJSON struct {
	Transitions []NodeTransitionJSON `json:"transitions"`
}

func (ns *NodeSet) MarshalJSON() ([]byte, error) {

	transitions := make([]NodeTransitionJSON, 0, len(ns.transitions))
	for _, transition := range ns.transitions {
		transitions = append(transitions, NodeTransitionJSON{
			NodeID:          transition.NodeID,
			CommunityID:     transition.CommunityID,
			ModularityDelta: transition.ModularityDelta,
		})
	}

	nodeSetJSON := NodeSetJSON{
		Transitions: transitions,
	}

	return json.Marshal(nodeSetJSON)
}

func (ns *NodeSet) UnmarshalJSON(data []byte) error {
	var nodeSetJSON NodeSetJSON
	if err := json.Unmarshal(data, &nodeSetJSON); err != nil {
		return err
	}

	ns.transitions = make(map[int]*NodeTransition)
	for _, transitionJSON := range nodeSetJSON.Transitions {
		transition := &NodeTransition{
			NodeID:          transitionJSON.NodeID,
			CommunityID:     transitionJSON.CommunityID,
			ModularityDelta: transitionJSON.ModularityDelta,
		}
		ns.transitions[transition.NodeID] = transition
	}

	return nil
}

func (nt *NodeTransition) MarshalJSON() ([]byte, error) {
	transitionJSON := NodeTransitionJSON{
		NodeID:          nt.NodeID,
		CommunityID:     nt.CommunityID,
		ModularityDelta: nt.ModularityDelta,
	}
	return json.Marshal(transitionJSON)
}

func (nt *NodeTransition) UnmarshalJSON(data []byte) error {
	var transitionJSON NodeTransitionJSON
	if err := json.Unmarshal(data, &transitionJSON); err != nil {
		return err
	}

	nt.NodeID = transitionJSON.NodeID
	nt.CommunityID = transitionJSON.CommunityID
	nt.ModularityDelta = transitionJSON.ModularityDelta

	return nil
}
