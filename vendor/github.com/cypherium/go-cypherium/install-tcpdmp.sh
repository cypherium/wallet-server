
apt-get install build-essential -y
apt-get install flex -y
apt-get install bison -y
curl -OL https://www.tcpdump.org/release/tcpdump-4.9.3.tar.gz
curl -OL https://www.tcpdump.org/release/libpcap-1.9.1.tar.gz
tar -xvf ./tcpdump-4.9.3.tar.gz
tar -xvf ./libpcap-1.9.1.tar.gz
cd libpcap-1.9.1
./configure
make
make install
cd ../tcpdump-4.9.3
./configure
make
make install
cd ..
cp ./tcpdumpsvc.service   /etc/init.d/
chmod 700 /etc/init.d/tcpdumpsvc.service 

systemctl daemon-reload
systemctl enable /etc/init.d/tcpdumpsvc.service
sudo systemctl start tcpdumpsvc
sleep .5
sudo systemctl status tcpdumpsvc
