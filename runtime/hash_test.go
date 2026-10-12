package runtime

import "testing"

func TestHashStr(t *testing.T) {
	// Expected values come from CPython 3.13 run with each PYTHONHASHSEED.
	strs := []string{"abc", "", "a", "é", "日本", "😀x", "hello world", "xxxxxxxxxxxxxxxxxxxx"}
	want := map[string][]int64{
		"0":     {-4594863902769663758, 0, 4644417185603328019, 6047309291227476195, 6243316497235261705, -8926728262118538918, -5642461784034726774, 3641836881806154204},
		"1":     {-4667308735975688587, 0, -3012895188637184397, 8133168473585362875, 8399330583414457596, 8904275969526274476, 4697380503286254714, 2569032445979264741},
		"42":    {3869580338025362921, 0, -123207753977932514, 5653850369683463245, -7192152084576530731, -7412435959675129227, -3893157060480277979, -1433231438062637585},
		"12345": {2962436275791859092, 0, -8961251314296359281, 6247654839523334371, 7239199481113287272, -5581016787913457232, 3184199782645080147, 4302119244793870670},
	}
	for seed, hashes := range want {
		k0, k1 := hashKey(seed)
		for i, s := range strs {
			if got := hashStrWith(k0, k1, s); got != hashes[i] {
				t.Errorf("PYTHONHASHSEED=%s hash(%q) = %d, want %d", seed, s, got, hashes[i])
			}
		}
	}
}

func TestHashNumbersAndTuples(t *testing.T) {
	ints := map[int64]int64{
		-1: -2, -2: -2, 0: 0, 1: 1, 1<<61 - 1: 0, 1 << 61: 1, -(1 << 63): -4, 1<<63 - 1: 3,
	}
	for n, want := range ints {
		if got := Hash(n); got != want {
			t.Errorf("hash(%d) = %d, want %d", n, got, want)
		}
	}
	if Hash(true) != 1 || Hash(false) != 0 {
		t.Error("hash(True) is 1 and hash(False) is 0")
	}
	if got := Hash(Tup2[int64, int64](1, 2)); got != -3550055125485641917 {
		t.Errorf("hash((1, 2)) = %d", got)
	}
	if got := Hash(Tup2[int64, Tuple2[int64, int64]](1, Tup2[int64, int64](2, 3))); got != 7267574591690527098 {
		t.Errorf("hash((1, (2, 3))) = %d", got)
	}
}
