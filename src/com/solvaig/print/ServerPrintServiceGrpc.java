package com.solvaig.print;

import static io.grpc.stub.ClientCalls.asyncUnaryCall;
import static io.grpc.stub.ClientCalls.asyncServerStreamingCall;
import static io.grpc.stub.ClientCalls.asyncClientStreamingCall;
import static io.grpc.stub.ClientCalls.asyncBidiStreamingCall;
import static io.grpc.stub.ClientCalls.blockingUnaryCall;
import static io.grpc.stub.ClientCalls.blockingServerStreamingCall;
import static io.grpc.stub.ClientCalls.futureUnaryCall;
import static io.grpc.MethodDescriptor.generateFullMethodName;
import static io.grpc.stub.ServerCalls.asyncUnaryCall;
import static io.grpc.stub.ServerCalls.asyncServerStreamingCall;
import static io.grpc.stub.ServerCalls.asyncClientStreamingCall;
import static io.grpc.stub.ServerCalls.asyncBidiStreamingCall;

@javax.annotation.Generated("by gRPC proto compiler")
public class ServerPrintServiceGrpc {

  private ServerPrintServiceGrpc() {}

  public static final String SERVICE_NAME = "print.ServerPrintService";

  // Static method descriptors that strictly reflect the proto.
  @io.grpc.ExperimentalApi
  public static final io.grpc.MethodDescriptor<com.solvaig.print.Print.Empty,
      com.solvaig.print.Print.PrintServices> METHOD_GET_PRINT_SERVICES =
      io.grpc.MethodDescriptor.create(
          io.grpc.MethodDescriptor.MethodType.UNARY,
          generateFullMethodName(
              "print.ServerPrintService", "GetPrintServices"),
          io.grpc.protobuf.ProtoUtils.marshaller(com.solvaig.print.Print.Empty.getDefaultInstance()),
          io.grpc.protobuf.ProtoUtils.marshaller(com.solvaig.print.Print.PrintServices.getDefaultInstance()));
  @io.grpc.ExperimentalApi
  public static final io.grpc.MethodDescriptor<com.solvaig.print.Print.PrintContent,
      com.solvaig.print.Print.PrintResponse> METHOD_PRINT =
      io.grpc.MethodDescriptor.create(
          io.grpc.MethodDescriptor.MethodType.CLIENT_STREAMING,
          generateFullMethodName(
              "print.ServerPrintService", "Print"),
          io.grpc.protobuf.ProtoUtils.marshaller(com.solvaig.print.Print.PrintContent.getDefaultInstance()),
          io.grpc.protobuf.ProtoUtils.marshaller(com.solvaig.print.Print.PrintResponse.getDefaultInstance()));

  public static ServerPrintServiceStub newStub(io.grpc.Channel channel) {
    return new ServerPrintServiceStub(channel);
  }

  public static ServerPrintServiceBlockingStub newBlockingStub(
      io.grpc.Channel channel) {
    return new ServerPrintServiceBlockingStub(channel);
  }

  public static ServerPrintServiceFutureStub newFutureStub(
      io.grpc.Channel channel) {
    return new ServerPrintServiceFutureStub(channel);
  }

  public static interface ServerPrintService {

    public void getPrintServices(com.solvaig.print.Print.Empty request,
        io.grpc.stub.StreamObserver<com.solvaig.print.Print.PrintServices> responseObserver);

    public io.grpc.stub.StreamObserver<com.solvaig.print.Print.PrintContent> print(
        io.grpc.stub.StreamObserver<com.solvaig.print.Print.PrintResponse> responseObserver);
  }

  public static interface ServerPrintServiceBlockingClient {

    public com.solvaig.print.Print.PrintServices getPrintServices(com.solvaig.print.Print.Empty request);
  }

  public static interface ServerPrintServiceFutureClient {

    public com.google.common.util.concurrent.ListenableFuture<com.solvaig.print.Print.PrintServices> getPrintServices(
        com.solvaig.print.Print.Empty request);
  }

  public static class ServerPrintServiceStub extends io.grpc.stub.AbstractStub<ServerPrintServiceStub>
      implements ServerPrintService {
    private ServerPrintServiceStub(io.grpc.Channel channel) {
      super(channel);
    }

    private ServerPrintServiceStub(io.grpc.Channel channel,
        io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected ServerPrintServiceStub build(io.grpc.Channel channel,
        io.grpc.CallOptions callOptions) {
      return new ServerPrintServiceStub(channel, callOptions);
    }

    @java.lang.Override
    public void getPrintServices(com.solvaig.print.Print.Empty request,
        io.grpc.stub.StreamObserver<com.solvaig.print.Print.PrintServices> responseObserver) {
      asyncUnaryCall(
          getChannel().newCall(METHOD_GET_PRINT_SERVICES, getCallOptions()), request, responseObserver);
    }

    @java.lang.Override
    public io.grpc.stub.StreamObserver<com.solvaig.print.Print.PrintContent> print(
        io.grpc.stub.StreamObserver<com.solvaig.print.Print.PrintResponse> responseObserver) {
      return asyncClientStreamingCall(
          getChannel().newCall(METHOD_PRINT, getCallOptions()), responseObserver);
    }
  }

  public static class ServerPrintServiceBlockingStub extends io.grpc.stub.AbstractStub<ServerPrintServiceBlockingStub>
      implements ServerPrintServiceBlockingClient {
    private ServerPrintServiceBlockingStub(io.grpc.Channel channel) {
      super(channel);
    }

    private ServerPrintServiceBlockingStub(io.grpc.Channel channel,
        io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected ServerPrintServiceBlockingStub build(io.grpc.Channel channel,
        io.grpc.CallOptions callOptions) {
      return new ServerPrintServiceBlockingStub(channel, callOptions);
    }

    @java.lang.Override
    public com.solvaig.print.Print.PrintServices getPrintServices(com.solvaig.print.Print.Empty request) {
      return blockingUnaryCall(
          getChannel().newCall(METHOD_GET_PRINT_SERVICES, getCallOptions()), request);
    }
  }

  public static class ServerPrintServiceFutureStub extends io.grpc.stub.AbstractStub<ServerPrintServiceFutureStub>
      implements ServerPrintServiceFutureClient {
    private ServerPrintServiceFutureStub(io.grpc.Channel channel) {
      super(channel);
    }

    private ServerPrintServiceFutureStub(io.grpc.Channel channel,
        io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected ServerPrintServiceFutureStub build(io.grpc.Channel channel,
        io.grpc.CallOptions callOptions) {
      return new ServerPrintServiceFutureStub(channel, callOptions);
    }

    @java.lang.Override
    public com.google.common.util.concurrent.ListenableFuture<com.solvaig.print.Print.PrintServices> getPrintServices(
        com.solvaig.print.Print.Empty request) {
      return futureUnaryCall(
          getChannel().newCall(METHOD_GET_PRINT_SERVICES, getCallOptions()), request);
    }
  }

  public static io.grpc.ServerServiceDefinition bindService(
      final ServerPrintService serviceImpl) {
    return io.grpc.ServerServiceDefinition.builder(SERVICE_NAME)
      .addMethod(
        METHOD_GET_PRINT_SERVICES,
        asyncUnaryCall(
          new io.grpc.stub.ServerCalls.UnaryMethod<
              com.solvaig.print.Print.Empty,
              com.solvaig.print.Print.PrintServices>() {
            @java.lang.Override
            public void invoke(
                com.solvaig.print.Print.Empty request,
                io.grpc.stub.StreamObserver<com.solvaig.print.Print.PrintServices> responseObserver) {
              serviceImpl.getPrintServices(request, responseObserver);
            }
          }))
      .addMethod(
        METHOD_PRINT,
        asyncClientStreamingCall(
          new io.grpc.stub.ServerCalls.ClientStreamingMethod<
              com.solvaig.print.Print.PrintContent,
              com.solvaig.print.Print.PrintResponse>() {
            @java.lang.Override
            public io.grpc.stub.StreamObserver<com.solvaig.print.Print.PrintContent> invoke(
                io.grpc.stub.StreamObserver<com.solvaig.print.Print.PrintResponse> responseObserver) {
              return serviceImpl.print(responseObserver);
            }
          })).build();
  }
}
