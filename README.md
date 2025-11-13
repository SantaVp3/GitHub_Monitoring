# GitHub Monitoring Platform

A comprehensive GitHub information leakage monitoring platform that helps organizations detect and prevent sensitive data exposure on GitHub repositories.

**Version**: 1.2.0
**Status**: Production Ready
**Language**: [中文文档](./README_CN.md)

---

## Overview

This platform automatically monitors GitHub for potential information leaks based on customizable rules and keywords. It supports token rotation, multiple notification channels, and provides a user-friendly web interface for management.

### Key Features

- **Automated Monitoring**: Continuous scanning of GitHub repositories based on custom rules
- **Token Pool Management**: Automatic rotation of GitHub API tokens to handle rate limits
- **Multi-Channel Notifications**: Support for WeCom, DingTalk, Feishu, and custom webhooks
- **Flexible Matching**: Both fuzzy and precise keyword matching algorithms
- **Whitelist System**: Filter out known safe repositories and users
- **Batch Operations**: Efficiently manage large numbers of search results
- **Proxy Support**: HTTP, HTTPS, and SOCKS5 proxy configuration
- **JWT Authentication**: Secure access control with password protection
- **Real-time Dashboard**: Monitor system status and statistics at a glance

---

## Technology Stack

### Frontend
- **Framework**: React 18 with TypeScript
- **Build Tool**: Vite 7.2.2
- **UI Components**: shadcn/ui
- **Styling**: Tailwind CSS 3.4.0
- **HTTP Client**: Axios
- **Icons**: Lucide React

### Backend
- **Language**: Go (Golang)
- **Web Framework**: Gin
- **ORM**: GORM
- **Database**: MySQL 8.x
- **Authentication**: JWT (golang-jwt/jwt/v5)
- **Configuration**: Viper
- **GitHub API**: google/go-github/v57

---

## Installation

### Prerequisites

- Node.js 16+ and npm
- Go 1.18+
- MySQL 8.0+

### Backend Setup

1. Clone the repository:
```bash
git clone <repository-url>
cd GitHub-Monitoring/backend
```

2. Install Go dependencies:
```bash
go mod download
```

3. Configure the database:
```bash
mysql -u root -p
CREATE DATABASE github_monitor CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

4. Configure the application:
```bash
cp config.yaml.example config.yaml
# Edit config.yaml with your settings
```

5. Run the backend:
```bash
go run main.go
```

The backend server will start on `http://localhost:8080`

### Frontend Setup

1. Navigate to the frontend directory:
```bash
cd ../frontend
```

2. Install dependencies:
```bash
npm install
```

3. Start the development server:
```bash
npm run dev
```

The frontend will be available at `http://localhost:5174`

---

## Configuration

### Backend Configuration (config.yaml)

```yaml
server:
  port: 8080
  mode: debug  # Use "release" in production

database:
  host: localhost
  port: 3306
  user: root
  password: your_password
  database: github_monitor

auth:
  enabled: true
  password: "admin123"  # Change this!
  jwt_secret: "your-secret-key"  # Change this!
  token_expiry: "24h"

github:
  tokens:
    - token: "ghp_your_token_1"
      name: "Token 1"
    - token: "ghp_your_token_2"
      name: "Token 2"

  # Proxy configuration (optional)
  proxy_enabled: false
  proxy_url: ""
  proxy_type: "http"  # http, https, or socks5
  proxy_username: ""
  proxy_password: ""

monitor:
  scan_interval: "5m"  # Scanning interval
  max_results_per_rule: 100
```

### Required GitHub Token Permissions

To use this platform, you need GitHub Personal Access Tokens with the following scope:
- `public_repo` - Search public repositories
- `repo` (optional) - If you need to search private repositories

Generate tokens at: https://github.com/settings/tokens

---

## Usage

### Initial Setup

1. Access the login page at `http://localhost:5174`
2. Login with the default password: `admin123`
3. Change the password in `backend/config.yaml` (recommended)

### Adding GitHub Tokens

1. Navigate to **Settings** page
2. Expand the **GitHub Tokens** section
3. Click **Add Token**
4. Enter token name and token value
5. Click **Add Token** to save

### Creating Monitor Rules

1. Navigate to **Monitor Rules** page
2. Click **Add Rule**
3. Fill in the form:
   - **Rule Name**: Descriptive name for the rule
   - **Match Type**: Choose Fuzzy or Precise matching
   - **Keywords**: Comma-separated keywords (e.g., `password, api_key, secret`)
   - **Description**: Optional description
   - **Active**: Check to enable immediately
4. Click **Create Rule**

### Managing Search Results

1. Navigate to **Search Results** page
2. View detected potential leaks
3. Use checkboxes to select multiple results
4. Use batch actions:
   - **Mark as Confirmed**: Flag as real leaks
   - **Mark as False Positive**: Mark as safe

### Configuring Notifications

1. Navigate to **Settings** page
2. Expand the **Notification Channels** section
3. Click **Add Channel**
4. Configure:
   - **Name**: Channel identifier
   - **Type**: Select WeCom, DingTalk, Feishu, or Webhook
   - **Webhook URL**: Your webhook endpoint
   - **Secret**: For DingTalk/Feishu signature verification
   - **Notify On**: Choose when to receive notifications
5. Click **Create Channel**
6. Test the notification with the **Test** button

### Using Whitelist

1. Navigate to **Whitelist** page
2. Click **Add to Whitelist**
3. Select type:
   - **User**: Whitelist a GitHub user
   - **Repository**: Whitelist a specific repository
4. Enter the value and optional description
5. Click **Add**

---

## API Documentation

### Authentication

All API endpoints (except `/api/v1/login`) require JWT authentication.

**Login**
```http
POST /api/v1/login
Content-Type: application/json

{
  "password": "your-password"
}

Response:
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "message": "Login successful"
}
```

**Authenticated Requests**
```http
GET /api/v1/dashboard/stats
Authorization: Bearer <your-token>
```

### API Endpoints

#### Dashboard
- `GET /api/v1/dashboard/stats` - Get dashboard statistics

#### Token Management
- `GET /api/v1/tokens` - List all tokens
- `POST /api/v1/tokens` - Create a new token
- `DELETE /api/v1/tokens/:id` - Delete a token
- `GET /api/v1/tokens/stats` - Get token usage statistics

#### Monitor Rules
- `GET /api/v1/rules` - List all rules
- `GET /api/v1/rules/:id` - Get a specific rule
- `POST /api/v1/rules` - Create a new rule
- `PUT /api/v1/rules/:id` - Update a rule
- `DELETE /api/v1/rules/:id` - Delete a rule

#### Search Results
- `GET /api/v1/results` - List search results (supports pagination)
- `PUT /api/v1/results/:id` - Update result status
- `POST /api/v1/results/batch` - Batch update result status

#### Whitelist
- `GET /api/v1/whitelist` - List whitelist entries
- `POST /api/v1/whitelist` - Add whitelist entry
- `DELETE /api/v1/whitelist/:id` - Remove whitelist entry

#### Monitor Control
- `GET /api/v1/monitor/status` - Get monitoring service status
- `POST /api/v1/monitor/start` - Start monitoring
- `POST /api/v1/monitor/stop` - Stop monitoring

#### Notifications
- `GET /api/v1/notifications` - List notification channels
- `POST /api/v1/notifications` - Create notification channel
- `PUT /api/v1/notifications/:id` - Update notification channel
- `DELETE /api/v1/notifications/:id` - Delete notification channel
- `POST /api/v1/notifications/:id/test` - Test notification channel

#### Scan History
- `GET /api/v1/history` - Get scan history (supports pagination)

---

## Architecture

### System Components

1. **Frontend (React + TypeScript)**
   - User interface for management and monitoring
   - Real-time data updates
   - Responsive design

2. **Backend (Go + Gin)**
   - RESTful API server
   - JWT authentication
   - Background monitoring service
   - Token pool management

3. **Database (MySQL)**
   - Data persistence
   - Search results storage
   - Configuration management

4. **GitHub API Integration**
   - Code search functionality
   - Rate limit handling
   - Proxy support

### Data Models

**GitHubToken**: Stores GitHub API tokens for rotation
**MonitorRule**: Defines monitoring rules and keywords
**SearchResult**: Stores detected potential leaks
**Whitelist**: Contains whitelisted users and repositories
**ScanHistory**: Records scanning activities
**NotificationConfig**: Notification channel configurations

---

## Security Considerations

### Password Protection
- Change the default password immediately
- Use strong passwords with mixed characters
- Store passwords securely in `config.yaml`

### JWT Secret
- Generate a random, long secret key
- Never commit secrets to version control
- Rotate secrets periodically

### Token Security
- Use tokens with minimal required permissions
- Rotate tokens regularly
- Monitor token usage

### HTTPS in Production
- Always use HTTPS in production environments
- Protect token transmission
- Use secure WebSocket connections

### Configuration File
- Set proper file permissions (600 or 400)
- Add `config.yaml` to `.gitignore`
- Use environment variables for sensitive data

---

## Troubleshooting

### Problem: Login fails immediately after successful authentication

**Possible Causes**:
- JWT secret mismatch
- Token format error
- System time incorrect

**Solutions**:
- Verify `jwt_secret` in `config.yaml`
- Check browser console for errors
- Ensure system time is correct

### Problem: 401 Unauthorized errors

**Possible Causes**:
- Token expired
- Invalid token
- Authentication middleware misconfigured

**Solutions**:
- Re-login to get a new token
- Check backend logs
- Verify `auth.enabled` configuration

### Problem: GitHub API returns 401 Bad Credentials

**Possible Causes**:
- Invalid GitHub tokens
- Expired tokens
- Wrong token permissions

**Solutions**:
- Generate new tokens at https://github.com/settings/tokens
- Add tokens via Settings page
- Ensure tokens have `public_repo` scope

### Problem: No search results

**Possible Causes**:
- No active monitoring rules
- Keywords too specific
- Whitelist filtering too broad

**Solutions**:
- Create and activate monitoring rules
- Use more common keywords
- Review whitelist entries

---

## Contributing

Contributions are welcome! Please follow these guidelines:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Write or update tests
5. Submit a pull request

---

## License

This project is licensed under the MIT License. See the LICENSE file for details.

---

## Changelog

### Version 1.2.0 (2025-11-13)
- Added GitHub token management in frontend
- Implemented JWT authentication system
- Added batch operations for search results
- Added proxy support (HTTP/HTTPS/SOCKS5)
- Optimized Settings page with collapsible sections
- Added Monitor Rules CRUD functionality
- Improved error handling and validation

### Version 1.1.0
- Added notification system (WeCom, DingTalk, Feishu, Webhook)
- Implemented whitelist management
- Added settings page
- Improved dashboard statistics

### Version 1.0.0
- Initial release
- Core monitoring functionality
- Token rotation system
- Basic web interface

---

**Last Updated**: 2025-11-13
**Project Status**: Production Ready
**Maintained**: Yes
