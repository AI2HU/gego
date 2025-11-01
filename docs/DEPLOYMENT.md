# Deploying Gego on a Server

This guide explains how to deploy Gego on a Server using the pre-built Docker image from Docker Hub.

## Prerequisites

- A Server with Docker and Docker Compose installed
- At least 2GB RAM (4GB recommended)
- Port 8989 available (or configure a different port)

## Quick Start with Docker Compose (Recommended)

The easiest way to deploy Gego is using the provided `docker-compose.yml` file.

### 1. Pull and Start Services

```bash
# Pull the latest image and start all services
docker-compose up -d

# View logs
docker-compose logs -f

# Check service status
docker-compose ps
```

This will:
- Pull the `slals/gego:latest` image from Docker Hub
- Pull the `mongo:latest` image
- Create and start both containers
- Set up persistent volumes for data

### 2. Verify Deployment

```bash
# Check if the API is responding
curl http://localhost:8989/api/v1/health

# Expected response:
# {"status":"ok"}
```

### 3. Access the API

The API will be available at:
- `http://your-vps-ip:8989/api/v1/health`

## Manual Docker Deployment

If you prefer to run containers manually:

### 1. Pull the Image

```bash
docker pull slals/gego:latest
```

### 2. Start MongoDB

```bash
docker run -d \
  --name gego-mongodb \
  --restart unless-stopped \
  -p 27017:27017 \
  -v mongodb_data:/data/db \
  mongo:latest
```

### 3. Create Configuration File

First, create a directory structure on your VPS:

```bash
mkdir -p /opt/gego/{data,config,logs}
```

Create the configuration file at `/opt/gego/config/config.yaml`:

```yaml
sql_database:
  provider: sqlite
  uri: /app/data/gego.db
  database: gego

nosql_database:
  provider: mongodb
  uri: mongodb://gego-mongodb:27017
  database: gego
```

**Note**: If MongoDB is running on the host (not in Docker), use:
```yaml
nosql_database:
  provider: mongodb
  uri: mongodb://your-vps-ip:27017
  database: gego
```

### 4. Run Gego Container

```bash
docker run -d \
  --name gego \
  --restart unless-stopped \
  --link gego-mongodb:mongodb \
  -p 8989:8989 \
  -v /opt/gego/data:/app/data \
  -v /opt/gego/config:/app/config \
  -v /opt/gego/logs:/app/logs \
  -e GEGO_CONFIG_PATH=/app/config/config.yaml \
  -e GEGO_DATA_PATH=/app/data \
  -e GEGO_LOG_PATH=/app/logs \
  slals/gego:latest
```

### 5. Verify Deployment

```bash
# Check container status
docker ps

# View logs
docker logs -f gego

# Test API
curl http://localhost:8989/api/v1/health
```

## Using a Docker Network (Recommended for Manual Deployment)

For better isolation, create a Docker network:

```bash
# Create network
docker network create gego-network

# Start MongoDB on the network
docker run -d \
  --name gego-mongodb \
  --network gego-network \
  --restart unless-stopped \
  -v mongodb_data:/data/db \
  mongo:latest

# Start Gego on the same network
docker run -d \
  --name gego \
  --network gego-network \
  --restart unless-stopped \
  -p 8989:8989 \
  -v /opt/gego/data:/app/data \
  -v /opt/gego/config:/app/config \
  -v /opt/gego/logs:/app/logs \
  -e GEGO_CONFIG_PATH=/app/config/config.yaml \
  -e GEGO_DATA_PATH=/app/data \
  -e GEGO_LOG_PATH=/app/logs \
  slals/gego:latest
```

In your config, use the container name for MongoDB URI:
```yaml
nosql_database:
  provider: mongodb
  uri: mongodb://gego-mongodb:27017
  database: gego
```

## Configuration

### Environment Variables

- `GEGO_CONFIG_PATH`: Path to configuration file (default: `/app/config/config.yaml`)
- `GEGO_DATA_PATH`: Path to SQLite data directory (default: `/app/data`)
- `GEGO_LOG_PATH`: Path to log directory (default: `/app/logs`)

### Config File Format

The configuration file should be located at the path specified in `GEGO_CONFIG_PATH`:

```yaml
sql_database:
  provider: sqlite
  uri: /app/data/gego.db
  database: gego

nosql_database:
  provider: mongodb
  uri: mongodb://gego-mongodb:27017
  database: gego
```

## Initializing Gego

After deployment, you need to initialize Gego:

```bash
# Enter the container
docker exec -it gego sh

# Run init command (interactive)
gego init

# Or configure manually by editing the config file
```

## Managing the Deployment

### View Logs

```bash
# Using docker-compose
docker-compose logs -f gego

# Using docker directly
docker logs -f gego
```

### Stop Services

```bash
# Using docker-compose
docker-compose down

# Using docker directly
docker stop gego gego-mongodb
```

### Update Gego

```bash
# Pull latest image
docker pull slals/gego:latest

# Stop current container
docker stop gego

# Remove old container
docker rm gego

# Start new container (same command as before)
docker run -d \
  --name gego \
  --restart unless-stopped \
  ...
```

Or with docker-compose:
```bash
docker-compose pull
docker-compose up -d
```

### Backup Data

```bash
# Backup MongoDB data
docker exec gego-mongodb mongodump --out /data/backup

# Backup Gego data volume
docker run --rm \
  -v gego_data:/data \
  -v $(pwd):/backup \
  alpine tar czf /backup/gego-data-backup.tar.gz /data
```

## API Endpoints

Once deployed, the API is available at `http://your-vps-ip:8989/api/v1`:

- `GET /api/v1/health` - Health check
- `GET /api/v1/llms` - List all LLMs
- `POST /api/v1/llms` - Create new LLM
- `GET /api/v1/prompts` - List all prompts
- `POST /api/v1/prompts` - Create new prompt
- `GET /api/v1/schedules` - List all schedules
- `POST /api/v1/schedules` - Create new schedule
- `GET /api/v1/stats` - Get statistics
- `POST /api/v1/search` - Search responses

## Troubleshooting

### Container won't start

```bash
# Check logs
docker logs gego

# Check if MongoDB is accessible
docker exec gego ping gego-mongodb
```

### Can't connect to MongoDB

1. Verify MongoDB is running: `docker ps | grep mongo`
2. Check network connectivity between containers
3. Verify MongoDB URI in config file matches container name
4. Check MongoDB logs: `docker logs gego-mongodb`

### Port already in use

If port 8989 is already in use, change it:

```bash
# Edit docker-compose.yml and change port mapping
ports:
  - "3000:8989"  # External:Internal

# Or use --publish flag with docker run
-p 3000:8989
```

### Data persistence

Ensure volumes are properly mounted:
```bash
docker volume ls
docker volume inspect gego_data
```

## Security Considerations

1. **Firewall**: Only expose port 8989 if needed, or use a reverse proxy
2. **MongoDB**: Don't expose MongoDB port (27017) publicly if using external MongoDB
3. **API Access**: Consider adding authentication/authorization if exposing publicly
4. **Updates**: Regularly update the Docker image for security patches

## Reverse Proxy (Optional)

To run behind Nginx or another reverse proxy:

```nginx
server {
    listen 80;
    server_name your-domain.com;

    location / {
        proxy_pass http://localhost:8989;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

