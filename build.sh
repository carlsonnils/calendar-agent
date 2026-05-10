# go env -w CGO_ENABLED=1
# export XAI_API_KEY="testapikey"
go build .
sudo setcap 'cap_net_bind_service=+ep' ./calendar
