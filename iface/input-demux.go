package iface

/*
#include "../csrc/iface/input-demux.h"
*/
import "C"

import (
	"unsafe"

	"github.com/usnistgov/ndn-dpdk/container/ndt"
)

// InputDemux is a demultiplexer for incoming packets of one L3 type.
type InputDemux C.InputDemux

// InputDemuxFromPtr converts *C.InputDemux pointer to InputDemux.
func InputDemuxFromPtr(ptr unsafe.Pointer) *InputDemux {
	return (*InputDemux)(ptr)
}

func (demux *InputDemux) ptr() *C.InputDemux {
	return (*C.InputDemux)(demux)
}

// InitDrop configures to drop all packets.
func (demux *InputDemux) InitDrop() {
	C.InputDemux_SetDispatchFunc_(demux.ptr(), C.InputDemux_DispatchDrop)
}

// InitFirst configures to pass all packets to the first and only destination.
func (demux *InputDemux) InitFirst() {
	demux.InitRoundrobin(1)
}

// InitRoundrobin configures to pass all packets to each destination in a round-robin fashion.
func (demux *InputDemux) InitRoundrobin(nDest int) {
	C.InputDemux_SetDispatchRoundrobin_(demux.ptr(), C.uint32_t(nDest))
}

// InitNdt configures to dispatch via NDT loopup.
func (demux *InputDemux) InitNdt(ndt *ndt.Ndt, ndtThread int) {
	demuxC := demux.ptr()
	C.InputDemux_SetDispatchFunc_(demuxC, C.InputDemux_DispatchByNdt)
	demuxC.ndt = (*C.Ndt)(unsafe.Pointer(ndt.Ptr()))
	demuxC.ndtt = demuxC.ndt.threads[ndtThread]
}

// InitToken configures to dispatch according to high 8 bits of PIT token.
func (demux *InputDemux) InitToken() {
	C.InputDemux_SetDispatchFunc_(demux.ptr(), C.InputDemux_DispatchByToken)
}

// SetDest assigns i-th destination.
func (demux *InputDemux) SetDest(i int, q *PktQueue) {
	demux.ptr().dest[i].queue = q.ptr()
}

// InputDemuxDestCounters contains counters of an InputDemux destination.
type InputDemuxDestCounters struct {
	NQueued  uint64
	NDropped uint64
}

// ReadDestCounters returns counters of i-th destination.
func (demux *InputDemux) ReadDestCounters(i int) (cnt InputDemuxDestCounters) {
	c := demux.ptr()
	cnt.NQueued = uint64(c.dest[i].nQueued)
	cnt.NDropped = uint64(c.dest[i].nDropped)
	return cnt
}
