package sketch

import (
	"encoding/binary"
	"fmt"
	"sync"
	"testing"
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
		cnt   CountType
		times CountType
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
	fmt.Println(sk.String())
	for _, c := range cases {
		for j := CountType(0); j < c.times; j++ {
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

func TestRace(t *testing.T) {
	sk := New(WidthDepth(0.0001, 0.001))
	fmt.Println(sk.String())

	var wg sync.WaitGroup
	for g := 0; g < 100; g++ {
		wg.Add(2)
		go func() {
			bs := make([]byte, 8)
			for i := 0; i < 10000; i++ {
				binary.BigEndian.PutUint64(bs, uint64(i)*65537)
				sk.Incr(bs)
			}
			wg.Done()
		}()
		go func() {
			bs := make([]byte, 8)
			for i := 0; i < 10000; i++ {
				binary.BigEndian.PutUint64(bs, uint64(i)*65537)
				sk.Query(bs)
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

const nsample = 7e6

func TestErrors(t *testing.T) {
	sk := New(WidthDepth(0.8/nsample, 0.001))
	fmt.Println(sk.String())
	bs := make([]byte, 4)
	for i := uint32(0); i < nsample; i++ {
		binary.BigEndian.PutUint32(bs, i)
		sk.Incr(bs)
	}

	errors := make([]int, 16)
	for i := uint32(0); i < nsample; i++ {
		binary.BigEndian.PutUint32(bs, i)
		v := sk.Query(bs)
		if v != 1 {
			errors[v]++
		}
	}

	for i, e := range errors {
		fmt.Printf("%2d %d\n", i, e)
	}
}

func BenchmarkIncr(b *testing.B) {
	sk := New(WidthDepth(1.0/nsample, 0.001))
	b.Log(sk.String())
	bs := make([]byte, 8)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		binary.BigEndian.PutUint64(bs, uint64(i)*65537)
		sk.Incr(bs)
	}
}

func BenchmarkQuery(b *testing.B) {
	sk := New(WidthDepth(1.0/nsample, 0.001))
	b.Log(sk.String())
	bs := make([]byte, 8)
	for i := 0; i < b.N; i++ {
		binary.BigEndian.PutUint64(bs, uint64(i)*65537)
		sk.Incr(bs)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		binary.BigEndian.PutUint64(bs, uint64(i)*65537)
		sk.Query(bs)
	}
}
