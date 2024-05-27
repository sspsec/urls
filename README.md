# urls

## 简介

urls是一个用于快速检查URL可用性并获取相关信息的工具。它可以快速测试多个URL，并输出每个URL的标题、状态码等信息，快速探测资产。

## 功能特点

* 支持批量测试URL。
* 输出每个URL的标题、状态码等信息。
* 输出支持高亮显示
* 基于GO高并发快速测试大量URL。

## 使用方法

从文件中读取资产，进行探活
```bash
urls -f url.txt
```

将结果保存到result.txt文件中
```bash
urls -f url.txt -o
```
<img width="910" alt="image" src="https://github.com/sspsec/urls/assets/142762749/563aa1ec-0e92-4f5e-82e5-bc61df6476f8">

 
