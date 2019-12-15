package proxy

import (
	"gitlab.com/seknox/trasa/trasadbproxy/dbstore"
	"gitlab.com/seknox/trasa/trasadbproxy/vitess/go/mysql"
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

	l, err := mysql.NewListener("tcp", "127.0.0.1:1999", authServer, handler, 0, 0)

	//	serverConfig, err := vttls.ServerConfig(
	//		"/Users/bhrg3se/go/src/Practice/certs/certs/node.crt",
	//		"/Users/bhrg3se/go/src/Practice/certs/certs/node.key",
	//		"/Users/bhrg3se/go/src/Practice/certs/certs/ca.crt",
	//	)
	//	if err != nil {
	//		panic("TLSServerConfig failed:  "+err.Error())
	//	}
	//
	l.AllowClearTextWithoutTLS = true
	l.RequireSecureTransport = false
	//serverConfig.InsecureSkipVerify=true
	//	l.TLSConfig=serverConfig

	if err != nil {
		panic(err)
	}

	l.Accept()

}
