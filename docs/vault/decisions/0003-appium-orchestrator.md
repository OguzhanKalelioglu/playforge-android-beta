---
tags: [adr, appium, automation, ui-testing]
status: accepted
date: 2026-05-22
---

# ADR-0003: Appium for UI automation

## Context

25 emülatör üzerinde gerçekçi UI otomasyonu gerekiyor:

- Play Store'da arama
- "Become a tester" akışı
- Uygulama indirme + açma
- Organik görünümlü engagement (swipe, tap, back)
- 5 yıldız + yorum yazma

**Alternatifler:**

1. **ADB shell scripting**
   - Sadece komut satırı işlemleri
   - UI element bulma yok
   - Gesture simülasyonu sınırlı

2. **uiautomator2 (Python)**
   - Tek dilde (Python)
   - Backend Go olunca context switch

3. **Selenium WebDriver**
   - Web için, Android native UI değil

4. **Appium (W3C WebDriver)**
   - Android native + WebView + Hybrid
   - Multi-language (Java, Python, JS, C#, Ruby, **Go**)
   - UiAutomator2 altta çalışır
   - Standart protokol, dokümantasyon iyi
   - Docker image mevcut (appium/appium)

## Decision

Appium 2.x kullanılacak. Orchestrator Go binary, appium server ayrı container.

## Architecture

```
┌─────────────────┐
│ Orchestrator Go │
│  (W3C client)   │◄──HTTP/JSON──►┌────────────────┐
│  port 9000      │                │ Appium server  │
└─────────────────┘                │ port 4723      │
                                   │ (UiAutomator2) │
                                   └────────┬───────┘
                                            │ ADB
                                            ▼
                                   ┌────────────────┐
                                   │ Emulator       │
                                   │ :5555          │
                                   └────────────────┘
```

## Consequences

### Olumlu

- Standard W3C WebDriver protokol
- Go SDK net (HTTP transport, kendi yazdık — 10 dosya, lightweight)
- UiAutomator2 backend → Android native desteği tam
- Container image hazır
- Inspector ile debug (geliştirme sırasında)

### Olumsuz

- Her test için yeni session (overhead ~3-5s)
- 25 emülatör × N test = N session/24h
- ADB bağlantı koptuğunda session kaybolur (retry logic gerekli)

### Mitigations

- `internal/appium/session.go` — idempotent Quit
- `internal/taskrunner/runner.go` — session defer + watchdog (10dk)
- Retry middleware: 3x exp backoff (30s/60s/120s)

## Implementation Notes

- `apps/orchestrator/internal/appium/` — 10 dosya
  - `client.go` — HTTP client
  - `session.go`, `locator.go`, `gestures.go`, `wait.go`, `screenshot.go`, `app.go`, `capabilities.go`, `errors.go`, `base64.go`
- `internal/taskrunner/antidetect.go` — Gaussian delay, Bezier, jitter
- `internal/profile/manager.go` — 10 cihaz profili (fingerprint randomization)
- `internal/task/comments.go` — 60+ TR yorum bank

## Status

**Accepted** — 2026-05-22. Implementasyon tamamlandı (commit 2b04ec9).
