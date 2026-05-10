go env -w CGO_ENABLED=1
$env:XAI_API_KEY="testapikey"
go build .
