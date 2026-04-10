# DominANT v3.0 - SSL & Domain Tracker Platform

<img src=".\ant-main.JPG" alt="alt text" width="880" height="330">

DominANT is a high-performance, enterprise-grade monitoring ecosystem built with **Go** and **React**. It provides real-time oversight of SSL/TLS certificate lifecycles and domain registration health, ensuring absolute continuity for your digital infrastructure.

![Aesthetics](https://img.shields.io/badge/Aesthetics-Dusk%20Grey-blueviolet?style=for-the-badge)
![Backend](https://img.shields.io/badge/Backend-Go%20Fiber-00ADD8?style=for-the-badge&logo=go)
![Frontend](https://img.shields.io/badge/Frontend-React%20%2B%20Vite-61DAFB?style=for-the-badge&logo=react)

## 🚀 Key Features

- **Parallel Scanning engine**: Scans hundreds of domains concurrently using Go routines for sub-second results.
- **Smart Expiry Tracking**:
  - **SSL/TLS**: Deep certificate inspection on Port 443.
  - **Domain WHOIS**: Multi-layered lookup (RDAP, WHOIS, specialized .tr support).
- **On-Prem & Internal Support**: Automatically detects and handles internal assets (IPs, .local, .lan) by skipping public WHOIS requests.
- **Advanced Workflows**:
  - **ACME (Let's Encrypt)**: Automated certificate issuance.
  - **SSH Deployment**: Seamless pushing of certificates to remote Linux servers.
- **Notification Ecosystem**: Instant alerts via **Telegram, Discord, Slack, Webhooks, and SMTP Email**.
- **Interactive Dashboard**: Modern "Dusk Grey" UI with real-time sync (SSE), visual health charts, and bulk management.

## 🛠 Tech Stack

- **Backend**: Go (Fiber Framework)
- **Database**: SQLite (high-concurrency WAL mode)
- **Frontend**: React (Vite, TypeScript, Lucide Icons)
- **Storage**: Highly optimized JSON-based domain storage with batch-update logic.

## 📦 Deployment & Setup

### For Development

1. **Backend**:

   ```bash
   cd backend-go
   go run main.go
   ```

2. **Frontend**:

   ```bash
   cd frontend
   npm install
   npm run dev
   ```

### For Production (Linux/Windows)

1. **Build the Interface**:

   ```bash
   cd frontend
   npm run build
   ```

2. **Build the Backend**:

   ```bash
   cd backend-go
   go build -o dominant-engine
   ```

3. **Run**:
   Ensure the `frontend/dist` folder is present near the binary, then run `./dominant-engine`.

## ⚙️ Configuration

Copy `backend-go/.env.example` to `backend-go/.env` and configure your preferences:

- `PORT`: Server port (default: 80)
- `DATA_PATH`: Directory for storage (default: ./data)
- `SMTP/Webhook Settings`: For notifications.

## 🛡 License

Commercial Project - All Rights Reserved.

---
*Developed with ❤️ for high-availability infrastructures.*
