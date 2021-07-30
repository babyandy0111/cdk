package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

var envType string

func main() {
	flag.StringVar(&envType, "env", "", "")
	flag.StringVar(&envType, "e", "", "")
	flag.Parse()

	switch strings.ToLower(envType) {
	case "production", "staging", "development":
		if err := generateGithubJson(strings.ToLower(envType)); err != nil {
			fmt.Println(err.Error())
		}
		return
	default:
		fmt.Println("Not a valid environment: " + envType)
	}
	return
}

func generateGithubJson(envType string) error {
	shortEnvName := envType[0:4]
	originFilename := fmt.Sprintf(os.Getenv("PWD")+"/%s-cdk.json", envType)
	destFileName := fmt.Sprintf(os.Getenv("PWD")+"/%s.json", envType)
	// 讀取 CDK 匯出來的內容
	f1, err := os.Open(originFilename)
	defer f1.Close()
	if err != nil {
		return err
	}
	stat1, err := f1.Stat()
	if err != nil {
		return err
	}
	content1 := make([]byte, stat1.Size())
	f1.Read(content1)
	fmt.Println(string(content1))
	//
	j1 := make(map[string]map[string]string, 0)
	if err := json.Unmarshal(content1, &j1); err != nil {
		return err
	}
	// 產生要準備匯至github的環境變數內容
	j2 := make(map[string]string, 0)
	c := j1[shortEnvName+"-RootStack"]
	j2["AWS_ACCESS_KEY"] = c["AWSACCESSKEY"]
	j2["AWS_REGION"] = c["AWSREGION"]
	j2["AWS_SECRET_KEY"] = c["AWSSECRETKEY"]
	j2["DOCKER_PASSWORD"] = c["DOCKERPASSWORD"]
	j2["CI_USER_TOKEN"] = c["CIUSERTOKEN"]
	j2["DOCKER_USERNAME"] = c["DOCKERUSERNAME"]
	j2["ECS_APIGATEWAY_SERVICE"] = j1[shortEnvName+"-ECSStack"]["ECSAPIGATEWAYSERVICE"]
	j2["ECS_BACKEND_SERVICE"] = j1[shortEnvName+"-ECSStack"]["ECSBACKENDSERVICE"]
	j2["ECS_CLUSTER"] = j1[shortEnvName+"-ECSStack"]["ECSCLUSTERARN"]
	j2["ECS_FRONTEND_SERVICE"] = j1[shortEnvName+"-ECSStack"]["ECSFRONTENDSERVICE"]
	j2["MYSQL_HOST"] = c["MYSQLHOST"]
	j2["MYSQL_PASSWORD"] = c["MYSQLPASSWORD"]
	j2["MYSQL_USER"] = c["MYSQLUSER"]
	output, err := json.Marshal(j2)
	if err != nil {
		return err
	}
	// 將 output 寫入檔案
	f2, err := os.Create(destFileName)
	defer f2.Close()
	if err != nil {
		return err
	}
	if _, err := f2.Write(output); err != nil {
		return err
	}
	// 執行 node script
	progPath := os.Getenv("PWD") + "/tools/sodium/index.js"
	cmd := exec.Command("node", []string{progPath, envType}...)
	outputContent, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	fmt.Println(outputContent)
	return nil
}
