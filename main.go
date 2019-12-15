package main

import "gitlab.com/seknox/trasa/trasadbproxy/proxy"

func main() {

	//params := mysql.ConnParams{
	//	Host:       "hostname",
	//	Port:       3306,
	//	Uname:      "username",
	//	Pass:       "clearPassword",
	//	Flavor:     "mariadb",
	//	ServerName: "localhost",
	//
	//	//	Flavor:mysql.MysqlNativePassword,
	//
	//}
	//ctx:=context.Background()
	//fmt.Println(ctx,"__________)))))))")
	//cc, err := mysql.Connect(nil, &params)
	//fmt.Println(cc,err)

	proxy.StartListner()
}
