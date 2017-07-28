package cache

import (
	"fmt"
	"time"

	"go-cache/metric/dummy"
)

func ExampleStructCache_Get() {
	structCache := NewStructCacheObject(8000, nil, dummy.NewMetric())

	k := &Key{
		Set: "set",
		Pk:  "1",
	}
	data := "The essential is invisible to the eyes, we can not truly see but with the eyes of the heart."

	ttl := 5*time.Minute

	structCache.Put(data, k, ttl)

	infResult, find := structCache.Get(k)
	if !find {
		panic("Key is not found")
	}

	result, ok := infResult.(string)
	if !ok {
		panic("Data have wrong type")
	}

	fmt.Println(string(result))
	// Output:
	// The essential is invisible to the eyes, we can not truly see but with the eyes of the heart.
}

func ExampleStructCache_Put() {
	structCache := NewStructCacheObject(8000, nil, dummy.NewMetric())

	k := &Key{
		Set: "set1",
		Pk:  "1",
	}
	data := "data"

	ttl := 5*time.Minute

	structCache.Put(data, k, ttl)

	data2 := "Your most unhappy customers are your greatest source of learning."
	structCache.Put(data2, k, ttl)

	infResult, find := structCache.Get(k)
	if !find {
		panic("Data was not found")
	}

	result, ok := infResult.(string)
	if !ok {
		panic("Data have wrong type")
	}

	fmt.Println(result)
	// Output:
	// Your most unhappy customers are your greatest source of learning.
}

func ExampleStructCache_Count() {
	structCache := NewStructCacheObject(8000, nil, dummy.NewMetric())

	k := &Key{
		Set: "set1",
		Pk:  "1",
	}
	data := "data"

	ttl := 5*time.Minute

	structCache.Put(data, k, ttl)

	k = &Key{
		Set: "set1",
		Pk:  "2",
	}
	structCache.Put(data, k, ttl)

	k = &Key{
		Set: "set2",
		Pk:  "1",
	}
	structCache.Put(data, k, ttl)

	count := structCache.Count()

	fmt.Println(count)
	// Output:
	// 3
}

func ExampleStructCache_Remove() {
	structCache := NewStructCacheObject(8000, nil, dummy.NewMetric())

	k := &Key{
		Set: "set1",
		Pk:  "1",
	}
	data := "data"

	ttl := 5*time.Minute

	structCache.Put(data, k, ttl)

	structCache.Remove(k)

	count := structCache.Count()

	fmt.Println(count)
	// Output:
	// 0
}

func ExampleStructCache_Flush() {
	structCache := NewStructCacheObject(8000, nil, dummy.NewMetric())

	k := &Key{
		Set: "set1",
		Pk:  "1",
	}
	data := "data"

	ttl := 5*time.Minute

	structCache.Put(data, k, ttl)

	k = &Key{
		Set: "set2",
		Pk:  "1",
	}
	structCache.Put(data, k, ttl)

	structCache.Flush()

	count := structCache.Count()

	fmt.Println(count)
	// Output:
	// 0
}

func ExampleStructCache_Find() {
	structCache := NewStructCacheObject(8000, nil, dummy.NewMetric())

	k := &Key{
		Set: "set1",
		Pk:  "mask",
	}
	data := "data"

	ttl := 5*time.Minute

	structCache.Put(data, k, ttl)

	cnt := structCache.Find("as", 1)

	fmt.Println(cnt)
	// Output:
	// [mask]
}

func ExampleStructCache_SetLimit() {
	structCache := NewStructCacheObject(1000, nil, dummy.NewMetric())
	structCache.SetLimit(1)

	k := &Key{
		Set: "set1",
		Pk:  "1",
	}
	data := "data"

	ttl := 5*time.Minute

	structCache.Put(data, k, ttl)

	k = &Key{
		Set: "set1",
		Pk:  "2",
	}
	structCache.Put(data, k, ttl)

	cnt := structCache.Count()

	fmt.Println(cnt)
	// Output:
	// 1
}