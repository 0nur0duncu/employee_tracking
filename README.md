# İş Takip Sistemi

Modern ve profesyonel bir iş takip sistemi. Personel ve yönetici panelleri ile iş süreçlerini kolayca takip edin.

## Özellikler

- Personel Paneli:
  - İş girişi (Yazılım/Video)
  - Video linki ekleme
  - İlk video/resize seçimi
  - Başlangıç ve bitiş zamanı takibi
  - Devam eden işleri görüntüleme ve tamamlama

- Yönetici Paneli:
  - Tüm işlerin listesi
  - İş durumu takibi
  - İstatistikler
  - Video işleri için ortalama süre hesaplama

## Gereksinimler

- Go 1.21 veya üzeri
- MongoDB
- Node.js ve npm (geliştirme için)

## Kurulum

1. Depoyu klonlayın:
```bash
git clone [repo-url]
cd is-takip-sistemi
```

2. Go bağımlılıklarını yükleyin:
```bash
go mod download
```

3. MongoDB'yi başlatın:
```bash
# MongoDB'nin çalıştığından emin olun
mongod
```

4. `.env` dosyasını düzenleyin:
```env
MONGODB_URI=mongodb://localhost:27017
DB_NAME=work_tracking_db
PORT=8080
```

5. Uygulamayı başlatın:
```bash
go run main.go
```

6. Tarayıcınızda `http://localhost:8080` adresine gidin

## Kullanım

### Personel Paneli

1. Ana sayfadan "Personel Girişi" butonuna tıklayın
2. Yeni iş girişi formunu doldurun:
   - Adınızı girin
   - İş türünü seçin (Yazılım/Video)
   - Video işi ise link ve türünü belirtin
   - Başlangıç zamanını seçin
3. İşi başlatın
4. İş bittiğinde "İşi Tamamla" butonuna tıklayıp bitiş zamanını girin

### Yönetici Paneli

1. Ana sayfadan "Yönetici Girişi" butonuna tıklayın
2. Tüm işleri görüntüleyin:
   - İstatistikleri inceleyin
   - Filtreleme butonlarını kullanın
   - Video işleri için ortalama süreleri görün

## Güvenlik

- Hassas bilgiler için .env dosyasını kullanın
- .env dosyasını asla git deposuna eklemeyin
- Üretim ortamında güvenli bir MongoDB bağlantısı kullanın

## Lisans

Bu proje MIT lisansı altında lisanslanmıştır. 