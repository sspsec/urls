package main

import (
	"urls/common"
	"urls/source"
)

func main() {
	common.Flag()
	//print(common.GetConfigFilePath())
	defer common.OutputFile.Close()
	source.UrlScan()

}
