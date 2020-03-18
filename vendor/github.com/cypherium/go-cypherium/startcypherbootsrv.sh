#!/usr/bin/env bash
systemctl start cypher.service
systemctl start boot.service
systemctl status cypher.service -l
systemctl status boot.service -l
