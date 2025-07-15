package ag_redis

import "time"

type RedisNodeConfig struct {
	ID       string        `value:"${id:}"`
	Type     string        `value:"${type:}"`
	Host     string        `value:"${host:localhost}"`
	Port     int           `value:"${port:6379}"`
	Password string        `value:"${password:}"`
	Timeout  time.Duration `value:"${timeout:5000}"`
	// 连接池配置
	PoolMaxWait   time.Duration `value:"${max-wait:5000}"`
	PoolMinIdle   int           `value:"${min-idle:8}"`
	PoolMaxActive int           `value:"${max-active:64}"`
}
