package proxy

import (
	"context"
	"errors"
	"fmt"
	"github.com/seknox/trasadbproxy/dbstore"
	"github.com/seknox/trasadbproxy/vitess/go/mysql"
	"github.com/seknox/trasadbproxy/vitess/go/sqltypes"
	"github.com/seknox/trasadbproxy/vitess/go/vt/proto/query"
	logger "github.com/sirupsen/logrus"
	"os"
	"strings"
	"time"
)

//proxyHandler is a custom implementation of mysql.Handler
type proxyHandler struct {
	connMap  map[*mysql.Conn]dbstore.ProxyMedata
	lastConn *mysql.Conn
}

func NewProxyHandler() *proxyHandler {
	return &proxyHandler{connMap: make(map[*mysql.Conn]dbstore.ProxyMedata)}
}

func (th *proxyHandler) NewConnection(c *mysql.Conn) {

	logger.Trace("New Connnnn ", c.User)
	//	th.lastConn = c

}

func (th *proxyHandler) ConnectionClosed(c *mysql.Conn) {
	logger.Trace("ConnectionClosed")
	proxyMeta := th.connMap[c]

	upstreamConn := proxyMeta.UpstreamConn

	if upstreamConn != nil && !upstreamConn.IsClosed() {
		upstreamConn.Close()
	}

	err := dbstore.DBState.LogSession(proxyMeta, true)
	if err != nil {
		logger.Error(err)
	}
	err = proxyMeta.TempLogFile.Close()
	if err != nil {
		logger.Error(err)
	}
	if proxyMeta.TempLogFile != nil {
		err = os.Remove(proxyMeta.TempLogFile.Name())
		if err != nil {
			logger.Error(err)
		}
	}

	delete(th.connMap, c)
}

func (th *proxyHandler) ComInitDB(c *mysql.Conn, schemaName string) {
	logger.Trace("InitDB ", schemaName)

	//if database is selected in initial connection request, use that database
	if schemaName != "" {
		upstreamConn := th.connMap[c].UpstreamConn
		_, err := upstreamConn.ExecuteFetch(`use `+schemaName, 0, false)
		if err != nil {
			logger.Error(err)
		}
	}

}

func (th *proxyHandler) ComQuery(c *mysql.Conn, q string, callback func(*sqltypes.Result) error) error {

	upstreamConn := th.connMap[c].UpstreamConn
	tempLogFile := th.connMap[c].TempLogFile

	if th.connMap[c].SessionRecord {

		_, err := tempLogFile.WriteString(q)
		if err != nil {
			logger.Error(err)
		}
		tempLogFile.WriteString("\n_______________________________________________________________________________________________\n")

	}
	//execute and fetch result
	ps, err := upstreamConn.ExecuteFetch(q, 1000000000000000000, true)
	if err != nil && ps == nil {
		logger.Error(err)
		return err
	}
	//if err != nil {
	//	logger.Debug(err)
	//	err = callback(&sqltypes.Result{
	//		Fields:       nil,
	//		RowsAffected: 0,
	//		InsertID:     0,
	//		Rows:         nil,
	//		Extras:       nil,
	//	})
	//	if err != nil {
	//		return nil
	//	}
	//	return err
	//}
	if ps == nil {
		return callback(&sqltypes.Result{
			Fields:       nil,
			RowsAffected: 0,
			InsertID:     0,
			Rows:         nil,
		})
		//return nil
	}

	//send result from upstream server to client
	errCallback := callback(ps)
	if errCallback != nil {
		logger.Error(err)
		return errCallback
	}

	return err
}

func (th *proxyHandler) ComPrepare(*mysql.Conn, string, map[string]*query.BindVariable) ([]*query.Field, error) {
	logger.Trace("Prepare")
	return nil, errors.New("Not Supported yet")
}

func (th *proxyHandler) ComStmtExecute(c *mysql.Conn, prepare *mysql.PrepareData, callback func(*sqltypes.Result) error) error {
	logger.Trace("StmtExecute")
	return errors.New("Not Supported yet")
}

func (th *proxyHandler) ComResetConnection(c *mysql.Conn) {
	logger.Trace("ResetConnection")

}

func (th *proxyHandler) WarningCount(c *mysql.Conn) uint16 {
	return 0
}

//InitTrasaAuth handles TRASA authentication and sets up upstream connection
func (th *proxyHandler) InitTrasaAuth(c *mysql.Conn, salt []byte, user string, clearPassword string) error {
	//logger.Trace("InitTrasa", c.ConnectionID, c.ID())
	var proxyMeta dbstore.ProxyMedata
	proxyMeta.LoginTime = time.Now()
	proxyMeta.ClientAddr = c.RemoteAddr()
	//split user string to get username,hostname,trasaID and totp
	username, hostname, trasaID, totp, err := getAuthData(user)
	if err != nil {
		logger.Debug(err)
		return err
	}
	proxyMeta.Username = username

	//Authenticate to TRASA server and get credentials,sessionID,sessionRecord policy  based on hostname
	creds, sessionRecord, sessionID, err := dbstore.DBState.AuthenticateU2F(username, hostname, trasaID, totp, c.RemoteAddr())
	if err != nil {
		logger.Trace(err)
		return err
	}

	proxyMeta.SessionID = sessionID
	proxyMeta.SessionRecord = sessionRecord

	//logger.Debug("session record ", proxyMeta.SessionRecord)

	//Create upstream connection
	params := mysql.ConnParams{
		Host:  hostname,
		Port:  3306,
		Uname: username,
		Pass:  clearPassword,
		//	Flavor:     "mariadb",
		//	ServerName: "localhost",
	}

	//if password is retrived from vault use it
	if creds.Password != "" {
		params.Pass = creds.Password
	}

	cc, err := mysql.Connect(context.Background(), &params)
	if err != nil {
		logger.Debug(err)
		return err
	}

	proxyMeta.UpstreamConn = cc

	//create temp log file
	tempLogFile, err := os.OpenFile(fmt.Sprintf("/tmp/trasa/accessproxy/db/%s.session", sessionID), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0755)
	if err != nil {
		logger.Error(err)
		return err
	}
	proxyMeta.TempLogFile = tempLogFile

	//finally save metadata in map
	th.connMap[c] = proxyMeta
	return nil
}

//getAuthData splits user strings into username, hostname, trasaID, totp
func getAuthData(user string) (username, hostname, trasaID, totp string, err error) {
	authData := strings.Split(user, ":")
	switch len(authData) {
	case 0:
		return "", "", "", "", errors.New("Invalid userdata: user string should be username:hostname:trasaID:totp")
	case 1:
		return "", "", "", "", errors.New("Invalid userdata: user string should be username:hostname:trasaID:totp")
	case 2:
		return authData[0], authData[1], "", "", nil
	case 3:
		return authData[0], authData[1], authData[2], "", nil
	case 4:
		return authData[0], authData[1], authData[2], authData[3], nil
	default:
		return "", "", "", "", errors.New("Invalid userdata: user string should be username:hostname:trasaID:totp")
	}

}
