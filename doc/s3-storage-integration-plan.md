# S3-Compatible Storage Integration Plan

## Overview
Add MinIO (S3-compatible) object storage for user-uploaded files (avatars, banners, content) with focus on simplicity, loose coupling, and maintainability.

## Architecture Design

### 1. Storage Layer (Loosely Coupled)
- **Interface-based design** in `internal/domain/storage.go`
- **MinIO implementation** in `internal/storage/minio.go`
- **File metadata tracking** in database (`uploaded_files` table)

### 2. Upload Service Layer
- **Upload service** in `internal/uploads/` (separate domain)
  - Validation, sanitization, image processing
  - Generates thumbnails/variants
  - Stores metadata in DB
  - Delegates actual storage to `storage.FileStorage` interface

### 3. File Types Supported
- **Images**: JPEG, PNG, WebP (avatars, banners)
- **Future extensibility**: Easy to add documents, videos, etc.

## Implementation Steps

### Phase 1: Foundation (Configuration & Storage Interface)
1. Add MinIO config to `internal/config/config.go`
2. Define `FileStorage` interface in `internal/domain/storage.go`
3. Implement MinIO client in `internal/storage/minio.go`
4. Add dependency: `github.com/minio/minio-go/v7`

### Phase 2: Database & Domain Models
1. Create migration `006_uploaded_files.sql`
2. Add `UploadedFile` domain model
3. Create uploads repository for metadata tracking

### Phase 3: Upload Service
1. Create `internal/uploads/` package with:
   - Service (validation, processing, storage orchestration)
   - Image processor (resize, format conversion, EXIF stripping)
   - Validator (MIME type, size limits, security)
2. Loosely coupled - depends only on `domain.FileStorage` interface

### Phase 4: HTTP Handlers
1. `POST /api/profile/avatar` - upload avatar
2. `POST /api/profile/banner` - upload banner
3. `GET /uploads/{type}/{profile-id}/{variant}/{filename}` - serve files via presigned URLs

### Phase 5: Integration
1. Wire up MinIO in `internal/app/app.go`
2. Initialize upload service
3. Add upload handlers to routes
4. Update profile service to handle avatar/banner URLs

## Key Design Decisions

### Loose Coupling Strategy
```
Upload Handler
    ↓
Upload Service (business logic)
    ↓
domain.FileStorage interface
    ↓
storage.MinIO implementation
```
- Easy to swap MinIO for local FS, S3, GCS, etc.
- Upload service doesn't know about MinIO specifics

### Security Measures
- Magic byte validation (not just extension)
- Size limits (5MB avatar, 10MB banner)
- Whitelist JPEG/PNG/WebP only
- Automatic image re-encoding (removes exploits)
- EXIF data stripping (privacy)
- UUID-based filenames (no user input)
- Rate limiting on upload endpoints

### File Organization
```
Buckets:
- circles-avatars/
  ├── {profile-id}/
  │   ├── original/
  │   ├── large/ (512x512)
  │   └── thumb/ (128x128)
- circles-banners/
  ├── {profile-id}/
  │   ├── original/
  │   └── optimized/ (1500x500)
- circles-content/  (future)
```

### Maintainability
- Clear separation of concerns
- Interface-based design (testable, swappable)
- Comprehensive logging
- Error handling at each layer
- Configuration via env vars or YAML
- Self-contained upload package

## Files to Create/Modify

**Create:**
- `internal/config/config.go` - Add `StorageConfig`
- `internal/domain/storage.go` - `FileStorage` interface
- `internal/domain/uploaded_file.go` - Domain model
- `internal/storage/minio.go` - MinIO implementation
- `internal/uploads/service.go` - Upload orchestration
- `internal/uploads/validator.go` - File validation
- `internal/uploads/processor.go` - Image processing
- `internal/uploads/repository.go` - Metadata persistence
- `internal/handlers/upload_handler.go` - HTTP handlers
- `internal/storage/migrations/006_uploaded_files.sql` - Migration

**Modify:**
- `go.mod` - Add MinIO SDK + image processing libs
- `internal/app/app.go` - Initialize storage & upload service
- `cmd/server/main.go` - Register upload routes
- `config.yaml` - Add storage configuration

## Configuration Example
```yaml
storage:
  provider: "minio"  # Future: s3, gcs, local
  endpoint: "localhost:9000"
  access_key: "minioadmin"
  secret_key: "minioadmin"
  use_ssl: false
  region: "us-east-1"
  buckets:
    avatars: "circles-avatars"
    banners: "circles-banners"
    content: "circles-content"
```

## Current Architecture Analysis

### Configuration Pattern
- Uses YAML config with env var overrides
- Struct-based config in `internal/config/config.go`
- Pattern: `Config` struct with nested configs (ServerConfig, DatabaseConfig, etc.)
- Validation in `Validate()` method
- Helper methods like `IsDevelopment()`, `DatabaseDSN()`

### Storage Pattern
- Storage implementations in `internal/storage/`
- Examples: `postgres.go`, `redis.go`
- Pattern: Struct with connection/pool + logger
- Constructor: `New{Service}(ctx, config, logger)` returns struct + error
- Health check method for each storage
- Connection initialized in `internal/app/app.go`

### Service Pattern
- Domain services in `internal/{domain}/`
- Pattern: Service struct with Repository + logger
- Constructor: `NewService(repo, logger)`
- Thin service layer with business logic
- Repository pattern for data access

### Dependencies
- Already using: `pgx/v5`, `go-redis/v9`, `zap`, `yaml.v3`
- No image processing libs yet
- No S3/MinIO SDK yet

## Implementation Details

### FileStorage Interface
```go
// internal/domain/storage.go
type FileStorage interface {
    // Upload stores a file and returns the storage path
    Upload(ctx context.Context, bucket, key string, data io.Reader, size int64, contentType string) (string, error)

    // Download retrieves a file
    Download(ctx context.Context, bucket, key string) (io.ReadCloser, error)

    // Delete removes a file
    Delete(ctx context.Context, bucket, key string) error

    // GetURL returns a presigned URL for accessing the file
    GetURL(ctx context.Context, bucket, key string, expiry time.Duration) (string, error)

    // Exists checks if a file exists
    Exists(ctx context.Context, bucket, key string) (bool, error)
}
```

### Database Schema
```sql
-- internal/storage/migrations/006_uploaded_files.sql
CREATE TABLE uploaded_files (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    profile_id UUID NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    file_type VARCHAR(20) NOT NULL, -- 'avatar', 'banner', 'content'
    variant VARCHAR(20), -- 'original', 'large', 'thumb', 'optimized'
    storage_bucket VARCHAR(100) NOT NULL,
    storage_key TEXT NOT NULL,
    original_filename VARCHAR(255),
    mime_type VARCHAR(100) NOT NULL,
    size_bytes INTEGER NOT NULL,
    width INTEGER,
    height INTEGER,
    uploaded_at TIMESTAMP NOT NULL DEFAULT NOW(),

    UNIQUE(profile_id, file_type, variant)
);

CREATE INDEX idx_uploaded_files_profile_id ON uploaded_files(profile_id);
CREATE INDEX idx_uploaded_files_type ON uploaded_files(file_type);
```

### MinIO Configuration
```go
// internal/config/config.go additions
type StorageConfig struct {
    Provider   string        `yaml:"provider"`    // "minio", "s3", "local"
    Endpoint   string        `yaml:"endpoint"`
    AccessKey  string        `yaml:"access_key"`
    SecretKey  string        `yaml:"secret_key"`
    UseSSL     bool          `yaml:"use_ssl"`
    Region     string        `yaml:"region"`
    Buckets    BucketConfig  `yaml:"buckets"`
}

type BucketConfig struct {
    Avatars string `yaml:"avatars"`
    Banners string `yaml:"banners"`
    Content string `yaml:"content"`
}
```

### Dependencies to Add
```
go get github.com/minio/minio-go/v7
go get github.com/disintegration/imaging  # Image resizing
go get github.com/h2non/filetype          # MIME detection
```

## Security Implementation Details

### File Validation Flow
1. Check file size (before reading entire file)
2. Read magic bytes (first 512 bytes)
3. Validate MIME type against whitelist
4. Verify file extension matches MIME type
5. Re-encode image (removes exploits, strips EXIF)
6. Generate UUID filename
7. Store with content-type header

### Rate Limiting
- Upload endpoints: 5 requests/minute per user
- Use existing rate limit middleware
- Add specific limits for upload endpoints

### Image Processing Security
- Always re-encode images (don't trust input)
- Strip EXIF data (privacy)
- Validate dimensions (prevent decompression bombs)
- Generate multiple sizes atomically
- If any variant fails, rollback all

## Testing Strategy

### Unit Tests
- FileStorage interface mock
- Upload service logic
- Validation rules
- Image processor

### Integration Tests
- MinIO test container
- Full upload flow
- Presigned URL generation
- File cleanup on errors

## Deployment Considerations

### MinIO Setup
```bash
# Docker compose addition
services:
  minio:
    image: minio/minio:latest
    command: server /data --console-address ":9001"
    ports:
      - "9000:9000"
      - "9001:9001"
    environment:
      MINIO_ROOT_USER: minioadmin
      MINIO_ROOT_PASSWORD: minioadmin
    volumes:
      - minio_data:/data
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:9000/minio/health/live"]
      interval: 30s
      timeout: 20s
      retries: 3
```

### Bucket Creation
- Auto-create buckets on startup if they don't exist
- Set appropriate bucket policies (public read for avatars/banners)
- Private by default for content

## Future Enhancements

1. **CDN Integration**
   - CloudFlare/BunnyCDN in front of MinIO
   - Cache presigned URLs

2. **Image Optimization**
   - WebP conversion
   - Lazy loading thumbnails
   - Progressive JPEG

3. **Storage Providers**
   - AWS S3 implementation
   - Google Cloud Storage
   - Local filesystem (development)

4. **Content Management**
   - Post attachments
   - File galleries
   - Document uploads (PDF, etc.)

5. **Analytics**
   - Track upload metrics
   - Storage usage per user
   - Bandwidth monitoring
