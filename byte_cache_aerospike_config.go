package cache

import "time"

// AerospikeConfig contains configuration for aerospike
type AerospikeConfig struct {
	Prefix    string   `config:"aerospike_prefix" default:"" description:"aerospike prefix"`
	NameSpace string   `config:"aerospike_namespace" description:"aerospike namespace"`
	Hosts     []string `config:"aerospike_hosts" description:"aerospike comma-separated host:port list"`

	MaxRetries          int           `config:"aerospike_max_retries" default:"3" description:"aerospike max retries for get and put operations"`
	SleepBetweenRetries time.Duration `config:"aerospike_sleep_between_retries" default:"500ms" description:"aerospike sleep between connection/read/write retries"`

	// tcp connection to aerospike server timeout
	ConnectionTimeout time.Duration `config:"aerospike_connection_timeout" default:"1s" description:"aerospike connection timeout"`

	// max unused connection lifetime
	IdleTimeout time.Duration `config:"aerospike_idle_timeout" default:"1m" description:"aerospike idle connection lifetime"`

	ReadTimeout   time.Duration `config:"aerospike_read_timeout" default:"100ms" description:"aerospike read timeout"`
	RemoveTimeout time.Duration `config:"aerospike_remove_timeout" default:"800ms" description:"aerospike remove timeout"`
	PutTimeout    time.Duration `config:"aerospike_put_timeout" default:"500ms" description:"aerospike put timeout"`

	// max connection pool (queue) size
	ConnectionQueueSize int `config:"aerospike_connection_queue_size" default:"256" description:"aerospike connection queue size"`

	// if true - wait for used connection to be released (up to 1ms)
	LimitConnectionsToQueueSize bool `config:"aerospike_limit_connections" default:"false" description:"aerospike limit connections count to queue size"`

	LogLevel int `config:"aerospike_log_level" default:"1" description:"aerospike logging level: DEBUG(-1), INFO(0), WARNING(1), ERR(2), OFF(999)"`

	FailIfNotConnected bool `config:"aerospike_fail_if_not_connected" default:"false" description:"aerospike fail if not connected"`

	// connection count metric update time interval
	UpdateConnectionCountMetricInterval time.Duration `config:"aerospike_update_connection_count_metric_interval" default:"1s" description:"aerospike update connection count metric interval"`
}
