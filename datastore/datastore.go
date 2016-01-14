package datastore

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/alexgear/checker/common"
	"github.com/boltdb/bolt"
	"github.com/montanaflynn/stats"
)

var err error
var db *bolt.DB

func InitDB() (*bolt.DB, error) {
	db, err = bolt.Open("my.db", 0600, &bolt.Options{Timeout: 5 * time.Second})
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

var cache = make(map[string]map[time.Time][]common.Response)

func average(c map[time.Time][]common.Response) (map[time.Time]common.Status, error) {
	result := make(map[time.Time]common.Status)
	for t, responses := range c {
		if time.Since(t).Seconds() > 5.0 {
			var s common.Status
			var latency []float64
			var isUps []bool
			for _, r := range responses {
				latency = append(latency, r.Latency.Seconds())
				isUps = append(isUps, r.IsUp)
			}
			s.Mean, err = stats.Mean(latency)
			if err != nil {
				return result, fmt.Errorf("Failed to calculate mean: %s", err.Error())
			}
			s.Percentile90, err = stats.Percentile(latency, 90.0)
			if err != nil {
				return result, fmt.Errorf("Failed to calculate 90th percentile: %s", err.Error())
			}
			s.Percentile99, err = stats.Percentile(latency, 99.0)
			if err != nil {
				return result, fmt.Errorf("Failed to calculate 99th percentile: %s", err.Error())
			}
			var uptimeUp int
			for _, up := range isUps {
				if up {
					uptimeUp += 1
				}
			}
			s.Uptime = float64(uptimeUp) * 100 / float64(len(isUps))
			result[t] = s
		}
	}
	return result, nil
}

func FlushCache() error {
	start := time.Now()
	for ief, c := range cache {
		status, err := average(c)
		if err != nil {
			return fmt.Errorf("Failed to caculate averages of cached data: %s", err.Error())
		}
		// Invalidate cache
		for t, _ := range c {
			if start.Sub(t).Seconds() > 5.0 {
				delete(cache[ief], t)
			}
		}
		// Write to db
		err = db.Update(func(tx *bolt.Tx) error {
			for t, s := range status {
				b := tx.Bucket([]byte(ief)).Bucket([]byte("status"))
				sEncoded, err := json.Marshal(s)
				if err != nil {
					return fmt.Errorf("Failed to encode to json: %s", err.Error())
				}
				err = b.Put([]byte(t.UTC().Format(time.RFC3339)), sEncoded)
				if err != nil {
					return fmt.Errorf("update bucket: %s", err.Error())
				}
			}
			return nil
		})
		if err != nil {
			fmt.Errorf("Failed to write to db: %s", err.Error())
		}
	}
}

func Write(ief string, r common.Response) error {
	if cache[ief] == nil {
		cache[ief] = make(map[time.Time][]common.Response)
	}
	cache[ief][r.Time.Round(time.Second)] = append(cache[ief][r.Time.Round(time.Second)], r)
	return nil
}

func Read(ief string, timeDelta time.Duration) (map[time.Time]common.Status, error) {
	min := time.Now().UTC().Add(-1 * timeDelta)
	status := make(map[time.Time]common.Status)
	err = db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(ief)).Bucket([]byte("status")).Cursor()
		for k, v := c.Seek([]byte(min.Format(time.RFC3339))); k != nil; k, v = c.Next() {
			var s common.Status
			err = json.Unmarshal(v, &s)
			if err != nil {
				return fmt.Errorf("Failed to decode bytes: %s", err.Error())
			}
			//fmt.Printf("time=%s, latency=%s, uptime=%s\n", k, s.Mean, s.Uptime)
			t, err := time.Parse(time.RFC3339, string(k))
			if err != nil {
				return fmt.Errorf("Failed to parse time: %s", err.Error())
			}
			status[t] = s
		}
		return nil
	})
	if err != nil {
		return status, fmt.Errorf("Failed to query db: %s", err.Error())
	}
	return status, nil
}
