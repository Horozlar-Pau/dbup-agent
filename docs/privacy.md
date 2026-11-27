# Ne Gönderilir, Ne Gönderilmez?

| Veri | Gönderilir mi? |
|---|---|
| Kod, dosya içerikleri | ❌ Asla |
| Prompt metni | ❌ Asla. Yalnızca "fix/hata" içeriyor mu bilgisi bilgisayarda çıkarılır |
| Dosya ve klasör adları | ❌ Hayır |
| Commit mesajı | ⚙️ Yalnızca izin verilirse, kısaltılmış olarak ("Ne Yapıyor Bu?" çevirisi için) |
| Olay türü ve zamanı | ✅ Evet (ör. "Claude çalışmaya başladı, 14:32") |
| Repo adı | ⚙️ Varsayılan kapalı |

## Yazılımcının kontrolü
- **Gizli mod:** `dbup pause` veya menüden. Takip tamamen durur.
- **Takip dışı repolar:** `dbup ignore` ile iş repoları dışarıda bırakılabilir.
- **Ayarlar:** Commit mesajı ve repo adı paylaşımı tek tek açılıp kapatılır.

## Güvenlik
- Cihaz token'ı macOS Keychain'de saklanır.
- Backend'le bağlantı yalnızca HTTPS/WSS üzerinden yapılır.
- `dbupd` ile arayüz arasındaki bağlantı yalnızca o bilgisayardan erişilebilir (Unix soket).
