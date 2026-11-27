# Git Workflow

> Bu dosya [Horozlar-Pau/.github/GIT-WORKFLOW.md](https://github.com/Horozlar-Pau/.github/blob/main/GIT-WORKFLOW.md) dosyasının kopyasıdır. Değişiklik önce orada yapılır.

## 0. Linear'da issue
Her iş [Linear](https://linear.app/donotbreakup)'da **DBUP** takımında bir issue ile başlar (ör. `DBUP-12`). Issue yoksa önce issue açılır.

Issue nasıl yazılır:

| Alan | Kural |
|---|---|
| **Başlık** | `<Alan>: <ne yapılacak>`. Alan `Web`, `Agent` veya `Backend` olur. Birden fazla alan varsa `/` ile ayrılır: `Web/Backend: ...` |
| **Etiket** | Alanla aynı: `web`, `agent`, `backend` (birden fazla olabilir) |
| **Açıklama** | Madde madde, ne yapılacağı. Kısa ve sade; dosya yolu veya gereksiz teknik detay yok |

Örnek:
```
Başlık:   Backend: Dürtme kuyruğunu ekle
Etiket:   backend
Açıklama:
- Gelen dürtmeleri sıraya al
- Yazılımcı odaktayken dürtmeleri beklet
- Commit atılınca veya Claude yanıt üretmeye başlayınca bekleyenleri gönder
- Gönderen kişiye "sırada" veya "iletildi" bilgisini göster
```

Linear, issue'yu branch ve PR'a numarasından bağlar ve durumunu kendisi ilerletir:

| Ne olunca | Issue durumu |
|---|---|
| Branch açılır veya PR açılır | **In Progress** |
| PR merge edilir | **Done** |

## 1. Branch aç
`main`'e doğrudan commit atılmaz. Her iş için `main`'den yeni bir branch açılır.

Format: `<ön ek>/DBUP-<no>-<kısa-ad>`

| Ön ek | Ne için | Örnek |
|---|---|---|
| `feat/` | Yeni özellik | `feat/DBUP-12-nudge-buttons` |
| `fix/` | Hata düzeltme | `fix/DBUP-15-ws-reconnect` |
| `docs/` | Doküman | `docs/DBUP-3-api-update` |
| `chore/` | Ayar, bağımlılık, temizlik | `chore/DBUP-7-eslint` |

Not: Linear'ın `Cmd + Shift + .` ile kopyaladığı branch adı bu formatta değildir. Ön eki ve kısa adı elle yazın.

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
Başlık: `DBUP-<no>: <kısa açıklama>`, ör. `DBUP-12: add coffee nudge button`

PR şablonu otomatik gelir. Sırasıyla doldurulur:

1. **Özet:** madde madde ne değişti
2. **Değişen dosyalar:** dosya yolları
3. **Riskler:** neyi bozabilir
4. **Notlar:** inceleyenin bilmesi gerekenler
5. **Linear:** `Fixes DBUP-12` (birden fazla issue varsa her biri ayrı satırda)

## 4. İnceleme ve merge
- En az 1 kişi onaylamadan merge edilmez.
- Merge yöntemi: **Squash and merge**. PR başlığı commit mesajı olur, böylece issue numarası `main` geçmişinde de görünür.
- Merge sonrası branch silinir.

## Kurulum (bir kez)
- Linear yöneticisi: **Settings → Integrations → GitHub** bölümünden `Horozlar-Pau` org'unu ve dbup repolarını bağlar.
- Her ekip üyesi: Linear'da **Settings → Account → Connected accounts** bölümünden kendi GitHub hesabını bağlar.
