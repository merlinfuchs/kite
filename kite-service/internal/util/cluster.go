package util

import (
	"encoding/binary"
)

// MurmurHash2_32 implements the 32-bit MurmurHash2 algorithm (Austin Appleby).
// This is the algorithm NGINX uses for split_clients.
// Seed is configurable; NGINX effectively uses seed=0 for split_clients.
func MurmurHash2_32(data []byte, seed uint32) uint32 {
	const (
		m uint32 = 0x5bd1e995
		r uint32 = 24
	)

	length := uint32(len(data))
	h := seed ^ length

	// Body: process 4 bytes at a time, little-endian
	for len(data) >= 4 {
		k := binary.LittleEndian.Uint32(data)

		k *= m
		k ^= k >> r
		k *= m

		h *= m
		h ^= k

		data = data[4:]
	}

	// Tail
	switch len(data) {
	case 3:
		h ^= uint32(data[2]) << 16
		fallthrough
	case 2:
		h ^= uint32(data[1]) << 8
		fallthrough
	case 1:
		h ^= uint32(data[0])
		h *= m
	}

	// Final avalanche
	h ^= h >> 13
	h *= m
	h ^= h >> 15

	return h
}

// CluserForKey returns a shard in [0, shardCount) using NGINX split_clients-compatible hashing.
// For dynamic shard count, it maps the 32-bit hash into N equal ranges.
// This avoids modulo bias and aligns with how split_clients percentages partition the hash space.
func CluserForKey(key string, clusterCount int) int {
	if clusterCount == 0 {
		return 0
	}

	// NGINX will hash the bytes of the variable/string; in Go strings are bytes (UTF-8 for typical IDs).
	h := MurmurHash2_32([]byte(key), 0)

	// Map to [0, shardCount) by splitting the 32-bit space into shardCount equal ranges:
	// shard = floor(h / 2^32 * shardCount) == (uint64(h) * uint64(shardCount)) >> 32
	return int(uint32((uint64(h) * uint64(clusterCount)) >> 32))
}
