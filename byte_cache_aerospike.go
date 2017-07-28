package cache

import (
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/aerospike/aerospike-client-go"
	aerospikeLogger "github.com/aerospike/aerospike-client-go/logger"
	"github.com/aerospike/aerospike-client-go/types"

	"go-cache/errors"
	"go-cache/metric"
)

const (
	dataBin = "data"
	tagsBin = "tags"

	defaultReadTimeout                         = 100 * time.Millisecond
	defaultUpdateConnectionCountMetricInterval = time.Second
)

// AerospikeCache implements cache that uses Aerospike as storage
type AerospikeCache struct {
	ns          string
	client      *aerospike.Client
	cachePrefix string
	getPolicy   *aerospike.BasePolicy
	maxRetries  int
	config      *AerospikeConfig
	logger      IAerospikeCacheLogger
	metric      metric.Metric

	// connection count metric
	updateConnectionCountMetricInterval time.Duration
	quitUpdateConnectionCountMetricChan chan struct{}
	mu                                  sync.Mutex
}

var _ IByteCache = &AerospikeCache{} // AerospikeCache implements IByteCache

// NewAerospikeCache initializes instance of Aerospike-based cache
func NewAerospikeCache(config *AerospikeConfig, client *aerospike.Client, logger IAerospikeCacheLogger, metric metric.Metric) *AerospikeCache {
	if client == nil {
		return nil
	}
	if logger == nil {
		logger = NewNilLogger()
	}

	aeroCache := newAerospike(config, client, logger, metric)
	return aeroCache
}

// CreateAerospikeClient creates AerospikeClient for AerospikeCache
func CreateAerospikeClient(config *AerospikeConfig, logger IAerospikeCacheLogger) (*aerospike.Client, error) {
	if config.NameSpace == "" {
		return nil, errors.New("need aerospike namespace")
	}

	if len(config.Hosts) == 0 || (len(config.Hosts) == 1 && config.Hosts[0] == "") {
		return nil, errors.New("aerospike host list is empty")
	}

	if logger == nil {
		logger = NewNilLogger()
	}
	aerospikeLogger.Logger.SetLogger(logger)
	aerospikeLogger.Logger.SetLevel(aerospikeLogger.LogPriority(config.LogLevel))

	hosts := make([]*aerospike.Host, len(config.Hosts))
	for i, connStr := range config.Hosts {
		hostStr, portStr, err := net.SplitHostPort(connStr)
		if err != nil {
			return nil, err
		}
		port, err := strconv.Atoi(portStr)
		if err != nil {
			return nil, err
		}
		hosts[i] = aerospike.NewHost(hostStr, port)
	}

	clientPolicy := aerospike.NewClientPolicy()

	if config.ConnectionTimeout > 0 {
		clientPolicy.Timeout = config.ConnectionTimeout
	}

	if config.IdleTimeout > 0 {
		clientPolicy.IdleTimeout = config.IdleTimeout
	}

	if config.ConnectionQueueSize > 0 {
		clientPolicy.ConnectionQueueSize = config.ConnectionQueueSize
	}

	if config.LimitConnectionsToQueueSize {
		clientPolicy.LimitConnectionsToQueueSize = true
	}

	clientPolicy.FailIfNotConnected = config.FailIfNotConnected

	client, err := aerospike.NewClientWithPolicyAndHost(clientPolicy, hosts...)
	if err != nil {
		return nil, err
	}

	return client, nil
}

// newAerospike internal constructor
func newAerospike(config *AerospikeConfig, client *aerospike.Client, logger IAerospikeCacheLogger, metric metric.Metric) *AerospikeCache {
	aerospikeLogger.Logger.SetLogger(logger)
	aerospikeLogger.Logger.SetLevel(aerospikeLogger.LogPriority(config.LogLevel))

	// if update connection count metric interval not set use default (1s)
	updateConnectionCountMetricInterval := defaultUpdateConnectionCountMetricInterval
	if config.UpdateConnectionCountMetricInterval > 0 {
		updateConnectionCountMetricInterval = config.UpdateConnectionCountMetricInterval
	}

	getPolicy := aerospike.NewPolicy()

	// if read timeout not set use default (100ms)
	getPolicy.Timeout = defaultReadTimeout
	if config.ReadTimeout > 0 {
		getPolicy.Timeout = config.ReadTimeout
	}

	if config.MaxRetries > 0 {
		getPolicy.MaxRetries = config.MaxRetries
	}

	if config.SleepBetweenRetries > 0 {
		getPolicy.SleepBetweenRetries = config.SleepBetweenRetries
	}

	ac := &AerospikeCache{
		ns:          config.NameSpace,
		client:      client,
		cachePrefix: config.Prefix,
		getPolicy:   getPolicy,
		maxRetries:  config.MaxRetries,
		config:      config,
		logger:      logger,
		metric:      metric,
		quitUpdateConnectionCountMetricChan: make(chan struct{}),
		updateConnectionCountMetricInterval: updateConnectionCountMetricInterval,
	}

	go ac.updateConnectionCountMetric()

	return ac
}

// Client returns aerospike client.
func (a *AerospikeCache) Client() *aerospike.Client {
	return a.client
}

// CreateTagsIndex creates tags index (indexName) by setName
func (a *AerospikeCache) CreateTagsIndex(aerospikeIndex AerospikeIndex) error {
	if aerospikeIndex.SetName == "" {
		return errors.New("setName cant be empty")
	}
	if aerospikeIndex.IndexName == "" {
		return errors.New("indexName cant be empty")
	}

	policy := aerospike.NewWritePolicy(0, 0)
	policy.MaxRetries = a.maxRetries

	createTask, err := a.client.CreateComplexIndex(
		policy,
		a.ns,
		aerospikeIndex.SetName,
		aerospikeIndex.IndexName,
		tagsBin,
		aerospikeIndex.IndexType,
		aerospike.ICT_LIST,
	)

	if err != nil {
		aerospikeError, ok := err.(types.AerospikeError)
		if ok && aerospikeError.ResultCode() == types.INDEX_FOUND {
			a.logger.Debugf(
				"Index %s already exists. Namespace: %s, setName: %s",
				aerospikeIndex.IndexName,
				a.ns,
				aerospikeIndex.SetName,
			)
			return nil
		}

		return errors.Errorf("Aerospike createIndex error: '%s'", err)
	}

	for err := range createTask.OnComplete() {
		if err != nil {
			return errors.Errorf("Aerospike createIndex error: '%s'", err)
		}
	}

	a.logger.Debugf(
		"Aerospike tags index added. Namespace: %s, setName: %s, indexName: %s",
		a.ns,
		aerospikeIndex.SetName,
		aerospikeIndex.IndexName,
	)

	return nil
}

// SetCachePrefix defines prefix for user key
func (a *AerospikeCache) SetCachePrefix(prefix string) {
	a.mu.Lock()
	a.cachePrefix = prefix
	a.mu.Unlock()
}

// Get returns data by given key
func (a *AerospikeCache) Get(key *Key) ([]byte, bool) {

	ts := time.Now()
	var (
		ok   bool
		node *aerospike.Node
		buf  []byte
		err  error
	)

	buf, node, ok, err = a.getByPk(key.Set, key.Pk)

	a.updateHitOrMissCount(ok, map[string]string{
		metric.LabelHost:      a.getNodeHostName(node),
		metric.LabelNamespace: a.ns,
		metric.LabelSet:       key.Set,
	})

	a.metric.ObserveRT(map[string]string{
		metric.LabelHost:      a.getNodeHostName(node),
		metric.LabelNamespace: a.ns,
		metric.LabelSet:       key.Set,
		metric.LabelOperation: "get",
		metric.LabelIsError:   metric.IsError(err),
	}, metric.SinceMs(ts))

	return buf, ok
}

func (a *AerospikeCache) updateHitOrMissCount(condition bool, labels map[string]string) {
	switch condition {
	case true:
		a.metric.RegisterHit(labels)
	case false:
		a.metric.RegisterMiss(labels)
	}
}

func (a *AerospikeCache) getByPk(set, pk string) ([]byte, *aerospike.Node, bool, error) {
	var (
		data []byte
		ok   bool
		err  error
	)

	key, err := a.createKey(set, pk)
	if err != nil {
		a.logger.Warning(err.Error())
		return data, nil, ok, err
	}

	rec, err := a.client.Get(a.getPolicy, key, dataBin)

	if err != nil {
		a.logger.Warningf("could not get data for set '%s' by primary key '%s', error: %q", set, pk, err.Error())
		return data, nil, ok, err
	}

	if rec == nil {
		return data, nil, ok, err
	}

	var bin interface{}
	bin, ok = rec.Bins[dataBin]
	if ok {
		data, ok = bin.([]byte)
	}

	return data, rec.Node, ok, err
}

// Put will delayed put cache in Aerospike
func (a *AerospikeCache) Put(data []byte, key *Key, ttl time.Duration) {
	a.put(data, key, ttl)
}

func (a *AerospikeCache) put(data []byte, key *Key, ttl time.Duration) {
	ts := time.Now()
	var err error

	if len(key.Tags) == 0 {
		err = a.putByPk(data, key.Set, key.Pk, ttl)
	} else {
		err = a.putByPkAndTags(data, key.Set, key.Pk, key.Tags, ttl)
	}

	a.metric.ObserveRT(map[string]string{
		metric.LabelNamespace: a.ns,
		metric.LabelSet:       key.Set,
		metric.LabelOperation: "put",
		metric.LabelIsError:   metric.IsError(err),
	}, metric.SinceMs(ts))
}

func (a *AerospikeCache) putByPk(data []byte, set, pk string, ttl time.Duration) error {
	aeroKey, err := a.createKey(set, pk)
	if err != nil {
		a.logger.Warning(err.Error())
		return err
	}

	bins := []*aerospike.Bin{
		aerospike.NewBin(dataBin, data),
	}

	policy := a.getWritePolice(ttl)

	if err = a.client.PutBins(policy, aeroKey, bins...); err != nil {
		a.logger.Warningf("could not put into set '%s' by primary key '%s': %+v", set, pk, err)
	}

	return err
}

func (a *AerospikeCache) putByPkAndTags(data []byte, set, pk string, tags []string, ttl time.Duration) error {
	var err error

	// build composite primary key for quick data fetching
	aeroKey, err := a.createKey(set, pk)
	if err != nil {
		a.logger.Warning(err.Error())
		return err
	}

	prefTags := make([]string, len(tags))
	for i := range prefTags {
		prefTags[i] = a.cachePrefix + tags[i]
	}

	// add tagsBin to be able to invalidate cache by tag
	bins := []*aerospike.Bin{
		aerospike.NewBin(dataBin, data),
		aerospike.NewBin(tagsBin, prefTags),
	}

	policy := a.getWritePolice(ttl)
	if a.config.PutTimeout > 0 {
		policy.Timeout = a.config.PutTimeout
	}

	if err = a.client.PutBins(policy, aeroKey, bins...); err != nil {
		a.logger.Warningf("could not put into set '%s' by primary key '%s': %+v", set, pk, err)
	}

	return nil
}

// ScanKeys return all keys for set
func (a *AerospikeCache) ScanKeys(set string) ([]Key, error) {

	// We do not know how many records will return
	var keys []Key

	policy := aerospike.NewScanPolicy()
	policy.Priority = aerospike.LOW
	policy.IncludeBinData = true

	// Works only with records containing the BINS `id` and `tags`
	fields := []string{"id", "tags"}

	var (
		pk          string
		pkInterface interface{}

		tags    []string
		binTags interface{}

		ok bool
	)

	r, err := a.client.ScanAll(policy, a.ns, set, fields...)
	if err != nil {
		return nil, err
	}

	for v := range r.Records {

		// We can't use v.Key because v.Key.Value() is nil
		if pkInterface, ok = v.Bins["id"]; !ok {
			a.logger.Warningf("BINS `id` not found for aerospike set: %q", set)
			continue
		}

		if pk, ok = pkInterface.(string); !ok {
			a.logger.Warningf("BINS `id` contain incorrect value: %v", pkInterface)
			continue
		}

		if binTags, ok = v.Bins["tags"]; !ok {
			a.logger.Warningf("BINS `tags` not found for aerospike set: %q", set)
			continue
		}

		tags = a.sliceInterfacesToString(binTags)

		keys = append(keys, Key{Set: set, Pk: pk, Tags: tags})
	}

	return keys, nil
}

func (a *AerospikeCache) sliceInterfacesToString(src interface{}) []string {
	var result []string

	if t, ok := src.([]interface{}); ok {
		result = make([]string, 0, len(t))
		for i := range t {
			if v, ok := t[i].(string); ok {
				result = append(result, v)
			}
		}
	}

	return result
}

// Remove removes data by given cache key
// If tags provided, all records having at least one of them will be removed
// Otherwise only an item with given Primary key will be removed
func (a *AerospikeCache) Remove(key *Key) (err error) {
	var ts = time.Now()

	if len(key.Pk) > 0 {
		err = a.removeByPk(key.Set, key.Pk)
	}

	if err == nil {
		for _, tag := range key.Tags {
			err = a.removeByTag(key.Set, tag)
			if err != nil {
				break
			}
		}
	}

	a.metric.ObserveRT(map[string]string{
		metric.LabelNamespace: a.ns,
		metric.LabelSet:       key.Set,
		metric.LabelOperation: "delete",
		metric.LabelIsError:   metric.IsError(err),
	}, metric.SinceMs(ts))

	return
}

func (a *AerospikeCache) removeByPk(set, pk string) error {
	aeroKey, err := a.createKey(set, pk)
	if err != nil {
		return err
	}

	writePolicy := a.getWritePolice(time.Duration(0))

	if _, err = a.client.Delete(writePolicy, aeroKey); err != nil {
		err = errors.Wrapf(err, "could not remove from set '%s' by primary key '%s'", set, pk)
	}

	return err
}

// removeByTag removes data from Aerospike by PK and tags
func (a *AerospikeCache) removeByTag(set, tag string) error {
	var err error

	defer func() {
		// once in a while query returns a key with empty digest hash
		// then delete command panics with given key (aerospike client bug?)
		if r := recover(); r != nil {
			err = errors.Errorf("removeByTag panic handled: '%v'", r)
		}
	}()

	queryPolicy := aerospike.NewQueryPolicy()
	queryPolicy.Timeout = a.config.RemoveTimeout
	queryPolicy.MaxRetries = a.config.MaxRetries
	queryPolicy.WaitUntilMigrationsAreOver = true

	writePolicy := a.getWritePolice(time.Duration(0))

	stm := aerospike.NewStatement(a.ns, set)
	stm.Addfilter(aerospike.NewContainsFilter(tagsBin, aerospike.ICT_LIST, a.cachePrefix+tag))

	recordSet, err := a.client.Query(queryPolicy, stm)
	if err != nil {
		return errors.Wrapf(err, "could not select data for deleting from for set '%s' and tag '%s'", set, tag)
	}
	defer recordSet.Close()

	ch := recordSet.Results()

	for d := range ch {

		if d.Err != nil {
			a.logger.Critical(d.Err.Error())
			continue
		}

		if d == nil || d.Record == nil {
			a.logger.Criticalf("Empty record %s : %s", set, tag)
			continue
		}
		a.logger.Debugf(
			"Clear record %s from input params {set: %s, tag: %s} with as prefix %s",
			d.Record.Key,
			set,
			tag,
			a.cachePrefix,
		)

		_, err = a.client.Delete(writePolicy, d.Record.Key)
	}

	if err != nil {
		err = errors.Wrapf(err, "could not delete cache by record")
	}

	return err
}

// Close cache storage Aerospike
func (a *AerospikeCache) Close() {
	// send quit message to updateConnectionCountMetric goroutine
	a.quitUpdateConnectionCountMetricChan <- struct{}{}
	a.client.Close()
}

//FIXME: This function should be implemented
// Flush cleans all data.
func (a *AerospikeCache) Flush() int {
	return 0
}

//FIXME: implement me
// Count returns count of data in cache
func (a *AerospikeCache) Count() int {
	return 0
}

// ClearSet removes all values in set
func (a *AerospikeCache) ClearSet(set string) error {
	result, err := a.client.ScanAll(nil, a.ns, set)
	if nil != err {
		return err
	}
	for record := range result.Results() {
		if record.Err != nil {
			a.logger.Warningf("Record error while clearing set %s: %s", set, err)
			continue
		}
		if _, err := a.client.Delete(nil, record.Record.Key); err != nil {
			return errors.Wrapf(err, "could not remove from set '%s'", set)
		}
	}
	a.logger.Debugf("Cache set %s cleared", set)
	return nil
}

func (a *AerospikeCache) createKey(set, key string) (aeroKey *aerospike.Key, err error) {
	a.mu.Lock()
	aeroKey, err = aerospike.NewKey(a.ns, set, a.cachePrefix+key)
	a.mu.Unlock()

	if err != nil {
		err = errors.Wrap(err, "could not create cache key")
	}
	return
}

func (a *AerospikeCache) updateConnectionCountMetric() {
	for {
		select {
		case <-time.After(a.updateConnectionCountMetricInterval):
			for _, node := range a.client.GetNodes() {
				host := node.GetHost().Name
				if node.IsActive() {
					nodeStatistics, err := aerospike.RequestNodeStats(node)
					if err != nil {
						a.logger.Warningf("Cannot get statistic for node"+node.String()+" Error: ", err.Error())
						continue
					}

					connectionsCount, err := strconv.Atoi(nodeStatistics["client_connections"])
					if err != nil {
						a.logger.Warningf("Cannot get statistic for node"+node.String()+" Error: ", err.Error())
						continue
					}

					a.metric.SetItemCount(host, connectionsCount)
				} else {
					a.metric.SetItemCount(host, 0)
				}
			}
		case <-a.quitUpdateConnectionCountMetricChan:
			return
		}
	}
}

func (a *AerospikeCache) getNodeHostName(node *aerospike.Node) string {
	if node == nil {
		return ""
	}
	host := node.GetHost()
	if host == nil {
		return ""
	}
	return host.Name
}

func (a *AerospikeCache) getWritePolice(ttl time.Duration) *aerospike.WritePolicy {
	policy := aerospike.NewWritePolicy(0, uint32(ttl.Seconds()))
	policy.RecordExistsAction = aerospike.REPLACE
	policy.MaxRetries = a.maxRetries

	return policy
}
