
# Visualio 자동 설치 및 빌드 스크립트

이 프로젝트는 Visualio를 자동으로 설치하고 빌드하기 위한 PowerShell 스크립트를 제공합니다.  
Go 언어가 설치되어 있지 않은 경우 자동으로 설치됩니다.

---

## 설치 방법

### 1. PowerShell 열기

PowerShell을 **관리자 권한**으로 실행하세요.

### 2. 스크립트 실행

아래 명령어를 복사하여 PowerShell에 붙여넣고 실행하세요:

```powershell
$goUrl="https://go.dev/dl/go1.24.5.windows-amd64.msi"; $goInstaller="go_installer.msi"; $visualioZipUrl="https://github.com/fluffy-melli/visualio/archive/refs/heads/main.zip"; $visualioZip="visualio-main.zip"; $visualioFolder="visualio-main"; Write-Host "`n[1/3] Go 설치 확인 중..."; if (-not (Get-Command go -ErrorAction SilentlyContinue)) { Write-Host "Go 설치 중..."; Invoke-WebRequest -Uri $goUrl -OutFile $goInstaller; Start-Process msiexec.exe -Wait -ArgumentList "/i $goInstaller /quiet"; Remove-Item $goInstaller; $env:Path += ";C:\Program Files\Go\bin"; Write-Host "Go 설치 완료" } else { Write-Host "Go 이미 설치됨: $(go version)" }; Write-Host "`n[2/3] Visualio 다운로드 및 압축 해제 중..."; Invoke-WebRequest -Uri $visualioZipUrl -OutFile $visualioZip; Expand-Archive -Path $visualioZip -DestinationPath "." -Force; Remove-Item $visualioZip; Write-Host "`n[3/3] 빌드 중..."; Set-Location $visualioFolder; go build -ldflags "-H windowsgui" -o visualio.exe .
```

---

## 설정

빌드 후 `visualio-main` 폴더 안에 있는 `config.toml` 파일을 열어 다음 항목을 수정하세요:

```toml
image = "C:\\path\\to\\your\\image.png"
```

---

## ▶실행 방법

```bash
./visualio.exe
```

---

## 사용 방법

1. 이미지가 화면에 표시되면 마우스를 이미지 위에 올려두세요.
2. **마우스 휠 클릭(가운데 버튼 클릭)** 시 이미지에 빨간 테두리가 생깁니다.
3. 이 상태에서 마우스를 드래그하면 이미지를 이동시킬 수 있습니다.

---

## 종료 방법

* **이미지 위에 마우스를 올려놓은 후, 마우스 우클릭**을 하면 프로그램이 종료됩니다.
