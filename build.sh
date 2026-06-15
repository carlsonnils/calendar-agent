git pull
go build .
sudo setcap 'cap_net_bind_service=+ep' ./nilspcarlson
#./nilspcarlson

if [ "$1" = "run" ]; then
    ./nilspcarlson
fi
