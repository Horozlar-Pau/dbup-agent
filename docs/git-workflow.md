# Git Workflow

## 1. Branch aç
`main`'e doğrudan commit atılmaz. Her iş için `main`'den yeni bir branch açılır:

| Ön ek | Ne için | Örnek |
|---|---|---|
| `feat/` | Yeni özellik | `feat/nudge-buttons` |
| `fix/` | Hata düzeltme | `fix/ws-reconnect` |
| `docs/` | Doküman | `docs/api-update` |
| `chore/` | Ayar, bağımlılık, temizlik | `chore/eslint` |

## 2. Commit at
Format: `<tür>: <kısa açıklama>`

```
feat: add coffee nudge button
fix: reconnect websocket after sleep
```

Türler: `feat`, `fix`, `docs`, `refactor`, `test`, `chore`, `perf`, `ci`

- Bir commit tek bir iş yapar.
- Commit ve PR'larda AI/araç imzası (`Co-Authored-By`, "Generated with …") bulunmaz.

## 3. PR aç
PR şablonu otomatik gelir. Sırasıyla doldurulur:

1. **Özet:** madde madde ne değişti
2. **Değişen dosyalar:** dosya yolları
3. **Riskler:** neyi bozabilir
4. **Notlar:** inceleyenin bilmesi gerekenler

## 4. İnceleme ve merge
- En az 1 kişi onaylamadan merge edilmez.
- Merge yöntemi: **Squash and merge**. PR başlığı commit formatında olur.
- Merge sonrası branch silinir.
