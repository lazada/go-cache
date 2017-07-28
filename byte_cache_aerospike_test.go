package cache

import (
	"fmt"
	"path"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/aerospike/aerospike-client-go"
	"github.com/vaughan0/go-ini"

	"go-cache/metric/dummy"
)

type AerospikeDummyLogger struct{}

func (l *AerospikeDummyLogger) Printf(format string, v ...interface{}) {}

func (l *AerospikeDummyLogger) Debugf(message string, args ...interface{}) {}

func (l *AerospikeDummyLogger) Errorf(message string, args ...interface{}) {
	fmt.Println("[ERROR]", fmt.Sprintf(message, args...))
}

func (l *AerospikeDummyLogger) Warningf(message string, args ...interface{}) {
	fmt.Println("[WARNING]", fmt.Sprintf(message, args...))
}

func (l *AerospikeDummyLogger) Warning(args ...interface{}) {
	fmt.Println("[WARNING]", fmt.Sprint(args...))
}

func (l *AerospikeDummyLogger) Criticalf(message string, args ...interface{}) {
	fmt.Println("[CRITICAL]", fmt.Sprintf(message, args...))
}

func (l *AerospikeDummyLogger) Critical(args ...interface{}) {
	fmt.Println("[CRITICAL]", fmt.Sprint(args...))
}

func newTestAerospikeByteCache(tb testing.TB, hosts []string, namespace string) *AerospikeCache {
	config := &AerospikeConfig{
		NameSpace:  namespace,
		Hosts:      hosts,
		MaxRetries: 5,
	}
	logger := &AerospikeDummyLogger{}

	client, err := CreateAerospikeClient(config, logger)
	if err != nil {
		tb.Fatalf("Can't create aerospike client: '%s'", err)
	}

	return NewAerospikeCache(config, client, logger, dummy.NewMetric())
}

func initAerospikeByteCache(tb testing.TB, prefix string) *AerospikeCache {
	file, err := ini.LoadFile("./test.ini")
	if err != nil {
		tb.Fatalf("Unable to parse ini file: '%s'", err)
	}

	namespace, ok := file.Get("aerospike", "namespace")
	if !ok {
		tb.Fatal("'namespace' variable missing from 'aerospike' section")
	}

	hosts, ok := file.Get("aerospike", "hosts")
	if !ok {
		tb.Fatal("'hosts' variable missing from 'aerospike' section")
	}

	cache := newTestAerospikeByteCache(tb, strings.Split(hosts, ","), namespace)
	cache.SetCachePrefix(prefix)

	cache.CreateTagsIndex(AerospikeIndex{
		"testset_withtags",
		"tags_testset_withtags",
		aerospike.STRING,
	})

	cache.CreateTagsIndex(AerospikeIndex{
		"benchset",
		"tags_benchset",
		aerospike.STRING,
	})

	return cache
}

func TestByteCacheAerospike_TestGet(t *testing.T) {
	cache := initAerospikeByteCache(t, "get")

	setData := []byte("test1")
	key := Key{
		Set: "testset",
		Pk:  "test1",
	}
	cache.Put(setData, &key, DefaultCacheTTL)

	assertByteCacheKeyHasValue(t, cache, &key, "test1")
}

func TestByteCacheAerospike_GetExpired(t *testing.T) {
	cache := initAerospikeByteCache(t, "getexpired")

	setData := []byte("testExpired")
	key := Key{
		Set: "testset",
		Pk:  "testExpired",
	}
	cache.Put(setData, &key, time.Second)

	assertByteCacheKeyHasValue(t, cache, &key, "testExpired")

	time.Sleep(time.Second * 2)

	assertByteCacheKeyEmpty(t, cache, &key)
}

func TestByteCacheAerospike_Remove(t *testing.T) {
	cache := initAerospikeByteCache(t, "remove")

	setData := []byte("test_remove")
	key := Key{
		Set: "testset",
		Pk:  "test_remove",
	}
	cache.Put(setData, &key, DefaultCacheTTL)

	assertByteCacheKeyHasValue(t, cache, &key, "test_remove")

	err := cache.Remove(&key)

	if err != nil {
		t.Error(err)
	}
	assertByteCacheKeyEmpty(t, cache, &key)
}

func TestByteCacheAerospike_RemoveByTag(t *testing.T) {
	cache := initAerospikeByteCache(t, "remove_by_tag")

	setData := []byte("test_tag")
	key := Key{
		Set:  "testset_withtags",
		Pk:   "test_remove_by_tag",
		Tags: []string{"tag1", "tag2"},
	}
	cache.Put(setData, &key, DefaultCacheTTL)

	assertByteCacheKeyHasValue(t, cache, &key, "test_tag")

	err := cache.Remove(&Key{
		Set:  "testset_withtags",
		Tags: []string{"tag2"},
	})

	if err != nil {
		t.Error(err)
	}
	assertByteCacheKeyEmpty(t, cache, &key)
}

func TestByteCacheAerospike_RemoveByTag_FewValues(t *testing.T) {
	cache := initAerospikeByteCache(t, "remove_by_tag_few_values")

	setData := []byte("test_tag")
	keyTag1Tag2 := Key{
		Set:  "testset_withtags",
		Pk:   "test_remove_by_tag_tags12",
		Tags: []string{"tag1", "tag2"},
	}
	cache.Put(setData, &keyTag1Tag2, DefaultCacheTTL)

	keyTag2Tag3 := Key{
		Set:  "testset_withtags",
		Pk:   "test_remove_by_tag_tags23",
		Tags: []string{"tag2", "tag3"},
	}
	cache.Put(setData, &keyTag2Tag3, DefaultCacheTTL)

	keyNoTags := Key{
		Set: "testset_withtags",
		Pk:  "test_remove_by_tag_notags",
	}
	cache.Put(setData, &keyNoTags, DefaultCacheTTL)

	// check all keys exists and have valid value
	assertByteCacheKeyHasValue(t, cache, &keyTag1Tag2, "test_tag")
	assertByteCacheKeyHasValue(t, cache, &keyTag2Tag3, "test_tag")
	assertByteCacheKeyHasValue(t, cache, &keyNoTags, "test_tag")

	err := cache.Remove(&Key{
		Set:  "testset_withtags",
		Tags: []string{"tag2"},
	})

	if err != nil {
		t.Error(err)
	}
	// check tagged keys removed, but untagged one still exists and has valid value
	assertByteCacheKeyEmpty(t, cache, &keyTag1Tag2)
	assertByteCacheKeyEmpty(t, cache, &keyTag2Tag3)
	assertByteCacheKeyHasValue(t, cache, &keyNoTags, "test_tag")
}

func TestByteCacheAerospike_RemoveByFewTags_FewValues(t *testing.T) {
	cache := initAerospikeByteCache(t, "remove_by_few_tags_few_values")

	setData := []byte("test_tag")
	keyTag1 := Key{
		Set:  "testset_withtags",
		Pk:   "test_remove_by_tag_tags1",
		Tags: []string{"tag1"},
	}
	cache.Put(setData, &keyTag1, DefaultCacheTTL)

	keyTag1Tag2 := Key{
		Set:  "testset_withtags",
		Pk:   "test_remove_by_tag_tags12",
		Tags: []string{"tag1", "tag2"},
	}
	cache.Put(setData, &keyTag1Tag2, DefaultCacheTTL)

	keyTag2Tag3 := Key{
		Set:  "testset_withtags",
		Pk:   "test_remove_by_tag_tags23",
		Tags: []string{"tag2", "tag3"},
	}
	cache.Put(setData, &keyTag2Tag3, DefaultCacheTTL)

	keyNoTags := Key{
		Set: "testset_withtags",
		Pk:  "test_remove_by_tag_notags",
	}
	cache.Put(setData, &keyNoTags, DefaultCacheTTL)

	// check all keys exists and have valid value
	assertByteCacheKeyHasValue(t, cache, &keyTag1, "test_tag")
	assertByteCacheKeyHasValue(t, cache, &keyTag1Tag2, "test_tag")
	assertByteCacheKeyHasValue(t, cache, &keyTag2Tag3, "test_tag")
	assertByteCacheKeyHasValue(t, cache, &keyNoTags, "test_tag")

	err := cache.Remove(&Key{
		Set:  "testset_withtags",
		Tags: []string{"tag1", "tag2"},
	})

	if err != nil {
		t.Error(err)
	}
	// check tagged keys removed, but untagged one still exists and has valid value
	assertByteCacheKeyEmpty(t, cache, &keyTag1)
	assertByteCacheKeyEmpty(t, cache, &keyTag1Tag2)
	assertByteCacheKeyEmpty(t, cache, &keyTag2Tag3)
	assertByteCacheKeyHasValue(t, cache, &keyNoTags, "test_tag")
}

func TestByteCacheAerospike_RemoveByTagWithPrefix(t *testing.T) {
	cache := initAerospikeByteCache(t, "remove_by_tag_with_prefix")

	prefixedCache := newTestAerospikeByteCache(t, cache.config.Hosts, cache.config.NameSpace)
	prefixedCache.SetCachePrefix("remove_by_tag_with_other_prefix")

	setData := []byte("test_tag")
	key := Key{
		Set:  "testset_withtags",
		Pk:   "test_remove_by_tag_with_prefix",
		Tags: []string{"tag1", "tag2"},
	}
	cache.Put(setData, &key, DefaultCacheTTL)
	prefixedCache.Put(setData, &key, DefaultCacheTTL)

	assertByteCacheKeyHasValue(t, cache, &key, "test_tag")
	assertByteCacheKeyHasValue(t, prefixedCache, &key, "test_tag")

	err := prefixedCache.Remove(&Key{
		Set:  "testset_withtags",
		Tags: []string{"tag2"},
	})

	if err != nil {
		t.Error(err)
	}
	assertByteCacheKeyHasValue(t, cache, &key, "test_tag")
	assertByteCacheKeyEmpty(t, prefixedCache, &key)
}

func assertByteCacheKeyHasValue(t *testing.T, cache IByteCache, key *Key, expectedValue string) {
	getData, success := cache.Get(key)
	if !success {
		t.Errorf("%s: value doesn't exist (%+v)", callerInfo(), key)
	}
	if string(getData) != expectedValue {
		t.Errorf("%s: %s is not equal to expected value: %s (%+v)", callerInfo(), string(getData), expectedValue, key)
	}
}

func assertByteCacheKeyEmpty(t *testing.T, cache IByteCache, key *Key) {
	getData, success := cache.Get(key)
	if success {
		t.Errorf("%s: value should not exist (key: %+v)", callerInfo(), key)
	}
	if string(getData) != "" {
		t.Errorf("%s: %s should be empty (key: %+v)", callerInfo(), string(getData), key)
	}
}

func callerInfo() string {
	_, file, line, _ := runtime.Caller(2)
	_, fileNameLine := path.Split(file)
	fileNameLine += ":" + strconv.Itoa(line)

	return fileNameLine
}
