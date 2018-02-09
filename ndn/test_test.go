package ndn_test

import (
	"os"
	"testing"

	"ndn-dpdk/dpdk/dpdktestenv"
	"ndn-dpdk/ndn"
)

func TestMain(m *testing.M) {
	dpdktestenv.MakeDirectMp(255, 0, 2000)

	os.Exit(m.Run())
}

var makeAR = dpdktestenv.MakeAR
var packetFromHex = dpdktestenv.PacketFromHex

func TlvBytesFromHex(input string) ndn.TlvBytes {
	return ndn.TlvBytes(dpdktestenv.PacketBytesFromHex(input))
}
