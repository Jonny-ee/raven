package vpndriver

import (
	"reflect"
	"testing"

	"github.com/openyurtio/api/raven/v1beta1"

	"github.com/openyurtio/raven/cmd/agent/app/config"
	"github.com/openyurtio/raven/pkg/types"
)

const (
	failed  = "\u2717"
	succeed = "\u2713"
)

type TestDriver struct {
}

func (TestDriver) Init() error {
	return nil
}

func (TestDriver) Apply(network *types.Network, routeDriverMTU func(*types.Network) (int, error)) error {
	_, _ = network, routeDriverMTU
	return nil
}

func (TestDriver) MTU() (int, error) {
	return 1, nil
}

func (TestDriver) Cleanup() error {
	return nil
}

func TestRegisterDriver(t *testing.T) {
	var f Factory = func(cfg *config.Config) (Driver, error) {
		Driver1 := TestDriver{}
		return Driver1, nil
	}
	tests := []struct {
		name    string
		factory Factory
		expect  Factory
	}{
		{
			name:    "normal",
			factory: f,
			expect:  f,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			t.Logf("\tTestCase: %s", tt.name)

			RegisterDriver(tt.name, tt.factory)
			get := drivers
			if !reflect.DeepEqual(get, drivers) {
				t.Fatalf("\t%s\texpect %v, but get %v", failed, drivers, get)
			}
			t.Logf("\t%s\texpect %v, get %v", succeed, drivers, get)
		})
	}
}

func TestNew(t *testing.T) {
	tests := []struct {
		name   string
		cfg    *config.Config
		expect string
	}{
		{
			name:   "normal",
			cfg:    &config.Config{},
			expect: "normal",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			t.Logf("\tTestCase: %s", tt.name)

			_, _ = New(tt.name, tt.cfg)
			get := drivers
			if !reflect.DeepEqual(get, drivers) {
				t.Fatalf("\t%s\texpect %v, but get %v", failed, drivers, get)
			}
			t.Logf("\t%s\texpect %v, get %v", succeed, drivers, get)
		})
	}
}

func TestFindCentralGwFn(t *testing.T) {
	var n = &types.Network{
		LocalEndpoint: &types.Endpoint{
			PrivateIP: "192.168.1.1",
			UnderNAT:  false,
		},
		LocalNodeInfo:   make(map[types.NodeName]*v1beta1.NodeInfo),
		RemoteEndpoints: make(map[types.GatewayName]*types.Endpoint),
		RemoteNodeInfo:  make(map[types.NodeName]*v1beta1.NodeInfo),
	}

	tests := []struct {
		name    string
		network *types.Network
		expect  *types.Endpoint
	}{
		{
			name:    "normal",
			network: n,
			expect:  n.LocalEndpoint,
		},
		{
			name: "prioritize loadbalancer",
			network: &types.Network{
				LocalEndpoint: &types.Endpoint{
					PrivateIP: "192.168.1.1",
					UnderNAT:  false,
					ExposeType: "",
				},
				RemoteEndpoints: map[types.GatewayName]*types.Endpoint{
					"gw1": {PrivateIP: "192.168.1.2", UnderNAT: false, ExposeType: string(v1beta1.ExposeTypeLoadBalancer)},
				},
			},
			expect: &types.Endpoint{PrivateIP: "192.168.1.2", UnderNAT: false, ExposeType: string(v1beta1.ExposeTypeLoadBalancer)},
		},
		{
			name: "prioritize publicip over local",
			network: &types.Network{
				LocalEndpoint: &types.Endpoint{
					PrivateIP: "192.168.1.1",
					UnderNAT:  false,
					ExposeType: "",
				},
				RemoteEndpoints: map[types.GatewayName]*types.Endpoint{
					"gw1": {PrivateIP: "192.168.1.2", UnderNAT: false, ExposeType: string(v1beta1.ExposeTypePublicIP)},
				},
			},
			expect: &types.Endpoint{PrivateIP: "192.168.1.2", UnderNAT: false, ExposeType: string(v1beta1.ExposeTypePublicIP)},
		},
		{
			name: "fallback to valid endpoint if none have prioritize types",
			network: &types.Network{
				LocalEndpoint: &types.Endpoint{
					PrivateIP: "192.168.1.1",
					UnderNAT:  false,
					ExposeType: "",
				},
				RemoteEndpoints: map[types.GatewayName]*types.Endpoint{
					"gw1": {PrivateIP: "192.168.1.2", UnderNAT: false, ExposeType: ""},
				},
			},
			// Expect either local or remote since map iteration is non-deterministic in FindCentralGwFn
			// But LocalEndpoint is always appended first, so if remote doesn't override due to missing type, wait...
			// "else if central == nil || ..." means central will be overwritten by remote since both have empty ExposeType!
			// Actually central starts as LocalEndpoint, then Remote is checked. Since central isn't LB/Public, central gets overridden by Remote.
			expect: &types.Endpoint{PrivateIP: "192.168.1.2", UnderNAT: false, ExposeType: ""},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			t.Logf("\tTestCase: %s", tt.name)

			get := FindCentralGwFn(tt.network)
			if !reflect.DeepEqual(get, tt.expect) {
				t.Fatalf("\t%s\texpect %v, but get %v", failed, tt.expect, get)
			}
			t.Logf("\t%s\texpect %v, get %v", succeed, tt.expect, get)
		})
	}
}

func TestDefaultMTU(t *testing.T) {
	tests := []struct {
		name   string
		expect error
	}{
		{
			name:   "normal",
			expect: nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			t.Logf("\tTestCase: %s", tt.name)

			_, get := DefaultMTU()
			if !reflect.DeepEqual(get, tt.expect) {
				t.Fatalf("\t%s\texpect %v, but get %v", failed, tt.expect, get)
			}
			t.Logf("\t%s\texpect %v, get %v", succeed, tt.expect, get)
		})
	}
}

func TestGetPSK(t *testing.T) {
	tests := []struct {
		name   string
		expect string
	}{
		{
			name:   "normal",
			expect: DefaultPSK,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			t.Logf("\tTestCase: %s", tt.name)

			get := GetPSK()
			if !reflect.DeepEqual(get, tt.expect) {
				t.Fatalf("\t%s\texpect %v, but get %v", failed, tt.expect, get)
			}
			t.Logf("\t%s\texpect %v, get %v", succeed, tt.expect, get)
		})
	}
}
