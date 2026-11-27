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
{ "backend_url": "https://dbup.ardadagdeviren.com.tr", "token": "" }
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

## Commit özeti (CLIProxy)
Commit atılınca `dbupd`, commit mesajının ilk satırını bilgisayardaki CLIProxy'ye gönderir ve gelen cümleyi `git.commit` olayına ekler ([docs/summary.md](docs/summary.md)). CLIProxy kapalıysa olay cümlesiz kaydedilir.

```bash
cli-proxy-api --claude-login                                    # CLIProxy'ye Claude hesabıyla giriş
security add-generic-password -a dbup -s dbup-cliproxy -w        # CLIProxy erişim anahtarı (sorar)
```

Ayarlar `~/.dbup/config.json` içinde (hepsi isteğe bağlı):

```json
{ "summary": { "enabled": true, "cliproxy_url": "http://127.0.0.1:8317", "model": "claude-haiku-5-5" } }
```

Yalnızca bu bilgisayardaki adresler kabul edilir. Ayar değişince `dbupd` yeniden başlatılır. Durum için `dbup status`.

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

## Backend olmadan deneme
Backend bağlanana kadar yakın çevre mesajları test komutlarıyla taklit edilir:

```bash
./bin/dbup test nudge --from Elif --kind "Kahve molası?" --note "Bugün erken çıkabilir misin?"   # kart hemen açılır
./bin/dbup test vip --from Elif --note "Kapıda kaldım, telefonu açar mısın?" # tam ekran, sesli
./bin/dbup test moral --from Can --note "Kolay gelsin!"                     # konfeti
```

Cevaplar ve "Gördüm" bilgisi şimdilik yalnızca dbupd'de (bellekte) tutulur. Backend bağlanınca (DBUP-34) gönderene iletilecek.
