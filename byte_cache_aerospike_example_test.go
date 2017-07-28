package cache

import (
	"fmt"
	"time"

	"go-cache/metric/dummy"
)

func ExampleAerospikeCache_Get() {
	config := &AerospikeConfig{
		NameSpace:  "test",
		Hosts:      []string{"localhost:3000"},
		MaxRetries: 5,
	}
	logger := &AerospikeDummyLogger{}

	client, err := CreateAerospikeClient(config, logger)
	if err != nil {
		panic(err)
	}

	aerospikeCache := NewAerospikeCache(config, client, logger, dummy.NewMetric())

	const wirth = "But quality of work can be expected only through personal satisfaction, dedication and enjoyment. In our profession, precision and perfection are not a dispensible luxury, but a simple necessity."
	data := []byte(wirth)
	key := &Key{
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
