package captcha

import (
	//"fmt"
	"github.com/beego/memcache"
	//"github.com/miguel-branco/goconfig"
	"time"
)

var gMemVitessExpireTime uint64 = 60 //default is 1 minute

type memcacheConnection struct {
	connection *memcache.Connection
}

func (mConn *memcacheConnection) Set(id string, digits []byte) {

	// item := new(memcache.Item)
	// item.Key = id
	// item.Value = digits
	// item.Expiration = gMemExpireTime
	stored, err := mConn.connection.Set(id, 0, gMemVitessExpireTime, digits)
	//defer mConn.connection.Close()
	if stored == false {

	}

	if err != nil {
		panic(err.Error())
	}

}

func (mConn *memcacheConnection) Get(id string, clear bool) (digits []byte) {
	result, err := mConn.connection.Get(id)
	//defer mConn.connection.Close()
	digits = result[0].Value
	if !clear {
		return
	}
	deleted, err := mConn.connection.Delete(id)
	if deleted == true {

	}
	if err != nil {

	}
	return
}

func GetMemcacheVitessConnection(host, port string, expire int64) Store {
	//configfile = NewConfigFile()
	//conf, err := goconfig.ReadConfigFile(HomeRoot + "/conf/config.cfg")
	//if err != nil {
	//	fmt.Println("error!")
	//	panic("error")
	//}
	//host, _ := conf.GetString("memcache", "host")
	//port, _ := conf.GetString("memcache", "port")
	////maxIdle, _ := conf.GetInt64("memcache", "maxidle")
	////maxActive, _ := conf.GetInt64("memcache", "maxactive")
	//expire, _ := conf.GetInt64("memcache", "expire")
	//fmt.Println(host, port, maxIdle, maxActive)

	var err error
	gMemVitessExpireTime = uint64(time.Duration(expire) * time.Minute)
	m := new(memcacheConnection)
	m.connection, err = memcache.Connect(host + ":" + port) // connection

	if err != nil {
		panic(err.Error())
	}
	//m.connection.Timeout = time.Duration(500) * time.Millisecond
	return m
}
