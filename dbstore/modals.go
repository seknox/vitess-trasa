package dbstore

import (
	minio "github.com/minio/minio-go"
	"github.com/olivere/elastic"
	geoip2 "github.com/oschwald/geoip2-golang"
	"gitlab.com/seknox/trasa/trasadbproxy/vitess/go/mysql"
	"net"
	"os"
	"time"
)

type DBCONN struct {
	elasticHostName string
	minioHostName   string
	elasticClient   *elastic.Client
	minioClient     *minio.Client
	geoDB           *geoip2.Reader
	orgId           string
	appID           string
	appSecret       string
	trasaServer     string
	ListenAddr      string
}

type AppLogin struct {
	AppID           string `json:"appID"`
	AppSecret       string `json:"appSecret"`
	TfaMethod       string `json:"tfaMethod"`
	TotpCode        string `json:"totpCode"`
	User            string `json:"user"`
	Password        string `json:"password"`
	DynamicAuthApp  bool   `json:"dynamicAuthApp"`
	IsSharedSession bool   `json:"isSharedSession"`
	AppType         string `json:"appType"`
	ClientIP        string `json:"clientIP"`
	OrgID           string `json:"orgID"`
	TrasaID         string `json:"trasaID"`
	Hostname        string `json:"hostname"`
}

type AppUser struct {
	PermissionID string   `json:"permissionID"`
	AppID        string   `json:"appID"`
	AppName      string   `json:"appName"`
	OrgID        string   `json:"orgID"`
	UserID       string   `json:"userID"`
	Email        string   `json:"email"`
	Permissions  string   `json:"permissions"`
	TfaMode      string   `json:"tfaMode"`
	UserAddedAt  string   `json:"userAddedAt"`
	Adhoc        bool     `json:"adhoc"`
	AppType      string   `json:"appType"`
	Hostname     string   `json:"hostname"`
	IsAdmin      bool     `json:"isAdmin"`
	Is2FAEnabled bool     `json:"is2FAEnabled"`
	Username     string   `json:"username"`
	Usernames    []string `json:"usernames"`
}

type errorStrings struct {
	Status string `json:"status"`
	Error  error  `json:"error,omitempty"`
	Reason string `json:"reason,omitempty"`
	Intent string `json:"intent,omitempty"`
}

type LogLogin struct {
	EventID            string   `json:"eventID"`
	Endpoint           string   `json:"endpoint"`
	SessionID          string   `json:"sessionID"`
	OrgID              string   `json:"orgID"`
	AppName            string   `json:"appName"`
	AppID              string   `json:"appID"`
	ServerIP           string   `json:"serverIP"`
	ServerName         string   `json:"serverName"`
	UserName           string   `json:"userName"`
	Email              string   `json:"email"`
	UserID             string   `json:"userID"`
	UserAgent          string   `json:"userAgent"`
	RegisteredDeviceID string   `json:"RegisteredDeviceID"`
	DeviceType         string   `json:"deviceType"`
	Commands           []string `json:"commands"`
	UserIP             string   `json:"userIP"`
	GeoLocation        struct {
		IsoCountryCode string    `json:"isoCountryCode"`
		City           string    `json:"city"`
		TimeZone       string    `json:"timeZone"`
		Location       []float64 `json:"location"`
	} `json:"geoLocation"`
	LoginMethod     string `json:"loginMethod"`
	Status          bool   `json:"status"`
	MarkedAs        string `json:"markedAs"`
	LoginTime       int64  `json:"loginTime"`
	LogoutTime      int64  `json:"logoutTime"`
	SessionDuration string `json:"sessionDuration"`
	FailedReason    string `json:"failedReason"`
	RecordedSession bool   `json:"recordedSession"`
}

type TrasaResponse struct {
	Status string         `json:"status"`
	Reason string         `json:"reason"`
	Data   []GuacResponse `json:"data"`
}

type GetAppResponse struct {
	Tokens struct {
		Session string `json:"session"`
		Csrf    string `json:"csrf"`
	} `json:"tokens"`
	AppUsers []AppUser `json:"appUsers"`
	UserID   string    `json:"userID"`
	Email    string    `json:"email"`
}

type GuacResponse struct {
	User          string `json:"user"`
	UserID        string `json:"userID"`
	Email         string `json:"email"`
	AppName       string `json:"appName"`
	AppID         string `json:"appID"`
	Password      string `json:"password"`
	Hostname      string `json:"hostname"`
	Type          string `json:"type"`
	RemoteAppName string `json:"remoteAppName"`
	SessionRecord bool   `json:"sessionRecord"`
}

type ProxyMedata struct {
	ClientAddr    net.Addr
	ServerAddr    net.Addr
	AppName       string
	AppID         string
	UserID        string
	Email         string
	SessionID     string
	Username      string
	ClientVersion string
	SessionRecord bool
	TempLogFile   *os.File
	UpstreamConn  *mysql.Conn
	LoginTime     time.Time
}
