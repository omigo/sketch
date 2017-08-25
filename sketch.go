package sketch

import (
	"fmt"
	"math"
	"sync"
	"unsafe"

	"github.com/spaolacci/murmur3"
)

type CountType uint8

const Max = ^(CountType(0))

type Sketch struct {
	width uint32
	depth uint32
	count [][]CountType
	mutex sync.RWMutex
}

// WidthDepth returns width and depth by formula in paper.
// Paper:
//   Approximating Data with the Count-Min Data Structure
//   Graham Cormode S. Muthukrishnan
//   August 12, 2011
// https://cs.stackexchange.com/questions/44803/what-is-the-correct-way-to-determine-the-width-and-depth-of-a-count-min-sketch
//
// w=⌈2/error⌉
// d=⌈−ln(uncertainty)/ln(2)⌉  # uncertainty = 1 - certainty
func WidthDepth(errorRatio, uncertainty float64) (width, depth uint32) {
	const (
		defaultErrorRatio  = 1.0 / 1e3 // 0.1%
		defaultUncertainty = 1.0 / 1e3 // 0.1%
	)
	if errorRatio < 1.0/1e9 || errorRatio > 0.1 {
		fmt.Printf("error ratio %g not in [1e-9,0.1], use default %g\n", errorRatio, defaultErrorRatio)
		errorRatio = defaultErrorRatio
	}
	if uncertainty < 1.0/1e9 || uncertainty > 0.1 {
		fmt.Printf("certainty %g not in [1e-9,0.1], use default %g\n", uncertainty, defaultUncertainty)
		uncertainty = defaultUncertainty
	}

	width = uint32(math.Ceil(2 / errorRatio))
	depth = uint32(math.Ceil(-math.Log(uncertainty) / math.Log(2)))
	return width, depth
}

func New(width, depth uint32) (sk *Sketch) {
	sk = &Sketch{
		width: width,
		depth: depth,
		count: make([][]CountType, depth),
	}
	for i := uint32(0); i < depth; i++ {
		sk.count[i] = make([]CountType, width)
	}
	return sk
}

func (sk *Sketch) Width() uint32 { return sk.width }
func (sk *Sketch) Depth() uint32 { return sk.depth }

func (sk *Sketch) String() string {
	space := float64(int64(sk.width)*int64(sk.depth)*int64(unsafe.Sizeof(sk.count[0][0]))) / 1e6
	return fmt.Sprintf("Count-Min Sketch(%p): width=%d, depth=%d, mem=%.3fm",
		sk, sk.width, sk.depth, space)
}

func (sk *Sketch) Clear() {
	sk.mutex.Lock()
	for i := uint32(0); i < sk.depth; i++ {
		for j := uint32(0); j < sk.width; j++ {
			sk.count[i][j] = 0
		}
	}
	sk.mutex.Unlock()
}

func (sk *Sketch) Incr(dat []byte) (min CountType) {
	return sk.Add(dat, 1)
}

func (sk *Sketch) Add(dat []byte, cnt CountType) (min CountType) {
	pos := sk.positions(dat)
	min = sk.query(pos)

	min += cnt

	sk.mutex.Lock()
	for i := uint32(0); i < sk.depth; i++ {
		v := sk.count[i][pos[i]]
		if v < min {
			sk.count[i][pos[i]] = min
		}
	}
	sk.mutex.Unlock()

	return min
}

func (sk *Sketch) Query(dat []byte) (min CountType) {
	pos := sk.positions(dat)
	return sk.query(pos)
}

func (sk *Sketch) positions(dat []byte) (pos []uint32) {
	// reference: https://github.com/addthis/stream-lib/blob/master/src/main/java/com/clearspring/analytics/stream/membership/Filter.java
	hash1 := murmur3.Sum32WithSeed(dat, 0)
	hash2 := murmur3.Sum32WithSeed(dat, hash1)
	pos = make([]uint32, sk.depth)
	for i := uint32(0); i < sk.depth; i++ {
		pos[i] = (hash1 + i*hash2) % sk.width
	}
	return pos
}

func (sk *Sketch) query(pos []uint32) (min CountType) {
	min = Max

	sk.mutex.RLock()
	for i := uint32(0); i < sk.depth; i++ {
		v := sk.count[i][pos[i]]
		if min > v {
			min = v
		}
	}
	sk.mutex.RUnlock()

	return min
}
