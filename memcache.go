package captcha

import (
	//"fmt"
	"github.com/bradfitz/gomemcache/memcache"
	//"github.com/miguel-branco/goconfig"
	"log"
	//"log/syslog"
	"os"
	"time"
)

var gMemExpireTime int32 = 60 //default is 1 minute

type memcacheClient struct {
	client *memcache.Client
	//logger *log.Logger
}

func (mClient *memcacheClient) Set(id string, digits []byte) {
	// item := new(memcache.Item)
	// item.Key = id
	// item.Value = digits
	// item.Expiration = gMemExpireTime
	err := mClient.client.Set(&memcache.Item{Key: id, Value: digits, Expiration: gMemExpireTime})

	if err != nil {
		panic(err)
		return
	}

}

func (mClient *memcacheClient) Get(id string, clear bool) (digits []byte) {
	item, err := mClient.client.Get(id)
	if err != nil {
		return
	}
	digits = item.Value
	if !clear {
		return
	}
	err = mClient.client.Delete(id)
	if err != nil {
		f, err := os.OpenFile("log.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			return
		}
		defer f.Close()
		log.SetOutput(f)
		log.Println("set value fail" + err.Error())
	}
	return
}

func GetMemcacheClient(host, port string, expire int64) Store {
	//configfile = NewConfigFile()
	//conf, err := goconfig.ReadConfigFile(HomeRoot + "/conf/config.cfg")
	//if err != nil {
	//fmt.Println("error!")
	//panic("error")
	//}
	//host, _ := conf.GetString("memcache", "host")
	//port, _ := conf.GetString("memcache", "port")
	//maxIdle, _ := conf.GetInt64("memcache", "maxidle")
	//maxActive, _ := conf.GetInt64("memcache", "maxactive")
	//expire, _ := conf.GetInt64("memcache", "expire")
	//fmt.Println(host, port, maxIdle, maxActive)
	gMemExpireTime = int32(time.Duration(expire) * time.Minute)
	m := new(memcacheClient)
	m.client = memcache.New(host + ":" + port) // client
	//m.logger = Dial("", "127.0.0.1", Log)
	m.client.Timeout = time.Duration(500) * time.Millisecond
	return m
}
