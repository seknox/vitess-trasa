package dbstore

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/minio/minio-go"
	"github.com/olivere/elastic"
	"github.com/oschwald/geoip2-golang"
	"github.com/spf13/viper"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

var DBState DBCONN

func (dbConn *DBCONN) Init() {

	//var dbConn DBCONN

	viper.SetConfigName("config")
	absPath, _ := filepath.Abs("config/")
	viper.AddConfigPath(absPath)
	fmt.Println(viper.ConfigFileUsed())

	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	dbConn.sshPort = viper.GetString("sshproxy.port")

	dbConn.trasaServer = viper.GetString("trasa.server")

	dbConn.orgId = viper.GetString("trasa.orgID")
	dbConn.appID = viper.GetString("trasa.appID")
	dbConn.appSecret = viper.GetString("trasa.appSecret")

	elasticHostName := viper.GetString("elastic.server")
	elasticPass := viper.GetString("elastic.password")
	elasticUser := viper.GetString("elastic.username")

	minioHostName := viper.GetString("minio.server")
	//minioAccessKeyID := "250PUJKB2AZ436RFO2T1"
	//minioSecretAccessKey := "QFVd5huA9OgbSTGQAk1cNan7GJAInViUmgifRefi"
	minioAccessKeyID := viper.GetString("minio.key")
	minioSecretAccessKey := viper.GetString("minio.secret")
	useSSL := viper.GetBool("minio.useSSL")

	//elasticport := viper.GetString("elastic.port")
	// commented out because our server is prroxied via nginx.	elasticport := viper.GetString("elastic.port")
	insecure := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	//_=insecure

	dbConn.elasticClient, err = elastic.NewClient(
		elastic.SetURL(elasticHostName),
		elastic.SetHttpClient(insecure),
		elastic.SetSniff(false),
		//elastic.SetSnifferTimeout(1000000000000000000),
		elastic.SetHealthcheckInterval(30*time.Second),
		elastic.SetBasicAuth(elasticUser, elasticPass),
		elastic.SetHealthcheckTimeout(time.Second*18),
		elastic.SetHealthcheck(true),
	)

	if err != nil {
		panic(err)
	}
	if _, err := dbConn.elasticClient.ElasticsearchVersion(elasticHostName); err != nil {
		panic("error")
	} else {
		//fmt.Println("Elastic version " + ver)
	}

	// Initialize minio client object.
	dbConn.minioClient, err = minio.New(minioHostName, minioAccessKeyID, minioSecretAccessKey, useSSL)
	if err != nil {
		panic(err)

	}

	exists, err := dbConn.minioClient.BucketExists("trasa-ssh-logs")
	if err != nil {
		panic(err)
	}

	if !exists {
		dbConn.minioClient.MakeBucket("trasa-ssh-logs", "")
	}

	dbConn.geoDB, err = geoip2.Open("static/GeoLite2-City.mmdb")
	if err != nil {
		panic(err)
	}

}

//sends u2f push notification and returns "success" or "failed"
func (dbConn *DBCONN) AuthenticateU2F(uname string, appID string, clientAddr net.Addr, totp, sessionID, csrfToken string, dynamicAuthApp bool) (*GuacResponse, error) {
	fmt.Println(appID, dynamicAuthApp)

	var cred AppLogin
	var trasaResp TrasaResponse
	cred.User = uname
	cred.DynamicAuthApp = dynamicAuthApp
	cred.IsSharedSession = false
	cred.TotpCode = totp
	cred.AppType = "ssh"
	if totp == "" {
		cred.TfaMethod = "u2f"
	} else {
		cred.TfaMethod = "totp"
	}
	//cred.Password = pass
	cred.AppID = appID
	//cred.AppSecret = dbConn.appSecret
	clientIP, _, err := net.SplitHostPort(clientAddr.String())

	mars, _ := json.Marshal(&cred)

	//fmt.Println(string(mars))

	url := dbConn.trasaServer + "/api/v1/remote/auth/httptunnel/" + clientIP //"http://192.168.0.100:3339/api/v1/remote/auth"
	fmt.Println(url)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(mars))
	req.Header.Set("X-SESSION", sessionID)
	req.Header.Set("X-CSRF", csrfToken)
	if err != nil {
		fmt.Printf("error sending request %s\n", err)
	}

	fmt.Printf("request sent was: %s\n", req.RequestURI)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	fmt.Printf("resp body was: %s\n", string(body))

	json.Unmarshal([]byte(body), &trasaResp)
	//fmt.Printf("status was: %s\n", result.Password)
	if trasaResp.Status == "success" {
		return &trasaResp.Data[0], nil
	} else {
		//fmt.Println(string(body))
		return nil, errors.New("failed")
	}
}

func (dbConn *DBCONN) LogSession(file *os.File, proxyMeta ProxyMedata, loginTime time.Time, success bool) (err error) {

	bucketName := "trasa-ssh-logs"
	objectNamePrefix := dbConn.orgId + "/" + strconv.Itoa(loginTime.Year()) + "/" + strconv.Itoa(int(loginTime.Month())) + "/" + strconv.Itoa(loginTime.Day()) + "/"
	if err != nil {
		objectNamePrefix = "mistake"
	}
	/*location := "us-east-1"

	err = minioClient.MakeBucket(bucketName, location)
	if err != nil {
		// Check to see if we already own this bucket (which happens if you run this twice)
		exists, err := minioClient.BucketExists(bucketName)
		if err == nil && exists {
			log.Printf("We already own %s\n", bucketName)
		} else {
			log.Fatalln(err)
		}
	}
	log.Printf("Successfully created %s\n", bucketName)
	*/

	if success && file != nil {
		// Upload the zip file
		objectName := objectNamePrefix + file.Name()
		filePath := file.Name()
		contentType :=

			"text/plain"

		// Upload log file to minio
		n, err := dbConn.minioClient.FPutObject(bucketName, objectName, filePath, minio.PutObjectOptions{ContentType: contentType})
		if err != nil {
			log.Println(err)
			return err
		}

		log.Printf("Successfully uploaded %s of size %d\n  to minio", objectName, n)

	}
	//TODO complete logging to elastic

	nep, _ := time.LoadLocation("Asia/Kathmandu")
	//ip = net.ParseIP(ip)
	locations, err := dbConn.geoDB.City(net.ParseIP(proxyMeta.ClientAddr.String()))
	//fmt.Println(conMeta.RemoteAddr(),conMeta.LocalAddr())
	if err != nil {
		fmt.Println("ip not found")
	}

	fmt.Println(locations)
	ctx := context.Background()
	eventID := make([]byte, 5)
	_, err = rand.Read(eventID)
	if err != nil {
		fmt.Println("err")
	}

	clientIP, _, err := net.SplitHostPort(proxyMeta.ClientAddr.String())
	fmt.Println(err)
	serverIP, _, err := net.SplitHostPort(proxyMeta.ClientAddr.String())
	fmt.Println(err)

	if err != nil {
		fmt.Println("Failed to parse remote ip in LogLogin sshproxy")
	}

	var log LogLogin
	log.EventID = hex.EncodeToString(eventID)
	log.Endpoint = "ssh"
	log.SessionID = proxyMeta.SessionID
	log.OrgID = dbConn.orgId
	log.UserID = proxyMeta.UserID //"330a43d-739b-489c-92a8"//conMeta.User()
	log.UserName = proxyMeta.Username
	log.Email = proxyMeta.Email
	log.AppID = proxyMeta.AppID
	log.AppName = proxyMeta.AppName
	log.UserAgent = proxyMeta.ClientVersion
	log.DeviceType = "hardcoded_DeviceType"
	log.UserIP = clientIP
	log.ServerIP = serverIP
	log.GeoLocation.IsoCountryCode = locations.Country.IsoCode
	log.GeoLocation.City = locations.City.Names["en"]
	log.GeoLocation.TimeZone = locations.Location.TimeZone

	log.GeoLocation.Location = []float64{locations.Location.Longitude, locations.Location.Latitude}
	//log.GeoLocation.Location[1]= locations.Location.Latitude
	log.Status = success
	log.LoginTime = loginTime.In(nep).Format(time.RFC3339)
	log.LogoutTime = time.Now().In(nep).Format(time.RFC3339)
	log.FailedReason = ""

	data, _ := json.Marshal(&log)
	//fmt.Println(string(data))

	putIndex, err := dbConn.elasticClient.Index().Index("orgloginsv1").Type("logins").Id(hex.EncodeToString(eventID)).BodyJson(string(data)).Do(ctx)
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println("Logged event", putIndex)
	//_ = putIndex

	return nil
}
