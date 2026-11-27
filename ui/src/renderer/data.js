// Fake data for the management window until dbup-backend is connected (DBUP-33, DBUP-34).
'use strict';

window.FAKE = {
  host: 'dbup.ornek.com',
  username: 'yazilimci',

  circle: [
    { id: 'elif', name: 'Elif', color: '#c0532b', lastSeen: '5 dk önce', vip: true, invite: 'active', code: 'k7Rq2vLp' },
    { id: 'can', name: 'Can', color: '#3a5bd9', lastSeen: 'Dün 21:40', vip: false, invite: 'active', code: 'w9Hs4nTe' },
    { id: 'annem', name: 'Annem', color: '#3f7f4f', lastSeen: '3 gün önce', vip: true, invite: 'active', code: 'p2Mx8cQa' },
    { id: 'mert', name: 'Mert', color: '#6a4fd6', lastSeen: 'Hiç girmedi', vip: false, invite: 'revoked', code: 'z5Bd1kYu' },
  ],

  nudges: {
    queued: [
      { from: 'elif', kind: 'Kahve molası?', note: 'Bugün erken çıkabilir misin?', at: '14:28' },
      { from: 'can', kind: 'Müsait olunca ara', note: '', at: '14:05' },
    ],
    history: [
      { from: 'annem', kind: 'Yemek yedin mi?', reply: 'Kalp gönderdin', at: '12:40', state: 'replied' },
      { from: 'elif', kind: 'Akşam planı?', reply: 'Akşam konuşalım', at: '11:15', state: 'replied' },
      { from: 'can', kind: 'Moral', reply: '', at: '10:02', state: 'delivered' },
      { from: 'elif', kind: 'Müsait olunca ara', reply: "10 dk'ya ararım", at: 'Dün 18:20', state: 'replied' },
      { from: 'annem', kind: 'Kahve molası?', reply: '', at: 'Dün 15:10', state: 'delivered' },
    ],
  },

  today: { aiMinutes: '3 sa 10 dk', breakMinutes: '25 dk', replies: 4 },

  privacy: [
    { id: 'summary', title: '"Ne yapıyor bu?" özeti', desc: 'Commit mesajı kısaltılıp esprili bir açıklamaya çevrilir', on: true },
    { id: 'repo', title: 'Repo adını paylaş', desc: 'Yakın çevre hangi projede çalıştığını görür', on: false },
    { id: 'duration', title: 'Çalışma süresini göster', desc: 'Durum kartında "42 dk\'dır" bilgisi görünür', on: true },
    { id: 'login', title: 'Bilgisayar açılınca başlat', desc: 'dbupd oturum açılınca kendiliğinden çalışır', on: true },
  ],

  ignoredRepos: ['~/code/is/odeme-servisi', '~/code/is/admin-panel'],
};
