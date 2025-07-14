# Visualio 자동 설치 및 빌드 스크립트

이 프로젝트는 Visualio를 자동으로 설치하고 빌드하기 위한 PowerShell 스크립트를 제공합니다.  
Go 언어가 설치되어 있지 않은 경우 자동으로 설치됩니다.

---

## 자동 설치 방법

### 1. PowerShell 열기
PowerShell을 **관리자 권한**으로 실행하세요.

### 2. 스크립트 실행
아래 명령어를 복사하여 PowerShell에 붙여넣고 실행하세요:

```powershell
$dp="$env:USERPROFILE\Downloads"; $goUrl="https://go.dev/dl/go1.24.5.windows-amd64.msi"; $goInstaller="$dp\go_installer.msi"; $vZipUrl="https://github.com/fluffy-melli/visualio/archive/refs/heads/main.zip"; $vZip="$dp\visualio-main.zip"; $vFolder="$dp\visualio-main"; Write-Host "`n[1/3] Go 설치 확인 중..."; if (-not (Get-Command go -ErrorAction SilentlyContinue)) { Write-Host "Go 설치 중..."; Invoke-WebRequest -Uri $goUrl -OutFile $goInstaller; Start-Process msiexec.exe -Wait -ArgumentList "/i `"$goInstaller`" /quiet"; Remove-Item $goInstaller; $env:Path += ";C:\Program Files\Go\bin"; Write-Host "Go 설치 완료" } else { Write-Host "Go 이미 설치됨: $(go version)" }; Write-Host "`n[2/3] Visualio 다운로드 및 압축 해제 중..."; Invoke-WebRequest -Uri $vZipUrl -OutFile $vZip; Expand-Archive -Path $vZip -DestinationPath $dp -Force; Remove-Item $vZip; Write-Host "`n[3/3] 빌드 중..."; Set-Location $vFolder; go build -ldflags "-H windowsgui" -o visualio.exe
```

---

## 수동 설치 방법

자동 스크립트를 사용하지 않고 직접 설치하고 싶은 경우, 아래 단계를 따라 수동으로 Visualio를 설치하고 빌드할 수 있습니다.

### 1. Go 설치
1. [Go 공식 사이트](https://go.dev/dl/)에 접속합니다.
2. Windows용 설치 파일 (`go1.24.5.windows-amd64.msi` 등)을 다운로드하여 설치합니다.
3. 설치 후 PowerShell 또는 CMD에서 아래 명령어로 설치 확인:
```ps1
go version
```
출력 예시: `go1.24.5 windows/amd64`

### 2. Visualio 다운로드
1. [Visualio GitHub 저장소](https://github.com/fluffy-melli/visualio)를 방문합니다.
2. **Code > Download ZIP**을 클릭해 저장소를 압축 파일로 다운로드합니다.
3. 압축을 풀고 폴더를 원하는 위치에 저장합니다. 예: `C:\Users\사용자이름\Downloads\visualio-main`

### 3. 빌드
1. PowerShell 또는 CMD를 열고, Visualio 폴더로 이동합니다:
```ps1
cd "C:\Users\사용자이름\Downloads\visualio-main"
```
2. 아래 명령어로 빌드합니다:
```ps1
go build -ldflags "-H windowsgui" -o visualio.exe
```
`visualio.exe` 파일이 생성되면 빌드 완료입니다.

---

## 설정

빌드 후 `visualio-main` 폴더 안에 있는 `config.toml` 파일을 열어 다음 항목을 수정하세요:

```toml
image = "C:\\path\\to\\your\\image.png"
```

---

## 실행 방법

터미널에서 실행하려면:
```bash
./visualio.exe
```
또는
```bash
visualio.exe
```
파일을 더블클릭해도 실행할 수 있습니다.

---

## 사용 방법

1. 이미지가 화면에 표시되면 마우스를 이미지 위에 올려두세요.
2. **마우스 휠 클릭(가운데 버튼 클릭)** 시 이미지에 빨간 테두리가 생깁니다.
3. 이 상태에서 마우스를 드래그하면 이미지를 이동시킬 수 있습니다.

---

## 종료 방법

* **이미지 위에 마우스를 올려놓은 후, 마우스 휠 클릭 (이동상태) + 마우스 우클릭**을 하면 프로그램이 종료됩니다.

---

## 문제 해결

### 프로그램이 실행되지 않는 경우
프로그램이 실행되지 않거나 예상대로 동작하지 않는 경우, 프로그램이 설치된 폴더에서 `error.log` 파일을 확인하세요.

```bash
# Windows에서 로그 파일 확인
notepad error.log
```

또는 메모장이나 다른 텍스트 에디터로 `error.log` 파일을 열어 오류 메시지를 확인할 수 있습니다. 이 로그 파일에는 프로그램 실행 중 발생한 오류에 대한 자세한 정보가 포함되어 있습니다.

### 일반적인 문제들
- **config.toml 파일이 없거나 잘못된 경우**: 설정 파일의 경로와 형식을 확인하세요.
- **이미지 파일 경로가 잘못된 경우**: config.toml에서 설정한 이미지 파일 경로가 올바른지 확인하세요.
- **권한 문제**: 프로그램을 관리자 권한으로 실행해보세요.
