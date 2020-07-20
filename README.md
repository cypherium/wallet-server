## wallet .service

浏览器后端服务
#####build
```
sh build.sh
```
   
#####deploy
修改配置文件：./conf/scan.conf
```
#####重要参数#####
Port     = "8359"                               #服务启动的端口，作为nginx的上游，为前端提供数据接口
Redis    = "xxx:8379"                           #redis缓存服务, 预留给加速用，当前可不配
Gate     = "gateway.inner.poc.com:8545"         #poc链网关节点rpc，浏览器通过此节点获取链数据
database = "xxx:xxx@2019@tcp(xxx:8306)/scan"    #数据库，格式化链上数据，以提供快速查询
```

#####API
参见：src/main.go 和 src/api

#####database
参见：src/model/create_table.go

#####编译运行测试

1.编译./build.sh
2.运行./load.sh restart
3.测试
curl --include \
     --request POST \
     --header "Content-Type: application/json; charset=utf-8" \
     --header "Authorization: Basic MzAwNWFlNDQtM2M5MS00MGFmLWI1NzktNDg4OTNhZDkxMGVm" \
     --data-binary "{\"app_id\": \"181c8c4b-27f8-4445-97c0-1e367c4a88ca\",
\"contents\": {\"en\": \"English Message\"},
\"filters\": [{\"field\": \"tag\", \"key\": \"keyname\", \"relation\": \"=\", \"value\": \"valuestr\"},{\"operator\": \"OR\"},{\"field\": \"amount_spent\", \"relation\": \">\",\"value\": \"0\"}]}" \
     https://onesignal.com/api/v1/notifications
