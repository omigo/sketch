package sketch

import (
	"testing"

	"innotechx.com/go-common/misc/random"
)

func TestWidthDepth(t *testing.T) {
	cases := []struct {
		errorRatio   float64
		uncertainty  float64
		width, depth uint32
	}{
		{0, 0, 2000, 10}, // error eratio, error certainty
		{1e-1, 1e-1, 2e1, 4},
		{1e-2, 1e-2, 2e2, 7},
		{1e-3, 1e-3, 2e3, 10},
		{1e-4, 1e-4, 2e4, 14},
		{1e-5, 1e-5, 2e5, 17},
		{1e-6, 1e-6, 2e6, 20},
		{1e-7, 1e-7, 2e7, 24},
		{1e-8, 1e-8, 2e8, 27},
		{1.0 / 7e6, 1e-3, 14e6 + 1, 10},
		{1.0 / 2e8, 1e-3, 4e8, 10},
	}

	for i, c := range cases {
		width, depth := WidthDepth(c.errorRatio, c.uncertainty)
		if width != c.width || depth != c.depth {
			t.Errorf("%d: %+v got width %d and depth %d", i, c, width, depth)
			continue
		}
		// t.Logf("%d: %+v => width %d and depth %d", i, c, width, depth)
	}
}

func TestSketch(t *testing.T) {
	cases := []struct {
		dat   []byte
		cnt   Type
		times Type
	}{
		{[]byte("notfound"), 1, 0},
		{[]byte("hello"), 1, 1},
		{[]byte("count"), 2, 1},
		{[]byte("min"), 2, 2},
		{[]byte("world"), 3, 1},
		{[]byte("antispam"), 3, 7},
		{[]byte("cheatcheat"), 10, 1},
		{[]byte("tigger"), 2, 34},
		{[]byte("flow"), 6, 39},
		{[]byte("miss"), 5, 81},
		{[]byte("ohoh"), 3, 1},
		{[]byte("haha"), 2, 9},
	}

	sk := New(WidthDepth(1.0/float64(len(cases)), 0.001))
	t.Log(sk.String())
	for _, c := range cases {
		for j := Type(0); j < c.times; j++ {
			sk.Add(c.dat, c.cnt)
		}
	}

	for i, c := range cases {
		expected := c.cnt * c.times
		got := sk.Query(c.dat)
		if expected != got {
			t.Logf("%d %s got %d, expect %d", i, c.dat, got, expected)
		}
	}
}

func BenchmarkIncr(b *testing.B) {
	sk := New(WidthDepth(1.0/1e6, 0.001))
	b.Log(sk.String())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bs := random.Bytes(16)
		sk.Incr(bs)
	}
}

func BenchmarkQuery(b *testing.B) {
	sk := New(WidthDepth(1.0/1e6, 0.001))
	b.Log(sk.String())
	for i := 0; i < b.N; i++ {
		bs := random.Bytes(16)
		sk.Incr(bs)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bs := random.Bytes(16)
		sk.Query(bs)
	}
}
