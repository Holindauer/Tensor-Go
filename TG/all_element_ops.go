package TG

// all_elements.go contains operations performed over all elements of a tensor

import (
	"math"
	"sync"
)

//=========================================================================================================== Operations on All Elements That Return a Single Value

// AllOperation is an interface representing an operation applied to all elements of a tensor.
type AllOperation interface {
	Apply(*Tensor, int, int) float64  // Apply performs the operation on a chunk of the tensor's data.
	CombineResults([]float64) float64 // CombineResults combines the results from all chunks.
}

// AllOperation applies a specified operation to all elements of the tensor.
//
// This function performs computations on all elements of the tensor, utilizing concurrency to
// speed up the process. The results from all chunks of data processed concurrently are then
// combined to produce the final result of the operation.
//
// The function takes one parameter:
//   - op: an AllOperation interface that defines the specific operation to be applied.
//     The operation is applied to chunks of the tensor's data and results are combined to produce a final result.
//
// It returns a float64 that is the result of applying the operation.
func (t *Tensor) AllOperation(op AllOperation) float64 {
	var wg sync.WaitGroup
	var mutex = &sync.Mutex{}

	numGoroutines := 4
	chunkSize := len(t.Data) / numGoroutines
	results := make([]float64, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)

		start := i * chunkSize
		end := start + chunkSize

		if i == numGoroutines-1 {
			end = len(t.Data) // Ensure the last chunk includes any remaining elements
		}

		go func(i int, start, end int) {
			defer wg.Done()

			chunkResult := op.Apply(t, start, end) // Apply the operation to the chunk of data.

			mutex.Lock()
			results[i] = chunkResult
			mutex.Unlock()
		}(i, start, end)
	}
	wg.Wait()

	return op.CombineResults(results) // Combine the results from all chunks.
}

//=========================================================================================================== Summation on All Elements

// SumAllOperation represents a summation operation over the entire tensor.
type SumAllOperation struct{}

// Apply performs the summation on a chunk of the tensor's data for a go routine.
func (s SumAllOperation) Apply(A *Tensor, start, end int) float64 {
	var sum float64
	for i := start; i < end; i++ { // sum the elements in the goroutine chunk
		sum += A.Data[i]
	}
	return sum
}

// CombineResults combines the summation results from each chunk of a go routine.
func (s SumAllOperation) CombineResults(results []float64) float64 {
	var sum float64
	for _, v := range results { // sum the results from each goroutine chunk
		sum += v
	}
	return sum
}

// Sum_All() calculates the sum of all elements in a tensor. It accepts a Tensor pointer and returns a float64.
func (A *Tensor) Sum_All() float64 {
	sumOp := SumAllOperation{}   // <--- create sum operation
	return A.AllOperation(sumOp) // <--- apply sum operation to all elements
}

//=========================================================================================================== Mean on All Elements

// MeanAllOperation represents a mean calculation operation over the entire tensor.
type MeanAllOperation struct{}

// Apply performs the mean calculation on a chunk of the tensor's data for a go routine.
func (m MeanAllOperation) Apply(A *Tensor, start, end int) float64 {
	sumOp := SumAllOperation{}        // <--- create sum operation
	sum := sumOp.Apply(A, start, end) // <--- sum goroutine chunk
	return sum / float64(end-start)   // <--- avg of chunk
}

// CombineResults combines the mean results from all chunks of a go routine.
func (m MeanAllOperation) CombineResults(results []float64) float64 {
	sumOp := SumAllOperation{}           // <-- create sum operation
	sum := sumOp.CombineResults(results) // <-- combine sum results from all chunks
	return sum / float64(len(results))   // <-- return avg of sum
}

func (A *Tensor) Mean_All() float64 {
	meanOp := MeanAllOperation{}  // <-- create mean operation
	return A.AllOperation(meanOp) // <-- apply mean operation to all elements
}

//=========================================================================================================== Variance on All Elements

// VarAllOperation represents a variance calculation operation over the entire tensor.
type VarAllOperation struct {
	mean float64
}

// Apply is a method of VarAllOperation that performs the variance calculation on a chunk of the tensor's data for a go routine.
func (v VarAllOperation) Apply(t *Tensor, start, end int) float64 {
	var variance float64
	for i := start; i < end; i++ {
		diff := t.Data[i] - v.mean // <--- var definition: sum((x - mean)^2) / n
		variance += diff * diff    // <--- Appy() performs: (x - mean)^2 for each x
	}
	return variance
}

// CombineResults combines the variance results from all chunks.
func (v VarAllOperation) CombineResults(results []float64) float64 {
	sumOp := SumAllOperation{}           // <-- create sum operation
	sum := sumOp.CombineResults(results) // <-- combine sum results from all chunks
	return sum / float64(len(results))   // <-- return avg of sum
}

func (A *Tensor) Var_All() float64 {
	mean := A.Mean_All()                 // <-- calculate mean
	varOp := VarAllOperation{mean: mean} // <-- pass mean to variance operation
	return A.AllOperation(varOp)         // <-- apply variance operation to all elements
}

//=========================================================================================================== Standard Deviation on All Elements

func (A *Tensor) Std_All() float64 {
	varOp := VarAllOperation{mean: A.Mean_All()} // <-- pass mean to variance operation
	variance := A.AllOperation(varOp)            // <-- calculate variance of all elements
	return math.Sqrt(variance)                   // <-- return sqrt(variance)
}
