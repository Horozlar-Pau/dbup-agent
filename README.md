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

## Hook'ları elle kurma
`dbup setup` gelene kadar hook'lar elle eklenir. `dbup` komutunun PATH'te olması gerekir (ör. `bin/` klasörünü PATH'e ekleyin).

**Claude Code:** `~/.claude/settings.json` içine:
```json
{
  "hooks": {
    "UserPromptSubmit": [{ "hooks": [{ "type": "command", "command": "dbup hook claude" }] }],
    "Stop":             [{ "hooks": [{ "type": "command", "command": "dbup hook claude" }] }],
    "SessionEnd":       [{ "hooks": [{ "type": "command", "command": "dbup hook claude" }] }]
  }
}
```

**Git:** takip edilecek repoda `.git/hooks/post-commit` dosyasına `dbup hook git-commit` satırını ekleyin ve dosyayı çalıştırılabilir yapın.

Gelen olayları görmek için: `dbup events`. Hook hatalarını görmek için `DBUP_DEBUG=1`.

## Arayüz (Electron)
Gereken: Node 20+

```bash
cd ui
npm install
npm start     # yönetim penceresini açar
npm test
```

- Pencere `dbupd`'ye Unix soketinden bağlanır. Teşhis sayfası ve Durum kartı gerçek olayları gösterir. Diğer sayfalar backend bağlanana kadar sahte verilerle (`ui/src/renderer/data.js`) çalışır.
- `npm start` "Electron failed to install correctly" derse, indirilen arşivi elle açın:
  `rm -rf node_modules/electron/dist && ditto -x -k ~/Library/Caches/electron/*/electron-v*-darwin-arm64.zip node_modules/electron/dist && printf 'Electron.app/Contents/MacOS/Electron' > node_modules/electron/path.txt`
