## EG-WALKER (GO)
This repository intends to be an implementation of the eg-walker algorithm for collaborative text editing (see [here](https://arxiv.org/pdf/2409.14252)) heavily inspired by the reference implementation by Joseph Gentle

# Milestones
1. Definition

This milestone involves the definition of intended behaviours and an understanding of the eg-walker algorithm. It will involve the creation of a test suite and some documents underlying the implementation used in this repository

## Types

### ID
Represents a globally unique identifier (GUID) for operations.
- `agent string` — the identifier of the originating agent
- `seq int` — sequence number of the operation from that agent
- `Unpack()` — returns `(agent, seq)` for convenience

### LV
Alias for `int`. Serves as a logical index into the operation log.

### OpType
Enum-like string for the type of operation.
- `Insert ("ins")`
- `Delete ("del")`

### InnerOp[T]
Encapsulates the core of an operation (generic over content type `T`).
- `optype OpType` — insert or delete
- `pos int` — target position for the operation
- `content T` — the content (only relevant for insert/retain)

### Op[T]
Represents a full operation, including metadata.
- `InnerOp[T]` — embeds the inner operation
- `id ID` — unique identifier for this operation
- `parents []LV` — list of parent operations (causal dependencies)

### RemoteVersion
Map of last known sequence numbers per agent:
- `map[string]int` where key = agent, value = last seq

### OpLog[T]
Append-only log of operations.
- `ops []Op[T]` — list of all operations
- `frontier []LV` — set of most recent operations per branch
- `version RemoteVersion` — causal version map

### CRDTItem
Represents a materialized item in the CRDT document.
- `lv LV` — the log index this item corresponds to
- `originLeft LV` — left neighbor origin (−1 if unset)
- `originRight LV` — right neighbor origin
- `deleted bool` — whether this item has been deleted
- `curState int` — insertion state (`NOT_YET_INSERTED = -1`, `INSERTED = 0`)

## EG-Walker Definition
*These are the key components to enable collaborative editing through the eg-walker algorithm*

### CRDTDoc
Represents the current editable document view derived from the op log.
- `items *[]*CRDTItem` — all CRDT items in sequence
- `currentVersion *[]LV` — current frontier
- `delTargets *[]LV` — LVs of delete operations
- `itemsByLV map[LV]*CRDTItem` — fast lookup from log index → item


## Operation Log
append only list that contains every single event
- ability to push local operations (insert/delete)
- ability to merge remote operations
- helper utilities

## Document Creation
Checkout: render a document from an operation log
Checkout consists of a for loop consistently modifying the state of a CRDT document.
Each operation is compared to the parent set of histories to determine what set of events
only exist in either the local operation log or remote operation log.

First operations that are common are applied. Those that only exist locally are retreated.
Those that only exist remotely are advanced.

In the apply helper, integrate is used to help determine where insertions should exist.
In the case that concurrent operations occur and conflict, the agent name is used as a tie-break

## Editor Attachment, Tests, Optimization
- may be adjusted w/ websocket server before creating integration

___
2. Implementation

This milestone involves the actual implementation of the eg-walker library in go by passing the test suite. It will also mean a reference implementation of a server/client (or mocking a server/client) using the algorithm to synchronize states
___
3. Verification

This milestone will mean a formal verification of the correctness of the implementation (whatever that means...)
