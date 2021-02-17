/*
Copyright 2019 The Vitess Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package grpctabletconn

import (
	"context"
	"io"
	"io/ioutil"
	"net"
	"os"
	"testing"

	"google.golang.org/grpc"

	"vitess.io/vitess/go/vt/callerid"
	"vitess.io/vitess/go/vt/servenv"
	"vitess.io/vitess/go/vt/vttablet/grpcqueryservice"
	"vitess.io/vitess/go/vt/vttablet/tabletconn"
	"vitess.io/vitess/go/vt/vttablet/tabletconntest"

	topodatapb "vitess.io/vitess/go/vt/proto/topodata"
)

func BenchmarkGRPCTabletConn(b *testing.B) {
	// fake service
	service := tabletconntest.CreateFakeServer(b)

	// listen on a random port
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		b.Fatalf("Cannot listen: %v", err)
	}
	host := listener.Addr().(*net.TCPAddr).IP.String()
	port := listener.Addr().(*net.TCPAddr).Port

	// Create a gRPC server and listen on the port
	server := grpc.NewServer()
	grpcqueryservice.Register(server, service)
	go server.Serve(listener)

	tablet := &topodatapb.Tablet{
		Keyspace: tabletconntest.TestTarget.Keyspace,
		Shard:    tabletconntest.TestTarget.Shard,
		Type:     tabletconntest.TestTarget.TabletType,
		Alias:    tabletconntest.TestAlias,
		Hostname: host,
		PortMap: map[string]int32{
			"grpc": int32(port),
		},
	}

	conn, err := tabletconn.GetDialer()(tablet, false)
	if err != nil {
		b.Fatalf("dial failed: %v", err)
	}
	defer conn.Close(context.Background())

	service.ExpectedTransactionID = tabletconntest.ExecuteBatchTransactionID

	b.SetParallelism(4)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ctx := context.Background()
			ctx = callerid.NewContext(ctx, tabletconntest.TestCallerID, tabletconntest.TestVTGateCallerID)
			_, err := conn.ExecuteBatch(ctx, tabletconntest.TestTarget, tabletconntest.ExecuteBatchQueries, tabletconntest.TestAsTransaction, tabletconntest.ExecuteBatchTransactionID, tabletconntest.TestExecuteOptions)
			if err != nil {
				b.Fatalf("ExecuteBatch failed: %v", err)
			}
		}
	})
}

// This test makes sure the go rpc service works
func TestGRPCTabletConn(t *testing.T) {
	// fake service
	service := tabletconntest.CreateFakeServer(t)

	// listen on a random port
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("Cannot listen: %v", err)
	}
	host := listener.Addr().(*net.TCPAddr).IP.String()
	port := listener.Addr().(*net.TCPAddr).Port

	// Create a gRPC server and listen on the port
	server := grpc.NewServer()
	grpcqueryservice.Register(server, service)
	go server.Serve(listener)

	// run the test suite
	tabletconntest.TestSuite(t, protocolName, &topodatapb.Tablet{
		Keyspace: tabletconntest.TestTarget.Keyspace,
		Shard:    tabletconntest.TestTarget.Shard,
		Type:     tabletconntest.TestTarget.TabletType,
		Alias:    tabletconntest.TestAlias,
		Hostname: host,
		PortMap: map[string]int32{
			"grpc": int32(port),
		},
	}, service, nil)
}

// This test makes sure the go rpc client auth works
func TestGRPCTabletAuthConn(t *testing.T) {
	// fake service
	service := tabletconntest.CreateFakeServer(t)

	// listen on a random port
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("Cannot listen: %v", err)
	}
	host := listener.Addr().(*net.TCPAddr).IP.String()
	port := listener.Addr().(*net.TCPAddr).Port

	// Create a gRPC server and listen on the port
	var opts []grpc.ServerOption
	opts = append(opts, grpc.StreamInterceptor(servenv.FakeAuthStreamInterceptor))
	opts = append(opts, grpc.UnaryInterceptor(servenv.FakeAuthUnaryInterceptor))
	server := grpc.NewServer(opts...)

	grpcqueryservice.Register(server, service)
	go server.Serve(listener)

	authJSON := `{
         "Username": "valid",
         "Password": "valid"
        }`

	f, err := ioutil.TempFile("", "static_auth_creds.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	if _, err := io.WriteString(f, authJSON); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	// run the test suite
	tabletconntest.TestSuite(t, protocolName, &topodatapb.Tablet{
		Keyspace: tabletconntest.TestTarget.Keyspace,
		Shard:    tabletconntest.TestTarget.Shard,
		Type:     tabletconntest.TestTarget.TabletType,
		Alias:    tabletconntest.TestAlias,
		Hostname: host,
		PortMap: map[string]int32{
			"grpc": int32(port),
		},
	}, service, f)
}
