package main

import (
	pb "DirectPrintServer/print"
	"github.com/procks/printer"
	"log"
	"net"
	"bytes"
	"fmt"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"io"
	"os/exec"
	"path/filepath"
	"os"
	"path"
)

const (
	port = ":9188"
)

func CheckError(err error) {
	if err  != nil {
		fmt.Println("Error: " , err)
	}
}

func discovery() {
	/* Lets prepare a address at any address at port 10001*/
	ServerAddr, err := net.ResolveUDPAddr("udp", port)
	CheckError(err)

	/* Now listen at selected port */
	ServerConn, err := net.ListenUDP("udp", ServerAddr)
	CheckError(err)

	defer ServerConn.Close()
	buf := make([]byte, 1024)
	for {
		n, addr, err := ServerConn.ReadFromUDP(buf)
		fmt.Println("Received ", string(buf[0 : n]), " from ", addr)
		if err != nil {
			fmt.Println("Error: ", err)
		}
		if bytes.Equal([]byte("DISCOVER_PRINT_SERVER_REQUEST"), buf[0 : n]) {
			response := []byte("DISCOVER_PRINT_SERVER_RESPONSE")
			n, err = ServerConn.WriteToUDP(response, addr);
			if err != nil {
				fmt.Println("Error: ", err)
			}
		}
	}
}

type dataReader struct{
	ch chan []byte
}

func (r *dataReader) Read(p []byte) (n int, err error) {
	data, ok := <- r.ch
	if ok {
		n = copy(p, data)
		return n, nil
	} else {
		return n, io.EOF
	}
}

func newDataReader() *dataReader {
	return &dataReader{make(chan []byte)}
}

type server struct{}

func (s *server) GetPrintServices(ctx context.Context, in *pb.Empty) (*pb.PrintServices, error) {
	names, err := printer.ReadNames()
	return &pb.PrintServices{Name: names}, err
}

func run(cmd *exec.Cmd) {
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func (s *server) Print(stream pb.ServerPrintService_PrintServer) error {
	var cmd *exec.Cmd = nil
	var out bytes.Buffer
	dataReader := newDataReader();
	for {
		content, err := stream.Recv()
		if err == io.EOF {
			close(dataReader.ch)
			stream.SendAndClose(&pb.PrintResponse{0})
			fmt.Printf("%q\n", out.String())
		}
		if err != nil {
			return err
		}

		if x, ok := content.GetPrintContentType().(*pb.PrintContent_PrintInfo); ok {
			dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
			if err != nil {
				log.Fatal(err)
			}
			cmd = exec.Command(path.Join(dir, "gswin32c.exe"),
				"-dNOPAUSE", "-dBATCH", "-sDEVICE=mswinpr2",
				fmt.Sprintf("-dNumCopies=%d", x.PrintInfo.Copies),
				fmt.Sprintf("-dDEVICEWIDTHPOINTS=%d", x.PrintInfo.PageSizeWidth),
				fmt.Sprintf("-dDEVICEHEIGHTPOINTS=%d", x.PrintInfo.PageSizeHeight),
				"-sOutputFile=%printer%" + x.PrintInfo.PrinterName,
				"-")
			fmt.Printf("cmd.Args %v\n", cmd.Args)
			cmd.Stdin = dataReader
			cmd.Stdout = &out
			go run(cmd)
		}
		if x, ok := content.GetPrintContentType().(*pb.PrintContent_Content); ok {
			if cmd != nil && cmd.Stdin != nil {
				dataReader.ch <- x.Content
			}
		}
	}
	return nil
}

func main() {
	go discovery()

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterServerPrintServiceServer(s, &server{})
	s.Serve(lis)
}