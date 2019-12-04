package proxy

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"vitess.io/vitess/go/mysql"
)

type TrasaAuthServer struct {
	mysql.AuthServerStatic
}

func NewTrasaAuthServer() *TrasaAuthServer {
	return &TrasaAuthServer{}
}

// ValidateHash is part of the AuthServer interface.
func (tas *TrasaAuthServer) ValidateHash(salt []byte, user string, authResponse []byte, remoteAddr net.Addr) (mysql.Getter, error) {
	//fmt.Println(string(salt))
	//fmt.Println(user)
	userData := &mysql.StaticUserData{}
	userData.Get()

	fmt.Println(string(authResponse), "Validate hasgh________________________")

	fmt.Println(user)
	username, hostname, _, totp, err := getAuthData(user)
	if err != nil {
		return userData, err
	}
	_ = username
	_ = hostname
	_ = totp
	//dbstore.DBState.AuthenticateU2F(username,hostname,nil,totp,"sakjhd","asdsa",false)

	return userData, nil //errors.New("Suck it validate hash")
}

// Salt is part of the AuthServer interface.
func (tas *TrasaAuthServer) Salt() ([]byte, error) {
	return mysql.NewSalt()
}

// AuthMethod is part of the AuthServer interface.
//user string should be username:hostname:trasaID:totp
//func (tas *TrasaAuthServer) AuthMethod(user string) (string, error) {
//
//
//	return mysql.MysqlNativePassword,nil// errors.New("Suck it Auth method")
//}

//func (tas *TrasaAuthServer)Negotiate(c *mysql.Conn, user string, remoteAddr net.Addr) (mysql.Getter, error){
//	fmt.Println(c)
//	fmt.Println(user)
//	fmt.Println("Negotitaing ______")
//	return &mysql.StaticUserData{},errors.New("Suck it")
//}
//

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
