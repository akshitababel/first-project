/*
Package data - Handles functions related to data source access e.g. cache, databases
*/
package data

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"io/ioutil"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

var ENV = strings.ToLower(os.Getenv("Environment"))

func InitS3() *s3.S3 {
	sess, err := session.NewSession()
	if err != nil {
		fmt.Println("failed to create session,", err)
		return nil
	}
	sess.Config.Region = aws.String("ap-southeast-2")

	svc := s3.New(sess)
	return svc
}

func S3PutItem(bucket, folder, key string, obj []byte) {
	svc := InitS3()

	if ENV != "production" {
		folder = "_dev/" + folder
	}

	payload := bytes.NewReader(obj)
	path := key
	if folder != "" {
		path = folder + "/" + key
	}

	params := &s3.PutObjectInput{
		Bucket:               aws.String(bucket), // Required
		Key:                  aws.String(path),   // Required
		Body:                 payload,
		ContentLength:        aws.Int64(payload.Size()),
		ContentType:          aws.String("text/html"),
		ServerSideEncryption: aws.String("AES256"),
	}
	resp, err := svc.PutObject(params)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println(resp)

}

func S3GetItem(bucket, folder, key string) ([]byte, error) {
	svc := InitS3()

	if ENV != "production" {
		folder = "_dev/" + folder
	}

	path := key
	if folder != "" {
		path = folder + "/" + key
	}
	params := &s3.GetObjectInput{
		Bucket: aws.String(bucket), // Required
		Key:    aws.String(path),   // Required

	}
	resp, err := svc.GetObject(params)

	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	fmt.Println(resp)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	return body, nil

}
