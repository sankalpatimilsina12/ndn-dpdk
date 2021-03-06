package ethface_test

import (
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"testing"
	"time"

	"github.com/usnistgov/ndn-dpdk/iface/ethface"
	"github.com/usnistgov/ndn-dpdk/iface/ifacetestenv"
	"github.com/usnistgov/ndn-dpdk/ndn/memiftransport"
)

func TestMemif(t *testing.T) {
	assert, require := makeAR(t)

	dir, e := ioutil.TempDir("", "ethface-test")
	require.NoError(e)
	defer os.RemoveAll(dir)
	socketName := path.Join(dir, "memif.sock")

	fixture := ifacetestenv.NewFixture(t)
	defer fixture.Close()

	var locA ethface.Locator
	locA.Local = memiftransport.AddressDPDK
	locA.Remote = memiftransport.AddressApp
	locA.Memif = &memiftransport.Locator{
		SocketName: socketName,
		ID:         7655,
	}
	faceA, e := locA.CreateFace()
	require.NoError(e)
	assert.Equal("memif", faceA.Locator().Scheme())

	var locB ethface.Locator
	locB.Local = memiftransport.AddressDPDK
	locB.Remote = memiftransport.AddressApp
	locB.Memif = &memiftransport.Locator{
		SocketName: socketName,
		ID:         1891,
	}
	faceB, e := locB.CreateFace()
	require.NoError(e)

	helper := exec.Command(os.Args[0], memifbridgeArg, socketName)
	helperIn, e := helper.StdinPipe()
	require.NoError(e)
	helper.Stdout = os.Stdout
	helper.Stderr = os.Stderr
	require.NoError(helper.Start())
	time.Sleep(1 * time.Second)

	fixture.RunTest(faceA, faceB)
	fixture.CheckCounters()

	helperIn.Write([]byte("."))
	assert.NoError(helper.Wait())
}

const memifbridgeArg = "memifbridge"

func memifbridgeHelper() {
	socketName := os.Args[2]
	var locA, locB memiftransport.Locator
	locA.SocketName = socketName
	locA.ID = 7655
	locB.SocketName = socketName
	locB.ID = 1891

	bridge, e := memiftransport.NewBridge(locA, locB, false)
	if e != nil {
		panic(e)
	}

	io.ReadAtLeast(os.Stdin, make([]byte, 1), 1)
	bridge.Close()
}
