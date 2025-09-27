package config_test

import (
	"testing"

	"github.com/lefeck/gonginx/parser"
	"gotest.tools/v3/assert"
)

func TestSplitClientsParsing(t *testing.T) {
	t.Parallel()

	configString := `
http {
	split_clients $remote_addr $variant {
		0.5%     .one;
		2.0%     .two;
		10%      .three;
		*        "";
	}
	
	split_clients $request_id $ab_test {
		20%      version_a;
		30%      version_b;
		*        version_c;
	}
	
	split_clients $http_user_agent $mobile {
		15.5%    mobile_app;
		*        desktop;
	}
	
	server {
		listen 80;
		location / {
			proxy_pass http://backend$variant;
		}
	}
}
`

	p := parser.NewStringParser(configString)
	conf, err := p.Parse()
	assert.NilError(t, err)

	// Test finding split_clients
	splitClients := conf.FindSplitClients()
	assert.Equal(t, len(splitClients), 3)

	// Test first split_clients
	variantSplit := splitClients[0]
	assert.Equal(t, variantSplit.Variable, "$remote_addr")
	assert.Equal(t, variantSplit.MappedVariable, "$variant")
	assert.Equal(t, len(variantSplit.Entries), 4)
	assert.Equal(t, variantSplit.HasWildcard(), true)
	assert.Equal(t, variantSplit.GetWildcardValue(), `""`)

	// Test percentage validation
	total, err := variantSplit.GetTotalPercentage()
	assert.NilError(t, err)
	assert.Equal(t, total, 12.5) // 0.5 + 2.0 + 10

	// Test second split_clients
	abTestSplit := splitClients[1]
	assert.Equal(t, abTestSplit.Variable, "$request_id")
	assert.Equal(t, abTestSplit.MappedVariable, "$ab_test")
	assert.Equal(t, len(abTestSplit.Entries), 3)

	// Test entries
	expectedEntries := map[string]string{
		"20%": "version_a",
		"30%": "version_b",
		"*":   "version_c",
	}

	for _, entry := range abTestSplit.Entries {
		expectedValue, exists := expectedEntries[entry.Percentage]
		assert.Assert(t, exists, "Unexpected percentage: %s", entry.Percentage)
		assert.Equal(t, entry.Value, expectedValue)
	}

	// Test third split_clients
	mobileSplit := splitClients[2]
	assert.Equal(t, mobileSplit.Variable, "$http_user_agent")
	assert.Equal(t, mobileSplit.MappedVariable, "$mobile")
	assert.Equal(t, len(mobileSplit.Entries), 2)

	total, err = mobileSplit.GetTotalPercentage()
	assert.NilError(t, err)
	assert.Equal(t, total, 15.5)
}

func TestFindSplitClientsByVariable(t *testing.T) {
	t.Parallel()

	configString := `
http {
	split_clients $remote_addr $variant {
		0.5%     .one;
		*        "";
	}
	
	split_clients $request_id $ab_test {
		50%      version_a;
		*        version_b;
	}
}
`

	p := parser.NewStringParser(configString)
	conf, err := p.Parse()
	assert.NilError(t, err)

	// Test finding existing split_clients by variable
	variantSplit := conf.FindSplitClientsByVariable("$variant")
	assert.Assert(t, variantSplit != nil)
	assert.Equal(t, variantSplit.Variable, "$remote_addr")
	assert.Equal(t, variantSplit.MappedVariable, "$variant")

	// Test finding split_clients by both variables
	abTestSplit := conf.FindSplitClientsByVariables("$request_id", "$ab_test")
	assert.Assert(t, abTestSplit != nil)
	assert.Equal(t, abTestSplit.Variable, "$request_id")
	assert.Equal(t, abTestSplit.MappedVariable, "$ab_test")

	// Test finding non-existent split_clients
	nonExistent := conf.FindSplitClientsByVariable("$nonexistent")
	assert.Assert(t, nonExistent == nil)
}

func TestSplitClientsManipulation(t *testing.T) {
	t.Parallel()

	configString := `
http {
	split_clients $remote_addr $variant {
		0.5%     .one;
		*        "";
	}
}
`

	p := parser.NewStringParser(configString)
	conf, err := p.Parse()
	assert.NilError(t, err)

	splitClients := conf.FindSplitClients()
	assert.Equal(t, len(splitClients), 1)

	variantSplit := splitClients[0]

	// Test adding new entry
	originalCount := len(variantSplit.Entries)
	err = variantSplit.AddEntry("5.0%", ".two")
	assert.NilError(t, err)
	assert.Equal(t, len(variantSplit.Entries), originalCount+1)

	// Verify the new entry was added
	found := false
	for _, entry := range variantSplit.Entries {
		if entry.Percentage == "5.0%" && entry.Value == ".two" {
			found = true
			break
		}
	}
	assert.Assert(t, found, "New entry was not added correctly")

	// Test total percentage calculation
	total, err := variantSplit.GetTotalPercentage()
	assert.NilError(t, err)
	assert.Equal(t, total, 5.5) // 0.5 + 5.0

	// Test removing entry
	removed := variantSplit.RemoveEntry("0.5%")
	assert.Assert(t, removed, "Entry should have been removed")
	assert.Equal(t, len(variantSplit.Entries), originalCount) // back to original count

	// Test getting entries by value
	entries := variantSplit.GetEntriesByValue(`""`)
	assert.Equal(t, len(entries), 1) // wildcard entry
	assert.Equal(t, entries[0].Percentage, "*")
}

func TestSplitClientsValidation(t *testing.T) {
	t.Parallel()

	configString := `
http {
	split_clients $remote_addr $variant {
		0.5%     .one;
	}
}
`

	p := parser.NewStringParser(configString)
	conf, err := p.Parse()
	assert.NilError(t, err)

	splitClients := conf.FindSplitClients()
	assert.Equal(t, len(splitClients), 1)

	variantSplit := splitClients[0]

	// Test valid percentage entry
	err = variantSplit.AddEntry("10.5%", ".two")
	assert.NilError(t, err)

	// Test valid wildcard entry
	err = variantSplit.AddEntry("*", ".default")
	assert.NilError(t, err)

	// Test invalid percentage format (no %)
	err = variantSplit.AddEntry("10", ".invalid")
	assert.Error(t, err, "percentage must end with %")

	// Test invalid percentage format (not a number)
	err = variantSplit.AddEntry("abc%", ".invalid")
	assert.Error(t, err, "invalid percentage format: abc%")

	// Test invalid percentage range
	err = variantSplit.AddEntry("150%", ".invalid")
	assert.Error(t, err, "percentage must be between 0% and 100%")

	err = variantSplit.AddEntry("-5%", ".invalid")
	assert.Error(t, err, "percentage must be between 0% and 100%")
}

func TestSplitClientsWithComments(t *testing.T) {
	t.Parallel()

	configString := `
http {
	# A/B testing configuration
	split_clients $remote_addr $variant { # inline comment
		0.5%     .one; # variant one
		2.0%     .two; # variant two
		# Default variant
		*        "";
	}
}
`

	p := parser.NewStringParser(configString)
	conf, err := p.Parse()
	assert.NilError(t, err)

	splitClients := conf.FindSplitClients()
	assert.Equal(t, len(splitClients), 1)

	variantSplit := splitClients[0]

	// Test that comments are preserved
	assert.Assert(t, len(variantSplit.GetComment()) > 0)

	// Test that entry comments are preserved
	for _, entry := range variantSplit.Entries {
		if entry.Percentage == "*" {
			assert.Assert(t, len(entry.GetComment()) > 0)
			break
		}
	}
}

func TestSplitClientsPercentageCalculations(t *testing.T) {
	t.Parallel()

	configString := `
http {
	split_clients $remote_addr $variant {
		10%      .one;
		20%      .two;
		30%      .three;
		*        "";
	}
}
`

	p := parser.NewStringParser(configString)
	conf, err := p.Parse()
	assert.NilError(t, err)

	splitClients := conf.FindSplitClients()
	assert.Equal(t, len(splitClients), 1)

	variantSplit := splitClients[0]

	// Test total percentage calculation
	total, err := variantSplit.GetTotalPercentage()
	assert.NilError(t, err)
	assert.Equal(t, total, 60.0) // 10 + 20 + 30

	// Test wildcard detection
	assert.Equal(t, variantSplit.HasWildcard(), true)
	assert.Equal(t, variantSplit.GetWildcardValue(), `""`)

	// Test entries by value
	oneEntries := variantSplit.GetEntriesByValue(".one")
	assert.Equal(t, len(oneEntries), 1)
	assert.Equal(t, oneEntries[0].Percentage, "10%")
}

func TestSplitClientsExceedingPercentage(t *testing.T) {
	t.Parallel()

	// This should be allowed in nginx (total > 100% is valid)
	configString := `
http {
	split_clients $remote_addr $variant {
		50%      .one;
		60%      .two;
		*        "";
	}
}
`

	p := parser.NewStringParser(configString)
	_, err := p.Parse()
	// Should return an error due to total > 100%
	assert.Error(t, err, "total percentage cannot exceed 100%")
}

func TestEmptySplitClientsBlock(t *testing.T) {
	t.Parallel()

	configString := `
http {
	split_clients $remote_addr $variant {
	}
}
`

	p := parser.NewStringParser(configString)
	conf, err := p.Parse()
	assert.NilError(t, err)

	splitClients := conf.FindSplitClients()
	assert.Equal(t, len(splitClients), 1)

	variantSplit := splitClients[0]
	assert.Equal(t, len(variantSplit.Entries), 0)
	assert.Equal(t, variantSplit.HasWildcard(), false)

	total, err := variantSplit.GetTotalPercentage()
	assert.NilError(t, err)
	assert.Equal(t, total, 0.0)
}

func TestInvalidSplitClientsDirective(t *testing.T) {
	t.Parallel()

	// Test split_clients directive with insufficient parameters
	configString := `
http {
	split_clients $remote_addr {
		0.5%     .one;
	}
}
`

	p := parser.NewStringParser(configString)
	_, err := p.Parse()
	// Should return an error due to insufficient parameters
	assert.Error(t, err, "split_clients directive requires exactly 2 parameters: source and target variables")
}

func TestSplitClientsWithTooManyParameters(t *testing.T) {
	t.Parallel()

	// Test split_clients directive with too many parameters
	configString := `
http {
	split_clients $remote_addr $variant $extra {
		0.5%     .one;
	}
}
`

	p := parser.NewStringParser(configString)
	_, err := p.Parse()
	// Should return an error due to too many parameters
	assert.Error(t, err, "split_clients directive accepts exactly 2 parameters only")
}

func TestSplitClientsEdgeCasePercentages(t *testing.T) {
	t.Parallel()

	configString := `
http {
	split_clients $remote_addr $variant {
		0%       .zero;
		0.001%   .small;
		99.999%  .large;
		*        "";
	}
}
`

	p := parser.NewStringParser(configString)
	conf, err := p.Parse()
	assert.NilError(t, err)

	splitClients := conf.FindSplitClients()
	assert.Equal(t, len(splitClients), 1)

	variantSplit := splitClients[0]

	// Test edge case percentages
	total, err := variantSplit.GetTotalPercentage()
	assert.NilError(t, err)
	assert.Equal(t, total, 100.0) // 0 + 0.001 + 99.999

	// Test that all entries are valid
	assert.Equal(t, len(variantSplit.Entries), 4)
	assert.Equal(t, variantSplit.HasWildcard(), true)
}
