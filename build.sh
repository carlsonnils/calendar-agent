
if [ "$1" = "git" ]; then 
    git pull
fi

go build .

sudo setcap 'cap_net_bind_service=+ep' ./nilspcarlson

if [ "$1" = "run" ]; then
    ./nilspcarlson
fi
