    
#!/usr/bin/env bash
rm -rf /out.log
systemctl stop wallet-server.service
systemctl status wallet-server.service -l
./load.sh stop
