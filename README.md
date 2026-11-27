# dbup-agent
DoNotBreakUp (dbup) — yazılımcının bilgisayarında çalışan, odak/AI/git aktivitesini takip eden ajan

İki parçadan oluşur: Go ile yazılan arka plan servisi (`dbupd`) ve Electron ile yazılan masaüstü arayüzü.

## Dokümantasyon
Nasıl çalıştığını anlamak için [docs/](docs/README.md) klasörüne bakın.

## Geliştirme
Gereken: Go 1.27+

```bash
make build    # bin/dbupd ve bin/dbup
make test
make run      # dbupd'yi başlatır
./bin/dbup status
```

Ayarlar `~/.dbup/config.json` dosyasından okunur (`DBUP_HOME` ile değiştirilebilir). Dosya yoksa varsayılanlar kullanılır:

```json
{ "backend_url": "http://localhost:8080", "token": "" }
```
