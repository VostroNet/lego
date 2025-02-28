package netcup

import (
	"fmt"
	"testing"

	"github.com/vostronet/lego/challenge/dns01"
	"github.com/vostronet/lego/platform/tester"
	"github.com/stretchr/testify/require"
)

var envTest = tester.NewEnvTest(
	"NETCUP_CUSTOMER_NUMBER",
	"NETCUP_API_KEY",
	"NETCUP_API_PASSWORD").
	WithDomain("NETCUP_DOMAIN")

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				"NETCUP_CUSTOMER_NUMBER": "A",
				"NETCUP_API_KEY":         "B",
				"NETCUP_API_PASSWORD":    "C",
			},
		},
		{
			desc: "missing credentials",
			envVars: map[string]string{
				"NETCUP_CUSTOMER_NUMBER": "",
				"NETCUP_API_KEY":         "",
				"NETCUP_API_PASSWORD":    "",
			},
			expected: "netcup: some credentials information are missing: NETCUP_CUSTOMER_NUMBER,NETCUP_API_KEY,NETCUP_API_PASSWORD",
		},
		{
			desc: "missing customer number",
			envVars: map[string]string{
				"NETCUP_CUSTOMER_NUMBER": "",
				"NETCUP_API_KEY":         "B",
				"NETCUP_API_PASSWORD":    "C",
			},
			expected: "netcup: some credentials information are missing: NETCUP_CUSTOMER_NUMBER",
		},
		{
			desc: "missing API key",
			envVars: map[string]string{
				"NETCUP_CUSTOMER_NUMBER": "A",
				"NETCUP_API_KEY":         "",
				"NETCUP_API_PASSWORD":    "C",
			},
			expected: "netcup: some credentials information are missing: NETCUP_API_KEY",
		},
		{
			desc: "missing api password",
			envVars: map[string]string{
				"NETCUP_CUSTOMER_NUMBER": "A",
				"NETCUP_API_KEY":         "B",
				"NETCUP_API_PASSWORD":    "",
			},
			expected: "netcup: some credentials information are missing: NETCUP_API_PASSWORD",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer envTest.RestoreEnv()
			envTest.ClearEnv()

			envTest.Apply(test.envVars)

			p, err := NewDNSProvider()

			if len(test.expected) == 0 {
				require.NoError(t, err)
				require.NotNil(t, p)
				require.NotNil(t, p.config)
				require.NotNil(t, p.client)
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}

func TestNewDNSProviderConfig(t *testing.T) {
	testCases := []struct {
		desc     string
		customer string
		key      string
		password string
		expected string
	}{
		{
			desc:     "success",
			customer: "A",
			key:      "B",
			password: "C",
		},
		{
			desc:     "missing credentials",
			expected: "netcup: credentials missing",
		},
		{
			desc:     "missing customer",
			customer: "",
			key:      "B",
			password: "C",
			expected: "netcup: credentials missing",
		},
		{
			desc:     "missing key",
			customer: "A",
			key:      "",
			password: "C",
			expected: "netcup: credentials missing",
		},
		{
			desc:     "missing password",
			customer: "A",
			key:      "B",
			password: "",
			expected: "netcup: credentials missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.Customer = test.customer
			config.Key = test.key
			config.Password = test.password

			p, err := NewDNSProviderConfig(config)

			if len(test.expected) == 0 {
				require.NoError(t, err)
				require.NotNil(t, p)
				require.NotNil(t, p.config)
				require.NotNil(t, p.client)
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}

func TestLivePresentAndCleanup(t *testing.T) {
	if !envTest.IsLiveTest() {
		t.Skip("skipping live test")
	}

	envTest.RestoreEnv()
	p, err := NewDNSProvider()
	require.NoError(t, err)

	fqdn, _ := dns01.GetRecord(envTest.GetDomain(), "123d==")

	zone, err := dns01.FindZoneByFqdn(fqdn)
	require.NoError(t, err, "error finding DNSZone")

	zone = dns01.UnFqdn(zone)

	testCases := []string{
		zone,
		"sub." + zone,
		"*." + zone,
		"*.sub." + zone,
	}

	for _, test := range testCases {
		t.Run(fmt.Sprintf("domain(%s)", test), func(t *testing.T) {
			err = p.Present(test, "987d", "123d==")
			require.NoError(t, err)

			err = p.CleanUp(test, "987d", "123d==")
			require.NoError(t, err, "Did not clean up! Please remove record yourself.")
		})
	}
}
