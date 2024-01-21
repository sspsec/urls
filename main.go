package main

import (
	"urls/common"
	"urls/source"
)

func main() {
	var FofaInfo common.FofaQuery
	common.Flag(&FofaInfo)
	source.Fofa(&FofaInfo)
	//print(common.GetConfigFilePath())

	source.UrlScan()

}
