package hash

/*
Contains number of most effective hash finalizers
*/

// MurmurFinalizerHash64 Basic murmur hash - https://gist.github.com/dnbaker/0fc1d4edbbdb24069eb063dc2559e4f5
func MurmurFinalizerHash64(hash uint64) uint64 {
	hash ^= hash >> 33
	hash *= 0xff51afd7ed558ccd
	hash ^= hash >> 33
	hash *= 0xc4ceb9fe1a85ec53
	hash ^= hash >> 33
	return hash
}

func IntHash32Finalizer(key uint64, salt uint64) uint32 {
	key ^= salt
	key = (^key) + (key << 18)
	key = key ^ ((key >> 31) | (key << 33))
	key = key * 21
	key = key ^ ((key >> 11) | (key << 53))
	key = key + (key << 6)
	key = key ^ ((key >> 22) | (key << 42))
	return uint32(key)
}
