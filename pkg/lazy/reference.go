package lazy

// Reference evaluates recursively and mutates a private heap. It is independent
// of the explicit continuation machinery and used for bounded differential tests.
func Reference(image Image, limit int) (Word, []Word) {
	heap := append([]Word{}, image.Nodes...)
	var eval func(uint16, int, int) Word
	eval = func(a uint16, depth, hops int) Word {
		if int(a) >= len(heap) {
			return Fault(AddressFault)
		}
		w := heap[a]
		if !w.Canonical() {
			return Fault(TypeFault)
		}
		switch w.Tag() {
		case Int, Error:
			return w
		case Ind:
			if hops == IndirectionLimit {
				return Fault(IndirectionCycle)
			}
			return eval(w.A(), depth, hops+1)
		case Blackhole:
			return Fault(CyclicThunk)
		case Thunk:
			if depth == limit {
				return Fault(StackOverflow)
			}
			heap[a] = Node(Blackhole, 0, 0)
			v := eval(w.A(), depth+1, 0)
			heap[a] = v
			return v
		case Add, Mul:
			if depth == limit {
				return Fault(StackOverflow)
			}
			left := eval(w.A(), depth+1, 0)
			if left.Tag() == Error {
				return left
			}
			right := eval(w.B(), depth+1, 0)
			return Evaluate(w.Tag(), left, right)
		default:
			return Fault(TypeFault)
		}
	}
	return eval(image.Root, 0, 0), heap
}
