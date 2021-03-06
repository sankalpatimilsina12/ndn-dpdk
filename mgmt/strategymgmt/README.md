# ndn-dpdk/mgmt/strategymgmt

This package implements [strategy table](../../container/strategycode/) management.

## Strategy

**Strategy.List** lists strategies.

**Strategy.Get** retrieves strategy information by ID.

**Strategy.Load** loads a strategy ELF object.
It requires every strategy to have a unique short name.

**Strategy.Unload** unloads a strategy.
If the strategy is in use by a FIB entry, this operation simply unreferences the strategy, and the strategy will be unloaded when the last FIB entry stops using it.
