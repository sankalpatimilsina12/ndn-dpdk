#ifndef NDN_DPDK_IFACE_FACE_H
#define NDN_DPDK_IFACE_FACE_H

/// \file

#include "faceid.h"
#include "rx-proc.h"
#include "rxburst.h"
#include "tx-proc.h"

#include "../core/running_stat/running-stat.h"
#include "../core/urcu/urcu.h"
#include <urcu/rcuhlist.h>

typedef struct Face Face;
typedef struct FaceCounters FaceCounters;

/** \brief Transmit a burst of L2 frames.
 *  \param pkts L2 frames
 *  \return successfully queued frames
 *  \post FaceImpl owns queued frames, but does not own remaining frames
 */
typedef uint16_t (*FaceImpl_TxBurst)(Face* face, struct rte_mbuf** pkts,
                                     uint16_t nPkts);

typedef struct FaceImpl
{
  RxProc rx;
  TxProc tx;

  /** \brief Statistics of L3 latency.
   *
   *  Latency counting starts from packet arrival or generation, and ends when
   *  packet is queuing for transmission; this counts per L3 packet.
   */
  RunningStat latencyStat;

  char priv[0];
} FaceImpl;

/** \brief Generic network interface.
 */
typedef struct Face
{
  FaceImpl* impl;
  FaceImpl_TxBurst txBurstOp;
  FaceId id;
  int numaSocket;

  struct rte_ring* threadSafeTxQueue;
  struct cds_hlist_node threadSafeTxNode;
} __rte_cache_aligned Face;

static void*
Face_GetPriv(Face* face)
{
  return face->impl->priv;
}

#define Face_GetPrivT(face, T) ((T*)Face_GetPriv((face)))

/** \brief Array of all faces.
 */
extern Face __gFaces[FACEID_MAX];

static Face*
__Face_Get(FaceId faceId)
{
  Face* face = &__gFaces[faceId];
  assert(face->id != FACEID_INVALID);
  return face;
}

// ---- functions invoked by user of face system ----

/** \brief Return whether the face is DOWN.
 */
static bool
Face_IsDown(FaceId faceId)
{
  // TODO implement
  return false;
}

/** \brief Callback upon packet arrival.
 *
 *  Face base type does not directly provide RX function. Each face
 *  implementation shall have an RxLoop function that accepts this callback.
 */
typedef void (*Face_RxCb)(FaceId faceId, FaceRxBurst* burst, void* cbarg);

/** \brief Send a burst of packets (non-thread-safe).
 */
void Face_TxBurst_Nts(Face* face, Packet** npkts, uint16_t count);

/** \brief Send a burst of packets.
 *  \param npkts array of L3 packets; face takes ownership
 *  \param count size of \p npkts array
 *
 *  This function is non-thread-safe by default.
 *  Invoke Face.EnableThreadSafeTx in Go API to make this thread-safe.
 */
static void
Face_TxBurst(FaceId faceId, Packet** npkts, uint16_t count)
{
  Face* face = __Face_Get(faceId);
  if (likely(face->threadSafeTxQueue != NULL)) {
    uint16_t nQueued = rte_ring_mp_enqueue_burst(face->threadSafeTxQueue,
                                                 (void**)npkts, count, NULL);
    uint16_t nRejects = count - nQueued;
    FreeMbufs((struct rte_mbuf**)&npkts[nQueued], nRejects);
    // TODO count nRejects
  } else {
    Face_TxBurst_Nts(face, npkts, count);
  }
}

/** \brief Send a packet.
 *  \param npkt an L3 packet; face takes ownership
 */
static void
Face_Tx(FaceId faceId, Packet* npkt)
{
  Face_TxBurst(faceId, &npkt, 1);
}

/** \brief Retrieve face counters.
 */
void Face_ReadCounters(FaceId faceId, FaceCounters* cnt);

// ---- functions invoked by face implementation ----

typedef struct FaceMempools
{
  /** \brief mempool for indirect mbufs
   */
  struct rte_mempool* indirectMp;

  /** \brief mempool for name linearize upon RX
   *
   *  Dataroom must be at least NAME_MAX_LENGTH.
   */
  struct rte_mempool* nameMp;

  /** \brief mempool for NDNLP headers upon TX
   *
   *  Dataroom must be at least transport-specific-headroom +
   *  PrependLpHeader_GetHeadroom().
   */
  struct rte_mempool* headerMp;
} FaceMempools;

/** \brief Initialize face RX and TX.
 *  \param mtu transport MTU available for NDNLP packets.
 *  \param headroom headroom before NDNLP header, as required by transport.
 */
void FaceImpl_Init(Face* face, uint16_t mtu, uint16_t headroom,
                   FaceMempools* mempools);

/** \brief Process received frames and invoke upper layer callback.
 *  \param burst FaceRxBurst_GetScratch(burst) shall contain received frames,
 *               and each frame should have timestamp set.
 */
void FaceImpl_RxBurst(Face* face, FaceRxBurst* burst, uint16_t nFrames,
                      Face_RxCb cb, void* cbarg);

/** \brief Update counters after a frame is transmitted.
 */
static void
FaceImpl_CountSent(Face* face, struct rte_mbuf* pkt)
{
  TxProc_CountSent(&face->impl->tx, pkt);
}

#endif // NDN_DPDK_IFACE_FACE_H
