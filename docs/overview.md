# Ajan Ne Yapar?

Ajan yazılımcının bilgisayarında çalışır. Üç görevi vardır:

1. **İzler.** Claude'un ne zaman çalıştığını, commit'leri ve odak süresini yakalar.
2. **Bildirir.** Bunları olay olarak backend'e gönderir.
3. **Gösterir.** Backend'den gelen dürtmeleri doğru anda ekranda gösterir.

## İki parça
```
Claude Code hook'ları ─┐
git hook'ları ─────────┼──▶  dbupd (Go servis)  ◀──WebSocket──▶  dbup-backend
odak takibi ───────────┘          │
                                  │ yerel bağlantı
                                  ▼
                        dbup arayüzü (Electron)
                   menü çubuğu · bekleme kartı · VIP uyarısı
```

| Parça | Dil | Görevi |
|---|---|---|
| **`dbupd`** | Go | Arka planda sürekli çalışır. Olayları toplar, backend'le konuşur. Ayrıca hook'ların çağırdığı `dbup` komutunu da içerir. |
| **Arayüz** | Electron | Menü çubuğu ikonu, bekleme kartı ve VIP uyarısı. Kendi başına karar vermez, sadece `dbupd`'nin söylediğini gösterir. |

Arayüz kapalı olsa bile `dbupd` çalışmaya ve durumu göndermeye devam eder.

## İlkeler
- **Hızlı ve görünmez:** Hook'lar yazılımcının işini asla yavaşlatmaz. Olay gönderilemezse sessizce sıraya alınır.
- **Gizlilik önce:** Kod ve prompt metni bilgisayardan çıkmaz (bkz. [privacy.md](privacy.md)).
- **Başta yalnızca macOS:** Windows ve Linux sonra eklenecek.
