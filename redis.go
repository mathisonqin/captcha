package captcha

import (
	"fmt"
	stdRedis "github.com/garyburd/redigo/redis"
	//"github.com/miguel-branco/goconfig"
	"time"
)

var gExpireTime = time.Duration(10) * time.Minute //default is 10 minute

type redisPool struct {
	redisPool stdRedis.Pool
}

func (redis *redisPool) Set(id string, digits []byte) {
	r := redis.redisPool.Get()
	defer r.Close()
	_, err := r.Do("SET", id, digits)
	if err != nil {
		//panic(err.Error())
	}
	_, err = r.Do("EXPIRE", id, gExpireTime)
	if err != nil {
		//panic(err.Error())
	}
}

func (redis *redisPool) Get(id string, clear bool) (digits []byte) {
	fmt.Println("get from redis")
	r := redis.redisPool.Get()
	defer r.Close()
	value, err := r.Do("GET", id)
	if err != nil {

	}
	digits, err = stdRedis.Bytes(value, err)
	if !clear {
		return
	}
	_, err = r.Do("DEL", id)
	if err != nil {

	}
	return
}

func GetRedisPool(host, port string, expire, maxIdle, maxActive int64) Store {
	//configfile = NewConfigFile()
	//conf, err := goconfig.ReadConfigFile(HomeRoot + "/conf/config.cfg")
	//if err != nil {
	//fmt.Println("error!")
	//panic("error")
	//}
	//host, _ := conf.GetString("redis", "host")
	//port, _ := conf.GetString("redis", "port")
	//maxIdle, _ := conf.GetInt64("redis", "maxidle")
	//maxActive, _ := conf.GetInt64("redis", "maxactive")
	//expire, _ := conf.GetInt64("redis", "expire")
	//fmt.Println(host, port, maxIdle, maxActive)
	gExpireTime = time.Duration(expire) * time.Minute
	r := new(redisPool)
	r.redisPool = stdRedis.Pool{
		MaxIdle:   int(maxIdle),
		MaxActive: int(maxActive), // max number of connections
		Dial: func() (stdRedis.Conn, error) {
			c, err := stdRedis.Dial("tcp", host+":"+port)
			if err != nil {
				return nil, err
			}
			return c, err
		},
	}
	return r
}
