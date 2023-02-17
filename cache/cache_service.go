package cache

import (
	"encoding/json"

	"fmt"
	"reflect"
	"strings"
	"time"

	"errors"

	"github.com/beego/beego/v2/core/logs"
	beego "github.com/beego/beego/v2/server/web"
	"github.com/go-redis/redis/v7"
	"github.com/mobilemindtec/go-utils/support"
	"github.com/mobilemindtec/go-utils/v2/lists"
	"github.com/mobilemindtec/go-utils/v2/optional"
)

const (
	DefaultDuration = 5 * 60 * 1000 // 5 min
)

type CacheService struct {
	rdb            *redis.Client
	duration       int // milisec
	sessionKashKey string
}

func New() *CacheService {
	v := &CacheService{duration: DefaultDuration}
	v.init()
	return v
}

func (this *CacheService) init() {

	sessionproviderconfig, _ := beego.AppConfig.String("sessionproviderconfig")
	this.sessionKashKey, _ = beego.AppConfig.String("cachesessionhashkey")

	//logs.Debug("CONNECT REDIS")

	this.rdb = redis.NewClient(&redis.Options{
		Addr:     sessionproviderconfig,
		Password: "",
		DB:       0,
		PoolSize: 5,
	})

}

func (this *CacheService) Close() {
	this.rdb.Close()
}

func (this *CacheService) ExpiresMill(duration int) *CacheService {
	this.duration = duration
	return this
}

func (this *CacheService) ExpiresSec(duration int) *CacheService {
	this.duration = duration * 1000
	return this
}

func (this *CacheService) ExpiresMin(duration int) *CacheService {
	this.duration = duration * 1000 * 60
	return this
}

func (this *CacheService) NewExpiresMin(duration int) *CacheService {
	return &CacheService{duration: duration * 60 * 1000, rdb: this.rdb}
}

func (this *CacheService) NewExpiresSec(duration int) *CacheService {
	return &CacheService{duration: duration * 1000, rdb: this.rdb}
}

func (this *CacheService) NewExpiresMill(duration int) *CacheService {
	return &CacheService{duration: duration, rdb: this.rdb}
}

func (this *CacheService) getSessionKey(key string) string {
	return fmt.Sprintf("%v_%v", this.sessionKashKey, key)
}

func (this *CacheService) Put(key string, value interface{}) {

	payload, err := json.Marshal(value)

	if err != nil {
		logs.Error("error convert value to json: %v", err)
		return
	}

	if err := this.rdb.Set(this.getSessionKey(key), string(payload), time.Duration(this.duration)*time.Millisecond).Err(); err != nil {
		logs.Error("error save cache on redis: %v", err)
		return
	}

	logs.Debug("%v cached saved on redis", key)
}

func (this *CacheService) Get(key string, value interface{}) interface{} {
	r, err := this.rdb.Get(this.getSessionKey(key)).Result()

	//logs.Debug("REDIS RESULT: %v", r)
	//logs.Debug("REDIS ERR: %v", err)

	if err == nil {
		err = json.Unmarshal([]byte(r), value)
	} else if err == redis.Nil {
		return optional.NewNone()
	}

	return optional.Make(value, err)
}

func (this *CacheService) GetVal(key string) interface{} {
	value, err := this.rdb.Get(this.getSessionKey(key)).Result()

	if err == redis.Nil {
		return optional.NewNone()
	}

	return optional.Make(value, err)
}

func (this *CacheService) Memoize(key string, value interface{}, cacheable func() interface{}) (interface{}, error) {
	r := this.MemoizeOpt(key, value, cacheable)

	switch r.(type) {
	case *optional.Some:
		return r.(*optional.Some).Item, nil
	case *optional.Fail:
		return nil, r.(*optional.Fail).Error
	case error:
		return nil, r.(error)
	default:
		return nil, errors.New("empty result")
	}
}

func (this *CacheService) MemoizeOpt(key string, value interface{}, cacheable func() interface{}) interface{} {
	v := this.Get(key, value)

	switch v.(type) {
	case *optional.Some, *optional.Fail:
		logs.Debug("CACHE: get key %v from cache", key)
		return v
	default:
		data := cacheable()

		logs.Debug("CACHE: try get key %v from cacheble func: %v", key, data)

		if data == nil {
			return optional.NewNone()
		}

		switch data.(type) {
		case *optional.Some:
			this.Put(key, data.(*optional.Some).Item)
			return data
		case *optional.Fail, *optional.None:
			return data
		case error:
			return optional.NewFail(data.(error))
		default:
			this.Put(key, data)
			return optional.NewSome(data)
		}
	}
}

func (this *CacheService) Delete(keys ...string) *CacheService {
	for _, key := range keys {
		this.rdb.Del(this.getSessionKey(key))
	}
	return this
}

func Cached[T any](value interface{}) (T, bool) {
	var x T
	switch value.(type) {
	case *optional.Some:
		return value.(*optional.Some).Item.(T), true
	case *optional.Fail:
		return x, false
	default:
		if v, ok := value.(T); ok {
			return v, true
		}
		return x, false
	}
}

func CacheKey(args ...interface{}) string {

	replacements := lists.Map[interface{}, string](args, func(v interface{}) string { return fmt.Sprintf("%v", v) })

	return fmt.Sprintf("key_%v", strings.Join(replacements, "_"))
}

func Memoize[T any](srv *CacheService, key string, value interface{}, cacheable func() (T, error)) (T, error) {
	r, err := srv.Memoize(key, value, func() interface{} {
		v, err := cacheable()
		if err != nil {
			return err
		}
		return v
	})

	var x T

	if err != nil {
		return x, err
	}

	if r == nil {
		return x, nil
	}

	if t, ok := r.(T); ok {
		return t, nil
	}

	if reflect.ValueOf(r).Kind() == reflect.Ptr {
		return reflect.ValueOf(r).Elem().Interface().(T), nil
	}

	return x, errors.New("can't convert result to T")

}

func MemoizeOpt[T any](srv *CacheService, key string, value interface{}, cacheable func() *optional.Optional[T]) *optional.Optional[T] {
	r := srv.MemoizeOpt(key, value, func() interface{} {
		return cacheable()
	})

	return r.(*optional.Optional[T])
}

func MemoizeVal[T any](srv *CacheService, key string, parser func(string) T, cacheable func() T) T {
	v := srv.GetVal(key)

	switch v.(type) {
	case *optional.Some:

		logs.Debug("CACHE: get key %v from cache", key)

		raw := v.(*optional.Some).Item.(string)
		return parser(raw)

	default:
		value := cacheable()
		srv.Put(key, value)
		return value
	}
}

func TryGet[T any](srv *CacheService, key string, value interface{}) (T, bool) {
	v := srv.Get(key, value)
	var x T

	switch v.(type) {
	case *optional.Some:

		logs.Debug("CACHE: get key %v from cache", key)

		r := v.(*optional.Some).Item

		if t, ok := r.(T); ok {
			return t, true
		}

		if reflect.ValueOf(r).Kind() == reflect.Ptr {
			return reflect.ValueOf(r).Elem().Interface().(T), true
		}
		return x, false

	default:
		return x, false
	}
}

func KeyHash(data interface{}) (string, error) {
	r, err := json.Marshal(data)

	if err != nil {
		return "", nil
	}

	return support.TextToSha1(string(r)), nil
}
