// Copyright 2012, Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package router

import (
	"fmt"
	"github.com/dataux/dataux/vendored/mixer/hack"
	"hash/crc32"
	"strconv"
)

type KeyError string

func NewKeyError(format string, args ...interface{}) KeyError {
	return KeyError(fmt.Sprintf(format, args...))
}

func (ke KeyError) Error() string {
	return string(ke)
}

func handleError(err *error) {
	if x := recover(); x != nil {
		*err = x.(KeyError)
	}
}

func EncodeValue(value interface{}) string {
	switch val := value.(type) {
	case int:
		return Uint64Key(val).String()
	case uint64:
		return Uint64Key(val).String()
	case int64:
		return Uint64Key(val).String()
	case string:
		return val
	case []byte:
		return hack.String(val)
	}
	panic(NewKeyError("Unexpected key variable type %T", value))
}

func HashValue(value interface{}) uint64 {
	switch val := value.(type) {
	case int:
		return uint64(val)
	case uint64:
		return uint64(val)
	case int64:
		return uint64(val)
	case string:
		return uint64(crc32.ChecksumIEEE(hack.Slice(val)))
	case []byte:
		return uint64(crc32.ChecksumIEEE(val))
	}
	panic(NewKeyError("Unexpected key variable type %T", value))
}

func NumValue(value interface{}) int64 {
	switch val := value.(type) {
	case int:
		return int64(val)
	case uint64:
		return int64(val)
	case int64:
		return int64(val)
	case string:
		if v, err := strconv.ParseInt(val, 10, 64); err != nil {
			panic(NewKeyError("invalid num format %s", v))
		} else {
			return v
		}
	case []byte:
		if v, err := strconv.ParseInt(hack.String(val), 10, 64); err != nil {
			panic(NewKeyError("invalid num format %s", v))
		} else {
			return v
		}
	}
	panic(NewKeyError("Unexpected key variable type %T", value))
}

type Shard interface {
	FindForKey(key interface{}) int
}

type RangeShard interface {
	Shard
	EqualStart(key interface{}, index int) bool
	EqualStop(key interface{}, index int) bool
}

type HashShard struct {
	ShardNum int
}

func (s *HashShard) FindForKey(key interface{}) int {
	h := HashValue(key)

	return int(h % uint64(s.ShardNum))
}

type NumRangeShard struct {
	Shards []NumKeyRange
}

func (s *NumRangeShard) FindForKey(key interface{}) int {
	v := NumValue(key)
	for i, r := range s.Shards {
		if r.Contains(v) {
			return i
		}
	}
	panic(NewKeyError("Unexpected key %v, not in range", key))
}

func (s *NumRangeShard) EqualStart(key interface{}, index int) bool {
	v := NumValue(key)
	return s.Shards[index].Start == v
}
func (s *NumRangeShard) EqualStop(key interface{}, index int) bool {
	v := NumValue(key)
	return s.Shards[index].End == v
}

type KeyRangeShard struct {
	Shards []KeyRange
}

func (s *KeyRangeShard) FindForKey(key interface{}) int {
	v := KeyspaceId(EncodeValue(key))
	for i, r := range s.Shards {
		if r.Contains(v) {
			return i
		}
	}
	panic(NewKeyError("Unexpected key %v, not in range", key))
}

func (s *KeyRangeShard) EqualStart(key interface{}, index int) bool {
	v := KeyspaceId(EncodeValue(key))
	return s.Shards[index].Start == v
}
func (s *KeyRangeShard) EqualStop(key interface{}, index int) bool {
	v := KeyspaceId(EncodeValue(key))
	return s.Shards[index].End == v
}

type DefaultShard struct {
}

func (s *DefaultShard) FindForKey(key interface{}) int {
	return 0
}
