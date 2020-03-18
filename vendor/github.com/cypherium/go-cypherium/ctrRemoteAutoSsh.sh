#!/bin/bash
localSshPath=/var/root/.ssh/
localPubFile=/var/root/.ssh/id_rsa.pub
localPrivateFile=/var/root/.ssh/id_rsa
localAuthKeys=/var/root/.ssh/id_rsa.pub
userPath=/root
userName="root"
password="1"

rm -f ./authorized_keys; touch ./authorized_keys
cat ~/.ssh/id_dsa.pub >> ~/.ssh/authorized_keys
[ ! -f $localPubFile ] && ssh-keygen -t rsa -p '' &>/dev/null  # 密钥对不存在则创建密钥
for i in $(cat ./hostname.txt)
do
    ip=$(echo "${i}" |awk -F":" '{print $1}')
    echo "ip:$ip,userName:$userName,password:$password"
#    ssh-copy-id -i $localPubFile $userName@$ip
#
##expect -c"
##         spawn ssh-copy-id -i $localPubFile $userName@$ip
##         expect {
##                "yes/no" { send "yes\n";exp_continue}     # expect 实现自动输入密码
##                "password" { send "$password\n"}
##        }
###expect eof
###EOF
##"
#     scp $localSshPath/* $userName@$ip:~/.ssh/
#     scp $localAuthKeys $userName@$ip:~/.ssh/authorized_keys
    scp ./sshd_config $userName@$ip:/etc/ssh/

     ssh  $userName@$ip "/etc/init.d/ssh restart"

done

