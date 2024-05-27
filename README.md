# urls

## 简介

urls是一个用于快速检查URL可用性并获取相关信息的工具。它可以快速测试多个URL，并输出每个URL的标题、状态码等信息，快速探测资产。

## 功能特点

* 支持批量测试URL。
* 输出每个URL的标题、状态码等信息。
* 支持从fofa提取资产，进行一键探活。
* 基于GO高并发快速测试大量URL。

## 使用方法

修改config.yaml文件，填入fofa信息

从文件中读取资产，进行探活
```bash
urls -f url.txt
```

从fofa中提取资产，进行探活
```bash
urls -ffq domain="xxx.edu.cn" -p 1
```

 
