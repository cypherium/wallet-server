#!/bin/bash
# install boot.service
#

cp boot.service  /etc/init.d/
chmod 700 /etc/init.d/boot.service

systemctl daemon-reload
systemctl enable /etc/init.d/boot.service