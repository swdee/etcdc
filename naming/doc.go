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

// Package naming provides an etcd-backed gRPC resolver for discovering gRPC services.
//
// To use, first import the packages:
//
//	import (
//		"github.com/swdee/etcdc"
//		etcdnaming "github.com/swdee/etcdc/naming"
//
//		"google.golang.org/grpc"
//		"google.golang.org/grpc/naming"
//	)
//
// First, register new endpoint addresses for a service:
//
//	func etcdAdd(c *etcdc.Client, service, addr string) error {
//		r := &etcdnaming.GRPCResolver{Client: c}
//		return r.Update(c.Ctx(), service, naming.Update{Op: naming.Add, Addr: addr})
//	}
//
// Dial an RPC service using the etcd gRPC resolver and a gRPC Balancer:
//
//	func etcdDial(c *etcdc.Client, service string) (*grpc.ClientConn, error) {
//		r := &etcdnaming.GRPCResolver{Client: c}
//		b := grpc.RoundRobin(r)
//		return grpc.Dial(service, grpc.WithBalancer(b))
//	}
//
// Optionally, force delete an endpoint:
//
//	func etcdDelete(c *etcdc. service, addr string) error {
//		r := &etcdnaming.GRPCResolver{Client: c}
//		return r.Update(c.Ctx(), service, naming.Update{Op: naming.Delete, Addr: "1.2.3.4"})
//	}
//
// Or register an expiring endpoint with a lease:
//
//	func etcdLeaseAdd(c *etcdc.Client, lid etcdc.LeaseID, service, addr string) error {
//		r := &etcdnaming.GRPCResolver{Client: c}
//		return r.Update(c.Ctx(), service, naming.Update{Op: naming.Add, Addr: addr}, etcdc.WithLease(lid))
//	}
//
package naming
