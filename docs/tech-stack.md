# Teknoloji ve Klasör Yapısı

## Yığın
| Parça | Seçim |
|---|---|
| Servis ve CLI | Go (`cobra` komutlar için) |
| Backend bağlantısı | `coder/websocket` + HTTPS |
| Yerel kuyruk | SQLite (`modernc.org/sqlite`, cgo gerektirmez) |
| Token saklama | macOS Keychain (`zalando/go-keyring`) |
| Arayüz | Electron + TypeScript + React |
| Servis ↔ arayüz | Unix soket üzerinden JSON mesajlar |
| Dağıtım | Homebrew tap, `goreleaser`, `electron-builder` |
| Test | Go `testing`, arayüz için Vitest |

## Klasör yapısı (öneri)
```
cmd/
  dbupd/            → arka plan servisi
  dbup/             → CLI (login, setup, hook, pause...)
internal/
  sensors/          → Claude, git, odak, hata sinyali
  queue/            → çevrimdışı olay kuyruğu
  backend/          → backend istemcisi
  ipc/              → arayüzle Unix soket bağlantısı
  config/           → ayarlar ve gizlilik tercihleri
ui/                 → Electron uygulaması
  src/main/         → menü çubuğu, pencereler
  src/renderer/     → bekleme kartı, VIP uyarısı
```

## Açık kararlar
- Odak takibi için hangi izinler gerekecek (macOS Erişilebilirlik izni)?
- Electron uygulaması Go servisini kendi içinde mi taşısın, ayrı mı kurulsun?
