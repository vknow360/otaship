# OTAShip

OTAShip is an open-source, self-hostable alternative to EAS Updates. It provides a full implementation of the Expo Updates protocol, allowing developers to deploy over-the-air (OTA) updates to React Native applications without relying on third-party cloud services.

The system is designed for high performance and data integrity, utilizing a Go-based backend and PostgreSQL for metadata management.

## Project Structure

The repository is organized into four main components:

*   **backend/**: A Go service built with the Chi router and SQLC. It handles manifest generation, asset management, and project coordination.
*   **cli/**: A Cobra-based command-line interface for managing projects, zipping assets, and publishing updates.
*   **admin-dashboard/**: A SvelteKit application for visualizing release history, managing API keys, and monitoring storage usage.
*   **expo-client/**: A reference implementation showing how to integrate the Expo Updates client with an OTAShip instance.

## Core Features

*   **Native Protocol Support**: Full compatibility with the modern Expo Updates protocol (multipart/mixed manifest support).
*   **Storage Abstraction**: Support for multiple storage providers including Cloudinary and AWS S3 via a unified provider interface.
*   **Security**: Per-project API key authentication and support for RSA manifest signing.
*   **Database Reliability**: Relational data integrity using PostgreSQL with automated migrations.
*   **Detailed Analytics**: Event-sourced download tracking and storage usage monitoring.

## Getting Started

Refer to the individual README files in each directory for specific setup instructions:

1.  **Backend Setup**: Configure your PostgreSQL database and environment variables in the `backend/` directory.
2.  **CLI Configuration**: Install the OTAShip CLI to manage your projects and publish releases.
3.  **Dashboard Deployment**: Deploy the SvelteKit dashboard to manage your OTAShip instance through a web interface.

## Technical Architecture

*   **Language**: Go (Backend/CLI), JavaScript/Svelte (Dashboard)
*   **Database**: PostgreSQL 16+
*   **Router**: Chi (Backend)
*   **Client Communication**: Expo Updates Protocol 0/1

## License

This project is licensed under the Apache License 2.0. See the LICENSE file for details.
