#ifndef NDNDPDK_DISKSTORE_DISKSTORE_H
#define NDNDPDK_DISKSTORE_DISKSTORE_H

/** @file */

#include "../dpdk/bdev.h"
#include "../dpdk/spdk-thread.h"
#include "../ndni/packet.h"

/**
 * @brief Expected block size of the underlying block device.
 */
#define DISK_STORE_BLOCK_SIZE 512

/** @brief Disk-backed Data packet store. */
typedef struct DiskStore
{
  struct spdk_thread* th;
  struct spdk_bdev_desc* bdev;
  struct spdk_io_channel* ch;
  uint64_t nBlocksPerSlot;
  uint32_t blockSize;
} DiskStore;

/**
 * @brief Store a Data packet.
 * @param slotID disk slot number; slot 0 cannot be used.
 * @param npkt a Data packet. DiskStore takes ownership.
 *
 * This function may be invoked on any thread, including non-SPDK thread.
 */
__attribute__((nonnull)) void
DiskStore_PutData(DiskStore* store, uint64_t slotID, Packet* npkt);

/**
 * @brief Retrieve a Data packet.
 * @param slotID disk slot number.
 * @param dataLen Data packet length.
 * @param npkt an Interest packet. DiskStore takes ownership.
 * @param dataBuf mbuf for Data packet. DiskStore takes ownership.
 * @param reply where to return results.
 *
 * This function asynchronously reads from a specified slot of the underlying disk, and parses
 * the content as a Data packet. It then assigns @c interest->diskSlot and @c interest->diskData
 * on the Interest @p npkt , and enqueues @p npkt into @p reply.
 *
 * This function may be invoked on any thread, including non-SPDK thread.
 */
__attribute__((nonnull)) void
DiskStore_GetData(DiskStore* store, uint64_t slotID, uint16_t dataLen, Packet* npkt,
                  struct rte_mbuf* dataBuf, struct rte_ring* reply);

__attribute__((nonnull)) static __rte_always_inline uint64_t
DiskStore_ComputeBlockOffset_(DiskStore* store, uint64_t slotID)
{
  return slotID * store->nBlocksPerSlot;
}

__attribute__((nonnull)) static __rte_always_inline uint64_t
DiskStore_ComputeBlockCount_(DiskStore* store, Packet* npkt)
{
  uint64_t pktLen = Packet_ToMbuf(npkt)->pkt_len;
  return pktLen / DISK_STORE_BLOCK_SIZE + (int)(pktLen % DISK_STORE_BLOCK_SIZE > 0);
}

#endif // NDNDPDK_DISKSTORE_DISKSTORE_H
