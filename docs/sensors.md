# Neyi Nasıl Takip Eder?

Her kaynak bir olay üretir. Olay adları backend'deki [api.md](https://github.com/Horozlar-Pau/dbup-backend/blob/main/docs/api.md) ile aynıdır.

| Kaynak | Nasıl | Ürettiği olay |
|---|---|---|
| **Claude Code** | Hook'lar `dbup hook <ad>` komutunu çağırır | `ai.prompt_started` (prompt gönderildi), `ai.prompt_finished` (yanıt bitti), `ai.session_ended` (oturum kapandı) |
| **Git** | `post-commit` hook'u | `git.commit` |
| **Odak** | Editör veya terminal ön planda mı, klavye aktif mi | `coding.active`, `coding.idle` |
| **Hata sinyali** | Art arda başarısız komutlar, "fix/hata" içeren promptlar | `error.signal` |

## Bekleme penceresi
- `ai.prompt_started` geldiğinde Claude yanıt üretmeye başlamıştır. Bu, yazılımcının 30–60 saniye beklediği andır.
- Backend bu anda bekleyen dürtmeleri ajana yollar, ajan da bekleme kartını açar.
- `ai.prompt_finished` gelince kart kapanır.

## Kurallar
- Hook komutu hemen döner. Asıl iş `dbupd` içinde yapılır, yazılımcı gecikme hissetmez.
- İnternet yoksa olaylar yerel kuyrukta bekler, bağlantı gelince gönderilir.
- Takip dışı repolardan ve Gizli moddayken olay üretilmez.
