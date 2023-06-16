// Copyright (c) 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/**
*@describe:
*@author wfl19/Kristas
*@date 2021/09/01
 */

package list

type Set interface {
	Put(s string)
	PutAll(arr ...string)
	ToArray() (arr []string)
	Exists(t string) (ok bool)
	ExistsAny(t ...string) (ok bool)
	ExistsAll(t ...string) (ok bool)
	Remove(s string)
	RemoveAll(arr ...string)
	Length() int
	ForEach(accept func(key string))
}

type GenericSet[T any] interface {
	Put(s T)
	PutAll(arr ...T)
	ToArray() (arr []T)
	Exists(t T) (ok bool)
	ExistsAny(t ...T) (ok bool)
	ExistsAll(t ...T) (ok bool)
	Remove(s T)
	RemoveAll(arr ...T)
	Length() int
	ForEach(accept func(key T))
}
