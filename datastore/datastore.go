package datastore

import (
	"bytes"
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
		err = b.Put([]byte(r.Time.Format(time.RFC3339)), []byte(r.Latency.String()))
		if err != nil {
			return fmt.Errorf("update bucket: %s", err.Error())
		}
		b = tx.Bucket([]byte(ief)).Bucket([]byte("status"))
		err = b.Put([]byte(r.Time.Format(time.RFC3339)), []byte(fmt.Sprint(r.Status)))
		if err != nil {
			return fmt.Errorf("update bucket: %s", err.Error())
		}
		return nil
	})
	return err
}

func Read(ief string) ([]bool, []float64, error) {
	max := time.Now()
	min := max.Add(-1 * time.Hour)
	latency := make([]float64, 3600)
	status := make([]bool, 3600)
	err = db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(ief)).Bucket([]byte("latency")).Cursor()
		for k, v := c.Seek([]byte(min.Format(time.RFC3339))); k != nil && bytes.Compare(k, []byte(max.Format(time.RFC3339))) <= 0; k, v = c.Next() {
			//fmt.Printf("time=%s, latency=%s", k, v)
			v1, err := time.ParseDuration(string(v))
			if err != nil {
				return fmt.Errorf("Failed to parse duration: %s", err.Error())
			}
			t, err := time.Parse(time.RFC3339, string(k))
			if err != nil {
				return fmt.Errorf("Failed to parse time: %s", err.Error())
			}
			i := int(t.Sub(min).Seconds())
			if latency[i] == 0.0 {
				latency[i] = v1.Seconds()
			} else {
				latency[i] = math.Max(v1.Seconds(), latency[i])
			}

			c := tx.Bucket([]byte(ief)).Bucket([]byte("status")).Cursor()
			_, v = c.Seek([]byte(k))
			v2, err := strconv.ParseBool(string(v))
			if err != nil {
				return fmt.Errorf("Failed to parse status: %s", err.Error())
			}
			status[i] = v2
			//fmt.Printf(", status=%s\n", v)
		}
		return nil
	})
	if err != nil {
		return status, latency, fmt.Errorf("Failed to query db: %s", err.Error())
	}
	return status, latency, nil
}
