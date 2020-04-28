package util

import cmap "github.com/orcaman/concurrent-map"

type SQLCache struct {
	cache cmap.ConcurrentMap
}

func (s *SQLCache) Set(sampleName, sql string, result []interface{}) {
	s.cache.SetIfAbsent(sampleName, cmap.New())
	cached, _ := s.cache.Get(sampleName)
	cached.(cmap.ConcurrentMap).SetIfAbsent(sql, result)
}

func (s SQLCache) Get(sampleName, sql string) ([]interface{}, bool) {
	cached, ok := s.cache.Get(sampleName)
	if !ok {
		return nil, ok
	}
	res, ok := cached.(cmap.ConcurrentMap).Get(sql)
	return res.([]interface{}), ok
}

func New() SQLCache {
	return SQLCache{cmap.New()}
}
