package proxy

import (
	"context"
	"fmt"
	"vitess.io/vitess/go/mysql"
	"vitess.io/vitess/go/sqltypes"
	querypb "vitess.io/vitess/go/vt/proto/query"
)

type proxyHandler struct {
	connMap      map[*mysql.Conn]*mysql.Conn
	lastConn     *mysql.Conn
	upstreamConn *mysql.Conn
}

func NewProxyHandler() *proxyHandler {
	return &proxyHandler{connMap: make(map[*mysql.Conn]*mysql.Conn)}
}

func (th *proxyHandler) NewConnection(c *mysql.Conn) {

	c.CloseResult()

	fmt.Println("________New Connnnn____________", c.User)
	th.lastConn = c
	fmt.Println(c.User, c.RemoteAddr())

	params := mysql.ConnParams{
		Host:       "127.0.0.1",
		Port:       3306,
		Uname:      "root",
		Pass:       "password",
		Flavor:     "mariadb",
		ServerName: "localhost",

		//	Flavor:mysql.MysqlNativePassword,

	}

	cc, err := mysql.Connect(context.Background(), &params)
	fmt.Println(cc.ServerVersion)

	if err != nil {

		fmt.Println(err.Error(), "_________")
		c.Close()
		return
	}
	th.connMap[c] = cc

}

func (th *proxyHandler) ConnectionClosed(c *mysql.Conn) {
	fmt.Println("ConnectionClosed")
	upstreamConn := th.connMap[c]
	upstreamConn.Close()

}

func (th *proxyHandler) ComInitDB(c *mysql.Conn, schemaName string) {
	fmt.Println("________com______InitDB ", schemaName)
	if schemaName != "" {

		upstreamConn := th.connMap[c]

		_, err := upstreamConn.ExecuteFetch(`use `+schemaName, 0, false)
		if err != nil {
			fmt.Println(err)
		}
	}

}

func (th *proxyHandler) ComQuery(c *mysql.Conn, q string, callback func(*sqltypes.Result) error) error {

	fmt.Println(q)
	//fmt.Println(c.User)
	//fmt.Println(c.String())

	upstreamConn := th.connMap[c]

	ps, err := upstreamConn.ExecuteFetch(q, 1000, true)
	if err != nil {

		fmt.Println("exe err_____________", err, "__END")
		return err
	}
	if ps == nil {
		return callback(&sqltypes.Result{
			Fields:       nil,
			RowsAffected: 0,
			InsertID:     0,
			Rows:         nil,
			Extras:       nil,
		})
		//return nil
	}
	err = callback(ps)

	if err != nil {
		fmt.Println("callback err__________", err)
		return err
	}

	return nil
}

func (th *proxyHandler) ComPrepare(c *mysql.Conn, q string) ([]*querypb.Field, error) {
	fmt.Println("________com______Prepare")
	return nil, nil
}

func (th *proxyHandler) ComStmtExecute(c *mysql.Conn, prepare *mysql.PrepareData, callback func(*sqltypes.Result) error) error {
	fmt.Println("________com______StmtExecute")
	return nil
}

func (th *proxyHandler) ComResetConnection(c *mysql.Conn) {
	fmt.Println("________com______ResetConnection")

}

func (th *proxyHandler) WarningCount(c *mysql.Conn) uint16 {
	return 0
}
