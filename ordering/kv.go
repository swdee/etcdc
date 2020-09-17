// Copyright 2017 The etcd Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ordering

import (
	"context"
	"sync"

	"github.com/swdee/etcdc"
)

// kvOrdering ensures that serialized requests do not return
// get with revisions less than the previous
// returned revision.
type kvOrdering struct {
	etcdc.KV
	orderViolationFunc OrderViolationFunc
	prevRev            int64
	revMu              sync.RWMutex
}

func NewKV(kv etcdc.KV, orderViolationFunc OrderViolationFunc) *kvOrdering {
	return &kvOrdering{kv, orderViolationFunc, 0, sync.RWMutex{}}
}

func (kv *kvOrdering) getPrevRev() int64 {
	kv.revMu.RLock()
	defer kv.revMu.RUnlock()
	return kv.prevRev
}

func (kv *kvOrdering) setPrevRev(currRev int64) {
	kv.revMu.Lock()
	defer kv.revMu.Unlock()
	if currRev > kv.prevRev {
		kv.prevRev = currRev
	}
}

func (kv *kvOrdering) Get(ctx context.Context, key string, opts ...etcdc.OpOption) (*etcdc.GetResponse, error) {
	// prevRev is stored in a local variable in order to record the prevRev
	// at the beginning of the Get operation, because concurrent
	// access to kvOrdering could change the prevRev field in the
	// middle of the Get operation.
	prevRev := kv.getPrevRev()
	op := etcdc.OpGet(key, opts...)
	for {
		r, err := kv.KV.Do(ctx, op)
		if err != nil {
			return nil, err
		}
		resp := r.Get()
		if resp.Header.Revision == prevRev {
			return resp, nil
		} else if resp.Header.Revision > prevRev {
			kv.setPrevRev(resp.Header.Revision)
			return resp, nil
		}
		err = kv.orderViolationFunc(op, r, prevRev)
		if err != nil {
			return nil, err
		}
	}
}

func (kv *kvOrdering) Txn(ctx context.Context) etcdc.Txn {
	return &txnOrdering{
		kv.KV.Txn(ctx),
		kv,
		ctx,
		sync.Mutex{},
		[]etcdc.Cmp{},
		[]etcdc.Op{},
		[]etcdc.Op{},
	}
}

// txnOrdering ensures that serialized requests do not return
// txn responses with revisions less than the previous
// returned revision.
type txnOrdering struct {
	etcdc.Txn
	*kvOrdering
	ctx     context.Context
	mu      sync.Mutex
	cmps    []etcdc.Cmp
	thenOps []etcdc.Op
	elseOps []etcdc.Op
}

func (txn *txnOrdering) If(cs ...etcdc.Cmp) etcdc.Txn {
	txn.mu.Lock()
	defer txn.mu.Unlock()
	txn.cmps = cs
	txn.Txn.If(cs...)
	return txn
}

func (txn *txnOrdering) Then(ops ...etcdc.Op) etcdc.Txn {
	txn.mu.Lock()
	defer txn.mu.Unlock()
	txn.thenOps = ops
	txn.Txn.Then(ops...)
	return txn
}

func (txn *txnOrdering) Else(ops ...etcdc.Op) etcdc.Txn {
	txn.mu.Lock()
	defer txn.mu.Unlock()
	txn.elseOps = ops
	txn.Txn.Else(ops...)
	return txn
}

func (txn *txnOrdering) Commit() (*etcdc.TxnResponse, error) {
	// prevRev is stored in a local variable in order to record the prevRev
	// at the beginning of the Commit operation, because concurrent
	// access to txnOrdering could change the prevRev field in the
	// middle of the Commit operation.
	prevRev := txn.getPrevRev()
	opTxn := etcdc.OpTxn(txn.cmps, txn.thenOps, txn.elseOps)
	for {
		opResp, err := txn.KV.Do(txn.ctx, opTxn)
		if err != nil {
			return nil, err
		}
		txnResp := opResp.Txn()
		if txnResp.Header.Revision >= prevRev {
			txn.setPrevRev(txnResp.Header.Revision)
			return txnResp, nil
		}
		err = txn.orderViolationFunc(opTxn, opResp, prevRev)
		if err != nil {
			return nil, err
		}
	}
}
