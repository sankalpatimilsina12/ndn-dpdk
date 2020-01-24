import { Counter } from "../../core/mod.js";

export interface Counters {
  NEntries: Counter;
  NInsert: Counter;
  NFound: Counter;
  NCsMatch: Counter;
  NAllocErr: Counter;
  NDataHit: Counter;
  NDataMiss: Counter;
  NNackHit: Counter;
  NNackMiss: Counter;
  NExpired: Counter;
}