package sketch

import (
	"testing"

	"innotechx.com/go-common/misc/random"
)

func TestWidthDepth(t *testing.T) {
	cases := []struct {
		eratio, certainty float64
		width, depth      uint32
	}{
		{0.001, 0.999, 2000, 10},
		{0, 1, 20000, 10}, // error eratio, error certainty
		{0.001, 0.99, 2000, 7},
		{0.0001, 0.9999, 20000, 14},
		{0.0001, 0.99999, 20000, 17},
		{0.0001, 0.999999, 20000, 20},
		{1.0 / 7e6, 0.999, 14000001, 10},
		{1.0 / 2e8, 0.999, 400000000, 10},
		{1.0 / 1e6, 0.999, 2000000, 10},
	}

	for i, c := range cases {
		width, depth := WidthDepth(c.eratio, c.certainty)
		if width != c.width || depth != c.depth {
			t.Errorf("%d: %+v got width %d and depth %d", i, c, width, depth)
		}
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

	sk := New(WidthDepth(1.0/float64(len(cases)), 0.999))
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
	sk := New(WidthDepth(1.0/1e6, 0.999))
	b.Log(sk.String())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bs := random.Bytes(16)
		sk.Incr(bs)
	}
}

func BenchmarkQuery(b *testing.B) {
	sk := New(WidthDepth(1.0/1e6, 0.999))
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
