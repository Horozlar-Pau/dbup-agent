# UI Tasarımı (Ekran Ekran)

Kaynak: [dbup Masaüstü Uygulaması — UI](https://claude.ai/artifact/2wRUKAt5aJVESNXH76pcF4) tasarımı.
Bu dosya o tasarımın yazılı hâlidir. Genel özet için [ui.md](ui.md)'ye bak.

> **Önemli fark:** Tasarımda yakın çevre kişileri şifreyle giriyor. Bu yanlış.
> Giriş **davet linkiyle** olur. Bu yüzden şifre alanları ve "Şifre değiştir" diyaloğu **yapılmaz**.
> Aşağıda bu kısımlar ~~üstü çizili~~ ve "Yapılmaz" notuyla gösterildi.

Tasarım iki gruptan oluşur:

1. **Yönetim penceresi:** Kenar çubuklu normal pencere. Durum, Yakın çevre, Dürtmeler, Ayarlar, Teşhis ve Kurulum sihirbazı.
2. **Anlık kartlar ve diyaloglar:** Menü çubuğu, Bekleme kartı, VIP uyarısı.

Pencere boyutu: 1120 × 720.

---

## Görsel dil

### Yazı tipi
- Arayüz: `Inter, system-ui, sans-serif`
- Log ve kod: `'JetBrains Mono', ui-monospace, monospace`

### Temel renkler
| Kullanım | Renk |
|---|---|
| Ana metin, koyu buton, seçili çip | `#1C1B19` |
| İkincil metin | `#5F5D57` |
| Kenar çubuğu metni | `#3D3B37` |
| Kart arka planı | `#FFFFFF` |
| Pencere / kenar çubuğu arka planı | `#F7F7F5`, `#F4F2EE` |
| Kenarlık | `#E4E2DC`, `#D9D6CF` |
| Kapalı anahtar (switch) | `#CFCBC2` |
| Gri etiket arka planı | `#EFEDE8` |

### Durum renkleri
| Durum | Renk |
|---|---|
| AI ile çalışıyor (Otomatik) | `#6E44C9` (mor) |
| Müsait | `#1E7F46` (yeşil) |
| Derin odak | `#2F5BEA` (mavi) |
| Rahatsız etme | `#B3261E` (kırmızı) |
| Gizli mod | `#5F5D57` (gri) |

### Etiketler (pill)
- Açık / Cevaplandı: arka plan `#DDF1E4`, yazı `#165E34`
- VIP açık: arka plan `#FBE3E1`, yazı `#8E1F18`
- Kapalı / İletildi: arka plan `#EFEDE8`, yazı `#5F5D57`
- Şekil: `font-size 12px`, `font-weight 600`, `padding 3px 9px`, tam yuvarlak.

### Kişi avatarları
- Yuvarlak, içinde ismin baş harfi, beyaz kalın yazı.
- Örnek renkler: Elif `#C2410C`, Can `#2F5BEA`, Annem `#1E7F46`, Mert `#6E44C9`.

### Genel ölçüler
- Kart köşesi 12px, buton köşesi 8px, çipler tam yuvarlak.
- Seçili kart: 2px durum renginde kenarlık. Seçili değil: 1px `#E4E2DC`.
- Hafif gölge: `0 1px 2px rgba(0,0,0,0.06–0.08)`.

---

## 1. Yönetim penceresi

### 1.0 Kenar çubuğu (bileşen)
Genişlik 232px. Her ekranın solunda durur.

- Üstte logo ("db") ve kullanıcı: **yazilimci**, altında anlık durum: "AI ile çalışıyor".
- Menü:
  - Durum
  - Yakın çevre (rozet: `4`)
  - Dürtmeler (rozet: `2`)
  - Ayarlar
  - Teşhis
- Seçili öğe: beyaz arka plan, kalın yazı, hafif gölge.
- Altta bağlantı bilgisi: "Sunucuya bağlı · dbup.ornek.com".

### 1.1 Durum
Başlık: **Durum**. Sağ üstte iki parçalı seçim: **Otomatik | Elle**.

**Büyük durum kartı**
- Üst satır: "Yakın çevren şu an bunu görüyor".
- 56px renkli daire (durum rengi, etrafında açık hâle).
- Durum adı ve alt metni. Sağda "Süre".
- Duruma göre içerik:
  - Otomatik: "AI ile çalışıyor" · "Claude yanıt üretiyor · Otomatik: ajan belirliyor" · "42 dk'dır"
  - Elle: seçilen durum · "Elle · 1 saat sonra Otomatik moda döner" · "az önce"
  - Elle + süresiz: "Elle · sen kapatana kadar"
  - Gizli mod: "Gizli mod" · "Takip durdu · yakın çevre son durumu görür" · "—"

**Elle durum seç** ("Seçince Elle moda geçilir")
Üç kart yan yana:
- Müsait: "Dürtmeler hemen iletilir."
- Derin odak: "Dürtmeler bekleme penceresine kadar bekler."
- Rahatsız etme: "Süre bitene kadar her şey bekler. VIP yine gelir."

**Ne kadar süre?** Çipler: `30 dk` · `1 saat` (varsayılan) · `Akşama kadar` (19:00) · `Ben kapatana kadar`.

**Gizli mod** anahtarı: "Takip tamamen durur. Yakın çevre "Gizli" yerine son durumu ve saatini görür."

**Bekleyen dürtmeler** kartı: "2 sırada".
Örnek: Elif: "Kahve molası?" · Can: "Müsait olunca ara". "Claude yanıt üretmeye başlayınca gösterilecek."
Buton: **Dürtmeleri gör**.

**Bugün** özeti: AI ile çalışma 3 sa 10 dk · Mola 25 dk · Cevaplanan dürtme 4.

### 1.2 Yakın çevre
Başlık: **Yakın çevre**.

Tablo sütunları: **Kişi · Son giriş · VIP · Erişim · İşlemler**
| Kişi | Son giriş | VIP | Erişim |
|---|---|---|---|
| Elif (@elif) | 5 dk önce | Açık | Açık |
| Can (@can) | Dün 21:40 | Kapalı | Açık |
| Annem (@anne) | 3 gün önce | Açık | Açık |
| Mert (@mert) | Hiç girmedi | Kapalı | Kapalı |

- İşlemler: Erişim **Kapat / Aç**. ~~Şifre~~ butonu **yapılmaz**.
- Alt not: "Erişim kapanınca kişinin açık web oturumları kapanır. Silinen kişinin dürtme geçmişi de silinir."

**Kişi ekle** formu:
- Görünen ad
- Kullanıcı adı ("Küçük harf, rakam ve alt çizgi")
- ~~Şifre / Güçlü şifre üret~~ → **Yapılmaz.** Yerine davet linki oluşturulur.
- "VIP uyarısı gönderebilsin" onay kutusu
- Buton: **Hesabı oluştur**

> Tasarımdaki "Bu kişiler dbup.ornek.com adresine kendi kullanıcı adı ve şifreleriyle girer." metni, davet linkine göre değiştirilmeli.

### 1.3 Dürtmeler
Başlık: **Dürtmeler**. Kişi filtresi çipleri: `Hepsi` · `Elif` · `Can` · `Annem` · `Mert`.

**Sırada** bölümü
- Her satır: avatar, "tür · gönderen", varsa not (tırnak içinde), saat, "Sırada" etiketi.
- Örnek: "Kahve molası? · Elif" — "Bugün erken çıkabilir misin?" — 14:28
- Açıklama: "Claude yanıt üretmeye başlayınca, commit atınca veya Müsait olunca bekleme kartında gösterilir."

**Geçmiş** bölümü
- Her satır: avatar, "tür · gönderen", "Cevabın: …", saat, durum etiketi.
- Durum etiketleri: **Cevaplandı** (yeşil) veya **İletildi** (gri).
- Kalp cevabı "kalp gönderdin" olarak yazılır.
- Dürtme türleri: Kahve molası?, Müsait olunca ara, Yemek yedin mi?, Akşam planı?, Moral.

### 1.4 Ayarlar
Başlık: **Ayarlar**.

**Gizlilik** (anahtarlar)
| Ayar | Açıklama | Varsayılan |
|---|---|---|
| "Ne yapıyor bu?" özeti | Commit mesajının kısaltılmış hâli esprili açıklama için gönderilir. | Açık |
| Repo adını paylaş | Kapalıyken yakın çevre yalnızca "bir projede" görür. | Kapalı |
| Çalışma süresini göster | Durum kartında "42 dk'dır" bilgisi görünür. | Açık |
| Bilgisayar açılınca başlat | Uygulama menü çubuğunda açılır. Servis her durumda çalışır. | Açık |

Anahtar: 44 × 26px. Açık `#1C1B19`, kapalı `#CFCBC2`.

**Takip dışı repolar**: liste + "+ Klasör ekle". Her satırın sağında `×`.
Örnek: `~/code/is/musteri-api`, `~/code/is/odeme-servisi`.

**Sunucu ve cihaz**
- Sunucu: https://dbup.ornek.com
- Hesap: yazilimci (sahip)
- Bu cihaz: MacBook Pro · 27 Kas 2026'da eşleşti
- Butonlar: **Sunucuyu değiştir**, **Çıkış yap**

**Hook'lar**
- Claude Code: Kurulu
- Git post-commit: Kurulu
- Butonlar: **Yeniden kur**, **Kaldır**

**dbup'ı kaldır**: "Hook'ları temizler, cihaz anahtarını Keychain'den siler ve servisi durdurur." Buton: **Kaldır**.

### 1.5 Teşhis
Başlık: **Teşhis**. Sağ üstte: **Yeniden test et**, **Log'u dışa aktar**.

Dört durum kartı:
| Kart | Değer | Alt metin |
|---|---|---|
| Arka plan servisi | Çalışıyor | dbupd 0.1.0 · 6 sa'dir açık |
| Sunucu bağlantısı | Bağlı | WSS · 38 ms gecikme |
| Kuyruktaki olay | 0 | Hepsi gönderildi |
| macOS izinleri | 1 eksik | Buton: Bildirim iznini ver |

**Son olaylar** ("Kod ve prompt metni log'a yazılmaz")
- Koyu arka planda, JetBrains Mono ile: `saat  tür  mesaj`.
- Tür renkleri: olay `#9DB6FF`, sunucu `#C7B2FF`, kart `#9FE0B6`, bağlantı `#FFC48A`, servis `#D6D3CC`.
- Örnek satırlar:
  ```
  14:32:08  olay      ai.prompt_started gönderildi
  14:32:08  sunucu    nudges.release · 2 dürtme
  14:32:09  kart      bekleme kartı açıldı
  14:21:15  olay      git.commit gönderildi (özet açık)
  14:02:44  bağlantı  bağlantı koptu, olaylar kuyruğa alındı
  14:02:50  bağlantı  yeniden bağlandı · 38 ms
  ```

### 1.6 Kurulum sihirbazı
Başlık: **dbup kurulumu** — "dbup'a hoş geldin".

Solda adım listesi (bitti: yeşil ✓, şu an: koyu, sırada: beyaz):
1. Sunucu — Adres ve bağlantı testi
2. Giriş — Sahip hesabı, cihaz anahtarı
3. Hook'lar — Claude Code ve git
4. İzinler — Erişilebilirlik, bildirimler
5. Başlangıç — Açılışta başlat

Tasarımda gösterilen adım: **Adım 2 / 5 — Sahip hesabınla giriş yap**
- "Sunucuyu kurarken oluşturduğun hesabı kullan. Bu bilgisayar bu hesaba bağlanacak."
- "https://dbup.ornek.com adresine bağlanıldı"
- Alanlar: Kullanıcı adı, Şifre
- "Şifren saklanmaz. Bu bilgisayar için bir cihaz anahtarı alınır ve macOS Keychain'e kaydedilir."
- Butonlar: **Geri**, **Giriş yap ve devam et**

> Not: Buradaki şifre, yazılımcının kendi **sahip** hesabı içindir. Yakın çevre şifresiyle ilgili kural bu adımı etkilemez.

---

## 2. Anlık kartlar ve diyaloglar

### 2.1 Menü çubuğu (tray)
Boyut 320 × 470.

- Üstte: "AI ile çalışıyor" · "42 dk'dır · Otomatik"
- "Bekleyen dürtme 2"
- **Durum** listesi (seçili olanın yanında ✓, renkli nokta):
  - Otomatik ✓
  - Müsait — 1 saat
  - Derin odak — 1 saat
  - Rahatsız etme — 1 saat
- Gizli mod: Kapalı
- Bu repoyu takip etme
- dbup'ı aç…
- Çık

### 2.2 Bekleme kartı
Boyut 380 × 420. Ekran köşesinde açılır.

- Üst: "Claude düşünürken" · "2 dürtme seni bekliyordu"
- Her dürtme: avatar, tür, "gönderen · saat", varsa not.
- Hızlı cevap butonları: **10 dk'ya ararım** · **Akşam konuşalım** · **❤**
- Cevap verilince butonlar gider, yerine "Gönderildi: …" yazar.
- Alt not: "Claude yanıtı bitince kart kendiliğinden kapanır"

### 2.3 VIP uyarısı
Tam ekran, 1120 × 720 tasarım. Koyu kırmızı arka plan (`#5C120D`).

- Üst etiket: "Gerçekten acil"
- Avatar + "Elif sana ulaşmaya çalışıyor"
- Not: "Kapıda kaldım, anahtarı unuttum. Telefonu açar mısın?"
- "14:41 · Bugünkü VIP hakkını kullandı"
- Butonlar: **Gördüm**, **Gördüm, 10 dk'ya ararım**
- Alt not: "Elif "Gördüm" bilgisini görecek. Ses, sen basana kadar çalar."

### ~~2.4 Şifre değiştir diyaloğu~~ — Yapılmaz
Tasarımda var ama yakın çevre davet linkiyle girdiği için bu diyalog yapılmayacak.
