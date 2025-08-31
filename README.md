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
type Item struct{ ID int }

func main() {
    in := make(chan chanseq.Seq[Item])
    out := chanseq.ReorderByIndex(in)

    // Simulate out-of-order production.
    go func() {
        defer close(in)
        // index 0 arrives late
        go func() {
            time.Sleep(50 * time.Millisecond)
            v := Item{ID: 0}
            in <- chanseq.Seq[Item]{Index: 0, Val: &v}
        }()
        // index 1 arrives early
        go func() {
            v := Item{ID: 1}
            in <- chanseq.Seq[Item]{Index: 1, Val: &v}
        }()
        // index 2 is “missing” (skipped)
        in <- chanseq.Seq[Item]{Index: 2, Val: nil}
        // index 3 arrives
        v3 := Item{ID: 3}
        in <- chanseq.Seq[Item]{Index: 3, Val: &v3}
    }()

    // Output (in order): 0, 1, 3
    for v := range out {
        fmt.Println(v.ID)
    }
}
```
