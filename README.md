# export-github-issues

# [![Build Status](https://travis-ci.org/johnsonz/export-github-issues.svg?branch=master)](https://travis-ci.org/johnsonz/export-github-issues) [![GPLv3 License](https://img.shields.io/badge/license-GPLv3-blue.svg)](https://github.com/johnsonz/export-github-issues/blob/master/LICENS)

## 说明

用于导出github特定仓库的所有issues，导出格式为date_title_state_number.html。最后会生成一个index.html文件用于索引。

这里使用了github的REST API v3，是有次数限制的，可自行申请client_id和client_secret以增大API调用次数。

## 配置说明

`"author":""` github ID，如johnsonz

`"repo":""` 仓库名称，如export-github-issues

`"per_page":80` 最大不能超过100

`"client_id":""` 不填有API次数限制，"Settings"->"OAuth applications"中生成的Client ID

`"client_secret":""` 不填有API次数限制，"Settings"->"OAuth applications"中生成的Client Secret
