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

// Package etcdc.til contains utility functions derived from etcdc.
package etcdc.til

import (
	"github.com/swdee/etcdc"
)

// KeyExists returns a comparison operation that evaluates to true iff the given
// key exists. It does this by checking if the key `Version` is greater than 0.
// It is a useful guard in transaction delete operations.
func KeyExists(key string) etcdc.Cmp {
	return etcdc.Compare(etcdc.Version(key), ">", 0)
}

// KeyMissing returns a comparison operation that evaluates to true iff the
// given key does not exist.
func KeyMissing(key string) etcdc.Cmp {
	return etcdc.Compare(etcdc.Version(key), "=", 0)
}
