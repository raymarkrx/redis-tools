package instance

import (
	"context"
	"encoding/json"
	"github.com/go-redis/redis/v8"
	"github.com/wooos/redis-tools/internal/resolver"
	"io/ioutil"
	"log"
	"os"
	"time"
)

type Key struct {
	Name  string        `json:"name"`
	Type  string        `json:"type"`
	Value interface{}   `json:"value"`
	Ttl   time.Duration `json:"ttl"`
}

type Database struct {
	DbName int   `json:"db_name"`
	DbSize int64 `json:"db_size"`
	Keys   []Key `json:"keys"`
}

type Instance struct {
	Databases []Database `json:"databases"`
}

func (i *Instance) Restore(conn *redis.Conn) {
	ctx, _ := context.WithTimeout(context.TODO(), time.Second*3)

	for _, database := range i.Databases {
		if err := conn.Select(ctx, database.DbName).Err(); err != nil {
			log.Fatalf("Cannot change database %d, error: %v\n", database.DbName, err)
		}

		log.Printf("Restore db: %d, keys: %d\n", database.DbName, database.DbSize)
		for _, key := range database.Keys {
			switch key.Type {
			case "string":
				if err := conn.Set(ctx, key.Name, key.Value, 0).Err(); err != nil {
					log.Printf("Cannot exec set command, key: %s, val: %s, ttl: %d, error: %v\n", key.Name, key.Value, key.Ttl, err)
				}
			case "hash":
				val := key.Value.(map[string]interface{})
				for k, v := range val {
					if err := conn.HSetNX(ctx, key.Name, k, v).Err(); err != nil {
						log.Printf("Cannot exec hset command, key: %s, val: %v, error: %v\n", key.Name, key.Value, err)
					}
				}
			case "list":
				if err := conn.LPush(ctx, key.Name, key.Value).Err(); err != nil {
					log.Printf("Cannot exec lpush command, key: %s, val: %v, error: %v\n", key.Name, key.Value, err)
				}
			case "set":
				if err := conn.SAdd(ctx, key.Name, key.Value).Err(); err != nil {
					log.Printf("Cannot exec sadd command, key: %s, val: %v, error: %v\n", key.Name, key.Value, err)
				}
			case "zset":
				val := key.Value.([]interface{})
				zset := make([]*redis.Z, 0)
				for k, v := range val {
					zset = append(zset, &redis.Z{Score: float64(k), Member: v})
				}

				if err := conn.ZAdd(ctx, key.Name, zset...).Err(); err != nil {
					log.Printf("Cannot exec set command, key: %s, val: %v, error: %v\n", key.Name, key.Value, err)
				}
			}

			if key.Ttl > 0 {
				if err := conn.Expire(ctx, key.Name, key.Ttl).Err(); err != nil {
					log.Printf("Cannot exec expire command, key: %s, ttl: %d, error: %v\n", key.Name, key.Ttl, err)
				}
			}
		}
		log.Printf("Restore done.")
	}
}

func (i *Instance) Dump(file string) error {
	data, err := json.Marshal(i)
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(file, data, os.ModePerm); err != nil {
		return err
	}

	return nil
}

func LoadFromFile(file string) (Instance, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return Instance{}, err
	}

	instance := Instance{}
	if err = json.Unmarshal(data, &instance); err != nil {
		return Instance{}, err
	}

	return instance, nil
}

func LoadFromRedis(conn *redis.Conn, dbs []resolver.KeySpace) (Instance, error) {
	ctx, _ := context.WithTimeout(context.TODO(), time.Second*3)

	databases := make([]Database, 0)
	for _, db := range dbs {
		if _, err := conn.Select(ctx, db.DbName).Result(); err != nil {
			log.Fatalf("Cannot change database to %d, error: %v\n", db.DbName, err)
		}

		keys, _, err := conn.Scan(ctx, 0, "*", db.DbSize+1).Result()
		if err != nil {
			log.Fatalf("Cannot scan keys, error: %v\n", err)
		}

		keyss := make([]Key, 0)
		for _, key := range keys {
			t, err := conn.Type(ctx, key).Result()
			if err != nil {
				log.Fatalf("Connot get \"%s\" type, error: %v\n", key, err)
			}

			var val interface{}
			switch t {
			case "string":
				val, err = conn.Get(ctx, key).Result()
				if err != nil {
					log.Printf("Connot get value, key: %s, error: %v\n", key, err)
					break
				}
			case "hash":
				val, err = conn.HGetAll(ctx, key).Result()
				if err != nil {
					log.Printf("Connot get value, key: %s, error: %v\n", key, err)
					break
				}
			case "list":
				length, err := conn.LLen(ctx, key).Result()
				if err != nil {
					log.Printf("Cannot get length, key: %s, error: %v\n", key, err)
					break
				}
				val, err = conn.LRange(ctx, key, 0, length).Result()
				if err != nil {
					log.Printf("Connot get value, key: %s, error: %v\n", key, err)
					break
				}
			case "set":
				val, err = conn.SMembers(ctx, key).Result()
				if err != nil {
					log.Printf("Connot get value, key: %s, error: %v\n", key, err)
					break
				}
			case "zset":
				length, err := conn.ZCard(ctx, key).Result()
				if err != nil {
					log.Printf("Cannot get length, key: %s, error: %v\n", key, err)
					break
				}
				val, err = conn.ZRange(ctx, key, 0, length).Result()
				if err != nil {
					log.Printf("Connot get value, key: %s, error: %v\n", key, err)
					break
				}
			}

			ttl, err := conn.TTL(ctx, key).Result()
			if err != nil {
				log.Printf("Cannot get ttl for key: %s, error: %v\n", key, err)
			}
			keyss = append(keyss, Key{Name: key, Type: t, Value: val, Ttl: ttl})
		}
		databases = append(databases, Database{DbName: db.DbName, DbSize: db.DbSize, Keys: keyss})
	}

	return Instance{Databases: databases}, nil
}
