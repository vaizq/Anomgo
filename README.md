# Anomgo – Anonymous Marketplace in Golang

**Anomgo** is an open-source, anonymous online marketplace written in **Go**.  
This project was built as a personal exploration into topics like **cryptography**, **privacy**, and **secure backend development**.

> ⚠️ **Note:** This project is for educational and research purposes only.  
> It is **not intended for production use**.

---

## 🧠 Why I Built It

I wanted to reverse-engineer and better understand the architecture of anonymous marketplaces — how they work, where they fail, and what can we learn from them.  
By building one from scratch, I was able to dive deep into **Golang**, **web security**, and **secure transaction handling**.

---

## ✨ Features

- 💱 **Monero integration** using [moneropay](https://moneropay.eu/)
- 🔐 **Escrow system** for secure transactions
- 🔑 **Optional PGP-based 2FA** for improved security and privacy
- 🧩 **Vendor pledge system** to reduce scam
- 🧠 **LLM-generated multi-language support** (native English + Finnish)
- 🛡️ **Captcha system** to defend against automated abuse
- 🚫 **JavaScript-free frontend** for maximum security

---

## 🌍 Demo

You can explore a live demo running on monero-stagenet at:
➡️ [www.anomgo.online](http://www.anomgo.online)

---

## ⚙️ Tech Stack

- **Golang** – core backend logic and routing
- **PostgreSQL** – relational database
- **HTML templates** – server-rendered UI (no JS)
- **PGP & Monero integration** – crypto & encryption

---

## ⚙️ Build

- Run [moneropay](https://moneropay.eu).
- Install golang and docker
- Configure .env (see [docker-compose.yaml](./docker-compose.yml) for necessary variables)
```
docker compose up
cd backend && make run
```

---

## 📄 License

MIT
