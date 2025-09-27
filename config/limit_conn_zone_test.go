package config_test

import (
	"fmt"
	"testing"

	"github.com/lefeck/gonginx/parser"
	"gotest.tools/v3/assert"
)

func TestLimitConnZoneParsing(t *testing.T) {
	t.Parallel()

	configString := `
http {
	limit_conn_zone $binary_remote_addr zone=addr:10m;
	limit_conn_zone $server_name zone=perserver:5m;
	limit_conn_zone $request_uri zone=peruri:20m sync;
	limit_conn_zone $cookie_user_id zone=peruser:1g;
	
	server {
		listen 80;
		location / {
			limit_conn addr 10;
			proxy_pass http://backend;
		}
	}
}
`

	p := parser.NewStringParser(configString)
	conf, err := p.Parse()
	assert.NilError(t, err)

	// Test finding limit_conn_zone
	limitConnZones := conf.FindLimitConnZones()
	assert.Equal(t, len(limitConnZones), 4)

	// Test first limit_conn_zone
	zone1 := limitConnZones[0]
	assert.Equal(t, zone1.Key, "$binary_remote_addr")
	assert.Equal(t, zone1.ZoneName, "addr")
	assert.Equal(t, zone1.ZoneSize, "10m")
	assert.Equal(t, zone1.Sync, false)

	// Test zone size parsing
	sizeBytes, err := zone1.GetZoneSizeBytes()
	assert.NilError(t, err)
	assert.Equal(t, sizeBytes, int64(10*1024*1024)) // 10MB

	// Test connection estimation
	maxConnections, err := zone1.EstimateMaxConnections()
	assert.NilError(t, err)
	expectedConnections := int64(10*1024*1024) / 64 // 10MB / 64 bytes per connection
	assert.Equal(t, maxConnections, expectedConnections)

	// Test second limit_conn_zone
	zone2 := limitConnZones[1]
	assert.Equal(t, zone2.Key, "$server_name")
	assert.Equal(t, zone2.ZoneName, "perserver")
	assert.Equal(t, zone2.ZoneSize, "5m")

	// Test third limit_conn_zone (with sync)
	zone3 := limitConnZones[2]
	assert.Equal(t, zone3.Key, "$request_uri")
	assert.Equal(t, zone3.ZoneName, "peruri")
	assert.Equal(t, zone3.ZoneSize, "20m")
	assert.Equal(t, zone3.Sync, true)

	// Test fourth limit_conn_zone (with gigabyte size)
	zone4 := limitConnZones[3]
	assert.Equal(t, zone4.Key, "$cookie_user_id")
	assert.Equal(t, zone4.ZoneName, "peruser")
	assert.Equal(t, zone4.ZoneSize, "1g")

	// Test gigabyte size parsing
	sizeBytes4, err := zone4.GetZoneSizeBytes()
	assert.NilError(t, err)
	assert.Equal(t, sizeBytes4, int64(1024*1024*1024)) // 1GB
}

func TestFindLimitConnZoneByName(t *testing.T) {
	t.Parallel()

	configString := `
http {
	limit_conn_zone $binary_remote_addr zone=addr:10m;
	limit_conn_zone $server_name zone=perserver:5m;
}
`

	p := parser.NewStringParser(configString)
	conf, err := p.Parse()
	assert.NilError(t, err)

	// Test finding existing limit_conn_zone by name
	zone1 := conf.FindLimitConnZoneByName("addr")
	assert.Assert(t, zone1 != nil)
	assert.Equal(t, zone1.ZoneName, "addr")
	assert.Equal(t, zone1.Key, "$binary_remote_addr")

	// Test finding another zone
	zoneServer := conf.FindLimitConnZoneByName("perserver")
	assert.Assert(t, zoneServer != nil)
	assert.Equal(t, zoneServer.ZoneName, "perserver")
	assert.Equal(t, zoneServer.Key, "$server_name")

	// Test finding non-existent zone
	nonExistent := conf.FindLimitConnZoneByName("nonexistent")
	assert.Assert(t, nonExistent == nil)
}

func TestLimitConnZoneManipulation(t *testing.T) {
	t.Parallel()

	configString := `
http {
	limit_conn_zone $binary_remote_addr zone=addr:10m;
}
`

	p := parser.NewStringParser(configString)
	conf, err := p.Parse()
	assert.NilError(t, err)

	limitConnZones := conf.FindLimitConnZones()
	assert.Equal(t, len(limitConnZones), 1)

	zone := limitConnZones[0]

	// Test updating zone size
	err = zone.SetZoneSize("20m")
	assert.NilError(t, err)
	assert.Equal(t, zone.ZoneSize, "20m")

	// Verify new size in bytes
	newSizeBytes, err := zone.GetZoneSizeBytes()
	assert.NilError(t, err)
	assert.Equal(t, newSizeBytes, int64(20*1024*1024)) // 20MB

	// Test enabling sync
	zone.SetSync(true)
	assert.Equal(t, zone.Sync, true)

	// Test disabling sync
	zone.SetSync(false)
	assert.Equal(t, zone.Sync, false)
}

func TestLimitConnZoneValidation(t *testing.T) {
	t.Parallel()

	configString := `
http {
	limit_conn_zone $binary_remote_addr zone=addr:10m;
}
`

	p := parser.NewStringParser(configString)
	conf, err := p.Parse()
	assert.NilError(t, err)

	limitConnZones := conf.FindLimitConnZones()
	zone := limitConnZones[0]

	// Test valid size updates
	err = zone.SetZoneSize("1g")
	assert.NilError(t, err)

	err = zone.SetZoneSize("512k")
	assert.NilError(t, err)

	err = zone.SetZoneSize("100")
	assert.NilError(t, err) // bytes

	// Test invalid size formats
	err = zone.SetZoneSize("10x") // invalid unit
	assert.Error(t, err, "invalid size format: 10x (expected format: 10m, 1g, 512k, etc.)")

	err = zone.SetZoneSize("abc") // not a size
	assert.Error(t, err, "invalid size format: abc (expected format: 10m, 1g, 512k, etc.)")
}

func TestLimitConnZoneCompatibility(t *testing.T) {
	t.Parallel()

	configString := `
http {
	limit_conn_zone $binary_remote_addr zone=addr1:10m;
	limit_conn_zone $binary_remote_addr zone=addr2:20m;
	limit_conn_zone $server_name zone=perserver:5m;
}
`

	p := parser.NewStringParser(configString)
	conf, err := p.Parse()
	assert.NilError(t, err)

	zones := conf.FindLimitConnZones()
	assert.Equal(t, len(zones), 3)

	// Test compatibility between zones with same key
	zone1 := zones[0] // $binary_remote_addr
	zone2 := zones[1] // $binary_remote_addr
	zone3 := zones[2] // $server_name

	assert.Equal(t, zone1.IsCompatibleWith(zone2), true)  // same key
	assert.Equal(t, zone1.IsCompatibleWith(zone3), false) // different key
	assert.Equal(t, zone1.IsCompatibleWith(nil), false)   // nil check
}

func TestLimitConnZoneWithComments(t *testing.T) {
	t.Parallel()

	configString := `
http {
	# Connection limiting for IP addresses
	limit_conn_zone $binary_remote_addr zone=addr:10m; # 10MB for IP tracking
	
	# Connection limiting per server
	limit_conn_zone $server_name zone=perserver:5m;
}
`

	p := parser.NewStringParser(configString)
	conf, err := p.Parse()
	assert.NilError(t, err)

	limitConnZones := conf.FindLimitConnZones()
	assert.Equal(t, len(limitConnZones), 2)

	zone1 := limitConnZones[0]
	// Test that comments are preserved
	assert.Assert(t, len(zone1.GetComment()) > 0)
}

func TestLimitConnZoneSizeParsing(t *testing.T) {
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
	limit_conn_zone $binary_remote_addr zone=test:%s;
}
`, tc.configSize)

		p := parser.NewStringParser(configString)
		conf, err := p.Parse()
		assert.NilError(t, err)

		zones := conf.FindLimitConnZones()
		assert.Equal(t, len(zones), 1)

		zone := zones[0]
		sizeBytes, err := zone.GetZoneSizeBytes()
		assert.NilError(t, err)
		assert.Equal(t, sizeBytes, tc.expected, "Size parsing failed for %s", tc.configSize)
	}
}

func TestLimitConnZoneConnectionEstimation(t *testing.T) {
	t.Parallel()

	// Test connection estimation for different sizes
	testCases := []struct {
		size        string
		sizeBytes   int64
		connections int64
	}{
		{"1k", 1024, 16},                     // 1024 / 64
		{"1m", 1024 * 1024, 16384},           // 1MB / 64
		{"10m", 10 * 1024 * 1024, 163840},    // 10MB / 64
		{"1g", 1024 * 1024 * 1024, 16777216}, // 1GB / 64
	}

	for _, tc := range testCases {
		configString := fmt.Sprintf(`
http {
	limit_conn_zone $binary_remote_addr zone=test:%s;
}
`, tc.size)

		p := parser.NewStringParser(configString)
		conf, err := p.Parse()
		assert.NilError(t, err)

		zones := conf.FindLimitConnZones()
		zone := zones[0]

		sizeBytes, err := zone.GetZoneSizeBytes()
		assert.NilError(t, err)
		assert.Equal(t, sizeBytes, tc.sizeBytes)

		connections, err := zone.EstimateMaxConnections()
		assert.NilError(t, err)
		assert.Equal(t, connections, tc.connections, "Connection estimation failed for %s", tc.size)
	}
}

func TestLimitConnZoneRecommendations(t *testing.T) {
	t.Parallel()

	configString := `
http {
	limit_conn_zone $binary_remote_addr zone=test:10m;
}
`

	p := parser.NewStringParser(configString)
	conf, err := p.Parse()
	assert.NilError(t, err)

	zones := conf.FindLimitConnZones()
	zone := zones[0]

	recommendations, err := zone.GetRecommendedLimits()
	assert.NilError(t, err)

	// Check that all recommendation levels exist
	_, hasConservative := recommendations["conservative"]
	_, hasModerate := recommendations["moderate"]
	_, hasAggressive := recommendations["aggressive"]

	assert.Assert(t, hasConservative, "Conservative recommendation missing")
	assert.Assert(t, hasModerate, "Moderate recommendation missing")
	assert.Assert(t, hasAggressive, "Aggressive recommendation missing")

	// Check that conservative < moderate < aggressive
	assert.Assert(t, recommendations["conservative"] < recommendations["moderate"])
	assert.Assert(t, recommendations["moderate"] < recommendations["aggressive"])

	// Check reasonable values for 10MB zone (163840 max connections)
	maxConnections, _ := zone.EstimateMaxConnections()
	assert.Assert(t, recommendations["conservative"] == int(maxConnections*70/100))
	assert.Assert(t, recommendations["moderate"] == int(maxConnections*85/100))
	assert.Assert(t, recommendations["aggressive"] == int(maxConnections*95/100))
}

func TestInvalidLimitConnZoneDirective(t *testing.T) {
	t.Parallel()

	// Test limit_conn_zone directive with insufficient parameters
	configString := `
http {
	limit_conn_zone $binary_remote_addr;
}
`

	p := parser.NewStringParser(configString)
	_, err := p.Parse()
	// Should return an error due to missing zone parameter
	assert.Error(t, err, "limit_conn_zone directive requires at least 2 parameters: key and zone")
}

func TestInvalidConnZoneSpecification(t *testing.T) {
	t.Parallel()

	// Test invalid zone specification
	configString := `
http {
	limit_conn_zone $binary_remote_addr zone=invalid;
}
`

	p := parser.NewStringParser(configString)
	_, err := p.Parse()
	// Should return an error due to invalid zone format
	assert.Error(t, err, "invalid zone specification: zone=invalid (expected format: zone=name:size)")
}

func TestUnknownConnParameter(t *testing.T) {
	t.Parallel()

	// Test unknown parameter
	configString := `
http {
	limit_conn_zone $binary_remote_addr zone=addr:10m unknown=value;
}
`

	p := parser.NewStringParser(configString)
	_, err := p.Parse()
	// Should return an error due to unknown parameter
	assert.Error(t, err, "unknown parameter: unknown=value")
}

func TestLimitConnZoneSync(t *testing.T) {
	t.Parallel()

	configString := `
http {
	limit_conn_zone $binary_remote_addr zone=addr1:10m sync;
	limit_conn_zone $server_name zone=addr2:10m;
}
`

	p := parser.NewStringParser(configString)
	conf, err := p.Parse()
	assert.NilError(t, err)

	zones := conf.FindLimitConnZones()
	assert.Equal(t, len(zones), 2)

	// First zone with sync
	zone1 := zones[0]
	assert.Equal(t, zone1.Sync, true)

	// Second zone without sync
	zone2 := zones[1]
	assert.Equal(t, zone2.Sync, false)
}

func TestLimitConnZonesByKey(t *testing.T) {
	t.Parallel()

	configString := `
http {
	limit_conn_zone $binary_remote_addr zone=addr1:10m;
	limit_conn_zone $binary_remote_addr zone=addr2:20m;
	limit_conn_zone $server_name zone=perserver:5m;
	limit_conn_zone $request_uri zone=peruri:15m;
}
`

	p := parser.NewStringParser(configString)
	conf, err := p.Parse()
	assert.NilError(t, err)

	// Test finding zones by key
	ipZones := conf.FindLimitConnZonesByKey("$binary_remote_addr")
	assert.Equal(t, len(ipZones), 2)
	assert.Equal(t, ipZones[0].ZoneName, "addr1")
	assert.Equal(t, ipZones[1].ZoneName, "addr2")

	serverZones := conf.FindLimitConnZonesByKey("$server_name")
	assert.Equal(t, len(serverZones), 1)
	assert.Equal(t, serverZones[0].ZoneName, "perserver")

	// Test finding zones with non-existent key
	nonExistentZones := conf.FindLimitConnZonesByKey("$nonexistent")
	assert.Equal(t, len(nonExistentZones), 0)
}

func TestLimitConnZoneMemoryUsage(t *testing.T) {
	t.Parallel()

	configString := `
http {
	limit_conn_zone $binary_remote_addr zone=test:10m;
}
`

	p := parser.NewStringParser(configString)
	conf, err := p.Parse()
	assert.NilError(t, err)

	zones := conf.FindLimitConnZones()
	zone := zones[0]

	// Test memory usage estimate
	memoryEstimate := zone.GetMemoryUsageEstimate()
	assert.Equal(t, memoryEstimate, "~64 bytes per connection")
}
