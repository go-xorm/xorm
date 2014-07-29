package xorm

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/garyburd/redigo/redis"
	//"github.com/go-xorm/core"
	"reflect"
	"strconv"
	"time"
)

const (
	DEFAULT = time.Duration(0)
	FOREVER = time.Duration(-1)
)

// Wraps the Redis client to meet the Cache interface.
type RedisCacher struct {
	pool              *redis.Pool
	defaultExpiration time.Duration
}

// until redigo supports sharding/clustering, only one host will be in hostList
func NewRedisCacher(host string, password string, defaultExpiration time.Duration) *RedisCacher {
	var pool = &redis.Pool{
		MaxIdle:     5,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			// the redis protocol should probably be made sett-able
			c, err := redis.Dial("tcp", host)
			if err != nil {
				return nil, err
			}
			if len(password) > 0 {
				if _, err := c.Do("AUTH", password); err != nil {
					c.Close()
					return nil, err
				}
			} else {
				// check with PING
				if _, err := c.Do("PING"); err != nil {
					c.Close()
					return nil, err
				}
			}
			return c, err
		},
		// custom connection test method
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if _, err := c.Do("PING"); err != nil {
				return err
			}
			return nil
		},
	}
	return &RedisCacher{pool, defaultExpiration}
}

func exists(conn redis.Conn, key string) bool {
	existed, _ := redis.Bool(conn.Do("EXISTS", key))
	return existed
}

func (c *RedisCacher) getBeanKey(tableName string, id string) string {
	return fmt.Sprintf("bean:%s:%s", tableName, id)
}

func (c *RedisCacher) getSqlKey(tableName string, sql string) string {
	return fmt.Sprintf("sql:%s:%s", tableName, sql)
}

func (c *RedisCacher) Flush() error {
	conn := c.pool.Get()
	defer conn.Close()
	_, err := conn.Do("FLUSHALL")
	return err
}

func (c *RedisCacher) getObject(key string) interface{} {

	conn := c.pool.Get()
	defer conn.Close()
	raw, err := conn.Do("GET", key)
	if raw == nil {
		return nil
	}
	_, err = redis.Bytes(raw, err) // TODO: item, err := redis.Bytes(raw, err)
	if err != nil {
		return err
	}
	return nil //Deserialize(item, ptrValue) // TODO
}

func (c *RedisCacher) GetIds(tableName, sql string) interface{} {

	return c.getObject(c.getSqlKey(tableName, sql))
}

func (c *RedisCacher) GetBean(tableName string, id string) interface{} {
	return c.getObject(c.getBeanKey(tableName, id))
}

func (c *RedisCacher) putObject(key string, value interface{}) {
	c.invoke(c.pool.Get().Do, key, value, c.defaultExpiration)
}

func (c *RedisCacher) PutIds(tableName, sql string, ids interface{}) {
	c.putObject(c.getBeanKey(tableName, sql), ids)
}

func (c *RedisCacher) PutBean(tableName string, id string, obj interface{}) {
	c.putObject(c.getBeanKey(tableName, id), obj)
}

func (c *RedisCacher) delObject(key string) {
	conn := c.pool.Get()
	defer conn.Close()
	if !exists(conn, key) {
		return // core.ErrCacheMiss
	}
	conn.Do("DEL", key)

	// _, err := conn.Do("DEL", key)
	// return err
}

func (c *RedisCacher) DelIds(tableName, sql string) {
	c.delObject(c.getSqlKey(tableName, sql))
	// TODO
}

func (c *RedisCacher) DelBean(tableName string, id string) {
	c.delObject(c.getBeanKey(tableName, id))
}

func (c *RedisCacher) clearObjects(key string) {
	conn := c.pool.Get()
	defer conn.Close()
	if exists(conn, key) {
		// _, err := conn.Do("DEL", key)
		// return err
		conn.Do("DEL", key)
	} else {
		// return ErrCacheMiss
	}
}

func (c *RedisCacher) ClearIds(tableName string) {
	// TODO
	c.clearObjects(c.getSqlKey(tableName, "*"))
}

func (c *RedisCacher) ClearBeans(tableName string) {
	c.clearObjects(c.getBeanKey(tableName, "*"))
}

func (c *RedisCacher) invoke(f func(string, ...interface{}) (interface{}, error),
	key string, value interface{}, expires time.Duration) error {

	switch expires {
	case DEFAULT:
		expires = c.defaultExpiration
	case FOREVER:
		expires = time.Duration(0)
	}

	b, err := Serialize(value)
	if err != nil {
		return err
	}
	conn := c.pool.Get()
	defer conn.Close()
	if expires > 0 {
		_, err := f("SETEX", key, int32(expires/time.Second), b)
		return err
	} else {
		_, err := f("SET", key, b)
		return err
	}
}

// Serialize transforms the given value into bytes following these rules:
//   - If value is a byte array, it is returned as-is.
//   - If value is an int or uint type, it is returned as the ASCII representation
//   - Else, encoding/gob is used to serialize
func Serialize(value interface{}) ([]byte, error) {
	if bytes, ok := value.([]byte); ok {
		return bytes, nil
	}

	switch v := reflect.ValueOf(value); v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return []byte(strconv.FormatInt(v.Int(), 10)), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return []byte(strconv.FormatUint(v.Uint(), 10)), nil
	}

	var b bytes.Buffer
	encoder := gob.NewEncoder(&b)
	if err := encoder.Encode(value); err != nil {
		// revel.ERROR.Printf("revel/cache: gob encoding '%s' failed: %s", value, err)
		return nil, err
	}
	return b.Bytes(), nil
}

// Deserialize transforms bytes produced by Serialize back into a Go object,
// storing it into "ptr", which must be a pointer to the value type.
func Deserialize(byt []byte, ptr interface{}) (err error) {
	if bytes, ok := ptr.(*[]byte); ok {
		*bytes = byt
		return
	}

	if v := reflect.ValueOf(ptr); v.Kind() == reflect.Ptr {
		switch p := v.Elem(); p.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			var i int64
			i, err = strconv.ParseInt(string(byt), 10, 64)
			if err != nil {
				// revel.ERROR.Printf("revel/cache: failed to parse int '%s': %s", string(byt), err)
			} else {
				p.SetInt(i)
			}
			return

		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			var i uint64
			i, err = strconv.ParseUint(string(byt), 10, 64)
			if err != nil {
				// revel.ERROR.Printf("revel/cache: failed to parse uint '%s': %s", string(byt), err)
			} else {
				p.SetUint(i)
			}
			return
		}
	}

	b := bytes.NewBuffer(byt)
	decoder := gob.NewDecoder(b)
	if err = decoder.Decode(ptr); err != nil {
		// revel.ERROR.Printf("revel/cache: gob decoding failed: %s", err)
		return
	}
	return
}
