package ndni

/*
#include "../csrc/ndn/tlv-element.h"
*/
import "C"
import (
	"fmt"

	"github.com/usnistgov/ndn-dpdk/dpdk/pktmbuf"
	"github.com/usnistgov/ndn-dpdk/ndn/an"
)

type TlvElement struct {
	c C.TlvElement
}

// Decode a TLV element.
func ParseTlvElement(it pktmbuf.PacketIterator) (ele TlvElement, e error) {
	if res := C.TlvElement_Decode(&ele.c, (*C.MbufLoc)(it.GetPtr()), C.TtInvalid); res != C.NdnErrOK {
		return TlvElement{}, NdnError(res)
	}
	return ele, nil
}

// Get total length.
func (ele *TlvElement) Len() int {
	return int(ele.c.size)
}

// Get TLV-TYPE.
func (ele *TlvElement) GetType() an.TlvType {
	return an.TlvType(ele.c._type)
}

// Get TLV-LENGTH.
func (ele *TlvElement) GetLength() int {
	return int(ele.c.length)
}

// Get TLV-VALUE.
func (ele *TlvElement) GetValue() (v TlvBytes) {
	var it pktmbuf.PacketIterator
	C.TlvElement_MakeValueDecoder(&ele.c, (*C.MbufLoc)(it.GetPtr()))

	v = make(TlvBytes, ele.GetLength())
	it.Read(([]byte)(v)) // will always succeed on valid TLV
	return v
}

// Interpret TLV-VALUE as NonNegativeInteger.
func (ele *TlvElement) ReadNonNegativeInteger() (n uint64, ok bool) {
	var v C.uint64_t
	res := C.TlvElement_ReadNonNegativeInteger(&ele.c, &v)
	return uint64(v), res == C.NdnErrOK
}

func (ele *TlvElement) String() string {
	return fmt.Sprintf("%v(%d) %v", ele.GetType(), ele.GetLength(), ele.GetValue())
}
