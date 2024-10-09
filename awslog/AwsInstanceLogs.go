package awslog

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/rds"
)

func AwsInstanceLogs(_region string, _accessKeyId string, _accessKeySecret string, _dbInstanceIdentifier string, _slowquerylog string) {
	// 设置 AK 和 SK
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(_accessKeyId, _accessKeySecret, "")),
		config.WithRegion(_region),
	)
	if err != nil {
		fmt.Println("Error loading AWS config:", err)
		return
	}

	client := rds.NewFromConfig(cfg)

	// 获取慢查询日志文件名
	input := &rds.DescribeDBLogFilesInput{
		DBInstanceIdentifier: aws.String(_dbInstanceIdentifier),
		FilenameContains:     aws.String(_slowquerylog),
	}

	resp, err := client.DescribeDBLogFiles(context.TODO(), input)
	if err != nil {
		fmt.Println("Error describing DB log files:", err)
		return
	}

	for _, logFile := range resp.DescribeDBLogFiles {
		fmt.Println("Log File Name:", *logFile.LogFileName)

		// 下载日志文件内容
		downloadInput := &rds.DownloadDBLogFilePortionInput{
			DBInstanceIdentifier: aws.String(_dbInstanceIdentifier),
			LogFileName:          logFile.LogFileName,
			Marker:               nil,
			NumberOfLines:        aws.Int32(1000), // 可根据需要调整获取行数
		}

		downloadResp, err := client.DownloadDBLogFilePortion(context.TODO(), downloadInput)
		if err != nil {
			fmt.Println("Error downloading log file portion:", err)
			continue
		}

		fmt.Println("Log File Content:")
		fmt.Println(*downloadResp.LogFileData)
		// 指定保存路径和文件名
		filepath := "./" + *logFile.LogFileName + ".txt"

		// 创建保存路径（如果不存在）
		os.MkdirAll(filepath[:strings.LastIndex(filepath, "/")], 0755)

		// 将日志文件内容保存到指定路径的文件中
		ioutil.WriteFile(filepath, []byte(*downloadResp.LogFileData), 0644)

		fmt.Printf("Downloaded log file content saved to %s\n", filepath)

		fmt.Println("----------------------------------")
	}
}
