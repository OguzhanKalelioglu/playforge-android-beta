#!/usr/bin/env python3
"""Playwright design audit - PlayForge - tüm yüzeylerin screenshot'u + içerik analizi."""
import asyncio
import json
import os
from playwright.async_api import async_playwright

BASE = "http://localhost:3000"
SURFACES = [
    # path, label, viewport
    ("/tr", "tr-landing-desktop", (1280, 800)),
    ("/tr", "tr-landing-mobile", (375, 812)),
    ("/en", "en-landing-desktop", (1280, 800)),
    ("/tr/register", "tr-register-desktop", (1280, 800)),
    ("/tr/register", "tr-register-mobile", (375, 812)),
    ("/tr/login", "tr-login-desktop", (1280, 800)),
    ("/tr/login", "tr-login-mobile", (375, 812)),
    ("/tr/legal/terms", "tr-terms", (1280, 800)),
    ("/tr/dashboard", "tr-dashboard-redirect", (1280, 800)),
    ("/tr/dashboard/new", "tr-dashboard-new", (1280, 800)),
]

async def main():
    async with async_playwright() as p:
        browser = await p.chromium.launch(channel="chrome")
        out = "/tmp/pf-audit"
        os.makedirs(out, exist_ok=True)
        report = []
        for path, label, viewport in SURFACES:
            ctx = await browser.new_context(viewport={"width": viewport[0], "height": viewport[1]})
            page = await ctx.new_page()
            url = f"{BASE}{path}"
            print(f"--- {label} {url} {viewport} ---")
            try:
                resp = await page.goto(url, wait_until="networkidle", timeout=15000)
                status = resp.status if resp else None
                final = page.url
                await page.wait_for_timeout(400)
                # full page screenshot
                await page.screenshot(path=f"{out}/{label}-full.png", full_page=True)
                # viewport
                await page.screenshot(path=f"{out}/{label}-view.png", full_page=False)
                # body & h1
                h1 = await page.locator("h1").first.text_content() if await page.locator("h1").count() else None
                # any console errors
                # focus path - check if main interactive elements are focusable
                # count of buttons, links
                btns = await page.locator("button, a[href]").count()
                # form inputs
                inputs = await page.locator("input, textarea, select").count()
                # visible text density
                body_text = await page.locator("body").inner_text()
                # extract colors from rendered styles
                primary = await page.evaluate("getComputedStyle(document.documentElement).getPropertyValue('--primary')")
                report.append({
                    "label": label, "url": url, "finalUrl": final, "status": status,
                    "h1": h1, "buttons": btns, "inputs": inputs,
                    "bodyTextLen": len(body_text), "primary": primary.strip(),
                    "bodyTextSample": body_text[:500]
                })
                print(f"  h1: {h1!r}")
                print(f"  buttons: {btns}, inputs: {inputs}, text len: {len(body_text)}")
                print(f"  primary: {primary.strip()!r}")
            except Exception as e:
                print(f"  ERROR: {str(e).splitlines()[0]}")
                report.append({"label": label, "url": url, "error": str(e)})
            await ctx.close()

        with open(f"{out}/report.json", "w") as f:
            json.dump(report, f, indent=2, ensure_ascii=False)
        print(f"\nSaved to {out}/")
        await browser.close()

asyncio.run(main())
