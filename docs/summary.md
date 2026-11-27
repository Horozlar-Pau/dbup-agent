# "Ne Yapıyor Bu?" Cümlesi

Commit atılınca ajan, commit mesajını yakın çevrenin anlayacağı tek bir cümleye çevirir. Çeviri yazılımcının bilgisayarında yapılır. Backend'e yalnızca hazır cümle gider.

## Nasıl çalışır?
1. `post-commit` hook'u commit mesajının ilk satırını `dbupd`'ye verir.
2. `dbupd`, bilgisayarda çalışan **CLIProxy**'ye istek atar. CLIProxy, yazılımcının kendi Claude hesabıyla Claude'a bağlanır.
3. Gelen cümle `git.commit` olayının `humanSummary` alanıyla backend'e gönderilir.

## CLIProxy
- [CLIProxyAPI](https://github.com/router-for-me/CLIProxyAPI) yazılımcının bilgisayarına kurulur ve Claude hesabıyla giriş yapılır (`cli-proxy-api --claude-login`).
- Ajan, CLIProxy'nin Anthropic uyumlu `/v1/messages` ucunu kullanır.
- Varsayılan adres `http://127.0.0.1:8317`. Yalnızca bu bilgisayardan erişilen bir adres kabul edilir.
- CLIProxy'nin yerel erişim anahtarı Keychain'de saklanır.
- Varsayılan model `claude-haiku-5-5`.

Not: Claude aboneliğini başka bir uygulamada kullanmak Anthropic'in kullanım koşullarına aykırı olabilir. Sorumluluk yazılımcıdadır.

## Kurallar
- Gönderilen metin: commit mesajının ilk satırı, en fazla 200 karakter. Kod ve dosya adı gönderilmez.
- Cümle en fazla 140 karakter olur. Daha uzunsa kısaltılır.
- Aynı mesaj tekrar çevrilmez (yerel önbellek).
- Çeviri 8 saniyede bitmezse ya da CLIProxy kapalıysa commit olayı cümlesiz gönderilir. Kartta durum adı görünür.
- Ayarlarda "commit özeti" kapalıysa CLIProxy'ye istek atılmaz, cümle de gönderilmez.

## Ayarlar
| Ayar | Varsayılan |
|---|---|
| CLIProxy adresi | `http://127.0.0.1:8317` |
| CLIProxy erişim anahtarı | Boş (Keychain) |
| Model | `claude-haiku-5-5` |
| Commit özeti paylaşımı | Açık |
