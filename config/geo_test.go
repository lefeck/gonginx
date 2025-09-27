package config_test

import (
	"testing"

	"github.com/lefeck/gonginx/parser"
	"gotest.tools/v3/assert"
)

func TestGeoParsing(t *testing.T) {
	t.Parallel()

	configString := `
http {
	geo $country {
		default ZZ;
		127.0.0.0/8 US;
		10.0.0.0/8 US;
		192.168.1.0/24 CN;
		203.208.60.0/24 CN;
	}
	
	geo $remote_addr $region {
		ranges;
		default unknown;
		127.0.0.1-127.0.0.255 local;
		10.0.0.1-10.0.0.100 internal;
	}
	
	geo $city {
		proxy 192.168.1.0/24;
		proxy_recursive;
		delete 127.0.0.1;
		default unknown;
		192.168.0.0/16 Beijing;
		10.0.0.0/8 Shanghai;
	}
	
	server {
		listen 80;
		return 200 "Country: $country, Region: $region, City: $city";
	}
}
`

	p := parser.NewStringParser(configString)
	conf, err := p.Parse()
	assert.NilError(t, err)

	// Test finding geos
	geos := conf.FindGeos()
	assert.Equal(t, len(geos), 3)

	// Test first geo (simple CIDR mapping)
	countryGeo := geos[0]
	assert.Equal(t, countryGeo.Variable, "$country")
	assert.Equal(t, countryGeo.SourceAddress, "$remote_addr") // default
	assert.Equal(t, countryGeo.GetDefaultValue(), "ZZ")
	assert.Equal(t, len(countryGeo.Entries), 4) // excluding default

	// Test second geo (with explicit source and ranges)
	regionGeo := geos[1]
	assert.Equal(t, regionGeo.Variable, "$region")
	assert.Equal(t, regionGeo.SourceAddress, "$remote_addr")
	assert.Equal(t, regionGeo.Ranges, true)
	assert.Equal(t, regionGeo.GetDefaultValue(), "unknown")

	// Test third geo (with proxy settings)
	cityGeo := geos[2]
	assert.Equal(t, cityGeo.Variable, "$city")
	assert.Equal(t, cityGeo.ProxyRecursive, true)
	assert.Equal(t, len(cityGeo.Proxy), 1)
	assert.Equal(t, cityGeo.Proxy[0], "192.168.1.0/24")
	assert.Equal(t, len(cityGeo.Delete), 1)
	assert.Equal(t, cityGeo.Delete[0], "127.0.0.1")
}

func TestFindGeoByVariable(t *testing.T) {
	t.Parallel()

	configString := `
http {
	geo $country {
		default ZZ;
		127.0.0.0/8 US;
	}
	
	geo $remote_addr $region {
		default unknown;
		10.0.0.0/8 internal;
	}
}
`

	p := parser.NewStringParser(configString)
	conf, err := p.Parse()
	assert.NilError(t, err)

	// Test finding existing geo by variable
	countryGeo := conf.FindGeoByVariable("$country")
	assert.Assert(t, countryGeo != nil)
	assert.Equal(t, countryGeo.Variable, "$country")
	assert.Equal(t, countryGeo.SourceAddress, "$remote_addr")

	// Test finding geo with explicit source
	regionGeo := conf.FindGeoByVariables("$remote_addr", "$region")
	assert.Assert(t, regionGeo != nil)
	assert.Equal(t, regionGeo.Variable, "$region")
	assert.Equal(t, regionGeo.SourceAddress, "$remote_addr")

	// Test finding non-existent geo
	nonExistent := conf.FindGeoByVariable("$nonexistent")
	assert.Assert(t, nonExistent == nil)
}

func TestGeoManipulation(t *testing.T) {
	t.Parallel()

	configString := `
http {
	geo $country {
		default ZZ;
		127.0.0.0/8 US;
	}
}
`

	p := parser.NewStringParser(configString)
	conf, err := p.Parse()
	assert.NilError(t, err)

	geos := conf.FindGeos()
	assert.Equal(t, len(geos), 1)

	countryGeo := geos[0]

	// Test adding new entry
	originalCount := len(countryGeo.Entries)
	err = countryGeo.AddEntry("192.168.1.0/24", "CN")
	assert.NilError(t, err)
	assert.Equal(t, len(countryGeo.Entries), originalCount+1)

	// Verify the new entry was added
	found := false
	for _, entry := range countryGeo.Entries {
		if entry.Network == "192.168.1.0/24" && entry.Value == "CN" {
			found = true
			break
		}
	}
	assert.Assert(t, found, "New entry was not added correctly")

	// Test updating default value
	countryGeo.SetDefaultValue("XX")
	assert.Equal(t, countryGeo.GetDefaultValue(), "XX")

	// Test adding proxy
	countryGeo.AddProxy("10.0.0.0/8")
	assert.Equal(t, len(countryGeo.Proxy), 1)
	assert.Equal(t, countryGeo.Proxy[0], "10.0.0.0/8")

	// Test adding delete entry
	countryGeo.AddDelete("127.0.0.1")
	assert.Equal(t, len(countryGeo.Delete), 1)
	assert.Equal(t, countryGeo.Delete[0], "127.0.0.1")

	// Test setting ranges mode
	countryGeo.SetRanges(true)
	assert.Equal(t, countryGeo.Ranges, true)

	// Test setting proxy recursive
	countryGeo.SetProxyRecursive(true)
	assert.Equal(t, countryGeo.ProxyRecursive, true)
}

func TestGeoWithComments(t *testing.T) {
	t.Parallel()

	configString := `
http {
	# Geo configuration for country detection
	geo $country { # inline comment
		default ZZ; # default country
		127.0.0.0/8 US; # localhost
		# Chinese networks
		192.168.1.0/24 CN;
	}
}
`

	p := parser.NewStringParser(configString)
	conf, err := p.Parse()
	assert.NilError(t, err)

	geos := conf.FindGeos()
	assert.Equal(t, len(geos), 1)

	countryGeo := geos[0]

	// Test that comments are preserved
	assert.Assert(t, len(countryGeo.GetComment()) > 0)

	// Test that entry comments are preserved
	for _, entry := range countryGeo.Entries {
		if entry.Network == "192.168.1.0/24" {
			assert.Assert(t, len(entry.GetComment()) > 0)
			break
		}
	}
}

func TestGeoRangeFormat(t *testing.T) {
	t.Parallel()

	configString := `
http {
	geo $region {
		ranges;
		default unknown;
		127.0.0.1-127.0.0.255 local;
		10.0.0.1-10.255.255.254 internal;
		192.168.1.1-192.168.1.100 dmz;
	}
}
`

	p := parser.NewStringParser(configString)
	conf, err := p.Parse()
	assert.NilError(t, err)

	geos := conf.FindGeos()
	assert.Equal(t, len(geos), 1)

	regionGeo := geos[0]
	assert.Equal(t, regionGeo.Ranges, true)
	assert.Equal(t, len(regionGeo.Entries), 3)

	// Check range entries
	expectedRanges := map[string]string{
		"127.0.0.1-127.0.0.255":     "local",
		"10.0.0.1-10.255.255.254":   "internal",
		"192.168.1.1-192.168.1.100": "dmz",
	}

	for _, entry := range regionGeo.Entries {
		expectedValue, exists := expectedRanges[entry.Network]
		assert.Assert(t, exists, "Unexpected range: %s", entry.Network)
		assert.Equal(t, entry.Value, expectedValue)
	}
}

func TestGeoValidation(t *testing.T) {
	t.Parallel()

	configString := `
http {
	geo $country {
		default ZZ;
	}
}
`

	p := parser.NewStringParser(configString)
	conf, err := p.Parse()
	assert.NilError(t, err)

	geos := conf.FindGeos()
	assert.Equal(t, len(geos), 1)

	countryGeo := geos[0]

	// Test valid CIDR entry
	err = countryGeo.AddEntry("192.168.1.0/24", "CN")
	assert.NilError(t, err)

	// Test valid IP address entry
	err = countryGeo.AddEntry("127.0.0.1", "US")
	assert.NilError(t, err)

	// Test valid range entry
	err = countryGeo.AddEntry("10.0.0.1-10.0.0.100", "internal")
	assert.NilError(t, err)

	// Test invalid network format
	err = countryGeo.AddEntry("invalid-network", "XX")
	assert.Error(t, err, "invalid network format: invalid-network")

	// Test special entries (should not be validated)
	err = countryGeo.AddEntry("default", "ZZ")
	assert.NilError(t, err)

	err = countryGeo.AddEntry("ranges", "")
	assert.NilError(t, err)
}

func TestEmptyGeoBlock(t *testing.T) {
	t.Parallel()

	configString := `
http {
	geo $country {
	}
}
`

	p := parser.NewStringParser(configString)
	conf, err := p.Parse()
	assert.NilError(t, err)

	geos := conf.FindGeos()
	assert.Equal(t, len(geos), 1)

	countryGeo := geos[0]
	assert.Equal(t, len(countryGeo.Entries), 0)
	assert.Equal(t, countryGeo.GetDefaultValue(), "") // No default value
}

func TestInvalidGeoDirective(t *testing.T) {
	t.Parallel()

	// Test geo directive with no parameters
	configString := `
http {
	geo {
		default ZZ;
	}
}
`

	p := parser.NewStringParser(configString)
	_, err := p.Parse()
	// Should return an error due to missing parameters
	assert.Error(t, err, "geo directive requires at least 1 parameter: target variable")
}

func TestGeoWithTooManyParameters(t *testing.T) {
	t.Parallel()

	// Test geo directive with too many parameters
	configString := `
http {
	geo $addr $country $extra {
		default ZZ;
	}
}
`

	p := parser.NewStringParser(configString)
	_, err := p.Parse()
	// Should return an error due to too many parameters
	assert.Error(t, err, "geo directive accepts 1 or 2 parameters only")
}
