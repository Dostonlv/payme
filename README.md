# 🚀 Payment Kisuke.uz - PayMe API Package

Ushbu paket PayMe.uz merchant API bilan ishlash uchun yaratilgan Go dasturlash tilida yozilgan to'liq paket.

## ✨ Xususiyatlar

- **To'liq Receipts API** - Barcha PayMe receipts methodlari qo'llab-quvvatlanadi
- **Webhook Handler** - PayMe webhook so'rovlarini avtomatik qayta ishlash
- **Type Safety** - Kuchli type system bilan xatolarni oldini olish
- **Error Handling** - To'liq xatolarni qayta ishlash
- **Test Mode** - Test va production muhitlari uchun alohida endpointlar
- **Documentation** - To'liq hujjatlar va misollar

## 📦 O'rnatish

```bash
go get payment.kisuke.uz
```

## 🚀 Tezkor Boshlash

### 1. Client Yaratish

```go
package main

import (
    "context"
    "payment.kisuke.uz/pkg/payme"
)

func main() {
    // PayMe client yaratish
    client := payme.NewClient(payme.ClientConfig{
        PaymeID:        "your-payme-id",
        PaymeKey:       "your-payme-key",
        IsTestMode:     true,
        RequisiteName:  "order_id", // charge_id, order_id, yoki id
        Timeout:        30,
    })

    // Receipt yaratish - RequisiteName ga qarab account map yaratish
    account := map[string]interface{}{
        client.RequisiteName: "123", // charge_id, order_id, yoki id
        "card_id":            "9999",
        "reason":             "test_payment",
    }

    // Receipt yaratish
    req := &payme.CreateReceiptRequest{
        Amount:      10000, // 100 so'm (tiyinda)
        Account:     map[string]interface{}{"charge_id": "123"},
        Description: "Test receipt",
        Detail:      map[string]interface{}{"method": "test"},
    }

    ctx := context.Background()
    resp, err := client.CreateReceipt(ctx, req)
    if err != nil {
        panic(err)
    }

    fmt.Printf("Receipt yaratildi: %s\n", resp.Receipt.ID)
}
```

### 2. Webhook Handler

```go
package main

import (
    "github.com/gofiber/fiber/v2"
    "payment.kisuke.uz/pkg/payme"
)

// SimpleReceiptProvider webhook handler
type SimpleReceiptProvider struct{}

func (rp *SimpleReceiptProvider) CreateReceipt(params map[string]interface{}) (*payme.CreateReceiptResponse, error) {
    // Receipt yaratish logic
    receipt := &payme.Receipt{
        ID:          "receipt_123",
        CreateTime:  time.Now().UnixMilli(),
        State:       payme.ReceiptStateCreated,
        Amount:      10000,
        Currency:    payme.CurrencyUZS,
        Description: "Test receipt",
    }
    
    return &payme.CreateReceiptResponse{Receipt: receipt}, nil
}

// Boshqa methodlarni ham implement qilish kerak...

func main() {
    app := fiber.New()
    
    // Webhook endpoint ni sozlash
    payme.SetupWebhookWithPath(app, "your-payme-key", &SimpleReceiptProvider{}, "/webhook/payme")
    
    app.Listen(":8080")
}
```

## ⚙️ Sozlamalar

### RequisiteName

`RequisiteName` field PayMe API ga yuboriladigan account object da qaysi nom ishlatilishini belgilaydi:

- **`charge_id`** - PayMe merchant ID (default)
- **`order_id`** - Order ID
- **`id`** - Universal ID

Bu field ga qarab, account map da to'g'ri nom ishlatiladi:

```go
// charge_id bilan
client.RequisiteName = "charge_id"
account := map[string]interface{}{
    "charge_id": "123",
    "card_id":   "9999",
    "reason":    "test_payment",
}

// order_id bilan  
client.RequisiteName = "order_id"
account := map[string]interface{}{
    "order_id": "123",
    "card_id":  "9999", 
    "reason":   "test_payment",
}
```

### ClientConfig

| Field | Type | Description | Default |
|-------|------|-------------|---------|
| `PaymeID` | `string` | PayMe merchant ID | - |
| `PaymeKey` | `string` | PayMe merchant key | - |
| `IsTestMode` | `bool` | Test mode (true = test endpoint, false = production) | `false` |
| `RequisiteName` | `string` | Rekvizit nomi (`charge_id`, `order_id`, yoki `id`). Bu field PayMe API ga yuboriladigan account object da qaysi nom ishlatilishini belgilaydi | `id` |
| `Timeout` | `time.Duration` | HTTP request timeout | `30s` |
| `Logger` | `*log.Logger` | Logger instance | `nil` |
| `HTTPClient` | `http.Client` | Custom HTTP client | `http.DefaultClient` |
| `BaseURL` | `string` | Custom base URL | Avtomatik (test/production) |

## 💰 Currency Codes

PayMe API qo'llab-quvvatlaydigan currency codes:

| Code | Currency | Symbol | Description |
|------|----------|--------|-------------|
| `860` | UZS | сум | Uzbekistan Som |
| `840` | USD | $ | US Dollar |
| `978` | EUR | € | Euro |

## 📚 API Methodlari

### Receipts API

| Method | Maqsad | Description |
|--------|--------|-------------|
| `receipts.create` | Chek yaratish | Yangi to'lov cheki yaratish |
| `receipts.pay` | Chekni to'lash | Mavjud chekni to'lash |
| `receipts.send` | Chekni yuborish | Chekni mijozga yuborish |
| `receipts.cancel` | Chekni bekor qilish | Chekni bekor qilish |
| `receipts.check` | Chek holatini tekshirish | Chek holatini tekshirish |
| `receipts.get` | Chek ma'lumotini olish | Chek ma'lumotlarini olish |
| `receipts.get_all` | Barcha cheklarni olish | Barcha cheklarni olish (vaqt oralig'i va soni bilan) |
| `receipts.set_fiscal_data` | Fiscal ma'lumotlarni o'rnatish | Fiscal ma'lumotlarni o'rnatish |

### Sizning `receipts.create` Requestingiz

```json
{
    "charge_id": 123,
    "method": "receipts.create",
    "params": {
        "amount": 10000,
        "account": {
            "charge_id": 108,
            "card_id": "9999",
            "reason": "1232132132132131"
        },
        "description": "DESC",
        "detail": {
            "method": "refill_driver_balance"
        }
    }
}
```

**Response:**
```json
{
    "result": {
        "receipt": {
            "_id": "receipt_108_1755113788",
            "create_time": 1755113788877,
            "state": 0,
            "amount": 10000,
            "currency": 860,
            "description": "DESC"
        }
    }
}
```

### GetAllReceipts Method

```go
// Barcha receiptlarni olish
from := time.Now().AddDate(0, -1, 0).UnixMilli() // 1 oy oldin
to := time.Now().UnixMilli()                        // Hozirgi vaqt
count := 10                                         // Maksimal soni

allReceipts, err := client.GetAllReceipts(ctx, from, to, count)
if err != nil {
    log.Printf("Xatolik: %v", err)
} else {
    fmt.Printf("Jami receiptlar: %d\n", len(allReceipts.Receipts))
    
    for _, receipt := range allReceipts.Receipts {
        if receipt != nil {
            fmt.Printf("ID: %s, Summa: %s, Holat: %d\n",
                receipt.ID,
                payment.FormatAmount(receipt.Amount, receipt.Currency),
                receipt.State)
        }
    }
}
```

**Parameters:**
- `from` - Boshlang'ich vaqt (Unix timestamp, millisecond)
- `to` - Tugash vaqti (Unix timestamp, millisecond)  
- `count` - Maksimal qaytariladigan receipt soni

**Response:**
```json
{
    "result": [
        {
            "_id": "receipt_1",
            "create_time": 1755113788877,
            "state": 1,
            "amount": 10000,
            "currency": 860
        },
        {
            "_id": "receipt_2", 
            "create_time": 1755113788878,
            "state": 0,
            "amount": 20000,
            "currency": 860
        }
    ]
}
```

## 🔧 Konfiguratsiya

### Environment Variables

```bash
PAYME_ID=your-payme-id
PAYME_KEY=your-payme-key
PAYME_TEST_MODE=true
PAYME_TIMEOUT=30
```

### Client Config

```go
type ClientConfig struct {
    PaymeID    string `json:"payme_id"`     // PayMe merchant ID
    PaymeKey   string `json:"payme_key"`    // PayMe secret key
    IsTestMode bool   `json:"is_test_mode"` // Test mode (true/false)
    Timeout    int    `json:"timeout"`      // Request timeout (seconds)
}
```

## 🌐 Endpointlar

- **Test:** `https://checkout.test.paycom.uz/api`
- **Production:** `https://checkout.paycom.uz/api`

## 📊 Receipt States

| State | Description |
|-------|-------------|
| `0` | Yaratildi (Created) |
| `1` | To'langan (Paid) |
| `-1` | Bekor qilingan (Canceled) |
| `-2` | Muddat o'tgan (Expired) |

## 💰 Currency Codes

| Code | Currency |
|------|----------|
| `860` | O'zbek so'mi (UZS) |
| `840` | AQSH dollari (USD) |
| `978` | Yevro (EUR) |
| `810` | Rossiya rubli (RUB) |

## 🚨 Error Codes

| Code | Description |
|------|-------------|
| `-31611` | Noto'g'ri summa |
| `-32602` | Noto'g'ri parametrlar |
| `-31400` | Karta topilmadi |
| `-32504` | Ruxsat berilmagan |
| `-32700` | Parse xatosi |
| `-32601` | Method topilmadi |

## 📁 Fayl Strukturasi

```
payment-kisuke-uz/
├── pkg/
│   └── payme/
│       ├── constants/     # Konstantalar
│       ├── types/         # Data turlari
│       ├── client/        # HTTP client
│       ├── handlers/      # Webhook handler
│       └── payme.go       # Asosiy paket
├── example/               # Misollar
├── go.mod                 # Go moduli
└── README.md             # Hujjat
```

## 🧪 Test Qilish

```bash
cd payment-kisuke-uz
go mod tidy
go run example/main.go
```

## 📝 Webhook Endpoint

Webhook endpoint: `POST /webhook/payme`

**Authorization:** Basic Auth
```
Authorization: Basic base64(merchant_id:key)
```

## 🔒 Xavfsizlik

- **API Key** - Har doim xavfsiz saqlang
- **HTTPS** - Faqat HTTPS orqali ishlatish
- **Validation** - Barcha inputlarni tekshirish
- **Rate Limiting** - So'rovlar sonini cheklash

## 🤝 Hissa Qo'shish

1. Repository ni fork qiling
2. Yangi branch yarating
3. O'zgarishlarni qo'shing
4. Pull request yuboring

## 📄 License

MIT License - [LICENSE](LICENSE) faylini ko'ring

## 📞 Aloqa

- **Email:** support@kisuke.uz
- **Website:** https://kisuke.uz
- **Documentation:** https://docs.kisuke.uz

## 🙏 Minnatdorchilik

- [PayMe.uz](https://payme.uz) - To'lov tizimi
- [Fiber](https://gofiber.io) - Web framework
- [Go](https://golang.org) - Dasturlash tili

---

**Made with ❤️ by Kisuke.uz Team**
