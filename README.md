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

## Usage

This project is to export all issues for specified repository. The exported file format looks like this: date_title_state_issue number.html. An index.html file which includes all issues will be generated before the application finishes its run.

## Configuration

`"author":""` github ID，such as "johnsonz"

`"repo":""` repesitory, such as "export-github-issues"

`"per_page":80` the maximum number is 100

`"client_id":""` there is API rate limit if empty. If you need a higher rate limit, please put in your OAuth application's client ID and secret.

`"client_secret":""` there is API rate limit if empty. If you need a higher rate limit, please put in your OAuth application's client ID and secret.
