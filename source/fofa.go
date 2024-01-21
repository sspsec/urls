package source

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v3"
	"log"
	"net/http"
	"os"
	"urls/common"
)

func GetFofaConfig() (common.FofaConfig, error) {

	file, err := os.ReadFile(common.GetConfigFilePath() + "/config.yaml")
	if err != nil {
		log.Fatal(err)
	}
	var config common.FofaConfig
	err = yaml.Unmarshal(file, &config)
	if err != nil {
		log.Fatal(err)
	}

	return config, nil
}

func Fofa(FofaInfo *common.FofaQuery) {

	query := base64.StdEncoding.EncodeToString([]byte(FofaInfo.Query))
	Config, _ := GetFofaConfig()
	for i := 1; i <= FofaInfo.Page; i++ {
		resp, err := http.Get(fmt.Sprintf("%s?email=%s&key=%s&qbase64=%s&page=%d&fields=host,ip,port,title", Config.Fofa.FOFAPI, Config.Fofa.EMAIL, Config.Fofa.TOKEN, query, i))
		if err != nil {
			log.Panic(err)
		}

		defer resp.Body.Close()

		var FofaResult common.FofaResult
		if err := json.NewDecoder(resp.Body).Decode(&FofaResult); err != nil {
			fmt.Println("解析 JSON 数据时出错:", err)
			log.Panic(err)
		}

		for _, result := range FofaResult.Results {
			url := fmt.Sprintf("%s", result[0])
			fmt.Printf("%s\n", url)
			common.Urls = append(common.Urls, url)
		}
	}
}
