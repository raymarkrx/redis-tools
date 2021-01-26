package resolver

import (
	"regexp"
	"strconv"
)

type KeySpace struct {
	DbName  int
	DbSize  int64
	Expires int64
	AvgTtl  int64
}

func NewKeySpaces(s string) []KeySpace {
	reg, _ := regexp.Compile("db([0-9]+):keys=([0-9]+),expires=([0-9]+),avg_ttl=([0-9]+)")
	res := reg.FindAllStringSubmatch(s, -1)
	keySpaces := make([]KeySpace, 0)
	for _, r := range res {
		dbname, _ := strconv.Atoi(r[1])
		dbsize, _ := strconv.Atoi(r[2])
		expires, _ := strconv.Atoi(r[3])
		avgTtl, _ := strconv.Atoi(r[4])
		keySpaces = append(keySpaces, KeySpace{
			DbName:  dbname,
			DbSize:  int64(dbsize),
			Expires: int64(expires),
			AvgTtl:  int64(avgTtl),
		})
	}

	return keySpaces
}
