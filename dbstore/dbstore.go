package dbstore

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/seknox/trasa/server/api/auth/serviceauth"
	"github.com/seknox/trasa/server/models"
	logger "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

var DBState DBCONN

func (dbConn *DBCONN) Init() {
	os.MkdirAll("/tmp/trasa/accessproxy/db", os.ModePerm)

	viper.SetConfigName("config")
	absPath, _ := filepath.Abs("/etc/trasa/config/")
	viper.AddConfigPath(absPath)

	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	dbConn.ListenAddr = viper.GetString("proxy.dbListenAddr")
	if dbConn.ListenAddr == "" {
		dbConn.ListenAddr = ":3306"
	}

	dbConn.trasaServer = viper.GetString("trasa.listenAddr")

	dbConn.orgId = viper.GetString("trasa.orgID")

	minioHostName := viper.GetString("minio.server")
	minioStatus := viper.GetBool("minio.status")
	minioAccessKeyID := viper.GetString("minio.key")
	minioSecretAccessKey := viper.GetString("minio.secret")
	useSSL := viper.GetBool("minio.useSSL")
	dbConn.insecureSkipVerify = viper.GetBool("security.insecureSkipVerify")

	if minioStatus {
		// Initialize minio client object.
		dbConn.minioClient, err = minio.New(minioHostName, &minio.Options{
			Creds:  credentials.NewStaticV4(minioAccessKeyID, minioSecretAccessKey, ""),
			Secure: useSSL,
		})
		if err != nil {
			panic(err)

		}

		exists, err := dbConn.minioClient.BucketExists(context.Background(), "trasa-db-logs")
		if err != nil {
			panic(err)
		}

		if !exists {
			dbConn.minioClient.MakeBucket(context.Background(), "trasa-db-logs", minio.MakeBucketOptions{})
		}

	}

}

//sends u2f push notification and returns "success" or "failed"
func (dbConn *DBCONN) AuthenticateU2F(username, hostname, trasaID, totp string, clientAddr net.Addr) (upCreds *models.UpstreamCreds, sessionRecord bool, sessionID string, e error) {

	var cred serviceauth.ServiceAgentLogin
	var trasaResp struct {
		models.TrasaResponseStruct
		Data []interface{} `json:"data"`
	}
	cred.User = username
	cred.TrasaID = trasaID
	cred.TotpCode = totp
	cred.ServiceType = "db"
	if totp == "" {
		cred.TfaMethod = "u2f"
	} else {
		cred.TfaMethod = "totp"
	}
	//cred.Password = pass
	cred.Hostname = hostname
	cred.OrgID = dbConn.orgId
	clientIP, _, err := net.SplitHostPort(clientAddr.String())
	cred.UserIP = clientIP
	mars, _ := json.Marshal(&cred)

	//logger.Trace(string(mars))

	url := fmt.Sprintf("https://%s/auth/agent/db", dbConn.trasaServer) //+ clientIP //"http://192.168.0.100:3339/api/v1/remote/auth"

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(mars))
	if err != nil {
		logger.Errorf("error sending request %v", err)
		return nil, false, "", err
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: dbConn.insecureSkipVerify},
	}

	client := &http.Client{Transport: tr}
	resp, err := client.Do(req)
	if err != nil {
		logger.Error(err)
		return nil, false, "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		e = err
		return
	}

	//fmt.Printf("resp body was: %s\n", string(body))

	err = json.Unmarshal([]byte(body), &trasaResp)
	if err != nil {
		logger.Error("invalid response from trasa server")
		return nil, false, "", errors.New("Failed to authenticate 2fa")
	}

	if trasaResp.Status == "success" && len(trasaResp.Data) == 3 {

		uc, _ := json.Marshal(trasaResp.Data[0])
		e = json.Unmarshal(uc, &upCreds)
		if e != nil {
			return
		}

		sessionID = trasaResp.Data[2].(string)
		sessionRecord = trasaResp.Data[1].(bool)

		return upCreds, sessionRecord, sessionID, nil
	} else {
		//logger.Trace(string(body))
		return nil, false, "", errors.New("Failed to authenticate 2fa")
	}
}

func (dbConn *DBCONN) LogSession(proxyMeta ProxyMedata, success bool) (err error) {

	bucketName := "trasa-db-logs"
	objectNamePrefix := dbConn.orgId + "/" + strconv.Itoa(proxyMeta.LoginTime.Year()) + "/" + strconv.Itoa(int(proxyMeta.LoginTime.Month())) + "/" + strconv.Itoa(proxyMeta.LoginTime.Day()) + "/"
	if success && proxyMeta.TempLogFile != nil {
		// Upload the zip file
		objectName := objectNamePrefix + filepath.Base(proxyMeta.TempLogFile.Name())
		filePath := proxyMeta.TempLogFile.Name()

		err := dbConn.PutIntoMinio(objectName, filePath, bucketName)
		if err != nil {
			logger.Error(err)
			return err
		}
	}

	return nil

}

func (dbConn *DBCONN) PutIntoMinio(objectName, logfilepath, bucketName string) error {
	// Upload log file to minio
	minioStatus := viper.GetBool("minio.status")

	if minioStatus {
		_, err := dbConn.minioClient.FPutObject(context.Background(), bucketName, objectName, logfilepath, minio.PutObjectOptions{})
		return err
	}
	newpath := filepath.Join("/var", "trasa", "sessions", bucketName, objectName)
	dir, _ := filepath.Split(newpath)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return err
	}
	return moveFile(logfilepath, newpath)

}

func moveFile(sourcePath, destPath string) error {
	inputFile, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("Couldn't open source file: %s", err)
	}
	outputFile, err := os.Create(destPath)
	if err != nil {
		inputFile.Close()
		return fmt.Errorf("Couldn't open dest file: %s", err)
	}
	defer outputFile.Close()
	_, err = io.Copy(outputFile, inputFile)
	inputFile.Close()
	if err != nil {
		return fmt.Errorf("Writing to output file failed: %s", err)
	}
	// The copy was successful, so now delete the original file
	err = os.Remove(sourcePath)
	if err != nil {
		return fmt.Errorf("Failed removing original file: %s", err)
	}
	return nil
}
