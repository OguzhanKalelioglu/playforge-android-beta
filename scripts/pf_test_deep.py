#!/usr/bin/env python3
"""Playwright deep test - hash anchor scroll, FAQ/Pricing sections, register/login pages."""
import asyncio
import os
from urllib.parse import urlparse
from playwright.async_api import async_playwright

BASE = "http://localhost:3000"
LOCALES = ["tr", "en", "es", "de", "fr", "pt"]

async def main():
    async with async_playwright() as p:
        browser = await p.chromium.launch(channel="chrome")
        ctx = await browser.new_context(viewport={"width": 1280, "height": 800})
        page = await ctx.new_page()
        os.makedirs("/tmp/pf-shots", exist_ok=True)

        for loc in LOCALES:
            url = f"{BASE}/{loc}"
            print(f"\n=== {loc.upper()} DEEP ({url}) ===")
            try:
                await page.goto(url, wait_until="networkidle", timeout=15000)
                await page.screenshot(path=f"/tmp/pf-shots/deep-01-landing-{loc}.png", full_page=True)

                # Hash anchor tıklamaları: #features, #how, #pricing, #faq
                for anchor in ["#features", "#how", "#pricing", "#faq"]:
                    link = page.locator(f"nav a[href$='{anchor}'], header a[href$='{anchor}']").first
                    if await link.count() == 0:
                        continue
                    before = page.url
                    await link.click()
                    await page.wait_for_timeout(500)  # smooth scroll
                    after = page.url
                    path = urlparse(after).path
                    locale_ok = path == f"/{loc}"
                    print(f"  {anchor}: {before} -> {after}  {'OK' if locale_ok else 'BUG'}")
                    if "deep" in anchor or anchor == "#features":
                        await page.screenshot(path=f"/tmp/pf-shots/deep-02-{loc}-{anchor[1:]}.png", full_page=False)

                # /tr/login açılıyor mu
                login_url = f"{BASE}/{loc}/login"
                r = await page.goto(login_url, wait_until="networkidle", timeout=10000)
                print(f"  /{loc}/login -> {r.status if r else 'no-resp'}")
                await page.screenshot(path=f"/tmp/pf-shots/deep-03-login-{loc}.png", full_page=False)

                # /tr/register
                reg_url = f"{BASE}/{loc}/register"
                r = await page.goto(reg_url, wait_until="networkidle", timeout=10000)
                print(f"  /{loc}/register -> {r.status if r else 'no-resp'}")
                await page.screenshot(path=f"/tmp/pf-shots/deep-04-register-{loc}.png", full_page=False)

                # /tr/dashboard (girişsiz, redirect beklenir)
                dash_url = f"{BASE}/{loc}/dashboard"
                r = await page.goto(dash_url, wait_until="networkidle", timeout=10000)
                print(f"  /{loc}/dashboard -> {r.status if r else 'no-resp'} -> {page.url}")

                # /tr/dashboard/legal/terms
                terms_url = f"{BASE}/{loc}/dashboard/legal/terms"
                r = await page.goto(terms_url, wait_until="networkidle", timeout=10000)
                print(f"  /{loc}/dashboard/legal/terms -> {r.status if r else 'no-resp'}")
            except Exception as e:
                print(f"  ERROR: {str(e).splitlines()[0]}")

        print("\nDone. Shots in /tmp/pf-shots/deep-*.png")
        await browser.close()

asyncio.run(main())
