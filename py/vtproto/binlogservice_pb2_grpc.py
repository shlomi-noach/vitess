# Generated by the gRPC Python protocol compiler plugin. DO NOT EDIT!
import grpc

import binlogdata_pb2 as binlogdata__pb2


class UpdateStreamStub(object):
  """UpdateStream is the RPC version of binlog.UpdateStream.
  """

  def __init__(self, channel):
    """Constructor.

    Args:
      channel: A grpc.Channel.
    """
    self.StreamKeyRange = channel.unary_stream(
        '/binlogservice.UpdateStream/StreamKeyRange',
        request_serializer=binlogdata__pb2.StreamKeyRangeRequest.SerializeToString,
        response_deserializer=binlogdata__pb2.StreamKeyRangeResponse.FromString,
        )
    self.StreamTables = channel.unary_stream(
        '/binlogservice.UpdateStream/StreamTables',
        request_serializer=binlogdata__pb2.StreamTablesRequest.SerializeToString,
        response_deserializer=binlogdata__pb2.StreamTablesResponse.FromString,
        )


class UpdateStreamServicer(object):
  """UpdateStream is the RPC version of binlog.UpdateStream.
  """

  def StreamKeyRange(self, request, context):
    """StreamKeyRange returns the binlog transactions related to
    the specified Keyrange.
    """
    context.set_code(grpc.StatusCode.UNIMPLEMENTED)
    context.set_details('Method not implemented!')
    raise NotImplementedError('Method not implemented!')

  def StreamTables(self, request, context):
    """StreamTables returns the binlog transactions related to
    the specified Tables.
    """
    context.set_code(grpc.StatusCode.UNIMPLEMENTED)
    context.set_details('Method not implemented!')
    raise NotImplementedError('Method not implemented!')


def add_UpdateStreamServicer_to_server(servicer, server):
  rpc_method_handlers = {
      'StreamKeyRange': grpc.unary_stream_rpc_method_handler(
          servicer.StreamKeyRange,
          request_deserializer=binlogdata__pb2.StreamKeyRangeRequest.FromString,
          response_serializer=binlogdata__pb2.StreamKeyRangeResponse.SerializeToString,
      ),
      'StreamTables': grpc.unary_stream_rpc_method_handler(
          servicer.StreamTables,
          request_deserializer=binlogdata__pb2.StreamTablesRequest.FromString,
          response_serializer=binlogdata__pb2.StreamTablesResponse.SerializeToString,
      ),
  }
  generic_handler = grpc.method_handlers_generic_handler(
      'binlogservice.UpdateStream', rpc_method_handlers)
  server.add_generic_rpc_handlers((generic_handler,))
