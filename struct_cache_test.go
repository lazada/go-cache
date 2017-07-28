package cache

import (
	"go-cache/metric/dummy"
	"testing"
	"time"
)

func TestStructCache_Get_Ok(t *testing.T) {
	structCache := NewStructCacheObject(8000, nil, dummy.NewMetric())
	k := &Key{
		Set: "set1",
		Pk:  "1",
	}
	data := "data"
	structCache.Put(data, k, time.Minute*5)

	infResult, find := structCache.Get(k)
	if !find {
		t.Error("Data was not found")
	}

	result, ok := infResult.(string)
	if !ok {
		t.Error("Data have wrong type")
	}

	if result != data {
		t.Error("Data is not expected")
	}
}

func TestStructCache_RenewExistingKey_Ok(t *testing.T) {
	structCache := NewStructCacheObject(8000, nil, dummy.NewMetric())
	k := &Key{
		Set: "set1",
		Pk:  "1",
	}
	data := "data"
	structCache.Put(data, k, time.Minute*5)

	data2 := "data2"
	structCache.Put(data2, k, time.Minute*5)

	infResult, find := structCache.Get(k)
	if !find {
		t.Error("Data was not found")
	}

	result, ok := infResult.(string)
	if !ok {
		t.Error("Data have wrong type")
	}

	if result != data2 {
		t.Error("Data is not expected")
	}

	cnt := structCache.Count()
	if cnt != 1 {
		t.Error("Count is not expeted")
	}
}

func TestStructCache_Get_KeyNotFoundInExistsingSet(t *testing.T) {
	structCache := NewStructCacheObject(8000, nil, dummy.NewMetric())
	k := &Key{
		Set: "set1",
		Pk:  "1",
	}
	data := "data"
	structCache.Put(data, k, time.Minute*5)

	k = &Key{
		Set: "set1",
		Pk:  "2",
	}

	infResult, find := structCache.Get(k)
	if find {
		t.Error("Key for this set shouldn't exist")
	}

	if infResult != nil {
		t.Error("Data should be nil")
	}
}

func TestStructCache_Count_Ok(t *testing.T) {
	structCache := NewStructCacheObject(8000, nil, dummy.NewMetric())
	k := &Key{
		Set: "set1",
		Pk:  "1",
	}
	data := "data"
	structCache.Put(data, k, time.Minute*5)

	k = &Key{
		Set: "set1",
		Pk:  "2",
	}
	structCache.Put(data, k, time.Minute*5)

	k = &Key{
		Set: "set2",
		Pk:  "1",
	}
	structCache.Put(data, k, time.Minute*5)

	count := structCache.Count()

	if count != 3 {
		t.Error("Count is not expected")
	}
}

func TestStructCache_Remove_Ok(t *testing.T) {
	structCache := NewStructCacheObject(8000, nil, dummy.NewMetric())
	k := &Key{
		Set: "set1",
		Pk:  "1",
	}
	data := "data"
	structCache.Put(data, k, time.Minute*5)

	structCache.Remove(k)

	count := structCache.Count()

	if count != 0 {
		t.Error("Count is not expected")
	}
}

func TestStructCache_Flush_Ok(t *testing.T) {
	structCache := NewStructCacheObject(8000, nil, dummy.NewMetric())
	k := &Key{
		Set: "set1",
		Pk:  "1",
	}
	data := "data"
	structCache.Put(data, k, time.Minute*5)

	k = &Key{
		Set: "set2",
		Pk:  "1",
	}
	structCache.Put(data, k, time.Minute*5)

	structCache.Flush()

	count := structCache.Count()

	if count != 0 {
		t.Error("Count is not expected")
	}
}

func TestStructCache_Find_Ok(t *testing.T) {
	structCache := NewStructCacheObject(8000, nil, dummy.NewMetric())
	k := &Key{
		Set: "set1",
		Pk:  "mask",
	}
	data := "data"
	structCache.Put(data, k, time.Minute*5)

	cnt := structCache.Find("as", 1)

	if len(cnt) != 1 {
		t.Error("Count is not expected")
	}

	if cnt[0] != "mask" {
		t.Error("Find return wrong key")
	}
}

func TestStructCache_RegisterCacheSet_Ok(t *testing.T) {
	structCache := NewStructCacheObject(8000, nil, dummy.NewMetric())

	structCache.RegisterCacheSet("set1", 1, nil)
	k := &Key{
		Set: "set1",
		Pk:  "1",
	}
	data := "data"
	structCache.Put(data, k, time.Minute*5)

	k = &Key{
		Set: "set1",
		Pk:  "2",
	}
	structCache.Put(data, k, time.Minute*5)

	cnt := structCache.Count()

	if cnt != 1 {
		t.Error("Count is not expected")
	}
}

func TestStructCache_RegisterCacheSet_AlreadyExists(t *testing.T) {
	structCache := NewStructCacheObject(8000, nil, dummy.NewMetric())

	err := structCache.RegisterCacheSet("set1", 1, nil)

	if err != nil {
		t.Error("CacheSet should be registered")
	}

	err = structCache.RegisterCacheSet("set1", 1, nil)

	if err == nil || err != ErrSetAlreadyExists {
		t.Error("Should error: cacheSet is already exists")
	}
}

func TestStructCache_Collector_Ok(t *testing.T) {
	structCache := NewStructCacheObject(2, nil, dummy.NewMetric())
	structCache.ticker = time.NewTicker(time.Millisecond * 1)
	k := &Key{
		Set: "set1",
		Pk:  "1",
	}
	data := "data"
	structCache.Put(data, k, time.Millisecond*1)

	time.Sleep(time.Millisecond * 20)

	cnt := structCache.Count()

	if cnt != 0 {
		t.Error("Count is not expected")
	}
}

func TestStructCache_Close_Ok(t *testing.T) {
	structCache := NewStructCacheObject(2, nil, dummy.NewMetric())
	structCache.ticker = time.NewTicker(time.Millisecond * 1)
	k := &Key{
		Set: "set1",
		Pk:  "1",
	}
	data := "data"
	structCache.Put(data, k, time.Millisecond*5)

	// Collector should be closed, so it will not clean expired cache
	structCache.Close()

	time.Sleep(time.Millisecond * 20)

	cnt := structCache.Count()

	if cnt != 1 {
		t.Error("Count is not expected")
	}
}

func TestStructCache_SetLimit_Ok(t *testing.T) {
	structCache := NewStructCacheObject(1000, nil, dummy.NewMetric())
	structCache.SetLimit(1)
	k := &Key{
		Set: "set1",
		Pk:  "1",
	}
	data := "data"
	structCache.Put(data, k, time.Minute*5)

	k = &Key{
		Set: "set1",
		Pk:  "2",
	}
	structCache.Put(data, k, time.Minute*5)

	cnt := structCache.Count()

	if cnt != 1 {
		t.Error("Count is not expected")
	}
}

func TestStructCache_Get_NotFound(t *testing.T) {
	structCache := NewStructCacheObject(1000, nil, dummy.NewMetric())
	k := &Key{
		Set: "set1",
		Pk:  "1",
	}

	data, exists := structCache.Get(k)

	if exists {
		t.Error("This key shouldn't exist")
	}

	if data != nil {
		t.Error("Data for this key should be nill")
	}
}

func TestStructCache_Get_Expired(t *testing.T) {
	structCache := NewStructCacheObject(1000, nil, dummy.NewMetric())
	structCache.SetLimit(1)
	k := &Key{
		Set: "set1",
		Pk:  "1",
	}
	data := "data"
	structCache.Put(data, k, time.Millisecond*1)

	time.Sleep(time.Millisecond * 5)

	result, exists := structCache.Get(k)

	if exists {
		t.Error("This key is expired")
	}

	if result != nil {
		t.Error("Data for this key should be nil")
	}

	cnt := structCache.Count()

	if cnt != 0 {
		t.Error("Expired key should be cleaned on Get operation")
	}
}
