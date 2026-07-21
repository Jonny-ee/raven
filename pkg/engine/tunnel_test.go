package engine

import (
	"testing"

	"github.com/openyurtio/api/raven/v1beta1"
	"github.com/openyurtio/raven/pkg/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestTunnelEngine_syncGateway(t *testing.T) {
	te := &TunnelEngine{
		network: &types.Network{
			LocalEndpoint: &types.Endpoint{
				NodeName: "local-node",
			},
			RemoteEndpoints: make(map[types.GatewayName]*types.Endpoint),
			LocalNodeInfo:   make(map[types.NodeName]*v1beta1.NodeInfo),
			RemoteNodeInfo:  make(map[types.NodeName]*v1beta1.NodeInfo),
		},
		localGateway: &v1beta1.Gateway{
			ObjectMeta: metav1.ObjectMeta{Name: "local-gateway"},
		},
		nodeInfos: map[types.NodeName]*v1beta1.NodeInfo{
			"test-node": {
				NodeName:  "test-node",
				PrivateIP: "1.2.3.4",
			},
		},
	}

	gw := &v1beta1.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-gateway",
		},
		Spec: v1beta1.GatewaySpec{
			ExposeType: v1beta1.ExposeTypeLoadBalancer,
		},
		Status: v1beta1.GatewayStatus{
			Nodes: []v1beta1.NodeInfo{
				{NodeName: "test-node", PrivateIP: "1.2.3.4", Subnets: []string{"10.0.0.0/24"}},
			},
			ActiveEndpoints: []*v1beta1.Endpoint{
				{
					NodeName:   "test-node",
					Type:       v1beta1.Tunnel,
					PublicIP:   "1.2.3.4",
					Port:       4500,
					PublicPort: 4500,
				},
			},
		},
	}

	te.syncGateway(gw)

	ep, ok := te.network.RemoteEndpoints["test-gateway"]
	if !ok {
		t.Fatalf("syncGateway failed to add the gateway to RemoteEndpoints")
	}

	if ep.ExposeType != string(v1beta1.ExposeTypeLoadBalancer) {
		t.Errorf("expected ExposeType %q, got %q", string(v1beta1.ExposeTypeLoadBalancer), ep.ExposeType)
	}

	if ep.PublicIP != "1.2.3.4" || ep.PublicPort != 4500 {
		t.Errorf("expected endpoint 1.2.3.4:4500, got %s:%d", ep.PublicIP, ep.PublicPort)
	}
}
