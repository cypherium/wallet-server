    
#!/usr/bin/env bash
systemctl stop wallet-server.service
systemctl status wallet-server.service -l
./load.sh stop
