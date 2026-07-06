set SERVER=root@takemeto.icu
set APP=pages

go mod tidy
set GOOS=linux
go build -ldflags="-s -w" -o app.out
set GOOS=windows

ssh %SERVER% "mkdir /apps 2>/dev/null"
ssh %SERVER% "mkdir /apps/%APP% 2>/dev/null"

ssh %SERVER% "rm -f /apps/%APP%/app.out 2>/dev/null"
scp .\app.out %SERVER%:/apps/%APP%/app.out
ssh %SERVER% "chmod +x /apps/%APP%/app.out"

ssh %SERVER% "kill $(ps aux | grep /apps/%APP%/app.out | grep -v grep | awk '{print $2}') 2>/dev/null"
ssh %SERVER% "cd /apps/%APP% && nohup /apps/%APP%/app.out > /dev/null 2>&1 &"
