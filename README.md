<div align="center">
    <h1 align="center">Film-Packager</h1>
<a href="https://film-packager.fly.dev/">

## About

<p>Film-Packager is a single-page application (SPA) that allows users to manage documents for film projects in development.</p>

## An SPA to manage documents for film projects in development

### Built With

![Go](https://img.shields.io/badge/Go-%23000000.svg?&style=for-the-badge&logo=go)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-%23000000.svg?&style=for-the-badge&logo=postgresql&logoColor=%23ffffff)
![AWS S3](https://img.shields.io/badge/AWS_S3-%23000000.svg?&style=for-the-badge&logo=s3&logoColor=%23ffffff)
![Fly](https://img.shields.io/badge/Fly.io-%23000000.svg?&style=for-the-badge&logo=fly.io&logoColor=%23ffffff)
![Docker](https://img.shields.io/badge/Docker-%23000000.svg?&style=for-the-badge&logo=docker&logoColor=%23ffffff)
![HTMX](https://img.shields.io/badge/HTMX-%23000000.svg?&style=for-the-badge&logo=htmx&logoColor=%23ffffff)

## Features

- **Project Management**: Create and organize film projects
- **Document Storage**: Upload, categorize, and version control film documents
- **Collaboration**: Share documents with team members
- **Search Functionality**: Quickly find documents across projects
- **Mobile Responsive**: Access your documents from any device

## Installation

### Prerequisites

- Go 1.19 or higher
- PostgreSQL
- AWS S3 account
- Docker (for development)

### Local Development Setup

```bash
# Clone the repository
git clone https://github.com/yourusername/film-packager.git
cd film-packager

# Set up environment variables (example)
cp .env.example .env
# Edit .env with your credentials

# Run with Docker Compose
docker-compose up

# Alternatively, run locally
go run cmd/app/main.go
```

## Usage

1. Create a new film project
2. Upload key documents (scripts, treatments, schedules, etc.)
3. Organize documents into categories
4. Share access with team members
5. Track document versions as your project develops
