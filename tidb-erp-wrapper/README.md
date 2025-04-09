# TiDB ERP Wrapper

A high-performance ERP system wrapper built on TiDB, designed to handle procurement and bookkeeping operations at scale. This system provides enterprise-grade features comparable to SAP S/4HANA, leveraging TiDB's distributed SQL capabilities.

## Features

- **Procurement Management**

  - Supplier management
  - Purchase order processing
  - Item tracking
  - Order status management

- **Bookkeeping & Financial Management**
  - Double-entry accounting
  - Journal entry management
  - Account balance tracking
  - Fiscal period management
  - Multi-currency support

## Prerequisites

- Docker and Docker Compose
- Go 1.20 or higher
- TiDB (automatically set up via Docker Compose)

## Setup

1. Clone the repository:

```bash
git clone https://github.com/yourusername/tidb-erp-wrapper.git
cd tidb-erp-wrapper
```

2. Start TiDB using Docker Compose:

```bash
docker-compose up -d
```

3. Install Go dependencies:

```bash
go mod tidy
```

4. Configure the application:

- The default configuration is in `config/config.yaml`
- Adjust database connection settings if needed

5. Run the application:

```bash
go run cmd/server/main.go
```

The server will start on port 8080 (default) or the port specified in your config.

## API Endpoints

### Procurement

- `POST /suppliers` - Create a new supplier
- `POST /purchase-orders` - Create a purchase order
- `GET /purchase-orders/:id` - Get purchase order details

### Bookkeeping

- `POST /accounts` - Create a new account
- `POST /journal-entries` - Create a journal entry
- `GET /journal-entries/:id` - Get journal entry details
- `GET /accounts/:id/balance` - Get account balance

## Database Schema

The system uses a comprehensive database schema that includes:

- Suppliers and purchase orders for procurement
- Accounts, journal entries, and fiscal periods for bookkeeping
- All tables include audit fields (created_at, updated_at)
- Proper indexing for high-performance queries

## Development

The project follows a clean architecture pattern:

- `cmd/server` - Application entry point
- `config` - Configuration handling
- `internal/db` - Database connection and schema
- `internal/models` - Data models
- `internal/services` - Business logic implementation
- `pkg/utils` - Shared utilities

## Production Considerations

- Set up proper authentication and authorization
- Configure TLS for secure communication
- Implement proper logging and monitoring
- Set up database backups
- Configure appropriate connection pooling

## License

MIT License - See LICENSE file for details
