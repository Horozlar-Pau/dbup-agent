# Kurulum ve Eşleştirme

## Kurulum
```bash
brew install horozlar-pau/tap/dbup
```
Bu komut `dbupd` servisini ve Electron arayüzünü kurar. `dbupd`, bilgisayar açılınca kendiliğinden başlar (macOS launchd).

## Eşleştirme
```bash
dbup login
```
1. Tarayıcı açılır, yazılımcı dbup hesabına giriş yapar.
2. Backend bu bilgisayar için bir **cihaz token'ı** verir.
3. Token macOS **Keychain**'e kaydedilir, düz dosyada tutulmaz.

## Hook'ları kurma
```bash
dbup setup
```
- Claude Code ayarlarına dbup hook'larını ekler.
- Git için genel bir `post-commit` hook'u ekler (mevcut hook'ları bozmadan).
- Ne değiştireceğini önce gösterir, onay ister.

## Kaldırma
```bash
dbup uninstall
```
Eklenen hook'ları temizler, token'ı Keychain'den siler ve servisi durdurur.

## Diğer komutlar
| Komut | Ne yapar |
|---|---|
| `dbup status` | Bağlantı ve güncel durum |
| `dbup pause 1h` | Takibi bir süre duraklatır (Gizli mod) |
| `dbup ignore` | Bulunulan repoyu takip dışı bırakır |
