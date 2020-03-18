#!/usr/bin/env bash
systemctl stop cypher.service
systemctl stop boot.service
systemctl stop tcpdumpsvc
systemctl status cypher.service -l
systemctl status boot.service -l
systemctl status tcpdumpsvc -l