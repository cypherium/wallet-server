#!/usr/bin/env bash
cp -rf wallet-server.service  /etc/init.d/
chmod 700 /etc/init.d/wallet-server.service

systemctl daemon-reload
systemctl enable /etc/init.d/wallet-server.service
systemctl start wallet-server.service
systemctl status wallet-server.service -l
