<div align="center">
    <h1 align="center">Film-Packager</h1>
    <a href="https://film-packager.fly.dev/">Visit Film-Packager</a>
</div>

## About

<p>Film-Packager is a single-page application (SPA) that allows users to manage documents for film projects in development.</p>

## An SPA to manage documents for film projects in development

### Built With

<div align="center">
  <img src="https://img.shields.io/badge/Go-%23000000.svg?&style=for-the-badge&logo=go" alt="Go">
  <img src="https://img.shields.io/badge/PostgreSQL-%23000000.svg?&style=for-the-badge&logo=postgresql&logoColor=%23ffffff" alt="PostgreSQL">
  <img src="https://img.shields.io/badge/AWS_S3-%23000000.svg?&style=for-the-badge&logo=s3&logoColor=%23ffffff" alt="AWS S3">
  <img src="https://img.shields.io/badge/Fly.io-%23000000.svg?&style=for-the-badge&logo=fly.io&logoColor=%23ffffff" alt="Fly">
  <img src="https://img.shields.io/badge/Docker-%23000000.svg?&style=for-the-badge&logo=docker&logoColor=%23ffffff" alt="Docker">
  <img src="https://img.shields.io/badge/HTMX-%23000000.svg?&style=for-the-badge&logo=htmx&logoColor=%23ffffff" alt="HTMX">
</div>

## Features

<ul align="left">
<li><strong>Project Management</strong>: Create and organize film projects</li>
<li><strong>Document Storage</strong>: Upload, categorize, and version control film documents</li>
<li><strong>Collaboration</strong>: Share documents with team members</li>
<li><strong>Search Functionality</strong>: Quickly find documents across projects</li>
<li><strong>Mobile Responsive</strong>: Access your documents from any device</li>
</ul>

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

## Deployment

<div align="left">
<p>The application is deployed on Fly.io with database on managed PostgreSQL and static files on AWS S3.</p>

</div>

## Project Structure

```
film-packager/
├── cmd/                   # Application entrypoint
├── internal/              # Private application code
│   ├── presentation/      # HTTP handlers
│   ├── domain/            # Data models
│   ├── infrastructure/    # Database access
│   └── application/       # Business logic
├── views/                 # Frontend templates
├── static/                # Frontend styles and assets
├── Dockerfile             # Container definition
└── docker-compose.yml     # Local development setup
```

## Contributing

<div align="left">
<p>Contributions are welcome! Please feel free to submit a Pull Request.</p>
</div>

## License

<div align="left">
<p>This project is licensed under the MIT License - see the LICENSE file for details.</p>
</div>
