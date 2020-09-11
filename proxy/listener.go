package proxy

import (
	"github.com/seknox/trasadbproxy/dbstore"
	"github.com/seknox/trasadbproxy/vitess/go/mysql"
	"github.com/seknox/trasadbproxy/vitess/go/sync2"
	"github.com/seknox/trasadbproxy/vitess/go/vt/vttls"
)

func StartListner() {

	dbstore.DBState.Init()

	handler := NewProxyHandler()

	authServer := NewTrasaAuthServer()
	//authServer.Method = mysql.MysqlNativePassword
	//authServer.Entries["root"] = []*mysql.AuthServerStaticEntry{
	//	{Password: "password"},
	//}

	//authServer := NewTrasaAuthServer()
	authServer.Method = mysql.MysqlClearPassword

	l, err := mysql.NewListener("tcp", dbstore.DBState.ListenAddr, authServer, handler, 0, 0, false)

	if err != nil {
		panic(err)
	}

	serverConfig, err := vttls.ServerConfig(
		"/etc/trasa/certs/node.crt",
		"/etc/trasa/certs/node.key",
		"/etc/trasa/certs/ca.crt",
	)
	if err != nil {
		panic("TLSServerConfig failed:  " + err.Error())

	}
	l.AllowClearTextWithoutTLS = sync2.AtomicBool{}
	l.AllowClearTextWithoutTLS.Set(true)
	l.RequireSecureTransport = false
	serverConfig.InsecureSkipVerify = true
	//l.TLSConfig = serverConfig
	l.TLSConfig.Store(serverConfig)

	l.Accept()

}
