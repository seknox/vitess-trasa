package proxy

import (
	"github.com/seknox/trasadbproxy/vitess/go/mysql"
	"net"
)

type TrasaAuthServer struct {
	Method string
	//mysql.AuthServerStatic
}

func NewTrasaAuthServer() *TrasaAuthServer {
	return &TrasaAuthServer{}
}

// ValidateHash is part of the AuthServer interface.
func (tas *TrasaAuthServer) ValidateHash(salt []byte, user string, authResponse []byte, remoteAddr net.Addr) (mysql.Getter, error) {

	userData := &mysql.StaticUserData{}

	return userData, nil //errors.New("Suck it validate hash")
}

// Salt is part of the AuthServer interface.
func (tas *TrasaAuthServer) Salt() ([]byte, error) {
	return mysql.NewSalt()
}

func (tas *TrasaAuthServer) AuthMethod(user string) (string, error) {

	return tas.Method, nil
}

func (tas *TrasaAuthServer) Negotiate(c *mysql.Conn, user string, remoteAddr net.Addr) (mysql.Getter, error) {

	password, err := mysql.AuthServerNegotiateClearOrDialog(c, tas.Method)
	_, _ = password, err
	//logger.Trace(password, err, "Nego")
	//c.ClientData=dbstore.TrasaUserData{
	//	Password:password,
	//}
	//if err != nil {
	//	return nil, err
	//}
	//
	//

	//	panic("Negotiate function is supposed to be commented. It should not be called")

	return &mysql.StaticUserData{}, nil
}
