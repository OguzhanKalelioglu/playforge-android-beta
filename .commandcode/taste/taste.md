# Taste (Continuously Learned by [CommandCode][cmd])

[cmd]: https://commandcode.ai/

# Payments
- Use Stripe (Global) instead of Iyzico or local TR providers for payment integration. Confidence: 0.90

# Workflow
- Use Commander Coder CLI with Stripe skills for fast Stripe integration implementation. Confidence: 0.75
- Maintain project knowledge in Obsidian Vault + Code Graph so the model can understand the full project context. Confidence: 0.80
- Run work in structured todo lists: create a todo, then proceed in order ("tekrar bir todo oluşturup, sırayla devam edelim"). Confidence: 0.75
- Use `gh` CLI for GitHub operations (auth, repo create, push, clone) without asking the user for credentials or URLs — it's pre-installed on both local and Mini PC. Confidence: 0.75

# Git
- In .gitignore, use `**/._*` and `**/.DS_Store` (not just `/._*` and `/.DS_Store`) to recursively ignore macOS AppleDouble files in all subdirectories. Confidence: 0.80

# Infrastructure
- Do not disrupt existing running stacks on the host (e.g. postiz uses 5432/6379 on the Mini PC). When adding a new stack, change host port mappings (e.g. `18080:8080`) and keep shared services (postgres/redis) on `expose:` only to avoid collisions. Confidence: 0.80

# Web / Next.js
- Use `/[locale]/` route segment for i18n (next-intl 4.x) with setRequestLocale + getTranslations in server components. Confidence: 0.75
