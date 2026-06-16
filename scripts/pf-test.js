// Playwright kapsamlı smoke test - PlayForge web
// - 6 dilde landing aç
// - Navbar'daki tüm linkleri tıkla, URL'i logla
// - Özellikle "Nasıl Çalışır" / "How it Works" linkini her dilde kontrol et
// - Ekran görüntüsü al

const { chromium } = require('playwright');
const fs = require('fs');

const BASE = 'http://localhost:3000';
const LOCALES = ['tr', 'en', 'es', 'de', 'fr', 'pt'];
const NAV_LABELS = {
  tr: ['Nasıl Çalışır', 'Fiyatlar', 'SSS', 'Giriş', 'Başla', 'Kayıt Ol'],
  en: ['How it Works', 'Pricing', 'FAQ', 'Login', 'Get Started', 'Sign Up'],
  es: ['Cómo Funciona', 'Precios', 'FAQ', 'Iniciar Sesión', 'Empezar', 'Registrarse'],
  de: ['Wie es funktioniert', 'Preise', 'FAQ', 'Anmelden', 'Starten', 'Registrieren'],
  fr: ['Comment ça marche', 'Tarifs', 'FAQ', 'Connexion', 'Commencer', 'Inscription'],
  pt: ['Como Funciona', 'Preços', 'FAQ', 'Entrar', 'Começar', 'Cadastre-se'],
};

(async () => {
  const browser = await chromium.launch();
  const context = await browser.newContext({ viewport: { width: 1280, height: 800 } });
  const page = await context.newPage();

  const results = [];
  const screenshotsDir = '/tmp/pf-shots';
  if (!fs.existsSync(screenshotsDir)) fs.mkdirSync(screenshotsDir, { recursive: true });

  for (const loc of LOCALES) {
    const url = `${BASE}/${loc}`;
    console.log(`\n=== ${loc.toUpperCase()} (${url}) ===`);
    try {
      const resp = await page.goto(url, { waitUntil: 'networkidle', timeout: 15000 });
      const finalUrl = page.url();
      const status = resp ? resp.status() : 'no-response';
      console.log(`  Final URL: ${finalUrl} [${status}]`);
      await page.screenshot({ path: `${screenshotsDir}/01-landing-${loc}.png`, fullPage: false });

      const links = await page.$$eval('nav a, header a', (els) =>
        els.map((e) => ({ text: e.innerText.trim(), href: e.getAttribute('href') })).filter((l) => l.text)
      );
      console.log(`  Nav links (${links.length}):`);
      links.forEach((l) => console.log(`    "${l.text}" -> ${l.href}`));

      results.push({ locale: loc, startUrl: url, finalUrl, status, linkCount: links.length, links });

      const howItWorksLabels = NAV_LABELS[loc];
      for (const label of howItWorksLabels) {
        try {
          const linkHandle = await page.$(`nav a:has-text("${label}"), header a:has-text("${label}")`);
          if (linkHandle) {
            const beforeUrl = page.url();
            await linkHandle.click();
            await page.waitForLoadState('networkidle', { timeout: 8000 });
            const afterUrl = page.url();
            const bug = !afterUrl.match(/\/(en|tr|es|de|fr|pt)(\/|$|\?)/);
            console.log(`  CLICK "${label}": ${beforeUrl} -> ${afterUrl} ${bug ? '  [BUG! locale lost]' : '  OK'}`);
            results.push({ test: 'nav-click', locale: loc, clickedLabel: label, from: beforeUrl, to: afterUrl, bug });
            await page.goto(url, { waitUntil: 'networkidle' });
          }
        } catch (e) {
          console.log(`  CLICK "${label}": error - ${e.message.split('\n')[0]}`);
        }
      }
    } catch (e) {
      console.log(`  ERROR: ${e.message.split('\n')[0]}`);
      results.push({ locale: loc, error: e.message });
    }
  }

  fs.writeFileSync('/tmp/pf-test-results.json', JSON.stringify(results, null, 2));
  console.log(`\n\nSaved: /tmp/pf-test-results.json + ${screenshotsDir}/01-landing-*.png`);

  await browser.close();
})();
