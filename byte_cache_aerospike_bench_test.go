package cache

import (
	"strconv"
	"testing"
)

func BenchmarkByteCacheAerospike_Set(b *testing.B) {
	cache := initAerospikeByteCache(b, "bench_set")

	setData := []byte("bench")
	key := Key{
		Set: "benchset",
		Pk:  "bench1",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key.Pk = strconv.Itoa(i)
		cache.Put(setData, &key, DefaultCacheTTL)
	}
}

func BenchmarkByteCacheAerospike_Get(b *testing.B) {
	cache := initAerospikeByteCache(b, "bench_get")

	setData := []byte("bench")
	key := Key{
		Set: "benchset",
		Pk:  "bench1",
	}

	for i := 0; i < b.N; i++ {
		key.Pk = strconv.Itoa(i)
		cache.Put(setData, &key, DefaultCacheTTL)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key.Pk = strconv.Itoa(i)
		cache.Get(&key)
	}
}

func BenchmarkByteCacheAerospike_Remove(b *testing.B) {
	cache := initAerospikeByteCache(b, "bench_remove")

	setData := []byte("bench")
	key := Key{
		Set: "benchset",
		Pk:  "bench1",
	}

	for i := 0; i < b.N; i++ {
		key.Pk = strconv.Itoa(i)
		cache.Put(setData, &key, DefaultCacheTTL)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key.Pk = strconv.Itoa(i)
		cache.Remove(&key)
	}
}

func BenchmarkByteCacheAerospike_SetTagged(b *testing.B) {
	cache := initAerospikeByteCache(b, "bench_set_tagged")

	setData := []byte("bench")
	key := Key{
		Set:  "benchset",
		Pk:   "bench1",
		Tags: []string{"tag"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key.Pk = strconv.Itoa(i)
		key.Tags[0] = "tag" + key.Pk
		cache.Put(setData, &key, DefaultCacheTTL)
	}
}

func BenchmarkByteCacheAerospike_GetTagged(b *testing.B) {
	cache := initAerospikeByteCache(b, "bench_get_tagged")

	setData := []byte("bench")
	key := Key{
		Set:  "benchset",
		Pk:   "bench1",
		Tags: []string{"tag"},
	}

	for i := 0; i < b.N; i++ {
		key.Pk = strconv.Itoa(i)
		key.Tags[0] = "tag" + key.Pk
		cache.Put(setData, &key, DefaultCacheTTL)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key.Pk = strconv.Itoa(i)
		cache.Get(&key)
	}
}

func BenchmarkByteCacheAerospike_RemoveTaggedByPk(b *testing.B) {
	cache := initAerospikeByteCache(b, "bench_remove_tagged_by_pk")

	setData := []byte("bench")
	key := Key{
		Set:  "benchset",
		Pk:   "bench1",
		Tags: []string{"tag"},
	}

	for i := 0; i < b.N; i++ {
		key.Pk = strconv.Itoa(i)
		key.Tags[0] = "tag" + key.Pk
		cache.Put(setData, &key, DefaultCacheTTL)
	}

	// remove tags to make removing by pk only
	key.Tags = []string{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key.Pk = strconv.Itoa(i)
		cache.Remove(&key)
	}
}

func BenchmarkByteCacheAerospike_RemoveTaggedByTag(b *testing.B) {
	cache := initAerospikeByteCache(b, "bench_remove_tagged_by_tag")

	setData := []byte("bench1")
	key := Key{
		Set:  "benchset",
		Pk:   "bench1",
		Tags: []string{"tag"},
	}

	for i := 0; i < b.N; i++ {
		key.Pk = strconv.Itoa(i)
		key.Tags[0] = "tag" + key.Pk
		cache.Put(setData, &key, DefaultCacheTTL)
	}

	// remove pk to make it remove by tags only
	key.Pk = ""

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key.Tags[0] = "tag" + key.Pk
		cache.Remove(&key)
	}
}
