package config_test

import (
	"fmt"
	"testing"

	"github.com/lefeck/gonginx/parser"
	"gotest.tools/v3/assert"
)

func TestLimitReqZoneParsing(t *testing.T) {
	t.Parallel()

	configString := `
http {
	limit_req_zone $binary_remote_addr zone=one:10m rate=1r/s;
	limit_req_zone $server_name zone=perserver:10m rate=10r/s;
	limit_req_zone $request_uri zone=peruri:5m rate=2r/m;
	limit_req_zone $http_user_agent zone=peragent:1g rate=0.5r/s sync;
	
	server {
		listen 80;
		location / {
			limit_req zone=one burst=5 nodelay;
			proxy_pass http://backend;
		}
	}
}
`

	p := parser.NewStringParser(configString)
	conf, err := p.Parse()
	assert.NilError(t, err)

	// Test finding limit_req_zone
	limitReqZones := conf.FindLimitReqZones()
	assert.Equal(t, len(limitReqZones), 4)

	// Test first limit_req_zone
	zone1 := limitReqZones[0]
	assert.Equal(t, zone1.Key, "$binary_remote_addr")
	assert.Equal(t, zone1.ZoneName, "one")
	assert.Equal(t, zone1.ZoneSize, "10m")
	assert.Equal(t, zone1.Rate, "1r/s")
	assert.Equal(t, zone1.Sync, false)

	// Test rate parsing
	rateNum, err := zone1.GetRateNumber()
	assert.NilError(t, err)
	assert.Equal(t, rateNum, 1.0)

	rateUnit, err := zone1.GetRateUnit()
	assert.NilError(t, err)
	assert.Equal(t, rateUnit, "s")

	// Test zone size parsing
	sizeBytes, err := zone1.GetZoneSizeBytes()
	assert.NilError(t, err)
	assert.Equal(t, sizeBytes, int64(10*1024*1024)) // 10MB

	// Test second limit_req_zone
	zone2 := limitReqZones[1]
	assert.Equal(t, zone2.Key, "$server_name")
	assert.Equal(t, zone2.ZoneName, "perserver")
	assert.Equal(t, zone2.Rate, "10r/s")

	// Test third limit_req_zone (rate per minute)
	zone3 := limitReqZones[2]
	assert.Equal(t, zone3.Key, "$request_uri")
	assert.Equal(t, zone3.ZoneName, "peruri")
	assert.Equal(t, zone3.Rate, "2r/m")

	rateUnit3, err := zone3.GetRateUnit()
	assert.NilError(t, err)
	assert.Equal(t, rateUnit3, "m")

	// Test fourth limit_req_zone (with sync)
	zone4 := limitReqZones[3]
	assert.Equal(t, zone4.Key, "$http_user_agent")
	assert.Equal(t, zone4.ZoneName, "peragent")
	assert.Equal(t, zone4.ZoneSize, "1g")
	assert.Equal(t, zone4.Rate, "0.5r/s")
	assert.Equal(t, zone4.Sync, true)

	// Test gigabyte size parsing
	sizeBytes4, err := zone4.GetZoneSizeBytes()
	assert.NilError(t, err)
	assert.Equal(t, sizeBytes4, int64(1024*1024*1024)) // 1GB

	// Test decimal rate
	rateNum4, err := zone4.GetRateNumber()
	assert.NilError(t, err)
	assert.Equal(t, rateNum4, 0.5)
}

func TestFindLimitReqZoneByName(t *testing.T) {
	t.Parallel()

	configString := `
http {
	limit_req_zone $binary_remote_addr zone=one:10m rate=1r/s;
	limit_req_zone $server_name zone=perserver:10m rate=10r/s;
}
`

	p := parser.NewStringParser(configString)
	conf, err := p.Parse()
	assert.NilError(t, err)

	// Test finding existing limit_req_zone by name
	zone1 := conf.FindLimitReqZoneByName("one")
	assert.Assert(t, zone1 != nil)
	assert.Equal(t, zone1.ZoneName, "one")
	assert.Equal(t, zone1.Key, "$binary_remote_addr")

	// Test finding another zone
	zoneServer := conf.FindLimitReqZoneByName("perserver")
	assert.Assert(t, zoneServer != nil)
	assert.Equal(t, zoneServer.ZoneName, "perserver")
	assert.Equal(t, zoneServer.Key, "$server_name")

	// Test finding non-existent zone
	nonExistent := conf.FindLimitReqZoneByName("nonexistent")
	assert.Assert(t, nonExistent == nil)
}

func TestLimitReqZoneManipulation(t *testing.T) {
	t.Parallel()

	configString := `
http {
	limit_req_zone $binary_remote_addr zone=one:10m rate=1r/s;
}
`

	p := parser.NewStringParser(configString)
	conf, err := p.Parse()
	assert.NilError(t, err)

	limitReqZones := conf.FindLimitReqZones()
	assert.Equal(t, len(limitReqZones), 1)

	zone := limitReqZones[0]

	// Test updating rate
	err = zone.SetRate("5r/s")
	assert.NilError(t, err)
	assert.Equal(t, zone.Rate, "5r/s")

	// Test updating zone size
	err = zone.SetZoneSize("20m")
	assert.NilError(t, err)
	assert.Equal(t, zone.ZoneSize, "20m")

	// Verify new size in bytes
	newSizeBytes, err := zone.GetZoneSizeBytes()
	assert.NilError(t, err)
	assert.Equal(t, newSizeBytes, int64(20*1024*1024)) // 20MB
}

func TestLimitReqZoneValidation(t *testing.T) {
	t.Parallel()

	configString := `
http {
	limit_req_zone $binary_remote_addr zone=one:10m rate=1r/s;
}
`

	p := parser.NewStringParser(configString)
	conf, err := p.Parse()
	assert.NilError(t, err)

	limitReqZones := conf.FindLimitReqZones()
	zone := limitReqZones[0]

	// Test valid rate updates
	err = zone.SetRate("10r/s")
	assert.NilError(t, err)

	err = zone.SetRate("5.5r/m")
	assert.NilError(t, err)

	// Test invalid rate formats
	err = zone.SetRate("10r/h") // invalid unit
	assert.Error(t, err, "invalid rate format: 10r/h (expected format: 10r/s or 5r/m)")

	err = zone.SetRate("abc") // not a rate
	assert.Error(t, err, "invalid rate format: abc (expected format: 10r/s or 5r/m)")

	err = zone.SetRate("10") // missing rate format
	assert.Error(t, err, "invalid rate format: 10 (expected format: 10r/s or 5r/m)")

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

func TestLimitReqZoneWithComments(t *testing.T) {
	t.Parallel()

	configString := `
http {
	# Rate limiting for IP addresses
	limit_req_zone $binary_remote_addr zone=one:10m rate=1r/s; # 1 request per second per IP
	
	# Rate limiting per server
	limit_req_zone $server_name zone=perserver:10m rate=10r/s;
}
`

	p := parser.NewStringParser(configString)
	conf, err := p.Parse()
	assert.NilError(t, err)

	limitReqZones := conf.FindLimitReqZones()
	assert.Equal(t, len(limitReqZones), 2)

	zone1 := limitReqZones[0]
	// Test that comments are preserved
	assert.Assert(t, len(zone1.GetComment()) > 0)
}

func TestLimitReqZoneSizeParsing(t *testing.T) {
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
	limit_req_zone $binary_remote_addr zone=test:%s rate=1r/s;
}
`, tc.configSize)

		p := parser.NewStringParser(configString)
		conf, err := p.Parse()
		assert.NilError(t, err)

		zones := conf.FindLimitReqZones()
		assert.Equal(t, len(zones), 1)

		zone := zones[0]
		sizeBytes, err := zone.GetZoneSizeBytes()
		assert.NilError(t, err)
		assert.Equal(t, sizeBytes, tc.expected, "Size parsing failed for %s", tc.configSize)
	}
}

func TestLimitReqZoneRateParsing(t *testing.T) {
	t.Parallel()

	// Test different rate formats
	testCases := []struct {
		rate         string
		expectedNum  float64
		expectedUnit string
	}{
		{"1r/s", 1.0, "s"},
		{"10r/s", 10.0, "s"},
		{"0.5r/s", 0.5, "s"},
		{"5r/m", 5.0, "m"},
		{"2.5r/m", 2.5, "m"},
	}

	for _, tc := range testCases {
		configString := fmt.Sprintf(`
http {
	limit_req_zone $binary_remote_addr zone=test:10m rate=%s;
}
`, tc.rate)

		p := parser.NewStringParser(configString)
		conf, err := p.Parse()
		assert.NilError(t, err)

		zones := conf.FindLimitReqZones()
		assert.Equal(t, len(zones), 1)

		zone := zones[0]

		rateNum, err := zone.GetRateNumber()
		assert.NilError(t, err)
		assert.Equal(t, rateNum, tc.expectedNum, "Rate number parsing failed for %s", tc.rate)

		rateUnit, err := zone.GetRateUnit()
		assert.NilError(t, err)
		assert.Equal(t, rateUnit, tc.expectedUnit, "Rate unit parsing failed for %s", tc.rate)
	}
}

func TestInvalidLimitReqZoneDirective(t *testing.T) {
	t.Parallel()

	// Test limit_req_zone directive with insufficient parameters
	configString := `
http {
	limit_req_zone $binary_remote_addr zone=one:10m;
}
`

	p := parser.NewStringParser(configString)
	_, err := p.Parse()
	// Should return an error due to missing rate parameter
	assert.Error(t, err, "limit_req_zone directive requires at least 3 parameters: key, zone, and rate")
}

func TestInvalidZoneSpecification(t *testing.T) {
	t.Parallel()

	// Test invalid zone specification
	configString := `
http {
	limit_req_zone $binary_remote_addr zone=invalid rate=1r/s;
}
`

	p := parser.NewStringParser(configString)
	_, err := p.Parse()
	// Should return an error due to invalid zone format
	assert.Error(t, err, "invalid zone specification: zone=invalid (expected format: zone=name:size)")
}

func TestInvalidRateSpecification(t *testing.T) {
	t.Parallel()

	// Test invalid rate specification
	configString := `
http {
	limit_req_zone $binary_remote_addr zone=one:10m rate=invalid;
}
`

	p := parser.NewStringParser(configString)
	_, err := p.Parse()
	// Should return an error due to invalid rate format
	assert.Error(t, err, "invalid rate format: invalid (expected format: 10r/s or 5r/m)")
}

func TestUnknownParameter(t *testing.T) {
	t.Parallel()

	// Test unknown parameter
	configString := `
http {
	limit_req_zone $binary_remote_addr zone=one:10m rate=1r/s unknown=value;
}
`

	p := parser.NewStringParser(configString)
	_, err := p.Parse()
	// Should return an error due to unknown parameter
	assert.Error(t, err, "unknown parameter: unknown=value")
}

func TestLimitReqZoneSync(t *testing.T) {
	t.Parallel()

	configString := `
http {
	limit_req_zone $binary_remote_addr zone=one:10m rate=1r/s sync;
	limit_req_zone $server_name zone=two:10m rate=5r/s;
}
`

	p := parser.NewStringParser(configString)
	conf, err := p.Parse()
	assert.NilError(t, err)

	zones := conf.FindLimitReqZones()
	assert.Equal(t, len(zones), 2)

	// First zone with sync
	zone1 := zones[0]
	assert.Equal(t, zone1.Sync, true)

	// Second zone without sync
	zone2 := zones[1]
	assert.Equal(t, zone2.Sync, false)
}
