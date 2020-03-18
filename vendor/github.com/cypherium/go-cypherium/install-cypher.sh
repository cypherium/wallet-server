#!/bin/bash
# install cypher.service
#

cp cypher.service  /etc/init.d/
chmod 700 /etc/init.d/cypher.service

systemctl daemon-reload
systemctl enable /etc/init.d/cypher.service