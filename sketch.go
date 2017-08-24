package sketch

import (
	"fmt"
	"math"
	"unsafe"

	"github.com/arstd/log"
	"github.com/spaolacci/murmur3"
)

type Type uint32

const Max = math.MaxUint32

type Sketch struct {
	width uint32
	depth uint32
	count [][]Type
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
		log.Warnf("eratio %g not in [1e-9,0.1], use default %g", errorRatio, defaultErrorRatio)
		errorRatio = defaultErrorRatio
	}
	if uncertainty < 1.0/1e9 || uncertainty > 0.1 {
		log.Warnf("certainty %g not in [1e-9,0.1], use default %g", uncertainty, defaultUncertainty)
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
		count: make([][]Type, depth),
	}
	for i := uint32(0); i < depth; i++ {
		sk.count[i] = make([]Type, width)
	}
	return sk
}

func (sk *Sketch) Width() uint32 { return sk.width }
func (sk *Sketch) Depth() uint32 { return sk.depth }

func (sk *Sketch) String() string {
	return fmt.Sprintf("Count-Min Sketch(%p): width=%d, depth=%d, mem=%d",
		sk, sk.width, sk.depth, int64(sk.width)*int64(sk.depth)*int64(unsafe.Sizeof(sk.count[0][0])))
}

func (sk *Sketch) Incr(dat []byte) (min Type) {
	return sk.Add(dat, 1)
}

func (sk *Sketch) Add(dat []byte, cnt Type) (min Type) {
	min = Max
	// reference: https://github.com/addthis/stream-lib/blob/master/src/main/java/com/clearspring/analytics/stream/membership/Filter.java
	hash1 := murmur3.Sum32WithSeed(dat, 0)
	hash2 := murmur3.Sum32WithSeed(dat, hash1)
	for i := uint32(0); i < sk.depth; i++ {
		pos := (hash1 + i*hash2) % sk.width
		v := sk.count[i][pos]
		v += cnt
		sk.count[i][pos] = v
		if min > v {
			min = v
		}
	}
	return min
}

func (sk *Sketch) Query(dat []byte) (min Type) {
	min = Max
	hash1 := murmur3.Sum32WithSeed(dat, 0)
	hash2 := murmur3.Sum32WithSeed(dat, hash1)
	for i := uint32(0); i < sk.depth; i++ {
		pos := (hash1 + i*hash2) % sk.width
		v := sk.count[i][pos]
		if min > v {
			min = v
		}
	}
	return min
}
