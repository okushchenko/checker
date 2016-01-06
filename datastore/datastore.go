package datastore

import (
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/alexgear/checker/common"
	"github.com/boltdb/bolt"
)

var err error
var db *bolt.DB

func InitDB() (*bolt.DB, error) {
	db, err = bolt.Open("my.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}
	buckets := []string{"wifi", "lan"}
	subbuckets := []string{"latency", "status"}
	err = db.Update(func(tx *bolt.Tx) error {
		for _, bucket := range buckets {
			b, err := tx.CreateBucketIfNotExists([]byte(bucket))
			if err != nil {
				return fmt.Errorf("create bucket: %s", err.Error())
			}
			for _, subbucket := range subbuckets {
				_, err = b.CreateBucketIfNotExists([]byte(subbucket))
				if err != nil {
					return fmt.Errorf("create subbucket: %s", err.Error())
				}

			}
		}
		return nil
	})
	return db, err
}

func Write(ief string, r common.Response) error {
	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(ief)).Bucket([]byte("latency"))
		err = b.Put([]byte(r.Time.Format(time.RFC3339Nano)), []byte(r.Latency.String()))
		if err != nil {
			return fmt.Errorf("update bucket: %s", err.Error())
		}
		b = tx.Bucket([]byte(ief)).Bucket([]byte("status"))
		err = b.Put([]byte(r.Time.Format(time.RFC3339Nano)), []byte(fmt.Sprint(r.Status)))
		if err != nil {
			return fmt.Errorf("update bucket: %s", err.Error())
		}
		return nil
	})
	return err
}

func Read(ief string) (map[time.Time]bool, map[time.Time]float64, error) {
	min := time.Now().Add(-24 * time.Hour)
	latency := make(map[time.Time]float64)
	status := make(map[time.Time]bool)
	err = db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(ief)).Bucket([]byte("latency")).Cursor()
		for k, v := c.Seek([]byte(min.Format(time.RFC3339Nano))); k != nil; k, v = c.Next() {
			//fmt.Printf("time=%s, latency=%s", k, v)
			v1, err := time.ParseDuration(string(v))
			if err != nil {
				return fmt.Errorf("Failed to parse duration: %s", err.Error())
			}
			t, err := time.Parse(time.RFC3339Nano, string(k))
			if err != nil {
				return fmt.Errorf("Failed to parse time: %s", err.Error())
			}
			latency[t] = math.Max(v1.Seconds(), latency[t])

			c := tx.Bucket([]byte(ief)).Bucket([]byte("status")).Cursor()
			_, v = c.Seek([]byte(k))
			v2, err := strconv.ParseBool(string(v))
			if err != nil {
				return fmt.Errorf("Failed to parse status: %s", err.Error())
			}
			status[t] = v2
			//fmt.Printf(", status=%s\n", v)
		}
		return nil
	})
	if err != nil {
		return status, latency, fmt.Errorf("Failed to query db: %s", err.Error())
	}
	return status, latency, nil
}
