type AlgorithmComplete struct {
	FinalModularity float64
	Iterations      int
}
func (m *AlgorithmComplete) Type() string { return "AlgorithmComplete" }

type Shutdown struct{}
func (m *Shutdown) Type() string { return "Shutdown" }