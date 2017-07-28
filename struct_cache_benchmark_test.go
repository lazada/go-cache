package cache

import (
	"go-cache/metric/dummy"
	"math/rand"
	"strconv"
	"testing"
	"time"
)

const (
	defaultTTL = 5 * time.Minute
)

func prepareDataForBench(structCache *StructCache) {

	for i := 0; i < 1000; i++ {
		k := &Key{
			Set: "set0",
			Pk:  strconv.Itoa(i),
		}
		structCache.Put(i, k, defaultTTL)
	}

	for i := 0; i < 1000; i++ {
		k := &Key{
			Set: "set1",
			Pk:  strconv.Itoa(i),
		}
		structCache.Put(i, k, defaultTTL)
	}

	for i := 0; i < 1000; i++ {
		k := &Key{
			Set: "set2",
			Pk:  strconv.Itoa(i),
		}
		structCache.Put(i, k, defaultTTL)
	}

	for i := 0; i < 1000; i++ {
		k := &Key{
			Set: "set3",
			Pk:  strconv.Itoa(i),
		}
		structCache.Put(i, k, defaultTTL)
	}

	for i := 0; i < 1000; i++ {
		k := &Key{
			Set: "set4",
			Pk:  strconv.Itoa(i),
		}
		structCache.Put(i, k, defaultTTL)
	}

	for i := 0; i < 1000; i++ {
		k := &Key{
			Set: "set5",
			Pk:  strconv.Itoa(i),
		}
		structCache.Put(i, k, defaultTTL)
	}

	for i := 0; i < 1000; i++ {
		k := &Key{
			Set: "set6",
			Pk:  strconv.Itoa(i),
		}
		structCache.Put(i, k, defaultTTL)
	}

}

func BenchmarkStructCache_ConcurentGetFrom7DifferentSets(b *testing.B) {
	structCache := NewStructCacheObject(8000, nil, dummy.NewMetric())
	prepareDataForBench(structCache)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			caseNum := rand.Intn(7)
			set := "set" + strconv.Itoa(caseNum)
			key := &Key{Set: set, Pk: "2"}

			v, ok := structCache.Get(key)
			if !ok {
				b.Fatalf("Get operation is unsuccessfull for %v", key)
			}
			if v == nil {
				b.Fatalf("Got nil value from cache for %v", key)
			}
		}
	})
}

func BenchmarkStructCache_ConcurentPutOverLimit(b *testing.B) {
	structCache := NewStructCacheObject(5, nil, dummy.NewMetric())

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			i++
			caseNum := rand.Intn(7)
			set := "set" + strconv.Itoa(caseNum)
			k := &Key{
				Set: set,
				Pk:  strconv.Itoa(i),
			}

			err := structCache.Put(i, k, defaultTTL)
			if err != nil {
				b.Fatalf("Put operation is unsuccessfull: %v", err)
			}
		}
	})
}

func BenchmarkStructCache_ConcurentPutToDifferentSets(b *testing.B) {
	structCache := NewStructCacheObject(8000, nil, dummy.NewMetric())

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			if i > 8000 {
				i = 1
			}

			i++
			caseNum := rand.Intn(7)
			set := "set" + strconv.Itoa(caseNum)
			k := &Key{
				Set: set,
				Pk:  strconv.Itoa(i),
			}

			err := structCache.Put(i, k, defaultTTL)
			if err != nil {
				b.Fatalf("Put operation is unsuccessfull: %v", err)
			}
		}
	})
}

func BenchmarkStructCache_ConcurentPutAndGet(b *testing.B) {
	structCache := NewStructCacheObject(8000, nil, dummy.NewMetric())

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			if i > 8000 {
				i = 1
			}
			i++

			k := &Key{
				Set: "set1",
				Pk:  strconv.Itoa(i),
			}

			err := structCache.Put(i, k, defaultTTL)
			if err != nil {
				b.Fatalf("Put operation is unsuccessfull: %v", err)
			}

			kGet := &Key{
				Set: "set1",
				Pk:  strconv.Itoa(i),
			}
			v, ok := structCache.Get(kGet)
			if !ok {
				b.Fatalf("Get operation is unsuccessfull for %v", kGet)
			}
			if v == nil {
				b.Fatalf("Got nil value from cache for %v", kGet)
			}
		}
	})
}

func BenchmarkStructCache_ConcurentGetWithoutGC(b *testing.B) {
	structCache := NewStructCacheObject(50000000, nil, dummy.NewMetric())
	for i := 0; i < 500000; i++ {
		k := &Key{
			Set: "set" + strconv.Itoa(i%50),
			Pk:  strconv.Itoa(i),
		}
		structCache.Put(i, k, defaultTTL)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			if i > 500000 {
				i = 1
			}
			i++

			k := &Key{
				Set: "set" + strconv.Itoa(i%50),
				Pk:  strconv.Itoa(i),
			}

			v, ok := structCache.Get(k)
			if !ok {
				b.Fatalf("Get operation is unsuccessfull for %v", k)
			}
			if v == nil {
				b.Fatalf("Got nil value from cache for %v", k)
			}
		}
	})
}

func BenchmarkStructCache_ConcurentPutWithGC(b *testing.B) {
	structCache := NewStructCacheObject(50000000, nil, dummy.NewMetric())
	structCache.ticker = time.NewTicker(time.Millisecond * 10)

	for i := 0; i < 500000; i++ {
		k := &Key{
			Set: "set" + strconv.Itoa(i%50),
			Pk:  strconv.Itoa(i),
		}
		structCache.Put(i, k, defaultTTL)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 500000
		for pb.Next() {
			i++
			k := &Key{
				Set: "set" + strconv.Itoa(i%50),
				Pk:  strconv.Itoa(i),
			}

			err := structCache.Put(i, k, defaultTTL)
			if err != nil {
				b.Fatalf("Put operation is unsuccessfull: %v", err)
			}
		}
	})
}

func BenchmarkStructCache_ConcurentGet_ManySets_WithGC(b *testing.B) {
	structCache := NewStructCacheObject(50000000, nil, dummy.NewMetric())
	structCache.ticker = time.NewTicker(time.Millisecond * 10)

	for i := 0; i < 500000; i++ {
		k := &Key{
			Set: "set" + strconv.Itoa(i%50),
			Pk:  strconv.Itoa(i),
		}
		structCache.Put(i, k, defaultTTL)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			if i > 500000 {
				i = 1
			}
			i++

			k := &Key{
				Set: "set" + strconv.Itoa(i%50),
				Pk:  strconv.Itoa(i),
			}

			v, ok := structCache.Get(k)
			if !ok {
				b.Fatalf("Get operation is unsuccessfull for %v", k)
			}
			if v == nil {
				b.Fatalf("Got nil value from cache for %v", k)
			}
		}
	})
}


func BenchmarkStructCache_ConcurentGet_SingleSet_WithGC(b *testing.B) {
	structCache := NewStructCacheObject(50000000, nil, dummy.NewMetric())
	structCache.ticker = time.NewTicker(time.Millisecond * 10)

	for i := 0; i < 500000; i++ {
		k := &Key{
			Set: "set1",
			Pk:  strconv.Itoa(i),
		}
		structCache.Put(i, k, defaultTTL)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			if i > 500000 {
				i = 1
			}
			i++

			k := &Key{
				Set: "set1",
				Pk:  strconv.Itoa(i),
			}

			v, ok := structCache.Get(k)
			if !ok {
				b.Fatalf("Get operation is unsuccessfull for %v", k)
			}
			if v == nil {
				b.Fatalf("Got nil value from cache for %v", k)
			}
		}
	})
}
