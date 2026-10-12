package runtime

// Tuples read from a sequence of objects: the key of a dict that is looked up
// with a tuple whose items are not known statically.

type anyTuple interface{ setFromSeq(s Seq) bool }

func (t *Tuple1[A]) setFromSeq(s Seq) bool {
	if s.Len() != 1 {
		return false
	}
	f0, ok0 := convOpt[A](s.At(0))
	if !(ok0) {
		return false
	}
	*t = Tuple1[A]{F0: f0}
	return true
}

func (t *Tuple2[A, B]) setFromSeq(s Seq) bool {
	if s.Len() != 2 {
		return false
	}
	f0, ok0 := convOpt[A](s.At(0))
	f1, ok1 := convOpt[B](s.At(1))
	if !(ok0 && ok1) {
		return false
	}
	*t = Tuple2[A, B]{F0: f0, F1: f1}
	return true
}

func (t *Tuple3[A, B, C]) setFromSeq(s Seq) bool {
	if s.Len() != 3 {
		return false
	}
	f0, ok0 := convOpt[A](s.At(0))
	f1, ok1 := convOpt[B](s.At(1))
	f2, ok2 := convOpt[C](s.At(2))
	if !(ok0 && ok1 && ok2) {
		return false
	}
	*t = Tuple3[A, B, C]{F0: f0, F1: f1, F2: f2}
	return true
}

func (t *Tuple4[A, B, C, D]) setFromSeq(s Seq) bool {
	if s.Len() != 4 {
		return false
	}
	f0, ok0 := convOpt[A](s.At(0))
	f1, ok1 := convOpt[B](s.At(1))
	f2, ok2 := convOpt[C](s.At(2))
	f3, ok3 := convOpt[D](s.At(3))
	if !(ok0 && ok1 && ok2 && ok3) {
		return false
	}
	*t = Tuple4[A, B, C, D]{F0: f0, F1: f1, F2: f2, F3: f3}
	return true
}

func (t *Tuple5[A, B, C, D, E]) setFromSeq(s Seq) bool {
	if s.Len() != 5 {
		return false
	}
	f0, ok0 := convOpt[A](s.At(0))
	f1, ok1 := convOpt[B](s.At(1))
	f2, ok2 := convOpt[C](s.At(2))
	f3, ok3 := convOpt[D](s.At(3))
	f4, ok4 := convOpt[E](s.At(4))
	if !(ok0 && ok1 && ok2 && ok3 && ok4) {
		return false
	}
	*t = Tuple5[A, B, C, D, E]{F0: f0, F1: f1, F2: f2, F3: f3, F4: f4}
	return true
}

func (t *Tuple6[A, B, C, D, E, F]) setFromSeq(s Seq) bool {
	if s.Len() != 6 {
		return false
	}
	f0, ok0 := convOpt[A](s.At(0))
	f1, ok1 := convOpt[B](s.At(1))
	f2, ok2 := convOpt[C](s.At(2))
	f3, ok3 := convOpt[D](s.At(3))
	f4, ok4 := convOpt[E](s.At(4))
	f5, ok5 := convOpt[F](s.At(5))
	if !(ok0 && ok1 && ok2 && ok3 && ok4 && ok5) {
		return false
	}
	*t = Tuple6[A, B, C, D, E, F]{F0: f0, F1: f1, F2: f2, F3: f3, F4: f4, F5: f5}
	return true
}

func (t *VarTuple[T]) setFromSeq(s Seq) bool {
	items := make([]T, s.Len())
	for i := range items {
		x, ok := convOpt[T](s.At(int64(i)))
		if !ok {
			return false
		}
		items[i] = x
	}
	*t = NewVarTuple(items...)
	return true
}
