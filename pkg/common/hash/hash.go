// Copyright 2021-present, StarRocks Inc.
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

package hash

import (
	"fmt"
	"hash"
	"hash/fnv"

	"github.com/davecgh/go-spew/spew"
)

// HashObject returns a hash of a given object using the 32-bit FNV-1 hash function
// and the spew library to print the object (see WriteHashObject).
// This is inspired by controller revisions in StatefulSets:
// https://github.com/kubernetes/kubernetes/blob/8de1569ddae62e8fab559fe6bd210a5d6100a277/pkg/controller/history/controller_history.go#L89-L101
//
//nolint:lll
func HashObject(object interface{}) string {
	objHash := fnv.New32a()
	WriteHashObject(objHash, object)
	return fmt.Sprint(objHash.Sum32())
}

// WriteHashObject writes specified object to hash using the spew library
// which follows pointers and prints actual values of the nested objects
// ensuring the hash does not change when a pointer changes.
// The hash can be used for object comparisons.
// Copy of https://github.com/kubernetes/kubernetes/blob/ea0764452222146c47ec826977f49d7001b0ea8c/pkg/util/hash/hash.go#L28
func WriteHashObject(hasher hash.Hash, objectToWrite interface{}) {
	printer := spew.ConfigState{
		Indent:         " ",
		SortKeys:       true,
		DisableMethods: true,
		SpewKeys:       true,
	}
	printer.Fprintf(hasher, "%#v", objectToWrite)
}
