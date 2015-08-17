package com.youtube.vitess.client;

import com.google.common.collect.Iterables;

import com.youtube.vitess.proto.Query.QueryResult;
import com.youtube.vitess.proto.Topodata.KeyRange;
import com.youtube.vitess.proto.Topodata.SrvKeyspace;
import com.youtube.vitess.proto.Topodata.TabletType;
import com.youtube.vitess.proto.Vtgate.BeginRequest;
import com.youtube.vitess.proto.Vtgate.BeginResponse;
import com.youtube.vitess.proto.Vtgate.BoundKeyspaceIdQuery;
import com.youtube.vitess.proto.Vtgate.BoundShardQuery;
import com.youtube.vitess.proto.Vtgate.ExecuteBatchKeyspaceIdsRequest;
import com.youtube.vitess.proto.Vtgate.ExecuteBatchKeyspaceIdsResponse;
import com.youtube.vitess.proto.Vtgate.ExecuteBatchShardsRequest;
import com.youtube.vitess.proto.Vtgate.ExecuteBatchShardsResponse;
import com.youtube.vitess.proto.Vtgate.ExecuteEntityIdsRequest;
import com.youtube.vitess.proto.Vtgate.ExecuteEntityIdsResponse;
import com.youtube.vitess.proto.Vtgate.ExecuteKeyRangesRequest;
import com.youtube.vitess.proto.Vtgate.ExecuteKeyRangesResponse;
import com.youtube.vitess.proto.Vtgate.ExecuteKeyspaceIdsRequest;
import com.youtube.vitess.proto.Vtgate.ExecuteKeyspaceIdsResponse;
import com.youtube.vitess.proto.Vtgate.ExecuteRequest;
import com.youtube.vitess.proto.Vtgate.ExecuteResponse;
import com.youtube.vitess.proto.Vtgate.ExecuteShardsRequest;
import com.youtube.vitess.proto.Vtgate.ExecuteShardsResponse;
import com.youtube.vitess.proto.Vtgate.GetSrvKeyspaceRequest;
import com.youtube.vitess.proto.Vtgate.GetSrvKeyspaceResponse;
import com.youtube.vitess.proto.Vtgate.SplitQueryRequest;
import com.youtube.vitess.proto.Vtgate.SplitQueryResponse;
import com.youtube.vitess.proto.Vtgate.StreamExecuteKeyRangesRequest;
import com.youtube.vitess.proto.Vtgate.StreamExecuteKeyspaceIdsRequest;
import com.youtube.vitess.proto.Vtgate.StreamExecuteRequest;
import com.youtube.vitess.proto.Vtgate.StreamExecuteShardsRequest;

import java.io.Closeable;
import java.io.IOException;
import java.util.List;
import java.util.Map;

/**
 * VTGateConn manages a VTGate connection.
 *
 * <p>Usage:
 *
 * <code>
 *   RpcClient client = RpcClientFactory.create(
 *     InetAddresses.forUriString("${VTGATE_ADDR}", Duration.millis(500)));
 *   VTGateConn conn = VTGateConn.WithRpcClient(client);
 *   Context ctx = Context.withDeadline(DateTime.now().plusMillis(500));
 *
 *   try {
 *
 *       CallerID callerId = CallerID.newBuilder().setPrincipal("username").build();
 *       BindVariable bindVars = BindVariable.newBuilder()
 *           .setType(Type.TYPE_INT)
 *           .setValueInt(12345)
 *           .build();
 *       BoundQuery.Builder queryBuilder = BoundQuery.newBuilder()
 *           .setSql(ByteString.copyFrom("INSERT INTO test_table VALUES(1, 2, 3)", Charsets.UTF_8));
 *       queryBuilder.getMutableBindVariables().put("keyspaceid_01", bindVars);
 *       BoundQuery query = queryBuilder.build();
 *       ExecuteKeyspaceIdsRequest.newBuilder()
 *           .setCallerId(callerId)
 *           .setQuery(query)
 *           .setKeyspace("my_keyspace")
 *           .setKeyspaceIds(0, ByteString.copyFrom("keyspaceid_01", Charsets.UTF_8))
 *           .setTabletType(TabletType.MASTER)
 *           .build();
 *
 *       ExecuteKeyspaceIdsResponse response = conn.executeKeyspaceIds(ctx, request);
 *       if (response.hasError()) {
 *         // handle error.
 *       }
 *       QueryResult result = response.getResult();
 *       for (Row row : result.getRowsList()) {
 *         // process each row.
 *       }
 *  } catch (VitessRpcException e) {
 *     // ...
 *   }
 * </code>
 * */
public class VTGateConn implements Closeable {
  private RpcClient client;

  private VTGateConn(RpcClient client) {
    this.client = client;
  }

  public static VTGateConn WithRpcClient(RpcClient client) {
    return new VTGateConn(client);
  }

  public QueryResult execute(Context ctx, String query, Map<String, ?> bindVars,
      TabletType tabletType) throws VitessException, VitessRpcException {
    ExecuteRequest request =
        ExecuteRequest.newBuilder()
            .setQuery(Proto.bindQuery(query, bindVars))
            .setTabletType(tabletType)
            .build();
    ExecuteResponse response = this.client.execute(ctx, request);
    Proto.checkError(response.getError());
    return response.getResult();
  }

  public QueryResult executeShards(Context ctx, String query, String keyspace,
      Iterable<String> shards, Map<String, ?> bindVars, TabletType tabletType)
      throws VitessException, VitessRpcException {
    ExecuteShardsRequest request =
        ExecuteShardsRequest.newBuilder()
            .setQuery(Proto.bindQuery(query, bindVars))
            .setKeyspace(keyspace)
            .addAllShards(shards)
            .setTabletType(tabletType)
            .build();
    ExecuteShardsResponse response = this.client.executeShards(ctx, request);
    Proto.checkError(response.getError());
    return response.getResult();
  }

  public QueryResult executeKeyspaceIds(Context ctx, String query, String keyspace,
      Iterable<byte[]> keyspaceIds, Map<String, ?> bindVars, TabletType tabletType)
      throws VitessException, VitessRpcException {
    ExecuteKeyspaceIdsRequest request =
        ExecuteKeyspaceIdsRequest.newBuilder()
            .setQuery(Proto.bindQuery(query, bindVars))
            .setKeyspace(keyspace)
            .addAllKeyspaceIds(Iterables.transform(keyspaceIds, Proto.BYTE_ARRAY_TO_BYTE_STRING))
            .setTabletType(tabletType)
            .build();
    ExecuteKeyspaceIdsResponse response = this.client.executeKeyspaceIds(ctx, request);
    Proto.checkError(response.getError());
    return response.getResult();
  }

  public QueryResult executeKeyRanges(Context ctx, String query, String keyspace,
      Iterable<? extends KeyRange> keyRanges, Map<String, ?> bindVars, TabletType tabletType)
      throws VitessException, VitessRpcException {
    ExecuteKeyRangesRequest request =
        ExecuteKeyRangesRequest.newBuilder()
            .setQuery(Proto.bindQuery(query, bindVars))
            .setKeyspace(keyspace)
            .addAllKeyRanges(keyRanges)
            .setTabletType(tabletType)
            .build();
    ExecuteKeyRangesResponse response = this.client.executeKeyRanges(ctx, request);
    Proto.checkError(response.getError());
    return response.getResult();
  }

  public QueryResult executeEntityIds(Context ctx, String query, String keyspace,
      String entityColumnName, Iterable<?> entityIds, Map<String, ?> bindVars,
      TabletType tabletType) throws VitessException, VitessRpcException {
    ExecuteEntityIdsRequest request =
        ExecuteEntityIdsRequest.newBuilder()
            .setQuery(Proto.bindQuery(query, bindVars))
            .setKeyspace(keyspace)
            .setEntityColumnName(entityColumnName)
            .addAllEntityKeyspaceIds(Iterables.transform(entityIds, Proto.OBJECT_TO_ENTITY_ID))
            .setTabletType(tabletType)
            .build();
    ExecuteEntityIdsResponse response = this.client.executeEntityIds(ctx, request);
    Proto.checkError(response.getError());
    return response.getResult();
  }

  public List<QueryResult> executeBatchShards(Context ctx,
      Iterable<? extends BoundShardQuery> queries, TabletType tabletType, boolean asTransaction)
      throws VitessException, VitessRpcException {
    ExecuteBatchShardsRequest request =
        ExecuteBatchShardsRequest.newBuilder()
            .addAllQueries(queries)
            .setTabletType(tabletType)
            .setAsTransaction(asTransaction)
            .build();
    ExecuteBatchShardsResponse response = this.client.executeBatchShards(ctx, request);
    Proto.checkError(response.getError());
    return response.getResultsList();
  }

  public List<QueryResult> executeBatchKeyspaceIds(Context ctx,
      Iterable<? extends BoundKeyspaceIdQuery> queries, TabletType tabletType,
      boolean asTransaction) throws VitessException, VitessRpcException {
    ExecuteBatchKeyspaceIdsRequest request =
        ExecuteBatchKeyspaceIdsRequest.newBuilder()
            .addAllQueries(queries)
            .setTabletType(tabletType)
            .setAsTransaction(asTransaction)
            .build();
    ExecuteBatchKeyspaceIdsResponse response = this.client.executeBatchKeyspaceIds(ctx, request);
    Proto.checkError(response.getError());
    return response.getResultsList();
  }

  public StreamIterator<QueryResult> streamExecute(Context ctx, String query,
      Map<String, ?> bindVars, TabletType tabletType) throws VitessRpcException {
    StreamExecuteRequest request =
        StreamExecuteRequest.newBuilder()
            .setQuery(Proto.bindQuery(query, bindVars))
            .setTabletType(tabletType)
            .build();
    return this.client.streamExecute(ctx, request);
  }

  public StreamIterator<QueryResult> streamExecuteShards(Context ctx, String query, String keyspace,
      Iterable<String> shards, Map<String, ?> bindVars, TabletType tabletType)
      throws VitessRpcException {
    StreamExecuteShardsRequest request =
        StreamExecuteShardsRequest.newBuilder()
            .setQuery(Proto.bindQuery(query, bindVars))
            .setKeyspace(keyspace)
            .addAllShards(shards)
            .setTabletType(tabletType)
            .build();
    return this.client.streamExecuteShards(ctx, request);
  }

  public StreamIterator<QueryResult> streamExecuteKeyspaceIds(Context ctx, String query,
      String keyspace, Iterable<byte[]> keyspaceIds, Map<String, ?> bindVars, TabletType tabletType)
      throws VitessRpcException {
    StreamExecuteKeyspaceIdsRequest request =
        StreamExecuteKeyspaceIdsRequest.newBuilder()
            .setQuery(Proto.bindQuery(query, bindVars))
            .setKeyspace(keyspace)
            .addAllKeyspaceIds(Iterables.transform(keyspaceIds, Proto.BYTE_ARRAY_TO_BYTE_STRING))
            .setTabletType(tabletType)
            .build();
    return this.client.streamExecuteKeyspaceIds(ctx, request);
  }

  public StreamIterator<QueryResult> streamExecuteKeyRanges(Context ctx, String query,
      String keyspace, Iterable<? extends KeyRange> keyRanges, Map<String, ?> bindVars,
      TabletType tabletType) throws VitessRpcException {
    StreamExecuteKeyRangesRequest request =
        StreamExecuteKeyRangesRequest.newBuilder()
            .setQuery(Proto.bindQuery(query, bindVars))
            .setKeyspace(keyspace)
            .addAllKeyRanges(keyRanges)
            .setTabletType(tabletType)
            .build();
    return this.client.streamExecuteKeyRanges(ctx, request);
  }

  public VTGateTx begin(Context ctx) throws VitessException, VitessRpcException {
    BeginRequest request = BeginRequest.newBuilder().build();
    BeginResponse response = this.client.begin(ctx, request);
    return VTGateTx.withRpcClientAndSession(this.client, response.getSession());
  }

  public List<SplitQueryResponse.Part> splitQuery(Context ctx, String keyspace, String query,
      Map<String, ?> bindVars, String splitColumn, long splitCount)
      throws VitessException, VitessRpcException {
    SplitQueryRequest request =
        SplitQueryRequest.newBuilder()
            .setKeyspace(keyspace)
            .setQuery(Proto.bindQuery(query, bindVars))
            .setSplitColumn(splitColumn)
            .setSplitCount(splitCount)
            .build();
    SplitQueryResponse response = this.client.splitQuery(ctx, request);
    return response.getSplitsList();
  }

  public SrvKeyspace getSrvKeyspace(Context ctx, String keyspace)
      throws VitessException, VitessRpcException {
    GetSrvKeyspaceRequest request =
        GetSrvKeyspaceRequest.newBuilder().setKeyspace(keyspace).build();
    GetSrvKeyspaceResponse response = this.client.getSrvKeyspace(ctx, request);
    return response.getSrvKeyspace();
  }

  @Override
  public void close() throws IOException {
    this.client.close();
  }
}
