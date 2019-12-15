package proxy

import (
	"fmt"
	"gitlab.com/seknox/trasa/trasadbproxy/vitess/go/mysql"
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

	fmt.Println(string(authResponse), "Validate hasgh________________________")

	//fmt.Println(user)
	//username, hostname, _, totp, err := getAuthData(user)
	//if err != nil {
	//	return userData, err
	//}
	//_ = username
	//_ = hostname
	//_ = totp
	////dbstore.DBState.AuthenticateU2F(username,hostname,nil,totp,"sakjhd","asdsa",false)

	return userData, nil //errors.New("Suck it validate hash")
}

// Salt is part of the AuthServer interface.
func (tas *TrasaAuthServer) Salt() ([]byte, error) {
	return mysql.NewSalt()
}

func (tas *TrasaAuthServer) AuthMethod(user string) (string, error) {

	return tas.Method, nil // errors.New("Suck it Auth method")
}

func (tas *TrasaAuthServer) Negotiate(c *mysql.Conn, user string, remoteAddr net.Addr) (mysql.Getter, error) {
	fmt.Println(c)
	fmt.Println(user)
	fmt.Println("Negotitaing ______")

	password, err := mysql.AuthServerNegotiateClearOrDialog(c, tas.Method)
	fmt.Println(password, err, "Nego")
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
