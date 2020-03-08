package proxy

import (
	"context"
	"errors"
	"fmt"
	uuid "github.com/satori/go.uuid"
	logger "github.com/sirupsen/logrus"
	"gitlab.com/seknox/trasa/trasadbproxy/dbstore"
	"gitlab.com/seknox/trasa/trasadbproxy/vitess/go/mysql"
	"gitlab.com/seknox/trasa/trasadbproxy/vitess/go/sqltypes"
	querypb "gitlab.com/seknox/trasa/trasadbproxy/vitess/go/vt/proto/query"
	"os"
	"strings"
	"time"
)

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

	dbstore.DBState.LogSession(proxyMeta, true)
	proxyMeta.TempLogFile.Close()
	err := os.Remove(proxyMeta.TempLogFile.Name())
	if err != nil {
		logger.Error(err)
	}
	delete(th.connMap, c)
}

func (th *proxyHandler) ComInitDB(c *mysql.Conn, schemaName string) {
	fmt.Println("________com______InitDB ", schemaName)

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

		tempLogFile.WriteString(q)
		tempLogFile.WriteString("\n_______________________________________________________________________________________________\n")

	}
	//execute and fetch result
	ps, err := upstreamConn.ExecuteFetch(q, 1000, true)
	if err != nil && ps == nil {
		logger.Error(err)
		return err
	}
	if err != nil {
		logger.Debug(err)
		err = callback(&sqltypes.Result{
			Fields:       nil,
			RowsAffected: 0,
			InsertID:     0,
			Rows:         nil,
			Extras:       nil,
		})
		if err != nil {
			return nil
		}
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

	//send result from upstream server to client
	errCallback := callback(ps)
	if errCallback != nil {
		logger.Error(err)
		return errCallback
	}

	return err
}

func (th *proxyHandler) ComPrepare(c *mysql.Conn, q string) ([]*querypb.Field, error) {
	logger.Debug("________com______Prepare")
	return nil, errors.New("Not Supported yet")
}

func (th *proxyHandler) ComStmtExecute(c *mysql.Conn, prepare *mysql.PrepareData, callback func(*sqltypes.Result) error) error {
	logger.Debug("________com______StmtExecute")
	return errors.New("Not Supported yet")
}

func (th *proxyHandler) ComResetConnection(c *mysql.Conn) {
	fmt.Println("________com______ResetConnection")

}

func (th *proxyHandler) WarningCount(c *mysql.Conn) uint16 {
	return 0
}

func (th *proxyHandler) InitTrasaAuth(c *mysql.Conn, salt []byte, user string, authResponse []byte, clearPassword string) error {
	logger.Trace("InitTrasa", c.ConnectionID, c.ID())
	var proxyMeta dbstore.ProxyMedata
	proxyMeta.LoginTime = time.Now()
	proxyMeta.ClientAddr = c.RemoteAddr()
	//split user string to get username,hostname,trasaID and totp
	username, hostname, trasaID, totp, err := getAuthData(user)
	if err != nil {
		logger.Trace(err)
		return err
	}
	proxyMeta.Username = username

	//Authenticate to TRASA server and get user/app details based on hostname
	resp, err := dbstore.DBState.AuthenticateU2F(username, hostname, trasaID, totp, c.RemoteAddr())
	if err != nil {
		logger.Trace(err)
		return err
	}
	proxyMeta.Email = resp.Email
	proxyMeta.AppName = resp.AppName
	proxyMeta.UserID = resp.UserID
	proxyMeta.SessionRecord = resp.SessionRecord

	//create session ID
	sessionID := ""
	tempuuid, err := uuid.NewV4()
	if err == nil {
		sessionID = tempuuid.String()
	} else {
		logger.Error(err)
	}
	proxyMeta.SessionID = sessionID

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
	if resp.Password != "" {
		params.Pass = resp.Password
	}

	cc, err := mysql.Connect(context.Background(), &params)
	if err != nil {
		logger.Debug(err)
		return err
	}

	proxyMeta.UpstreamConn = cc

	//create temp log file
	tempLogFile, err := os.OpenFile(sessionID+".session", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0755)
	if err != nil {
		logger.Error(err)
		return err
	}
	proxyMeta.TempLogFile = tempLogFile

	//finally save metadata in map
	th.connMap[c] = proxyMeta
	return nil
}

//splits user strings into username, hostname, trasaID, totp
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
