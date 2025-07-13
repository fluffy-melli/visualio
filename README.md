
## 설치 방법

### 1. Go 설치

먼저 Go 언어를 설치해야 합니다.

* Go 다운로드: [https://go.dev/dl/](https://go.dev/dl/)

설치가 완료되면 `go` 명령어가 동작하는지 확인하세요:

```bash
go version
```

---

### 2. Visualio 다운로드 및 빌드

다음 PowerShell 명령어를 실행하여 Visualio 소스를 다운로드하고 빌드합니다:

```ps1
Invoke-WebRequest -Uri "https://github.com/fluffy-melli/visualio/archive/refs/heads/main.zip" -OutFile "visualio-main.zip"; Expand-Archive -Path "visualio-main.zip" -DestinationPath "." -Force; Remove-Item "visualio-main.zip"; Set-Location "visualio-main"; go build -ldflags "-H windowsgui" -o visualio.exe .;
```

---

## 설정

`config.toml` 파일을 열어 이미지 경로를 원하는 파일로 변경하세요:

```toml
image = "C:\\path\\to\\your\\image.png"
```

---

## 실행

```bash
./visualio.exe
```

---

## 사용 방법

1. 이미지가 화면에 표시되면, **마우스를 이미지 위에 올려두세요.**
2. **마우스 휠 클릭(가운데 버튼 클릭)** 을 하면 이미지에 **빨간 테두리**가 표시됩니다.
3. 이 상태에서 마우스를 이동시키면 **이미지를 드래그하여 위치를 변경**할 수 있습니다.

---

## 종료 방법

```ps1
taskkill /IM visualio.exe /F
```