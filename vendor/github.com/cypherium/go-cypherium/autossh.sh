#!/usr/bin/env bash
#用户名
user='/var/root'
#密码
passwd='1'
#循环遍历服务器IP信息

for ip in $(awk -F' ' '{print $1}' ./hostname.txt); do
(
expect -c "
#返回超时设置
set timeout -1
#复制公钥到目标服务器
spawn ssh-copy-id -i /root/.ssh/id_rsa.pub $user@$ip
expect {
"*yes/no" { send "yesr"; exp_continue}
"*assword" { send "$passwd"}}
expect eof
EOF
"
)
#执行服务器远程copy
#scp /internal-hosts $user@$ip:/etc/hosts
done
