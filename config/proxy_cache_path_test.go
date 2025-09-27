package config_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/lefeck/gonginx/parser"
	"gotest.tools/v3/assert"
)

func TestProxyCachePathParsing(t *testing.T) {
	t.Parallel()

	configString := `
http {
	proxy_cache_path /data/nginx/cache levels=1:2 keys_zone=my_cache:10m max_size=10g inactive=60m use_temp_path=off;
	proxy_cache_path /var/cache/nginx keys_zone=simple:5m;
	proxy_cache_path /tmp/cache levels=2 keys_zone=temp_cache:20m inactive=1h max_size=1g min_free=100m;
	proxy_cache_path /cache/complex levels=1:1:2 keys_zone=complex:50m max_size=5g inactive=2h manager_files=1000 manager_sleep=50ms purger=on;
	
	server {
		listen 80;
		location / {
			proxy_cache my_cache;
			proxy_pass http://backend;
		}
	}
}
`

	p := parser.NewStringParser(configString)
	conf, err := p.Parse()
	assert.NilError(t, err)

	// Test finding proxy_cache_path
	proxyCachePaths := conf.FindProxyCachePaths()
	assert.Equal(t, len(proxyCachePaths), 4)

	// Test first proxy_cache_path (full configuration)
	cache1 := proxyCachePaths[0]
	assert.Equal(t, cache1.Path, "/data/nginx/cache")
	assert.Equal(t, cache1.Levels, "1:2")
	assert.Equal(t, cache1.KeysZoneName, "my_cache")
	assert.Equal(t, cache1.KeysZoneSize, "10m")
	assert.Equal(t, cache1.MaxSize, "10g")
	assert.Equal(t, cache1.Inactive, "60m")
	assert.Assert(t, cache1.UseTemPath != nil)
	assert.Equal(t, *cache1.UseTemPath, false)

	// Test size parsing
	maxSizeBytes, err := cache1.GetMaxSizeBytes()
	assert.NilError(t, err)
	assert.Equal(t, maxSizeBytes, int64(10*1024*1024*1024)) // 10GB

	keysZoneBytes, err := cache1.GetKeysZoneSizeBytes()
	assert.NilError(t, err)
	assert.Equal(t, keysZoneBytes, int64(10*1024*1024)) // 10MB

	// Test time parsing
	inactiveDuration, err := cache1.GetInactiveDuration()
	assert.NilError(t, err)
	assert.Equal(t, inactiveDuration, 60*time.Minute)

	// Test levels parsing
	levels, err := cache1.GetLevelsDepth()
	assert.NilError(t, err)
	assert.Equal(t, len(levels), 2)
	assert.Equal(t, levels[0], 1)
	assert.Equal(t, levels[1], 2)

	// Test key capacity estimation
	keyCapacity, err := cache1.EstimateKeyCapacity()
	assert.NilError(t, err)
	expectedCapacity := int64(10*1024*1024) / 256 // 10MB / 256 bytes per key
	assert.Equal(t, keyCapacity, expectedCapacity)

	// Test second proxy_cache_path (minimal configuration)
	cache2 := proxyCachePaths[1]
	assert.Equal(t, cache2.Path, "/var/cache/nginx")
	assert.Equal(t, cache2.KeysZoneName, "simple")
	assert.Equal(t, cache2.KeysZoneSize, "5m")
	assert.Equal(t, cache2.Levels, "")         // not set
	assert.Assert(t, cache2.UseTemPath == nil) // not set

	// Test third proxy_cache_path (with min_free)
	cache3 := proxyCachePaths[2]
	assert.Equal(t, cache3.Path, "/tmp/cache")
	assert.Equal(t, cache3.Levels, "2")
	assert.Equal(t, cache3.MinFree, "100m")

	minFreeBytes, err := cache3.GetMinFreeBytes()
	assert.NilError(t, err)
	assert.Equal(t, minFreeBytes, int64(100*1024*1024)) // 100MB

	// Test fourth proxy_cache_path (complex with manager and purger)
	cache4 := proxyCachePaths[3]
	assert.Equal(t, cache4.Path, "/cache/complex")
	assert.Equal(t, cache4.Levels, "1:1:2")
	assert.Assert(t, cache4.ManagerFiles != nil)
	assert.Equal(t, *cache4.ManagerFiles, 1000)
	assert.Equal(t, cache4.ManagerSleep, "50ms")
	assert.Assert(t, cache4.Purger != nil)
	assert.Equal(t, *cache4.Purger, true)

	// Test complex levels parsing
	complexLevels, err := cache4.GetLevelsDepth()
	assert.NilError(t, err)
	assert.Equal(t, len(complexLevels), 3)
	assert.Equal(t, complexLevels[0], 1)
	assert.Equal(t, complexLevels[1], 1)
	assert.Equal(t, complexLevels[2], 2)
}

func TestFindProxyCachePathByName(t *testing.T) {
	t.Parallel()

	configString := `
http {
	proxy_cache_path /data/cache1 keys_zone=cache1:10m;
	proxy_cache_path /data/cache2 keys_zone=cache2:20m;
}
`

	p := parser.NewStringParser(configString)
	conf, err := p.Parse()
	assert.NilError(t, err)

	// Test finding existing proxy_cache_path by zone name
	cache1 := conf.FindProxyCachePathByZone("cache1")
	assert.Assert(t, cache1 != nil)
	assert.Equal(t, cache1.KeysZoneName, "cache1")
	assert.Equal(t, cache1.Path, "/data/cache1")

	// Test finding another cache
	cache2 := conf.FindProxyCachePathByZone("cache2")
	assert.Assert(t, cache2 != nil)
	assert.Equal(t, cache2.KeysZoneName, "cache2")
	assert.Equal(t, cache2.Path, "/data/cache2")

	// Test finding non-existent cache
	nonExistent := conf.FindProxyCachePathByZone("nonexistent")
	assert.Assert(t, nonExistent == nil)
}

func TestProxyCachePathManipulation(t *testing.T) {
	t.Parallel()

	configString := `
http {
	proxy_cache_path /data/cache keys_zone=my_cache:10m max_size=1g inactive=30m;
}
`

	p := parser.NewStringParser(configString)
	conf, err := p.Parse()
	assert.NilError(t, err)

	proxyCachePaths := conf.FindProxyCachePaths()
	assert.Equal(t, len(proxyCachePaths), 1)

	cache := proxyCachePaths[0]

	// Test updating max_size
	err = cache.SetMaxSize("5g")
	assert.NilError(t, err)
	assert.Equal(t, cache.MaxSize, "5g")

	// Test updating inactive time
	err = cache.SetInactive("2h")
	assert.NilError(t, err)
	assert.Equal(t, cache.Inactive, "2h")

	// Test updating keys_zone size
	err = cache.SetKeysZoneSize("50m")
	assert.NilError(t, err)
	assert.Equal(t, cache.KeysZoneSize, "50m")

	// Test setting use_temp_path
	cache.SetUseTemPath(true)
	assert.Assert(t, cache.UseTemPath != nil)
	assert.Equal(t, *cache.UseTemPath, true)

	// Test setting purger
	cache.SetPurger(true)
	assert.Assert(t, cache.Purger != nil)
	assert.Equal(t, *cache.Purger, true)
}

func TestProxyCachePathValidation(t *testing.T) {
	t.Parallel()

	configString := `
http {
	proxy_cache_path /data/cache keys_zone=my_cache:10m;
}
`

	p := parser.NewStringParser(configString)
	conf, err := p.Parse()
	assert.NilError(t, err)

	proxyCachePaths := conf.FindProxyCachePaths()
	cache := proxyCachePaths[0]

	// Test valid size updates
	err = cache.SetMaxSize("10g")
	assert.NilError(t, err)

	err = cache.SetMaxSize("500m")
	assert.NilError(t, err)

	err = cache.SetMaxSize("") // empty is valid
	assert.NilError(t, err)

	// Test invalid size formats
	err = cache.SetMaxSize("10x") // invalid unit
	assert.Error(t, err, "invalid size format: 10x (expected format: 10m, 1g, 512k, etc.)")

	err = cache.SetMaxSize("abc") // not a size
	assert.Error(t, err, "invalid size format: abc (expected format: 10m, 1g, 512k, etc.)")

	// Test valid time updates
	err = cache.SetInactive("1h")
	assert.NilError(t, err)

	err = cache.SetInactive("30m")
	assert.NilError(t, err)

	err = cache.SetInactive("") // empty is valid
	assert.NilError(t, err)

	// Test invalid time formats
	err = cache.SetInactive("10x") // invalid unit
	assert.Error(t, err, "invalid time format: 10x (expected format: 60m, 1h, 30s, 50ms, 7d)")

	err = cache.SetInactive("abc") // not a time
	assert.Error(t, err, "invalid time format: abc (expected format: 60m, 1h, 30s, 50ms, 7d)")
}

func TestProxyCachePathWithComments(t *testing.T) {
	t.Parallel()

	configString := `
http {
	# Main cache configuration
	proxy_cache_path /data/cache keys_zone=main:10m max_size=1g; # 1GB cache
	
	# Temporary cache
	proxy_cache_path /tmp/cache keys_zone=temp:5m;
}
`

	p := parser.NewStringParser(configString)
	conf, err := p.Parse()
	assert.NilError(t, err)

	proxyCachePaths := conf.FindProxyCachePaths()
	assert.Equal(t, len(proxyCachePaths), 2)

	cache1 := proxyCachePaths[0]
	// Test that comments are preserved
	assert.Assert(t, len(cache1.GetComment()) > 0)
}

func TestProxyCachePathSizeParsing(t *testing.T) {
	t.Parallel()

	// Test different size formats
	testCases := []struct {
		configSize string
		expected   int64
	}{
		{"1k", 1024},
		{"10k", 10 * 1024},
		{"1m", 1024 * 1024},
		{"10m", 10 * 1024 * 1024},
		{"1g", 1024 * 1024 * 1024},
		{"100", 100}, // bytes
	}

	for _, tc := range testCases {
		configString := fmt.Sprintf(`
http {
	proxy_cache_path /data/cache keys_zone=test:%s max_size=%s;
}
`, tc.configSize, tc.configSize)

		p := parser.NewStringParser(configString)
		conf, err := p.Parse()
		assert.NilError(t, err)

		caches := conf.FindProxyCachePaths()
		assert.Equal(t, len(caches), 1)

		cache := caches[0]

		// Test keys_zone size
		keysZoneBytes, err := cache.GetKeysZoneSizeBytes()
		assert.NilError(t, err)
		assert.Equal(t, keysZoneBytes, tc.expected, "Keys zone size parsing failed for %s", tc.configSize)

		// Test max_size
		maxSizeBytes, err := cache.GetMaxSizeBytes()
		assert.NilError(t, err)
		assert.Equal(t, maxSizeBytes, tc.expected, "Max size parsing failed for %s", tc.configSize)
	}
}

func TestProxyCachePathTimeParsing(t *testing.T) {
	t.Parallel()

	// Test different time formats
	testCases := []struct {
		time     string
		expected time.Duration
	}{
		{"50ms", 50 * time.Millisecond},
		{"100ms", 100 * time.Millisecond},
		{"30s", 30 * time.Second},
		{"5m", 5 * time.Minute},
		{"2h", 2 * time.Hour},
		{"1d", 24 * time.Hour},
		{"7d", 7 * 24 * time.Hour},
	}

	for _, tc := range testCases {
		configString := fmt.Sprintf(`
http {
	proxy_cache_path /data/cache keys_zone=test:10m inactive=%s;
}
`, tc.time)

		p := parser.NewStringParser(configString)
		conf, err := p.Parse()
		assert.NilError(t, err)

		caches := conf.FindProxyCachePaths()
		assert.Equal(t, len(caches), 1)

		cache := caches[0]
		duration, err := cache.GetInactiveDuration()
		assert.NilError(t, err)
		assert.Equal(t, duration, tc.expected, "Time parsing failed for %s", tc.time)
	}
}

func TestProxyCachePathLevelsParsing(t *testing.T) {
	t.Parallel()

	// Test different levels formats
	testCases := []struct {
		levels   string
		expected []int
	}{
		{"1", []int{1}},
		{"2", []int{2}},
		{"1:2", []int{1, 2}},
		{"2:1", []int{2, 1}},
		{"1:1:2", []int{1, 1, 2}},
		{"2:2:1", []int{2, 2, 1}},
	}

	for _, tc := range testCases {
		configString := fmt.Sprintf(`
http {
	proxy_cache_path /data/cache levels=%s keys_zone=test:10m;
}
`, tc.levels)

		p := parser.NewStringParser(configString)
		conf, err := p.Parse()
		assert.NilError(t, err)

		caches := conf.FindProxyCachePaths()
		assert.Equal(t, len(caches), 1)

		cache := caches[0]
		levels, err := cache.GetLevelsDepth()
		assert.NilError(t, err)
		assert.Equal(t, len(levels), len(tc.expected))
		for i, expected := range tc.expected {
			assert.Equal(t, levels[i], expected, "Levels parsing failed for %s at position %d", tc.levels, i)
		}
	}
}

func TestProxyCachePathKeyCapacity(t *testing.T) {
	t.Parallel()

	// Test key capacity estimation for different sizes
	testCases := []struct {
		size        string
		sizeBytes   int64
		keyCapacity int64
	}{
		{"1m", 1024 * 1024, 4096},           // 1MB / 256
		{"10m", 10 * 1024 * 1024, 40960},    // 10MB / 256
		{"100m", 100 * 1024 * 1024, 409600}, // 100MB / 256
		{"1g", 1024 * 1024 * 1024, 4194304}, // 1GB / 256
	}

	for _, tc := range testCases {
		configString := fmt.Sprintf(`
http {
	proxy_cache_path /data/cache keys_zone=test:%s;
}
`, tc.size)

		p := parser.NewStringParser(configString)
		conf, err := p.Parse()
		assert.NilError(t, err)

		caches := conf.FindProxyCachePaths()
		cache := caches[0]

		sizeBytes, err := cache.GetKeysZoneSizeBytes()
		assert.NilError(t, err)
		assert.Equal(t, sizeBytes, tc.sizeBytes)

		keyCapacity, err := cache.EstimateKeyCapacity()
		assert.NilError(t, err)
		assert.Equal(t, keyCapacity, tc.keyCapacity, "Key capacity estimation failed for %s", tc.size)
	}
}

func TestInvalidProxyCachePathDirective(t *testing.T) {
	t.Parallel()

	// Test proxy_cache_path directive with insufficient parameters
	configString := `
http {
	proxy_cache_path /data/cache;
}
`

	p := parser.NewStringParser(configString)
	_, err := p.Parse()
	// Should return an error due to missing keys_zone parameter
	assert.Error(t, err, "proxy_cache_path directive requires at least 2 parameters: path and keys_zone")
}

func TestInvalidKeysZoneFormat(t *testing.T) {
	t.Parallel()

	// Test invalid keys_zone format
	configString := `
http {
	proxy_cache_path /data/cache keys_zone=invalid;
}
`

	p := parser.NewStringParser(configString)
	_, err := p.Parse()
	// Should return an error due to invalid keys_zone format
	assert.Error(t, err, "invalid keys_zone format: invalid (expected format: name:size)")
}

func TestInvalidLevelsFormat(t *testing.T) {
	t.Parallel()

	// Test invalid levels format
	configString := `
http {
	proxy_cache_path /data/cache levels=3 keys_zone=test:10m;
}
`

	p := parser.NewStringParser(configString)
	_, err := p.Parse()
	// Should return an error due to invalid levels format
	assert.Error(t, err, "level depth must be 1 or 2, got 3")
}

func TestUnknownCacheParameter(t *testing.T) {
	t.Parallel()

	// Test unknown parameter
	configString := `
http {
	proxy_cache_path /data/cache keys_zone=test:10m unknown=value;
}
`

	p := parser.NewStringParser(configString)
	_, err := p.Parse()
	// Should return an error due to unknown parameter
	assert.Error(t, err, "unknown parameter: unknown=value")
}

func TestProxyCachePathComplexConfiguration(t *testing.T) {
	t.Parallel()

	configString := `
http {
	proxy_cache_path /var/cache/nginx/complex 
		levels=1:2:2 
		keys_zone=complex_cache:128m 
		max_size=50g 
		inactive=7d 
		use_temp_path=off 
		manager_files=10000 
		manager_sleep=100ms 
		manager_threshold=500ms 
		loader_files=5000 
		loader_sleep=50ms 
		loader_threshold=200ms 
		purger=on 
		purger_files=1000 
		purger_sleep=10ms 
		purger_threshold=100ms 
		min_free=1g;
}
`

	p := parser.NewStringParser(configString)
	conf, err := p.Parse()
	assert.NilError(t, err)

	caches := conf.FindProxyCachePaths()
	assert.Equal(t, len(caches), 1)

	cache := caches[0]

	// Test all parameters
	assert.Equal(t, cache.Path, "/var/cache/nginx/complex")
	assert.Equal(t, cache.Levels, "1:2:2")
	assert.Equal(t, cache.KeysZoneName, "complex_cache")
	assert.Equal(t, cache.KeysZoneSize, "128m")
	assert.Equal(t, cache.MaxSize, "50g")
	assert.Equal(t, cache.Inactive, "7d")
	assert.Equal(t, cache.MinFree, "1g")

	assert.Assert(t, cache.UseTemPath != nil)
	assert.Equal(t, *cache.UseTemPath, false)

	assert.Assert(t, cache.ManagerFiles != nil)
	assert.Equal(t, *cache.ManagerFiles, 10000)
	assert.Equal(t, cache.ManagerSleep, "100ms")
	assert.Equal(t, cache.ManagerThreshold, "500ms")

	assert.Assert(t, cache.LoaderFiles != nil)
	assert.Equal(t, *cache.LoaderFiles, 5000)
	assert.Equal(t, cache.LoaderSleep, "50ms")
	assert.Equal(t, cache.LoaderThreshold, "200ms")

	assert.Assert(t, cache.Purger != nil)
	assert.Equal(t, *cache.Purger, true)
	assert.Assert(t, cache.PurgerFiles != nil)
	assert.Equal(t, *cache.PurgerFiles, 1000)
	assert.Equal(t, cache.PurgerSleep, "10ms")
	assert.Equal(t, cache.PurgerThreshold, "100ms")

	// Test size calculations
	maxSizeBytes, err := cache.GetMaxSizeBytes()
	assert.NilError(t, err)
	assert.Equal(t, maxSizeBytes, int64(50*1024*1024*1024)) // 50GB

	minFreeBytes, err := cache.GetMinFreeBytes()
	assert.NilError(t, err)
	assert.Equal(t, minFreeBytes, int64(1024*1024*1024)) // 1GB

	keysZoneBytes, err := cache.GetKeysZoneSizeBytes()
	assert.NilError(t, err)
	assert.Equal(t, keysZoneBytes, int64(128*1024*1024)) // 128MB

	// Test time calculation
	inactiveDuration, err := cache.GetInactiveDuration()
	assert.NilError(t, err)
	assert.Equal(t, inactiveDuration, 7*24*time.Hour) // 7 days

	// Test levels parsing
	levels, err := cache.GetLevelsDepth()
	assert.NilError(t, err)
	assert.Equal(t, len(levels), 3)
	assert.Equal(t, levels[0], 1)
	assert.Equal(t, levels[1], 2)
	assert.Equal(t, levels[2], 2)

	// Test key capacity
	keyCapacity, err := cache.EstimateKeyCapacity()
	assert.NilError(t, err)
	expectedCapacity := int64(128*1024*1024) / 256 // 128MB / 256 bytes per key
	assert.Equal(t, keyCapacity, expectedCapacity)
}
