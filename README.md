# GO-cache #
A nice and fast set of caches by LAZADA!
All cache types are thread safe and supports prometheus metrics.

## Have 3 cache interfaces: ##
1. IAutoCache
2. IByteCache
3. IStructCache

# **IAutoCache:** #

## Implementations: ##
* StorageAutoCacheFake
* StorageAutoCache

### StorageAutoCacheFake: ###
Can be used in tests

### StorageAutoCache: ###
Executes callback when record`s ttl expired.

#### Example: ####
```go
package main

import (
	"fmt"
	"math/rand"
	"time"
	
	"go-cache"
)

func init() {
	rand.Seed(42)
}

func getValue() (interface{}, error) {
	return rand.Intn(100), nil
}

func main() {
	autoCache := cache.NewStorageAutoCacheObject(true, nil)

	const key = "unique-key"

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
```

# **IByteCache:** #

## Implementations: ##
* BlackholeCache
* AerospikeCache

### BlackholeCache: ###
Can be used in tests

### AerospikeCache: ###
Use aerospike to store slices of bytes. Can be limited. Can store data into different sets. Also supports tags.

#### Example: ####
```go
package main

import (
	"fmt"
	"time"

	"go-cache"
	"go-cache/metric/dummy"
)

func main() {
	config := &cache.AerospikeConfig{
		NameSpace:  "test",
		Hosts:      []string{"localhost:3000"},
		MaxRetries: 5,
	}
	logger := &cache.AerospikeDummyLogger{}

	client, err := cache.CreateAerospikeClient(config, logger)
	if err != nil {
		panic(err)
	}

	aerospikeCache := cache.NewAerospikeCache(config, client, logger, dummy.NewMetric())

	const wirth = "But quality of work can be expected only through personal satisfaction, dedication and enjoyment. In our profession, precision and perfection are not a dispensible luxury, but a simple necessity."
	data := []byte(wirth)
	key := &cache.Key{
		Set: "testset",
		Pk:  "testExpired",
	}

	aerospikeCache.Put(data, key, time.Second)

	cachedData, ok := aerospikeCache.Get(key)
	if !ok {
		panic("Something went wrong")
	}

	fmt.Println(string(cachedData))
	// Output:
	// But quality of work can be expected only through personal satisfaction, dedication and enjoyment. In our profession, precision and perfection are not a dispensible luxury, but a simple necessity.
}
```

# **StructCache:** #
Can store type into local cache. Fast and tread safe.

## Supports: ##
* Limit
* LRU
* Cache sets

#### Example: ####
```go
package main

import (
	"fmt"
	"time"

	"go-cache"
	"go-cache/metric/dummy"
)

func main() {
	structCache := cache.NewStructCacheObject(8000, nil, dummy.NewMetric())

	k := &cache.Key{
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
```