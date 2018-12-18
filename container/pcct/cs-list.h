#ifndef NDN_DPDK_CONTAINER_PCCT_CS_LIST_H
#define NDN_DPDK_CONTAINER_PCCT_CS_LIST_H

/// \file

#include "cs-entry.h"

struct CsNode
{
  CsNode* prev;
  CsNode* next;
};

void CsList_Init(CsList* csl);

/** \brief Append an entry to back of list.
 */
void CsList_Append(CsList* csl, CsEntry* entry);

/** \brief Remove an entry from the list.
 */
void CsList_Remove(CsList* csl, CsEntry* entry);

/** \brief Move an entry to back of list.
 */
void CsList_MoveToLast(CsList* csl, CsEntry* entry);

typedef void (*CsList_EvictCb)(void* arg, CsEntry* entry);

/** \brief Evict up to \p max entries from front of list.
 *  \param cb callback to erase an entry; the callback must not invoke CsList_Remove.
 */
uint32_t CsList_EvictBulk(CsList* csl, uint32_t max, CsList_EvictCb cb,
                          void* cbarg);

/** \brief Identify a list within CS.
 */
typedef enum CsListId {
  CSL_MD, ///< in-memory direct entries
  CSL_MI, ///< in-memory indirect entries
} CsListId;

const char* CsListId_GetName(CsListId cslId);

#endif // NDN_DPDK_CONTAINER_PCCT_CS_LIST_H
