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

package namespace

import (
	"context"

	"github.com/swdee/etcdc"
	"go.etcd.io/etcd/etcdserver/api/v3rpc/rpctypes"
	pb "go.etcd.io/etcd/etcdserver/etcdserverpb"
)

type kvPrefix struct {
	etcdc.KV
	pfx string
}

// NewKV wraps a KV instance so that all requests
// are prefixed with a given string.
func NewKV(kv etcdc.KV, prefix string) etcdc.KV {
	return &kvPrefix{kv, prefix}
}

func (kv *kvPrefix) Put(ctx context.Context, key, val string, opts ...etcdc.OpOption) (*etcdc.PutResponse, error) {
	if len(key) == 0 {
		return nil, rpctypes.ErrEmptyKey
	}
	op := kv.prefixOp(etcdc.OpPut(key, val, opts...))
	r, err := kv.KV.Do(ctx, op)
	if err != nil {
		return nil, err
	}
	put := r.Put()
	kv.unprefixPutResponse(put)
	return put, nil
}

func (kv *kvPrefix) Get(ctx context.Context, key string, opts ...etcdc.OpOption) (*etcdc.GetResponse, error) {
	if len(key) == 0 {
		return nil, rpctypes.ErrEmptyKey
	}
	r, err := kv.KV.Do(ctx, kv.prefixOp(etcdc.OpGet(key, opts...)))
	if err != nil {
		return nil, err
	}
	get := r.Get()
	kv.unprefixGetResponse(get)
	return get, nil
}

func (kv *kvPrefix) Delete(ctx context.Context, key string, opts ...etcdc.OpOption) (*etcdc.DeleteResponse, error) {
	if len(key) == 0 {
		return nil, rpctypes.ErrEmptyKey
	}
	r, err := kv.KV.Do(ctx, kv.prefixOp(etcdc.OpDelete(key, opts...)))
	if err != nil {
		return nil, err
	}
	del := r.Del()
	kv.unprefixDeleteResponse(del)
	return del, nil
}

func (kv *kvPrefix) Do(ctx context.Context, op etcdc.Op) (etcdc.OpResponse, error) {
	if len(op.KeyBytes()) == 0 && !op.IsTxn() {
		return etcdc.OpResponse{}, rpctypes.ErrEmptyKey
	}
	r, err := kv.KV.Do(ctx, kv.prefixOp(op))
	if err != nil {
		return r, err
	}
	switch {
	case r.Get() != nil:
		kv.unprefixGetResponse(r.Get())
	case r.Put() != nil:
		kv.unprefixPutResponse(r.Put())
	case r.Del() != nil:
		kv.unprefixDeleteResponse(r.Del())
	case r.Txn() != nil:
		kv.unprefixTxnResponse(r.Txn())
	}
	return r, nil
}

type txnPrefix struct {
	etcdc.Txn
	kv *kvPrefix
}

func (kv *kvPrefix) Txn(ctx context.Context) etcdc.Txn {
	return &txnPrefix{kv.KV.Txn(ctx), kv}
}

func (txn *txnPrefix) If(cs ...etcdc.Cmp) etcdc.Txn {
	txn.Txn = txn.Txn.If(txn.kv.prefixCmps(cs)...)
	return txn
}

func (txn *txnPrefix) Then(ops ...etcdc.Op) etcdc.Txn {
	txn.Txn = txn.Txn.Then(txn.kv.prefixOps(ops)...)
	return txn
}

func (txn *txnPrefix) Else(ops ...etcdc.Op) etcdc.Txn {
	txn.Txn = txn.Txn.Else(txn.kv.prefixOps(ops)...)
	return txn
}

func (txn *txnPrefix) Commit() (*etcdc.TxnResponse, error) {
	resp, err := txn.Txn.Commit()
	if err != nil {
		return nil, err
	}
	txn.kv.unprefixTxnResponse(resp)
	return resp, nil
}

func (kv *kvPrefix) prefixOp(op etcdc.Op) etcdc.Op {
	if !op.IsTxn() {
		begin, end := kv.prefixInterval(op.KeyBytes(), op.RangeBytes())
		op.WithKeyBytes(begin)
		op.WithRangeBytes(end)
		return op
	}
	cmps, thenOps, elseOps := op.Txn()
	return etcdc.OpTxn(kv.prefixCmps(cmps), kv.prefixOps(thenOps), kv.prefixOps(elseOps))
}

func (kv *kvPrefix) unprefixGetResponse(resp *etcdc.GetResponse) {
	for i := range resp.Kvs {
		resp.Kvs[i].Key = resp.Kvs[i].Key[len(kv.pfx):]
	}
}

func (kv *kvPrefix) unprefixPutResponse(resp *etcdc.PutResponse) {
	if resp.PrevKv != nil {
		resp.PrevKv.Key = resp.PrevKv.Key[len(kv.pfx):]
	}
}

func (kv *kvPrefix) unprefixDeleteResponse(resp *etcdc.DeleteResponse) {
	for i := range resp.PrevKvs {
		resp.PrevKvs[i].Key = resp.PrevKvs[i].Key[len(kv.pfx):]
	}
}

func (kv *kvPrefix) unprefixTxnResponse(resp *etcdc.TxnResponse) {
	for _, r := range resp.Responses {
		switch tv := r.Response.(type) {
		case *pb.ResponseOp_ResponseRange:
			if tv.ResponseRange != nil {
				kv.unprefixGetResponse((*etcdc.GetResponse)(tv.ResponseRange))
			}
		case *pb.ResponseOp_ResponsePut:
			if tv.ResponsePut != nil {
				kv.unprefixPutResponse((*etcdc.PutResponse)(tv.ResponsePut))
			}
		case *pb.ResponseOp_ResponseDeleteRange:
			if tv.ResponseDeleteRange != nil {
				kv.unprefixDeleteResponse((*etcdc.DeleteResponse)(tv.ResponseDeleteRange))
			}
		case *pb.ResponseOp_ResponseTxn:
			if tv.ResponseTxn != nil {
				kv.unprefixTxnResponse((*etcdc.TxnResponse)(tv.ResponseTxn))
			}
		default:
		}
	}
}

func (kv *kvPrefix) prefixInterval(key, end []byte) (pfxKey []byte, pfxEnd []byte) {
	return prefixInterval(kv.pfx, key, end)
}

func (kv *kvPrefix) prefixCmps(cs []etcdc.Cmp) []etcdc.Cmp {
	newCmps := make([]etcdc.Cmp, len(cs))
	for i := range cs {
		newCmps[i] = cs[i]
		pfxKey, endKey := kv.prefixInterval(cs[i].KeyBytes(), cs[i].RangeEnd)
		newCmps[i].WithKeyBytes(pfxKey)
		if len(cs[i].RangeEnd) != 0 {
			newCmps[i].RangeEnd = endKey
		}
	}
	return newCmps
}

func (kv *kvPrefix) prefixOps(ops []etcdc.Op) []etcdc.Op {
	newOps := make([]etcdc.Op, len(ops))
	for i := range ops {
		newOps[i] = kv.prefixOp(ops[i])
	}
	return newOps
}
