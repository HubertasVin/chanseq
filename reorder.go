package chanseq

import "sort"

// Seq carries a value along with its original index in the sequence.
// If Val is nil, that index is considered "no item" and will be skipped.
type Seq[T any] struct {
	Index int
	Val   *T
}

// ReorderByIndex consumes Seq[T] items from 'in' and emits T values on 'out'
// in ascending Index order, streaming as soon as a contiguous prefix [0..next]
// is satisfied. Missing indices (Val == nil) are skipped.
func ReorderByIndex[T any](in <-chan Seq[T]) <-chan T {
	out := make(chan T, 16)

	go func() {
		defer close(out)

		next := 0               // next expected index
		buf := make(map[int]*T) // out-of-order or pending indices

		for it := range in {
			buf[it.Index] = it.Val
			// Flush any newly satisfied contiguous prefix.
			for {
				v, ok := buf[next]
				if !ok {
					break
				}
				if v != nil {
					out <- *v
				}
				delete(buf, next)
				next++
			}
		}

		// Input ended. Flush any remaining buffered items in index order.
		if len(buf) == 0 {
			return
		}
		keys := make([]int, 0, len(buf))
		for k := range buf {
			keys = append(keys, k)
		}
		sort.Ints(keys)
		for _, k := range keys {
			if v := buf[k]; v != nil {
				out <- *v
			}
		}
	}()

	return out
}
