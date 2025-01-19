# Personel Takip Yazılımı

Bu proje, personel ve stajyerlerin günlük iş takibini yapabilmek için geliştirilmiş bir web uygulamasıdır. Yazılım ve video işlerinin zaman bazlı takibini, video incelemelerini ve revizyon süreçlerini yönetmeyi sağlar.

## Özellikler

- 👥 Personel/stajyer yönetimi
- 📝 İş tanımlama ve takibi
- 🎥 Video işleri takibi
- 💻 Yazılım işleri takibi
- ⏱️ Gerçek zamanlı iş durumu güncellemeleri
- 📊 İş istatistikleri
- 🔄 Video revizyon sistemi
- ✅ Video inceleme ve onay süreci

## Teknolojiler

- Backend: Go (Fiber framework)
- Frontend: HTML, JavaScript, Bootstrap 5
- Veritabanı: MongoDB
- Containerization: Docker

## Kurulum

### Gereksinimler

- Docker
- Docker Compose

### Kurulum Adımları

1. Projeyi klonlayın:
   ```bash
   git clone [repo-url]
   cd personel-takip
   ```

2. `.env` dosyasını oluşturun:
   ```env
   MONGODB_URI=mongodb://mongodb:27017
   DB_NAME=personel_takip
   PORT=8080
   ```

3. Docker ile başlatın:
   ```bash
   docker-compose up -d
   ```

Uygulama varsayılan olarak `http://localhost:8080` adresinde çalışacaktır.

## Kullanım

### Yönetici Paneli

- Personel/stajyer ekleme ve yönetimi
- Günlük iş takibi
- İstatistikleri görüntüleme

### Personel Paneli

- İş başlatma ve tamamlama
- Video yükleme ve revizyon
- İş geçmişi görüntüleme

## Geliştirme

### Yerel Geliştirme Ortamı

1. Go'yu yükleyin (1.21 veya üstü)
2. MongoDB'yi yükleyin
3. Bağımlılıkları yükleyin:
   ```bash
   go mod download
   ```
4. Uygulamayı başlatın:
   ```bash
   go run main.go
   ```

### Docker ile Geliştirme

```bash
# Geliştirme ortamını başlatma
docker-compose up -d

# Logları izleme
docker-compose logs -f

# Servisleri durdurma
docker-compose down
```

## Lisans

Bu proje MIT lisansı altında lisanslanmıştır. Detaylar için [LICENSE](LICENSE) dosyasına bakınız. 