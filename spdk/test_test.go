package spdk_test

import (
	"os"
	"testing"

	"ndn-dpdk/dpdk/dpdktestenv"
	"ndn-dpdk/spdk"
)

func TestMain(m *testing.M) {
	eal := dpdktestenv.InitEal()
	spdk.MustInit(eal, eal.Slaves[0])
	os.Exit(m.Run())
}

var makeAR = dpdktestenv.MakeAR
