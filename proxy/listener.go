package proxy

import (
	"gitlab.com/seknox/trasa/trasadbproxy/dbstore"
	"gitlab.com/seknox/trasa/trasadbproxy/vitess/go/mysql"
	"gitlab.com/seknox/trasa/trasadbproxy/vitess/go/vt/vttls"
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

	l, err := mysql.NewListener("tcp", dbstore.DBState.ListenAddr, authServer, handler, 0, 0)

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

	l.AllowClearTextWithoutTLS = true
	l.RequireSecureTransport = false
	serverConfig.InsecureSkipVerify = true
	l.TLSConfig = serverConfig

	if err != nil {
		panic(err)
	}

	l.Accept()

}
