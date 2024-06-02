//
//Copyright 2019 The Vitess Authors.
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.

// This file contains the UpdateStream service definition, necessary
// to make RPC calls to VtTablet for the binlog protocol, used by
// filtered replication only.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.1
// 	protoc        v3.21.3
// source: binlogservice.proto

package binlogservice

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	binlogdata "vitess.io/vitess/go/vt/proto/binlogdata"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

var File_binlogservice_proto protoreflect.FileDescriptor

var file_binlogservice_proto_rawDesc = []byte{
	0x0a, 0x13, 0x62, 0x69, 0x6e, 0x6c, 0x6f, 0x67, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0d, 0x62, 0x69, 0x6e, 0x6c, 0x6f, 0x67, 0x73, 0x65, 0x72,
	0x76, 0x69, 0x63, 0x65, 0x1a, 0x10, 0x62, 0x69, 0x6e, 0x6c, 0x6f, 0x67, 0x64, 0x61, 0x74, 0x61,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x32, 0xc2, 0x01, 0x0a, 0x0c, 0x55, 0x70, 0x64, 0x61, 0x74,
	0x65, 0x53, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x12, 0x5b, 0x0a, 0x0e, 0x53, 0x74, 0x72, 0x65, 0x61,
	0x6d, 0x4b, 0x65, 0x79, 0x52, 0x61, 0x6e, 0x67, 0x65, 0x12, 0x21, 0x2e, 0x62, 0x69, 0x6e, 0x6c,
	0x6f, 0x67, 0x64, 0x61, 0x74, 0x61, 0x2e, 0x53, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x4b, 0x65, 0x79,
	0x52, 0x61, 0x6e, 0x67, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x22, 0x2e, 0x62,
	0x69, 0x6e, 0x6c, 0x6f, 0x67, 0x64, 0x61, 0x74, 0x61, 0x2e, 0x53, 0x74, 0x72, 0x65, 0x61, 0x6d,
	0x4b, 0x65, 0x79, 0x52, 0x61, 0x6e, 0x67, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x22, 0x00, 0x30, 0x01, 0x12, 0x55, 0x0a, 0x0c, 0x53, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x54, 0x61,
	0x62, 0x6c, 0x65, 0x73, 0x12, 0x1f, 0x2e, 0x62, 0x69, 0x6e, 0x6c, 0x6f, 0x67, 0x64, 0x61, 0x74,
	0x61, 0x2e, 0x53, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x54, 0x61, 0x62, 0x6c, 0x65, 0x73, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x20, 0x2e, 0x62, 0x69, 0x6e, 0x6c, 0x6f, 0x67, 0x64, 0x61,
	0x74, 0x61, 0x2e, 0x53, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x54, 0x61, 0x62, 0x6c, 0x65, 0x73, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x30, 0x01, 0x42, 0x2c, 0x5a, 0x2a, 0x76,
	0x69, 0x74, 0x65, 0x73, 0x73, 0x2e, 0x69, 0x6f, 0x2f, 0x76, 0x69, 0x74, 0x65, 0x73, 0x73, 0x2f,
	0x67, 0x6f, 0x2f, 0x76, 0x74, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x62, 0x69, 0x6e, 0x6c,
	0x6f, 0x67, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x33,
}

var file_binlogservice_proto_goTypes = []interface{}{
	(*binlogdata.StreamKeyRangeRequest)(nil),  // 0: binlogdata.StreamKeyRangeRequest
	(*binlogdata.StreamTablesRequest)(nil),    // 1: binlogdata.StreamTablesRequest
	(*binlogdata.StreamKeyRangeResponse)(nil), // 2: binlogdata.StreamKeyRangeResponse
	(*binlogdata.StreamTablesResponse)(nil),   // 3: binlogdata.StreamTablesResponse
}
var file_binlogservice_proto_depIdxs = []int32{
	0, // 0: binlogservice.UpdateStream.StreamKeyRange:input_type -> binlogdata.StreamKeyRangeRequest
	1, // 1: binlogservice.UpdateStream.StreamTables:input_type -> binlogdata.StreamTablesRequest
	2, // 2: binlogservice.UpdateStream.StreamKeyRange:output_type -> binlogdata.StreamKeyRangeResponse
	3, // 3: binlogservice.UpdateStream.StreamTables:output_type -> binlogdata.StreamTablesResponse
	2, // [2:4] is the sub-list for method output_type
	0, // [0:2] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_binlogservice_proto_init() }
func file_binlogservice_proto_init() {
	if File_binlogservice_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_binlogservice_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   0,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_binlogservice_proto_goTypes,
		DependencyIndexes: file_binlogservice_proto_depIdxs,
	}.Build()
	File_binlogservice_proto = out.File
	file_binlogservice_proto_rawDesc = nil
	file_binlogservice_proto_goTypes = nil
	file_binlogservice_proto_depIdxs = nil
}
