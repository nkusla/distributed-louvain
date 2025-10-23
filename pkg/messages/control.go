package messages

type AlgorithmComplete struct {
	FinalModularity float64 `json:"final_modularity"`
	Iterations      int `json:"iterations"`
}
func (m *AlgorithmComplete) Type() string { return "AlgorithmComplete" }

type Shutdown struct{}
func (m *Shutdown) Type() string { return "Shutdown" }