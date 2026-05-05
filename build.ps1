go env -w CGO_ENABLED=1
$env:ANTHROPIC_API_KEY="testapikey"
go build .
