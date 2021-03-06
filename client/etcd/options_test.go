// Copyright (c) 2016 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package etcd

import (
	"testing"
	"time"

	etcdsd "github.com/m3db/m3cluster/services/client/etcd"
	"github.com/m3db/m3x/instrument"

	"github.com/stretchr/testify/assert"
)

func TestCluster(t *testing.T) {
	c := NewCluster()
	assert.Equal(t, "", c.Zone())
	assert.Equal(t, 0, len(c.Endpoints()))

	c = c.SetZone("z")
	assert.Equal(t, "z", c.Zone())
	assert.Equal(t, 0, len(c.Endpoints()))

	c = c.SetEndpoints([]string{"e1"})
	assert.Equal(t, "z", c.Zone())
	assert.Equal(t, []string{"e1"}, c.Endpoints())
}

func TestOptions(t *testing.T) {
	opts := NewOptions()
	assert.Equal(t, "", opts.Zone())
	assert.Equal(t, "", opts.Env())
	assert.Equal(t, etcdsd.Configuration{}, opts.ServiceDiscoveryConfig())
	assert.Equal(t, "", opts.CacheDir())
	assert.Equal(t, "", opts.Service())
	assert.Equal(t, []Cluster{}, opts.Clusters())
	_, ok := opts.ClusterForZone("z")
	assert.False(t, ok)
	assert.Equal(t, instrument.NewOptions(), opts.InstrumentOptions())

	c1 := NewCluster().SetZone("z1")
	c2 := NewCluster().SetZone("z2")
	iopts := instrument.NewOptions().SetReportInterval(time.Minute)

	opts = opts.SetEnv("env").
		SetZone("zone").
		SetServiceDiscoveryConfig(etcdsd.Configuration{}).
		SetCacheDir("/dir").
		SetService("app").
		SetClusters([]Cluster{c1, c2}).
		SetInstrumentOptions(iopts)
	assert.Equal(t, "env", opts.Env())
	assert.Equal(t, "zone", opts.Zone())
	assert.Equal(t, etcdsd.Configuration{}, opts.ServiceDiscoveryConfig())
	assert.Equal(t, "/dir", opts.CacheDir())
	assert.Equal(t, "app", opts.Service())
	assert.Equal(t, 2, len(opts.Clusters()))
	c, ok := opts.ClusterForZone("z1")
	assert.True(t, ok)
	assert.Equal(t, c, c1)
	c, ok = opts.ClusterForZone("z2")
	assert.True(t, ok)
	assert.Equal(t, c, c2)
	assert.Equal(t, iopts, opts.InstrumentOptions())
}

func TestValidate(t *testing.T) {
	opts := NewOptions()
	assert.Error(t, opts.Validate())

	opts = opts.SetService("app")
	assert.Error(t, opts.Validate())

	c1 := NewCluster().SetZone("z1")
	c2 := NewCluster().SetZone("z2")
	opts = opts.SetClusters([]Cluster{c1, c2})
	assert.NoError(t, opts.Validate())

	opts = opts.SetInstrumentOptions(nil)
	assert.Error(t, opts.Validate())
}
