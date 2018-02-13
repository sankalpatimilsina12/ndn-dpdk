package ndn_test

import (
	"strings"
	"testing"

	"ndn-dpdk/dpdk"
	"ndn-dpdk/dpdk/dpdktestenv"
	"ndn-dpdk/ndn"
)

func TestName1Decode(t *testing.T) {
	assert, _ := makeAR(t)

	tests := []struct {
		input     string
		ok        bool
		nComps    int
		hasDigest bool
		str       string
	}{
		{"", false, 0, false, ""},
		{"0700", true, 0, false, "/"},
		{"0714 080141 080142 080100 0801FF 800141 0800 08012E", true, 7, false,
			"/A/B/%00/%FF/128=A/.../...."},
		{"0722 0120(DC6D6840C6FAFB773D583CDBF465661C7B4B968E04ACD4D9015B1C4E53E59D6A)", true, 1, true,
			"/sha256digest=dc6d6840c6fafb773d583cdbf465661c7b4b968e04acd4d9015b1c4e53e59d6a"},
		{"0763 " + strings.Repeat("080141 ", 32) + "080142", true, 33, false,
			strings.Repeat("/A", 32) + "/B"},
		{"0200", false, 0, false, ""},           // bad TLV-TYPE
		{"0704 0102 DDDD", false, 0, false, ""}, // wrong digest length
	}
	for _, tt := range tests {
		pkt := dpdktestenv.PacketFromHex(tt.input)
		defer pkt.Close()
		d := ndn.NewTlvDecodePos(pkt)

		n, e := d.ReadName1()
		if tt.ok {
			if assert.NoError(e, tt.input) {
				assert.Equal(tt.nComps, n.Len(), tt.input)
				assert.Equal(tt.hasDigest, n.HasDigest(), tt.input)
				assert.Equal(tt.str, n.String(), tt.input)
			}
		} else {
			assert.Error(e, tt.input)
		}
	}
}

func TestName1PrefixSize(t *testing.T) {
	assert, require := makeAR(t)

	pkt := dpdktestenv.PacketFromHex("0709 0800 080141 08024243")
	defer pkt.Close()
	d := ndn.NewTlvDecodePos(pkt)

	name, e := d.ReadName1()
	require.NoError(e)

	assert.Equal(3, name.Len())
	assert.Equal(9, name.Size())
	assert.Equal(0, name.GetPrefixSize(0))
	assert.Equal(2, name.GetPrefixSize(1))
	assert.Equal(5, name.GetPrefixSize(2))
	assert.Equal(9, name.GetPrefixSize(3))
}

func TestName1PrefixHash(t *testing.T) {
	assert, require := makeAR(t)

	nameStrs := []string{
		"0700",
		"0702 0200",
		"0702 0800",
		"0704 0800 0800",
		"0703 080141",
		"0705 080141 0800",
		"0707 080141 0800 0800",
		"0706 080141 080141",
		"0703 080142",
		"0704 08024100",
		"0704 08024101",
		"0702 0900",
	}
	pkts := make([]dpdk.Packet, len(nameStrs))
	names := make([]ndn.Name1, len(nameStrs))
	for i, nameStr := range nameStrs {
		pkts[i] = dpdktestenv.PacketFromHex(nameStr)
		defer pkts[i].Close()
		d := ndn.NewTlvDecodePos(pkts[i])

		var e error
		names[i], e = d.ReadName1()
		require.NoError(e, nameStr)
	}

	// (a,i,b,j) means names[a].ComputePrefixHash(i) should equal names[b].ComputePrefixHash(j)
	equalPairs := map[[4]int]bool{
		{2, 1, 3, 1}: true,
		{4, 1, 5, 1}: true,
		{4, 1, 6, 1}: true,
		{4, 1, 7, 1}: true,
		{5, 1, 6, 1}: true,
		{5, 2, 6, 2}: true,
		{5, 1, 7, 1}: true,
		{6, 1, 7, 1}: true,
	}
	for a, nameA := range names {
		for b, nameB := range names {
			if a >= b {
				continue
			}
			assert.Equal(nameA.ComputePrefixHash(0), nameB.ComputePrefixHash(0),
				"%d,%d-%d,%d", a, 0, b, 0)
			for i := 1; i <= nameA.Len(); i++ {
				for j := 1; j <= nameB.Len(); j++ {
					if equalPairs[[4]int{a, i, b, j}] {
						assert.Equal(nameA.ComputePrefixHash(i), nameB.ComputePrefixHash(j),
							"%d,%d-%d,%d", a, i, b, j)
					} else {
						assert.NotEqual(nameA.ComputePrefixHash(i), nameB.ComputePrefixHash(j),
							"%d,%d-%d,%d", a, i, b, j)
					}
				}
			}
		}
	}
}

func TestName1Compare(t *testing.T) {
	assert, require := makeAR(t)

	nameStrs := []string{
		"0700",
		"0702 0200",
		"0702 0800",
		"0704 0800 0800",
		"0703 080141",
		"0705 080141 0800",
		"0707 080141 0800 0800",
		"0706 080141 080141",
		"0703 080142",
		"0704 08024100",
		"0704 08024101",
		"0702 0900",
	}
	pkts := make([]dpdk.Packet, len(nameStrs))
	names := make([]ndn.Name1, len(nameStrs))
	for i, nameStr := range nameStrs {
		pkts[i] = dpdktestenv.PacketFromHex(nameStr)
		defer pkts[i].Close()
		d := ndn.NewTlvDecodePos(pkts[i])

		var e error
		names[i], e = d.ReadName1()
		require.NoError(e, nameStr)
	}

	relTable := [][]int{
		[]int{+0, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
		[]int{+1, +0, -2, -2, -2, -2, -2, -2, -2, -2, -2, -2},
		[]int{+1, +2, +0, -1, -2, -2, -2, -2, -2, -2, -2, -2},
		[]int{+1, +2, +1, +0, -2, -2, -2, -2, -2, -2, -2, -2},
		[]int{+1, +2, +2, +2, +0, -1, -1, -1, -2, -2, -2, -2},
		[]int{+1, +2, +2, +2, +1, +0, -1, -2, -2, -2, -2, -2},
		[]int{+1, +2, +2, +2, +1, +1, +0, -2, -2, -2, -2, -2},
		[]int{+1, +2, +2, +2, +1, +2, +2, +0, -2, -2, -2, -2},
		[]int{+1, +2, +2, +2, +2, +2, +2, +2, +0, -2, -2, -2},
		[]int{+1, +2, +2, +2, +2, +2, +2, +2, +2, +0, -2, -2},
		[]int{+1, +2, +2, +2, +2, +2, +2, +2, +2, +2, +0, -2},
		[]int{+1, +2, +2, +2, +2, +2, +2, +2, +2, +2, +2, +0},
	}
	assert.Equal(len(names), len(relTable))
	for i, relRow := range relTable {
		assert.Equal(len(names), len(relRow), i)
		for j, rel := range relRow {
			cmp := names[i].Compare(names[j])
			assert.Equal(ndn.NameCompareResult(rel), cmp, "%d=%s %d=%s", i, names[i], j, names[j])
		}
	}
}