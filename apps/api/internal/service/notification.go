package service

import (
	"context"
	"fmt"
	"net/smtp"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
)

// EmailService, SMTP üzerinden transactional email gönderir
// Production'da SendGrid/Postmark gibi servislere geçilebilir
type EmailService struct {
	host     string
	port     int
	user     string
	password string
	fromAddr string
	fromName string
	logger   *zap.Logger
}

func NewEmailService(logger *zap.Logger) *EmailService {
	port := 587
	if v := os.Getenv("SMTP_PORT"); v != "" {
		fmt.Sscanf(v, "%d", &port)
	}
	return &EmailService{
		host:     os.Getenv("SMTP_HOST"),
		port:     port,
		user:     os.Getenv("SMTP_USER"),
		password: os.Getenv("SMTP_PASSWORD"),
		fromAddr: envDefault("SMTP_FROM_ADDR", "no-reply@testerscomm.net"),
		fromName: envDefault("SMTP_FROM_NAME", "TestersCommunity"),
		logger:   logger,
	}
}

func envDefault(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func (s *EmailService) IsConfigured() bool {
	return s.host != "" && s.user != ""
}

// SendOrderPaid, sipariş ödemesi tamamlandığında
func (s *EmailService) SendOrderPaid(ctx context.Context, toEmail, toName, orderID, packageName string) {
	subject := "Ödemen alındı, test başlıyor"
	body := s.orderPaidHTML(toName, orderID, packageName)
	s.send(ctx, toEmail, subject, body)
}

// SendTestStarted, test başladığında
func (s *EmailService) SendTestStarted(ctx context.Context, toEmail, toName, testID, packageName string) {
	subject := "Test başladı: " + packageName
	body := s.testStartedHTML(toName, testID, packageName)
	s.send(ctx, toEmail, subject, body)
}

// SendTestCompleted, test tamamlandığında
func (s *EmailService) SendTestCompleted(ctx context.Context, toEmail, toName, testID, packageName string, reviews int) {
	subject := "Test tamamlandı: " + packageName
	body := s.testCompletedHTML(toName, testID, packageName, reviews)
	s.send(ctx, toEmail, subject, body)
}

// SendAccountBanned, hesap ban yediğinde (admin'e)
func (s *EmailService) SendAccountBanned(ctx context.Context, adminEmail, testerEmail string) {
	subject := "Hesap ban yedi: " + testerEmail
	body := fmt.Sprintf(`<!DOCTYPE html><html><body>
		<p>%s hesabı Google tarafından ban yedi. Lütfen durumu kontrol edip hesabı yenisiyle değiştirin.</p>
		<p>— TestersCommunity bot</p>
	</body></html>`, testerEmail)
	s.send(ctx, adminEmail, subject, body)
}

func (s *EmailService) send(ctx context.Context, to, subject, htmlBody string) {
	if !s.IsConfigured() {
		s.logger.Debug("smtp not configured, skipping email",
			zap.String("to", to), zap.String("subject", subject))
		return
	}
	go func() {
		addr := fmt.Sprintf("%s:%d", s.host, s.port)
		auth := smtp.PlainAuth("", s.user, s.password, s.host)

		msg := buildMessage(s.fromAddr, s.fromName, to, subject, htmlBody)
		if err := smtp.SendMail(addr, auth, s.fromAddr, []string{to}, []byte(msg)); err != nil {
			s.logger.Error("smtp send failed", zap.Error(err), zap.String("to", to))
		}
	}()
}

func buildMessage(fromAddr, fromName, to, subject, htmlBody string) string {
	headers := []string{
		fmt.Sprintf("From: %s <%s>", fromName, fromAddr),
		fmt.Sprintf("To: %s", to),
		fmt.Sprintf("Subject: =?UTF-8?B?%s?=", base64Encode(subject)),
		"MIME-Version: 1.0",
		"Content-Type: text/html; charset=UTF-8",
		fmt.Sprintf("Date: %s", time.Now().Format(time.RFC1123Z)),
	}
	return strings.Join(headers, "\r\n") + "\r\n\r\n" + htmlBody
}

func base64Encode(s string) string {
	const tbl = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	src := []byte(s)
	out := make([]byte, 0, ((len(src)+2)/3)*4)
	for i := 0; i < len(src); i += 3 {
		var b [3]byte
		n := copy(b[:], src[i:])
		out = append(out, tbl[b[0]>>2])
		out = append(out, tbl[(b[0]&3)<<4|b[1]>>4])
		if n > 1 {
			out = append(out, tbl[(b[1]&0xF)<<2|b[2]>>6])
		} else {
			out = append(out, '=')
		}
		if n > 2 {
			out = append(out, tbl[b[2]&0x3F])
		} else {
			out = append(out, '=')
		}
	}
	return string(out)
}

// --- Templates ---

func (s *EmailService) orderPaidHTML(name, orderID, packageName string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html><head><meta charset="utf-8"><title>Ödeme onayı</title></head>
<body style="font-family: -apple-system, sans-serif; max-width: 560px; margin: 0 auto; padding: 24px; color: #1a1a1a;">
<h2 style="font-weight: 600; margin-bottom: 8px;">Merhaba %s,</h2>
<p>Ödemen başarıyla alındı. %s için test 24 saat içinde başlayacak.</p>
<table style="width:100%%; border-collapse: collapse; margin: 24px 0;">
<tr><td style="padding: 8px 0; color: #666;">Sipariş</td><td style="text-align:right; font-family: monospace;">%s</td></tr>
<tr><td style="padding: 8px 0; color: #666;">Paket</td><td style="text-align:right;">%s</td></tr>
</table>
<p>Test ilerlemesini dashboard'dan takip edebilirsin:</p>
<p><a href="https://testerscomm.net/dashboard" style="display:inline-block; background:#1a56db; color:#fff; padding:10px 16px; border-radius:6px; text-decoration:none;">Dashboard'a Git</a></p>
<p style="color: #666; font-size: 13px; margin-top: 32px;">— TestersCommunity</p>
</body></html>`, name, packageName, orderID[:8], packageName)
}

func (s *EmailService) testStartedHTML(name, testID, packageName string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html><body style="font-family: -apple-system, sans-serif; max-width: 560px; margin: 0 auto; padding: 24px;">
<h2>Test başladı</h2>
<p>%s, %s için test başladı. 25 hesap şu anda opt-in ve indirme yapıyor.</p>
<p><a href="https://testerscomm.net/dashboard">İlerlemeyi Gör</a></p>
</body></html>`, name, packageName)
}

func (s *EmailService) testCompletedHTML(name, testID, packageName string, reviews int) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html><body style="font-family: -apple-system, sans-serif; max-width: 560px; margin: 0 auto; padding: 24px;">
<h2>Test tamamlandı</h2>
<p>%s, %s için 14 günlük test tamamlandı.</p>
<ul>
<li>%d hesap tamamladı</li>
<li>%d yorum yazıldı</li>
<li>Tüm aktivite logları dashboard'da</li>
</ul>
<p><a href="https://testerscomm.net/dashboard">Detaylı Raporu Gör</a></p>
</body></html>`, name, packageName, 25, reviews)
}
