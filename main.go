package main

//go:generate go-bindata data/

import (
	"bytes"
	"encoding/json"
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"syscall"

	_ "github.com/mattn/go-sqlite3"

	"github.com/anacrolix/dms/db"
	"github.com/anacrolix/dms/dlna/dms"
)

type dmsConfig struct {
	Path                string
	IfName              string
	Http                string
	FriendlyName        string
	MediaDBPath         string
	LogHeaders          bool
	NoTranscode         bool
	StallEventSubscribe bool
}

func (config *dmsConfig) load(configPath string) {
	file, err := os.Open(configPath)
	if err != nil {
		log.Printf("config error (config file: '%s'): %v\n", configPath, err)
		return
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		log.Printf("config error: %v\n", err)
		return
	}
}

//default config
var config = &dmsConfig{
	Path:         "",
	IfName:       "",
	Http:         ":1338",
	FriendlyName: "",
	LogHeaders:   false,
	MediaDBPath:  filepath.Join(getHomeDir(), ".dms.db"),
}

func getHomeDir() string {
	_user, err := user.Current()
	if err != nil {
		panic(err)
	}
	return _user.HomeDir
}

var mediaDb *db.Database

func main() {
	log.SetFlags(log.Ltime | log.Lshortfile)

	path := flag.String("path", config.Path, "browse root path")
	ifName := flag.String("ifname", config.IfName, "specific SSDP network interface")
	http := flag.String("http", config.Http, "http server port")
	friendlyName := flag.String("friendlyName", config.FriendlyName, "server friendly name")
	mediaDBPath := flag.String("mediadbpath", config.MediaDBPath, "catalogue db path")
	logHeaders := flag.Bool("logHeaders", config.LogHeaders, "log HTTP headers")
	configFilePath := flag.String("config", "", "json configuration file")
	flag.BoolVar(&config.NoTranscode, "noTranscode", false, "disable transcoding")
	flag.BoolVar(&config.StallEventSubscribe, "stallEventSubscribe", false, "workaround for some bad event subscribers")

	flag.Parse()
	if flag.NArg() != 0 {
		flag.Usage()
		log.Fatalf("%s: %s\n", "unexpected positional arguments", flag.Args())
	}

	config.Path = *path
	config.IfName = *ifName
	config.Http = *http
	config.FriendlyName = *friendlyName
	config.LogHeaders = *logHeaders

	if len(*configFilePath) > 0 {
		config.load(*configFilePath)
	}

	var err error
	mediaDb, err = db.Open(*mediaDBPath)
	if err != nil {
		panic(err)
	}

	dmsServer := &dms.Server{
		Interfaces: func(ifName string) (ifs []net.Interface) {
			var err error
			if ifName == "" {
				ifs, err = net.Interfaces()
			} else {
				var if_ *net.Interface
				if_, err = net.InterfaceByName(ifName)
				if if_ != nil {
					ifs = append(ifs, *if_)
				}
			}
			if err != nil {
				log.Fatal(err)
			}
			return
		}(config.IfName),
		HTTPConn: func() net.Listener {
			conn, err := net.Listen("tcp", config.Http)
			if err != nil {
				log.Fatal(err)
			}
			return conn
		}(),
		FriendlyName:   config.FriendlyName,
		RootObjectPath: filepath.Clean(config.Path),
		LogHeaders:     config.LogHeaders,
		NoTranscode:    config.NoTranscode,
		Icons: []dms.Icon{
			dms.Icon{
				Width:      48,
				Height:     48,
				Depth:      8,
				Mimetype:   "image/png",
				ReadSeeker: bytes.NewReader(MustAsset("data/VGC Sonic.png")),
			},
			dms.Icon{
				Width:      128,
				Height:     128,
				Depth:      8,
				Mimetype:   "image/png",
				ReadSeeker: bytes.NewReader(MustAsset("data/VGC Sonic 128.png")),
			},
		},
		StallEventSubscribe: config.StallEventSubscribe,
		Catalogue:           mediaDb,
	}
	go func() {
		if err := dmsServer.Serve(); err != nil {
			log.Fatal(err)
		}
	}()
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	<-sigs
	err = dmsServer.Close()
	if err != nil {
		log.Fatal(err)
	}
}
