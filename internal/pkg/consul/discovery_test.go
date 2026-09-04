package consul

import (
	"testing"

	"github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInstance_URL(t *testing.T) {
	tests := []struct {
		name    string // description of this test case
		ins     Instance
		want    string
		wantErr bool
	}{
		{name: "full", ins: Instance{Port: 8006, Address: "am", Meta: map[string]string{META_SCHEME: "https", META_PATH: "/synt"}}, want: "https://am:8006/synt", wantErr: false},
		{name: "http", ins: Instance{Port: 8006, Address: "am", Meta: map[string]string{META_SCHEME: "http", META_PATH: "/synt"}}, want: "http://am:8006/synt", wantErr: false},
		{name: "no path", ins: Instance{Port: 8006, Address: "am", Meta: map[string]string{META_SCHEME: "http"}}, want: "http://am:8006/", wantErr: false},
		{name: "ip", ins: Instance{Port: 8006, Address: "1.1.1.1", Meta: map[string]string{META_SCHEME: "http", META_PATH: "/synt"}}, want: "http://1.1.1.1:8006/synt", wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// TODO: construct the receiver type.
			got, gotErr := tt.ins.URL()
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("URL() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("URL() succeeded unexpectedly")
			}
			if tt.want != got {
				t.Errorf("URL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDiscovery_OneInstance(t *testing.T) {
	d, err := New(t.Context(), &Config{SkipConsul: true})
	require.NoError(t, err)
	d.instances, err = buildInstances([]*api.ServiceEntry{{Node: &api.Node{}, Service: &api.AgentService{Service: "olia", Port: 8080, Address: "1.1.1.1", Tags: []string{"aa", "bb"},
		Meta: map[string]string{META_PATH: "/olia"}}}})
	require.NoError(t, err)
	u, err := d.URL("cc")
	assert.Error(t, err)
	assert.Empty(t, u)

	u, err = d.URL("aa")
	assert.NoError(t, err)
	assert.Equal(t, "http://1.1.1.1:8080/olia", u)

	u, err = d.URL("aa")
	assert.NoError(t, err)
	assert.Equal(t, "http://1.1.1.1:8080/olia", u)

	u, err = d.URL("bb")
	assert.NoError(t, err)
	assert.Equal(t, "http://1.1.1.1:8080/olia", u)
}

func TestDiscovery_NoInstances(t *testing.T) {
	d, err := New(t.Context(), &Config{SkipConsul: true})
	require.NoError(t, err)
	d.instances, err = buildInstances([]*api.ServiceEntry{})
	require.NoError(t, err)
	u, err := d.URL("cc")
	assert.Error(t, err)
	assert.Empty(t, u)

	u, err = d.URL("aa")
	assert.Error(t, err)
	assert.Empty(t, u)
}

func TestDiscovery_SeveralInstances(t *testing.T) {
	d, err := New(t.Context(), &Config{SkipConsul: true})
	require.NoError(t, err)
	d.instances, err = buildInstances([]*api.ServiceEntry{
		{Node: &api.Node{}, Service: &api.AgentService{Service: "olia", Port: 8080, Address: "1.1.1.1", Tags: []string{"aa", "bb"},
			Meta: map[string]string{META_PATH: "/olia"}}},
		{Node: &api.Node{}, Service: &api.AgentService{Service: "olia1", Port: 8000, Address: "1.1.1.2", Tags: []string{"aa", "bb", "cc"},
			Meta: map[string]string{META_PATH: "/olia"}}},
		{Node: &api.Node{}, Service: &api.AgentService{Service: "olia3", Port: 8060, Address: "1.1.1.3", Tags: []string{"aa", "aaa"},
			Meta: map[string]string{META_PATH: "/olia"}}},
	})
	require.NoError(t, err)
	u, _ := d.URL("cc")
	assert.Equal(t, "http://1.1.1.2:8000/olia", u)

	// rr
	u, _ = d.URL("aa")
	assert.Equal(t, "http://1.1.1.1:8080/olia", u)
	u, _ = d.URL("aa")
	assert.Equal(t, "http://1.1.1.2:8000/olia", u)
	u, _ = d.URL("aa")
	assert.Equal(t, "http://1.1.1.3:8060/olia", u)
	u, _ = d.URL("aa")
	assert.Equal(t, "http://1.1.1.1:8080/olia", u)

	u, _ = d.URL("bb")
	assert.Equal(t, "http://1.1.1.1:8080/olia", u)
	u, _ = d.URL("bb")
	assert.Equal(t, "http://1.1.1.2:8000/olia", u)
	u, _ = d.URL("bb")
	assert.Equal(t, "http://1.1.1.1:8080/olia", u)
	u, _ = d.URL("bb")
	assert.Equal(t, "http://1.1.1.2:8000/olia", u)

	u, _ = d.URL("aaa")
	assert.Equal(t, "http://1.1.1.3:8060/olia", u)
	u, _ = d.URL("aaa")
	assert.Equal(t, "http://1.1.1.3:8060/olia", u)
}
