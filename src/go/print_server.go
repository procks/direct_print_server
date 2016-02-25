package main

import (
	pb "github.com/procks/direct_print_server/src/go/print"
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
	"math"
	"syscall"
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
	ServerAddr,err := net.ResolveUDPAddr("udp", port)
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
	printServs := make([]*pb.PrintServ, 0) //len(names)
	for _, name := range names {
		port, _ := printer.GetPrinterPort(name)
	 	settings, _ := printer.GetDefaultSettings(name, port)
		defPaperSize := settings[0]
		defResolutionX := settings[2]
		defResolutionY := settings[2]

		mediaNames, _ := printer.GetAllMediaNames(name, port)
		mediaSizes, _ := printer.GetAllMediaSizes(name, port)
		mediaIds, _ := printer.GetAllMediaIDs(name, port)
		resolArray, _ := printer.GetAllResolutions(name, port)
		//log.Print("settings ", settings)
		//log.Print("mediaNameArray ", mediaNames)
		//log.Printf("GetAllMediaIDs %v", mediaIds)
		pageSizes := make([]*pb.PageSize, 0)
		var defWidth int = 0
		var defHeight int = 0
		for i := 0; i < len(mediaNames); i++ {
			width := int32(mediaSizes[i * 2] * 1000 / 254)
			height := int32(mediaSizes[i * 2 + 1] * 1000 / 254)
			if (height > 0 && width > 0) && (math.Abs(float64(defWidth) - float64(width)) > 100 &&
					math.Abs(float64(defHeight) - float64(height)) > 100) {
				pageSize := pb.PageSize{Label: mediaNames[i], WidthMils: width,
					HeightMils: height, IsDefault: defPaperSize == mediaIds[i]}
				pageSizes = append(pageSizes, &pageSize)
				if pageSize.IsDefault {
					defWidth = int(width)
					defHeight = int(height)
				}
			}
		}

		resolutions := make([]*pb.Resolution, 0)
		for i := 0; i < len(resolArray) / 2; i++ {
			resol := pb.Resolution{HorizontalDpi: int32(resolArray[i * 2]),
				VerticalDpi: int32(resolArray[i * 2 + 1]), IsDefault: resolArray[i * 2] == defResolutionX &&
				resolArray[i * 2 + 1] == defResolutionY}
			resolutions = append(resolutions, &resol)
		}

		//printServ := new(pb.PrintServ)
		printServ := pb.PrintServ{Name: name, PageSize: pageSizes, Resolution: resolutions}

		log.Printf("printServ %v", printServ)
		//printServ.Name = name
		printServs = append(printServs, &printServ)
	}
	//return &pb.PrintServices{Name: names, PrintService:{}}, err
	return &pb.PrintServices{Name: names, PrintService: printServs}, err
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
			cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
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

func start_print_server() {
	go discovery()

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterServerPrintServiceServer(s, &server{})
	s.Serve(lis)
}