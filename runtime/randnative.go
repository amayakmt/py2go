package runtime

import (
	"math/big"
)

// MT is a Mersenne Twister, MT19937, seeded and read the way CPython's
// _random module does, so that a seeded program draws the same numbers.
type MT struct {
	state [624]uint32
	index int
}

func (m *MT) PyClassName() string { return "Random" }

func (m *MT) initGenrand(s uint32) {
	m.state[0] = s
	for i := 1; i < 624; i++ {
		m.state[i] = 1812433253*(m.state[i-1]^(m.state[i-1]>>30)) + uint32(i)
	}
	m.index = 624
}

func (m *MT) initByArray(key []uint32) {
	m.initGenrand(19650218)
	i, j := 1, 0
	k := 624
	if len(key) > k {
		k = len(key)
	}
	for ; k > 0; k-- {
		m.state[i] = (m.state[i] ^ ((m.state[i-1] ^ (m.state[i-1] >> 30)) * 1664525)) + key[j] + uint32(j)
		i++
		j++
		if i >= 624 {
			m.state[0] = m.state[623]
			i = 1
		}
		if j >= len(key) {
			j = 0
		}
	}
	for k = 623; k > 0; k-- {
		m.state[i] = (m.state[i] ^ ((m.state[i-1] ^ (m.state[i-1] >> 30)) * 1566083941)) - uint32(i)
		i++
		if i >= 624 {
			m.state[0] = m.state[623]
			i = 1
		}
	}
	m.state[0] = 0x80000000
	m.index = 624
}

func (m *MT) genrand() uint32 {
	if m.index >= 624 {
		const upper, lower = 0x80000000, 0x7fffffff
		mag := [2]uint32{0, 0x9908b0df}
		var kk int
		for kk = 0; kk < 624-397; kk++ {
			y := (m.state[kk] & upper) | (m.state[kk+1] & lower)
			m.state[kk] = m.state[kk+397] ^ (y >> 1) ^ mag[y&1]
		}
		for ; kk < 623; kk++ {
			y := (m.state[kk] & upper) | (m.state[kk+1] & lower)
			m.state[kk] = m.state[kk+397-624] ^ (y >> 1) ^ mag[y&1]
		}
		y := (m.state[623] & upper) | (m.state[0] & lower)
		m.state[623] = m.state[396] ^ (y >> 1) ^ mag[y&1]
		m.index = 0
	}
	y := m.state[m.index]
	m.index++
	y ^= y >> 11
	y ^= (y << 7) & 0x9d2c5680
	y ^= (y << 15) & 0xefc60000
	y ^= y >> 18
	return y
}

func (m *MT) random() float64 {
	a, b := m.genrand()>>5, m.genrand()>>6
	return (float64(a)*67108864.0 + float64(b)) * (1.0 / 9007199254740992.0)
}

// getrandbits is k random bits as an int.
func (m *MT) getrandbits(k int) any {
	if k <= 0 {
		return int64(0)
	}
	if k <= 32 {
		return int64(m.genrand() >> uint(32-k))
	}
	words := (k-1)/32 + 1
	buf := make([]byte, words*4)
	for i := 0; i < words; i++ {
		r := m.genrand()
		if k < 32 {
			r >>= uint(32 - k)
		}
		buf[i*4], buf[i*4+1], buf[i*4+2], buf[i*4+3] = byte(r), byte(r>>8), byte(r>>16), byte(r>>24)
		k -= 32
	}
	// Little-endian words: reverse into big-endian for big.Int.
	for i, j := 0, len(buf)-1; i < j; i, j = i+1, j-1 {
		buf[i], buf[j] = buf[j], buf[i]
	}
	return NormBig(new(big.Int).SetBytes(buf))
}

func init() {
	RegisterNative("random.mt_new", func(args []any, kw []KW) any {
		m := &MT{}
		m.initByArray([]uint32{0})
		return m
	})
	RegisterNative("random.mt_seed", func(args []any, kw []KW) any {
		m := Unbox(args[0]).(*MT)
		n := new(big.Int).Abs(intLike(args[1]))
		var key []uint32
		if n.Sign() == 0 {
			key = []uint32{0}
		}
		for n.Sign() > 0 {
			key = append(key, uint32(new(big.Int).And(n, big.NewInt(0xffffffff)).Uint64()))
			n.Rsh(n, 32)
		}
		m.initByArray(key)
		return nil
	})
	RegisterNative("random.mt_random", func(args []any, kw []KW) any { return Unbox(args[0]).(*MT).random() })
	RegisterNative("random.mt_bits", func(args []any, kw []KW) any {
		return Unbox(args[0]).(*MT).getrandbits(int(DynInt(args[1])))
	})
	RegisterNative("random.mt_getstate", func(args []any, kw []KW) any {
		m := Unbox(args[0]).(*MT)
		items := make([]any, 0, 625)
		for _, v := range m.state {
			items = append(items, int64(v))
		}
		items = append(items, int64(m.index))
		return NewVarTuple(items...)
	})
	RegisterNative("random.mt_setstate", func(args []any, kw []KW) any {
		m := Unbox(args[0]).(*MT)
		i := 0
		for c := DynIter(args[1]).Iter(); c.Next(); i++ {
			v := DynInt(c.Value())
			if i < 624 {
				m.state[i] = uint32(v)
			} else if i == 624 {
				m.index = int(v)
			}
		}
		return nil
	})
}
