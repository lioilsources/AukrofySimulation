# SPARK — LiteLLM endpoint pro auction-sim

Tento adresář se nasazuje na DGX Spark přes `make deploy-spark` (rsync + `docker compose up -d`).
Obsahuje LiteLLM proxy, která servíruje modelové aliasy `llm-dev` (rozhodování bidderů)
a `llm-lab` (insights do reportů), na které se odkazuje `.env` enginu.

## První nasazení

```bash
# z macu — klíče a přístup (jednorázově)
make ssh-keys
make ssh-copy-id-spark SPARK_HOST=<user@spark>
make ssh-add

# na Sparku — vytvořit reálný config (deploy ho nikdy nepřepíše)
mkdir -p /opt/auction-sim/litellm
cp config.example.yaml /opt/auction-sim/litellm/config.yaml   # a upravit api_base/modely

# z macu — deploy + kontrola
make deploy-spark SPARK_HOST=<user@spark>
make spark-status SPARK_HOST=<user@spark>
```

## Co tu je vs. co žije jen na Sparku

| soubor | synchronizuje se | poznámka |
|---|---|---|
| `docker-compose.yml` | ✅ | LiteLLM služba na :4000 |
| `config.example.yaml` | ✅ | šablona |
| `config.yaml` | ❌ (exclude v rsync) | reálné backendy/klíče, jen na Sparku |

Expozice na `https://llm.ol1n.com` běží přes cloudflared tunel + Cloudflare Access
(mimo tento compose). Engine se autentizuje service tokenem — hlavičky
`CF-Access-Client-Id` / `CF-Access-Client-Secret` z `.env` na NASu.
