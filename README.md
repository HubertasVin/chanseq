# chanseq

Stream reordering for concurrent pipelines in Go.  
Given items tagged with their original `Index`, `ReorderByIndex` preserves the input order while still streaming results as soon as a contiguous prefix is complete.

## Install

```sh
go get github.com/HubertasVin/chanseq@latest
```

## API

```go
type Seq[T any] struct {
    Index int  // original position
    Val   *T   // nil means “no item for this index”
}

func ReorderByIndex[T any](in <-chan Seq[T]) <-chan T
```
- Emits values in ascending Index order.
- Skips indices where Val == nil.
- Starts its own goroutine; the returned channel is closed when done.

## Example

```go
func main() {
	in := make(chan chanseq.Seq[int])
	out := chanseq.ReorderByIndex(in)

	// Simulate out-of-order production.
	go func() {
		var wg sync.WaitGroup

		wg.Add(1)
		go func() {
			defer wg.Done()
			time.Sleep(50 * time.Millisecond)

			// Send index 0 after a delay, simulating late arrival
			val0 := 0
			in <- chanseq.Seq[int]{Index: 0, Val: &val0}
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()

			// Send index 1 immediately, simulating early arrival
			val1 := 1
			in <- chanseq.Seq[int]{Index: 1, Val: &val1}
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()

			// Send index 2 with Val == nil, simulating a missing item
			in <- chanseq.Seq[int]{Index: 2, Val: nil}
		}()

		wg.Wait()
		close(in)
	}()

	// Output (in order): 0, 1
	for i := range out {
		fmt.Println(i)
	}
}
```
