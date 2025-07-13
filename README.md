
```
https://go.dev/dl/
```

```ps1
Invoke-WebRequest -Uri "https://github.com/fluffy-melli/visualio/archive/refs/heads/main.zip" -OutFile "visualio-main.zip"; Expand-Archive -Path "visualio-main.zip" -DestinationPath "." -Force; Remove-Item "visualio-main.zip"; Set-Location "visualio-main"; go build -o visualio.exe .;
```