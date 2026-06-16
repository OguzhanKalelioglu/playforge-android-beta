#!/usr/bin/env python3
"""Playwright smoke test - PlayForge web - 6 dilde nav linklerini test et."""
import asyncio
import json
import os
import sys
from playwright.async_api import async_playwright

BASE = "http://localhost:3000"
LOCALES = ["tr", "en", "es", "de", "fr", "pt"]
NAV_LABELS = {
    "tr": ["Nasıl Çalışır", "Fiyatlar", "SSS", "Giriş", "Başla", "Kayıt Ol", "Anasayfa"],
    "en": ["How it Works", "Pricing", "FAQ", "Login", "Get Started", "Sign Up", "Home"],
    "es": ["Cómo Funciona", "Precios", "FAQ", "Iniciar Sesión", "Empezar", "Registrarse", "Inicio"],
    "de": ["Wie es funktioniert", "Preise", "FAQ", "Anmelden", "Starten", "Registrieren", "Startseite"],
    "fr": ["Comment ça marche", "Tarifs", "FAQ", "Connexion", "Commencer", "Inscription", "Accueil"],
    "pt": ["Como Funciona", "Preços", "FAQ", "Entrar", "Começar", "Cadastre-se", "Início"],
}

async def main():
    async with async_playwright() as p:
        browser = await p.chromium.launch(channel="chrome")
        ctx = await browser.new_context(viewport={"width": 1280, "height": 800})
        page = await ctx.new_page()

        os.makedirs("/tmp/pf-shots", exist_ok=True)
        results = []

        for loc in LOCALES:
            url = f"{BASE}/{loc}"
            print(f"\n=== {loc.upper()} ({url}) ===")
            try:
                resp = await page.goto(url, wait_until="networkidle", timeout=15000)
                final_url = page.url
                status = resp.status if resp else "no-response"
                print(f"  Final URL: {final_url} [{status}]")
                await page.screenshot(path=f"/tmp/pf-shots/01-landing-{loc}.png", full_page=False)

                links = await page.eval_on_selector_all(
                    "nav a, header a",
                    "els => els.map(e => ({text: e.innerText.trim(), href: e.getAttribute('href')})).filter(l => l.text)"
                )
                print(f"  Nav links ({len(links)}):")
                for l in links:
                    print(f"    \"{l['text']}\" -> {l['href']}")

                results.append({"locale": loc, "startUrl": url, "finalUrl": final_url, "status": status, "links": links})

                # Her label'ı tıkla
                for label in NAV_LABELS.get(loc, []):
                    try:
                        link = page.locator(f"nav a:has-text(\"{label}\"), header a:has-text(\"{label}\")").first
                        if await link.count() == 0:
                            continue
                        before = page.url
                        await link.click()
                        try:
                            await page.wait_for_load_state("networkidle", timeout=5000)
                        except Exception:
                            pass
                        after = page.url
                        # Hash-linkler icin path kontrolu: /tr, /en, /es... icermesi gerek
                        from urllib.parse import urlparse
                        path = urlparse(after).path
                        bug = not any(path == f"/{l}" or path == f"/{l}/" or path == "" or path == "/" for l in LOCALES) and path not in [f"/{l}" for l in LOCALES]
                        # Daha basit: path bos veya /tr /en /es /de /fr /pt olmali
                        bug = path not in [f"/{l}" for l in LOCALES] + ["/", ""]
                        marker = "  [BUG: locale lost]" if bug else "  OK"
                        print(f"  CLICK \"{label}\": {before} -> {after} {marker}")
                        results.append({"test": "nav-click", "locale": loc, "label": label, "from": before, "to": after, "bug": bug})
                        await page.goto(url, wait_until="networkidle", timeout=10000)
                    except Exception as e:
                        print(f"  CLICK \"{label}\": {str(e).splitlines()[0]}")
            except Exception as e:
                print(f"  ERROR: {str(e).splitlines()[0]}")
                results.append({"locale": loc, "error": str(e)})

        with open("/tmp/pf-test-results.json", "w") as f:
            json.dump(results, f, indent=2)
        print(f"\nSaved: /tmp/pf-test-results.json + /tmp/pf-shots/01-landing-*.png")
        await browser.close()

asyncio.run(main())
