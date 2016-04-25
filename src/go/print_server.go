package main

import (
	pb "github.com/procks/direct_print_server/src/go/print"
	"github.com/procks/printer"

	"github.com/lxn/walk"

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

var (
	printServer *grpc.Server
	discoveryServerConn *net.UDPConn
	stoped bool
	startAction *walk.Action
	notifyIcon *walk.NotifyIcon
	iconPlay *walk.Icon
	iconStop *walk.Icon
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
	discoveryServerConn, err = net.ListenUDP("udp", ServerAddr)
	CheckError(err)

	//defer discoveryServerConn.Close() // closed in stop func
	buf := make([]byte, 1024)
	for {
		if stoped {
			break
		}
		n, addr, err := discoveryServerConn.ReadFromUDP(buf)
		fmt.Println("Received ", string(buf[0 : n]), " from ", addr)
		if err != nil {
			fmt.Println("Error: ", err)
		}
		if bytes.Equal([]byte("DISCOVER_PRINT_SERVER_REQUEST"), buf[0 : n]) {
			response := []byte("DISCOVER_PRINT_SERVER_RESPONSE")
			n, err = discoveryServerConn.WriteToUDP(response, addr);
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
	 	settings, err := printer.GetDefaultSettings(name, port)

		log.Printf("GetDefaultSettings err %v", err)
		var defPaperSize int
		var defResolutionX int
		var defResolutionY int
		if len(settings) >= 3 {
			defPaperSize = settings[0]
			defResolutionX = settings[2]
			defResolutionY = settings[2]
		}

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
				log.Print(err)
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
			go func(cmd *exec.Cmd) {
				err := cmd.Run()
				if err != nil {
					log.Print(err)
				}
			}(cmd)
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
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Printf("failed to listen: %v", err)
	}

	printServer = grpc.NewServer()
	pb.RegisterServerPrintServiceServer(printServer, &server{})
	printServer.Serve(lis)
}

func runNotify() {
	// We need either a walk.MainWindow or a walk.Dialog for their message loop.
	// We will not make it visible in this example, though.
	mw, err := walk.NewMainWindow()
	if err != nil {
		log.Print(err)
	}

	// We load our icon from a file.
	iconPlay, err = walk.NewIconFromFile("play.ico")
	if err != nil {
		log.Print(err)
	}

	// We load our icon from a file.
	iconStop, err = walk.NewIconFromFile("stop.ico")
	if err != nil {
		log.Print(err)
	}

	// Create the notify icon and make sure we clean it up on exit.
	notifyIcon, err = walk.NewNotifyIcon()
	if err != nil {
		log.Print(err)
	}
	defer notifyIcon.Dispose()

	if err := notifyIcon.SetToolTip("Direct Print Server"); err != nil {
		log.Print(err)
	}

	// We put an exit action into the context menu.
	startAction = walk.NewAction()
	startAction.Triggered().Attach(func() {
		if stoped {
			start()
		} else {
			stop()
		}
	})
	if err := notifyIcon.ContextMenu().Actions().Add(startAction); err != nil {
		log.Print(err)
	}

	// We put an exit action into the context menu.
	exitAction := walk.NewAction()
	if err := exitAction.SetText("E&xit"); err != nil {
		log.Print(err)
	}
	exitAction.Triggered().Attach(func() {
		//stop()
		walk.App().Exit(0)
	})
	if err := notifyIcon.ContextMenu().Actions().Add(exitAction); err != nil {
		log.Print(err)
	}

	// The notify icon is hidden initially, so we have to make it visible.
	if err := notifyIcon.SetVisible(true); err != nil {
		log.Print(err)
	}

	start()
	// Run the message loop.
	mw.Run()
}

func stop() {
	stoped = true

	if err := startAction.SetText("Start"); err != nil {
		log.Print(err)
	}
	// Set the icon and a tool tip text.
	if err := notifyIcon.SetIcon(iconStop); err != nil {
		log.Print(err)
	}

	discoveryServerConn.Close()
	printServer.Stop()
}

func start() {
	stoped = false

	if err := startAction.SetText("Stop"); err != nil {
		log.Print(err)
	}
	// Set the icon and a tool tip text.
	if err := notifyIcon.SetIcon(iconPlay); err != nil {
		log.Print(err)
	}

	go start_print_server()
	go discovery()
}

func runApp() {
	runNotify()
}