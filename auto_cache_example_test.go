package cache

import (
	"fmt"
	"math/rand"
	"time"
)

func ExampleStorageAutoCache_Get() {
	autoCache := NewStorageAutoCacheObject(true, nil)

	const key = "unique-key"
	getValue := func() (interface{}, error) {
		rand.Seed(42)
		return rand.Intn(100), nil
	}

	err := autoCache.Put(getValue, key, 100*time.Second)
	if err != nil {
		panic(err)
	}

	value, err := autoCache.Get(key)
	if err != nil {
		panic(err)
	}

	fmt.Print(value)
	// Output:
	// 5
}

func ExampleStorageAutoCache_Remove() {
	autoCache := NewStorageAutoCacheObject(true, nil)

	const key = "unique-key"
	getValue := func() (interface{}, error) {
		rand.Seed(42)
		return rand.Intn(100), nil
	}

	err := autoCache.Put(getValue, key, 100*time.Second)
	if err != nil {
		panic(err)
	}

	autoCache.Remove(key)

	_, err = autoCache.Get(key)

	fmt.Print(err)
	// Output:
	// Auto cache key unique-key nof found
}
