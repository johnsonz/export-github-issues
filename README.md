# export-github-issues

## 说明
用于导出github特定仓库的所有issues，导出格式为date_title_state_number.html。

这里使用了github的REST API v3，是有次数限制的，可自行申请client_id和client_secret以增大API调用次数。

## 配置说明

`"author":""` github ID，如johnsonz

``"repo":""` 仓库名称，如export-github-issues

`"per_page":80` 最大不能超过100

`"client_id":""` 不填有API次数限制，"Settings"->"OAuth applications"中生成的Client ID

`"client_secret":""` 不填有API次数限制，"Settings"->"OAuth applications"中生成的Client Secret
