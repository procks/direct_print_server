package com.solvaig.print;

import java.io.BufferedReader;
import java.io.IOException;
import java.io.InputStreamReader;
import java.io.OutputStream;
import java.net.DatagramPacket;
import java.net.DatagramSocket;
import java.net.InetAddress;
import java.util.ArrayList;
import java.util.List;
import java.util.logging.Level;
import java.util.logging.Logger;

import javax.print.PrintService;
import javax.print.PrintServiceLookup;

import com.solvaig.print.Print.Empty;
import com.solvaig.print.Print.PrintContent;
import com.solvaig.print.Print.PrintContent.PrintContentTypeCase;
import com.solvaig.print.Print.PrintInfo;
import com.solvaig.print.Print.PrintResponse;
import com.solvaig.print.Print.PrintServices;
import com.solvaig.print.Print.PrintServices.Builder;
import com.solvaig.print.ServerPrintServiceGrpc.ServerPrintService;

import io.grpc.Server;
import io.grpc.ServerBuilder;
import io.grpc.stub.StreamObserver;

/**
 * Server that manages startup/shutdown of a {@code Greeter} server.
 */
public class PrintServer {
	private static final Logger logger = Logger.getLogger(PrintServer.class.getName());

	private static Thread discoveryThread;
	private static PrintServer printServer;

	/* The port on which the server should run */
	private int port = 9188;
	private Server server;

	private void start() throws Exception {
		server = ServerBuilder.forPort(port)
				.addService(ServerPrintServiceGrpc.bindService(new ServerPrintServiceImpl())).build().start();
		logger.log(Level.INFO, "Server started, listening on " + port);
		Runtime.getRuntime().addShutdownHook(new Thread() {
			@Override
			public void run() {
				// Use stderr here since the logger may have been reset by its
				// JVM shutdown hook.
				System.err.println("*** shutting down gRPC server since JVM is shutting down");
				PrintServer.this.stop();
				System.err.println("*** server shut down");
			}
		});
	}

	private void stop() {
		if (server != null) {
			server.shutdown();
		}
	}

	/**
	 * Await termination on the main thread since the grpc library uses daemon
	 * threads.
	 */
	private void blockUntilShutdown() throws InterruptedException {
		if (server != null) {
			server.awaitTermination();
		}
	}

	/**
	 * Main launches the server from the command line.
	 */
	public static void main(String[] args) throws Exception {
		if (args.length > 0) {
			String mode = args[0];
			if ("stop".equals(mode)) {
				System.err.println("stop mode");
				DiscoveryThread.getInstance().terminate();
				
				printServer.stop();
				System.err.println("stop mode end");
				return;
			}
		}
		discoveryThread = new Thread(DiscoveryThread.getInstance());
		discoveryThread.start();

		printServer = new PrintServer();
		printServer.start();
		printServer.blockUntilShutdown();
	}

	private class ServerPrintServiceImpl implements ServerPrintService {
		@Override
		public StreamObserver<PrintContent> print(final StreamObserver<PrintResponse> responseObserver) {
			return new StreamObserver<PrintContent>() {
				private PrintInfo mPrintInfo;
				private OutputStream mOutputStream;
				private BufferedReader mInReader;

				@Override
				public void onCompleted() {
					try {
						mOutputStream.close();
						String line;
						while ((line = mInReader.readLine()) != null) {
							System.out.println(line);
						}
						mInReader.close();
					} catch (IOException e) {
						e.printStackTrace();
					}
					responseObserver.onCompleted();
				}

				@Override
				public void onError(Throwable t) {
					logger.log(Level.WARNING, "Encountered error in routeChat", t);
				}

				@Override
				public void onNext(PrintContent printContent) {
					if (printContent.getPrintContentTypeCase() == PrintContentTypeCase.PRINTINFO) {
						mPrintInfo = printContent.getPrintInfo();
						List<String> list = new ArrayList<String>();
						list.add("gswin32c.exe");
						list.add("-dNOPAUSE");
						list.add("-dBATCH");
						list.add("-sDEVICE=mswinpr2");
						list.add("-dNumCopies=" + mPrintInfo.getCopies());

						list.add("-dDEVICEWIDTHPOINTS=" + mPrintInfo.getPageSizeWidth());
						list.add("-dDEVICEHEIGHTPOINTS=" + mPrintInfo.getPageSizeHeight());
						list.add("-sOutputFile=%printer%" + mPrintInfo.getPrinterName());

						list.add("-");
						String[] cmd = new String[list.size()];
						list.toArray(cmd);

						ProcessBuilder pb = new ProcessBuilder(cmd);
						Process p;
						try {
							p = pb.start();
							mOutputStream = p.getOutputStream();
							mInReader = new BufferedReader(new InputStreamReader(p.getInputStream()));
						} catch (IOException e1) {
							e1.printStackTrace();
						}
					} else {
						try {
							mOutputStream.write(printContent.getContent().toByteArray());
						} catch (IOException e) {
							e.printStackTrace();
						}
					}
				}
			};
		}

		@Override
		public void getPrintServices(Empty request, StreamObserver<PrintServices> responseObserver) {
			PrintService[] services = PrintServiceLookup.lookupPrintServices(null, null);
			Builder builder = PrintServices.newBuilder();
			for (PrintService service : services) {
				if (service != null) {
					String printServiceName = service.getName();
					builder.addName(printServiceName);
					System.out.println("Print Service Name is " + printServiceName);
				} else {
					System.out.println("No default print service found");
				}
			}
			PrintServices printServices = builder.build();
			responseObserver.onNext(printServices);
			responseObserver.onCompleted();
		}
	}

	// http://michieldemey.be/blog/network-discovery-using-udp-broadcast/
	public static class DiscoveryThread implements Runnable {
		DatagramSocket socket;
		private volatile boolean running = true;
		
		public void terminate() {
			System.err.println("terminate");
			running = false;
			socket.close();
		}

		@Override
		public void run() {
			try {
				// Keep a socket open to listen to all the UDP trafic that is
				// destined for this port
				socket = new DatagramSocket(9188, InetAddress.getByName("0.0.0.0"));
				socket.setBroadcast(true);

				while (running) {
					System.out.println(getClass().getName() + ">>>Ready to receive broadcast packets!");

					// Receive a packet
					byte[] recvBuf = new byte[15000];
					DatagramPacket packet = new DatagramPacket(recvBuf, recvBuf.length);
					socket.receive(packet);

					// Packet received
					System.out.println(getClass().getName() + ">>>Discovery packet received from: "
							+ packet.getAddress().getHostAddress());
//					System.out.println(
//							getClass().getName() + ">>>Packet received; data: " + new String(packet.getData()));

					// See if the packet holds the right command (message)
					String message = new String(packet.getData()).trim();
					if (message.equals("DISCOVER_PRINT_SERVER_REQUEST")) {
						byte[] sendData = "DISCOVER_PRINT_SERVER_RESPONSE".getBytes();

						// Send a response
						DatagramPacket sendPacket = new DatagramPacket(sendData, sendData.length, packet.getAddress(),
								packet.getPort());
						socket.send(sendPacket);

						System.out.println(getClass().getName() + ">>>Sent packet to: "
								+ sendPacket.getAddress().getHostAddress());
					}
				}
			} catch (IOException ex) {
				Logger.getLogger(DiscoveryThread.class.getName()).log(Level.SEVERE, null, ex);
			}

			System.err.println("end run");
		}

		public static DiscoveryThread getInstance() {
			return DiscoveryThreadHolder.INSTANCE;
		}

		private static class DiscoveryThreadHolder {
			private static final DiscoveryThread INSTANCE = new DiscoveryThread();
		}
	}
}
