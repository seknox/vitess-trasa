package dbstore

import (
	minio "github.com/minio/minio-go/v7"
	geoip2 "github.com/oschwald/geoip2-golang"
	"github.com/seknox/trasadbproxy/vitess/go/mysql"
	"net"
	"os"
	"time"
)

type DBCONN struct {
	minioHostName      string
	minioClient        *minio.Client
	geoDB              *geoip2.Reader
	orgId              string
	insecureSkipVerify bool
	trasaServer        string
	ListenAddr         string
}

type ProxyMedata struct {
	ClientAddr    net.Addr
	SessionID     string
	Username      string
	ClientVersion string
	SessionRecord bool
	TempLogFile   *os.File
	UpstreamConn  *mysql.Conn
	LoginTime     time.Time
}
