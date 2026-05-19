go env -w CGO_ENABLED=1
go build .
sudo setcap 'cap_net_bind_service=+ep' ./nilspcarlson
