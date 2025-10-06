# Echoforge Architecture Diagrams

This document contains visual representations of Echoforge's architecture using Mermaid diagrams to help developers understand the system structure and data flow.

## Hexagonal Architecture Overview

The following diagram illustrates Echoforge's hexagonal (ports and adapters) architecture pattern:

```mermaid
graph TB
    subgraph "External Actors"
        User[👤 User/Client]
        DB[(🗄️ PostgreSQL Database)]
        Cache[(🔄 Redis Cache)]
        FileStorage[📁 File Storage]
        EmailService[📧 Email Service]
        Analytics[📊 Analytics Service]
    end
    
    subgraph "Adapters Layer (Infrastructure)"
        subgraph "Inbound Adapters"
            HTTPAdapter[🌐 HTTP/REST Adapter<br/>Gin Router]
            GraphQLAdapter[📊 GraphQL Adapter]
            CLIAdapter[💻 CLI Adapter]
        end
        
        subgraph "Outbound Adapters"
            DatabaseAdapter[🗄️ Database Adapter<br/>GORM Repository]
            CacheAdapter[🔄 Cache Adapter<br/>Redis Client]
            StorageAdapter[📁 Storage Adapter<br/>S3/Local Files]
            EmailAdapter[📧 Email Adapter<br/>SMTP Client]
            AnalyticsAdapter[📊 Analytics Adapter<br/>External APIs]
        end
    end
    
    subgraph "Core Domain (Business Logic)"
        subgraph "Application Layer"
            UserUseCase[👤 User Use Cases<br/>Registration, Authentication]
            PostUseCase[📝 Post Use Cases<br/>CRUD, Publishing]
            ConfigUseCase[⚙️ Config Use Cases<br/>Site Management]
            AuthUseCase[🔐 Auth Use Cases<br/>JWT, Sessions]
        end
        
        subgraph "Domain Layer"
            UserEntity[👤 User Entity<br/>ID, Email, Profile]
            PostEntity[📝 Post Entity<br/>Title, Content, Metadata]
            ConfigEntity[⚙️ Config Entity<br/>Site Settings]
            AuthEntity[🔐 Auth Entity<br/>Tokens, Permissions]
        end
        
        subgraph "Ports (Interfaces)"
            UserPort[👤 User Repository Port]
            PostPort[📝 Post Repository Port]
            ConfigPort[⚙️ Config Repository Port]
            CachePort[🔄 Cache Port]
            StoragePort[📁 Storage Port]
            EmailPort[📧 Email Port]
        end
    end
    
    %% Connections
    User --> HTTPAdapter
    User --> GraphQLAdapter
    User --> CLIAdapter
    
    HTTPAdapter --> UserUseCase
    HTTPAdapter --> PostUseCase
    HTTPAdapter --> ConfigUseCase
    HTTPAdapter --> AuthUseCase
    
    GraphQLAdapter --> UserUseCase
    GraphQLAdapter --> PostUseCase
    
    CLIAdapter --> ConfigUseCase
    
    UserUseCase --> UserEntity
    PostUseCase --> PostEntity
    ConfigUseCase --> ConfigEntity
    AuthUseCase --> AuthEntity
    
    UserUseCase --> UserPort
    PostUseCase --> PostPort
    ConfigUseCase --> ConfigPort
    UserUseCase --> CachePort
    PostUseCase --> StoragePort
    AuthUseCase --> EmailPort
    
    UserPort --> DatabaseAdapter
    PostPort --> DatabaseAdapter
    ConfigPort --> DatabaseAdapter
    CachePort --> CacheAdapter
    StoragePort --> StorageAdapter
    EmailPort --> EmailAdapter
    
    DatabaseAdapter --> DB
    CacheAdapter --> Cache
    StorageAdapter --> FileStorage
    EmailAdapter --> EmailService
    AnalyticsAdapter --> Analytics
    
    %% Styling
    classDef external fill:#e1f5fe,stroke:#0277bd,stroke-width:2px
    classDef adapter fill:#f3e5f5,stroke:#7b1fa2,stroke-width:2px
    classDef usecase fill:#e8f5e8,stroke:#2e7d32,stroke-width:2px
    classDef entity fill:#fff3e0,stroke:#ef6c00,stroke-width:2px
    classDef port fill:#fce4ec,stroke:#c2185b,stroke-width:2px
    
    class User,DB,Cache,FileStorage,EmailService,Analytics external
    class HTTPAdapter,GraphQLAdapter,CLIAdapter,DatabaseAdapter,CacheAdapter,StorageAdapter,EmailAdapter,AnalyticsAdapter adapter
    class UserUseCase,PostUseCase,ConfigUseCase,AuthUseCase usecase
    class UserEntity,PostEntity,ConfigEntity,AuthEntity entity
    class UserPort,PostPort,ConfigPort,CachePort,StoragePort,EmailPort port
```

## Multi-Tenant Data Flow

This diagram shows how multi-tenant isolation works throughout the system:

```mermaid
sequenceDiagram
    participant Client as 👤 Client App
    participant Router as 🌐 Gin Router
    participant Middleware as 🔒 Site Middleware
    participant Handler as 📋 HTTP Handler
    participant UseCase as 🎯 Use Case
    participant Repository as 🗄️ Repository
    participant Database as 💾 PostgreSQL
    
    Client->>Router: GET /api/v1/posts
    Note over Client,Router: Request with site identifier
    
    Router->>Middleware: Process request
    Middleware->>Middleware: Extract site_id from header/subdomain
    Middleware->>Handler: Request with site context
    
    Handler->>UseCase: Call business logic
    Note over Handler,UseCase: site_id passed in context
    
    UseCase->>Repository: Query with site isolation
    Note over UseCase,Repository: All queries include site_id filter
    
    Repository->>Database: SELECT * FROM posts WHERE site_id = 'blog-001'
    Note over Repository,Database: Automatic tenant isolation
    
    Database-->>Repository: Filtered results
    Repository-->>UseCase: Domain entities
    UseCase-->>Handler: Business objects
    Handler-->>Router: JSON response
    Router-->>Client: Site-specific data
    
    Note over Client,Database: Multi-tenant isolation maintained at every layer
```

## Configuration Management Flow

This diagram illustrates how Echoforge handles multi-site configuration:

```mermaid
flowchart TD
    Start([🚀 Application Start]) --> LoadConfig[📄 Load Base Config]
    LoadConfig --> CheckSiteID{🔍 Site ID Present?}
    
    CheckSiteID -->|Yes| LoadSiteConfig[📋 Load Site-Specific Config]
    CheckSiteID -->|No| DefaultSite[🌐 Use Default Site]
    
    LoadSiteConfig --> MergeConfigs[🔄 Merge Configurations]
    DefaultSite --> MergeConfigs
    
    MergeConfigs --> ValidateConfig[✅ Validate Configuration]
    ValidateConfig --> ConfigValid{Valid?}
    
    ConfigValid -->|No| ConfigError[❌ Configuration Error]
    ConfigValid -->|Yes| InitDB[🗄️ Initialize Database]
    
    InitDB --> RunMigrations[🔧 Run DB Migrations]
    RunMigrations --> InitServices[⚙️ Initialize Services]
    
    InitServices --> StartServer[🌐 Start HTTP Server]
    StartServer --> Ready([✅ Server Ready])
    
    ConfigError --> Exit([❌ Exit])
    
    subgraph "Configuration Sources"
        BaseConfig[📄 config.yaml<br/>Default settings]
        SiteConfig[📋 site-specific.yaml<br/>Override settings]
        EnvVars[🌍 Environment Variables<br/>Runtime overrides]
    end
    
    LoadConfig --> BaseConfig
    LoadSiteConfig --> SiteConfig
    MergeConfigs --> EnvVars
    
    %% Styling
    classDef process fill:#e8f5e8,stroke:#2e7d32,stroke-width:2px
    classDef decision fill:#fff3e0,stroke:#ef6c00,stroke-width:2px
    classDef error fill:#ffebee,stroke:#c62828,stroke-width:2px
    classDef success fill:#e1f5fe,stroke:#0277bd,stroke-width:2px
    classDef config fill:#f3e5f5,stroke:#7b1fa2,stroke-width:2px
    
    class LoadConfig,LoadSiteConfig,MergeConfigs,ValidateConfig,InitDB,RunMigrations,InitServices,StartServer process
    class CheckSiteID,ConfigValid decision
    class ConfigError,Exit error
    class Start,Ready success
    class BaseConfig,SiteConfig,EnvVars config
```

## Site Extension Architecture

This diagram shows how new site types can be added to Echoforge:

```mermaid
graph LR
    subgraph "Core Echoforge"
        CoreConfig[⚙️ Core Configuration]
        CoreAuth[🔐 Core Authentication]
        CoreDB[🗄️ Core Database Layer]
        CoreHTTP[🌐 Core HTTP Framework]
    end
    
    subgraph "Site Extensions"
        subgraph "Blog Extension"
            BlogDomain[📝 Blog Domain<br/>Post, Category, Tag]
            BlogRepository[📚 Blog Repository<br/>Multi-tenant queries]
            BlogHandlers[🎯 Blog API Handlers<br/>/api/v1/blog/*]
            BlogConfig[📄 blog-site.yaml]
        end
        
        subgraph "Manga Extension"
            MangaDomain[📖 Manga Domain<br/>Series, Chapter, Page]
            MangaRepository[📚 Manga Repository<br/>Multi-tenant queries]
            MangaHandlers[🎯 Manga API Handlers<br/>/api/v1/manga/*]
            MangaConfig[📄 manga-site.yaml]
        end
        
        subgraph "Portfolio Extension"
            PortfolioDomain[🎨 Portfolio Domain<br/>Project, Skill, Experience]
            PortfolioRepository[📚 Portfolio Repository<br/>Multi-tenant queries]
            PortfolioHandlers[🎯 Portfolio API Handlers<br/>/api/v1/portfolio/*]
            PortfolioConfig[📄 portfolio-site.yaml]
        end
    end
    
    subgraph "Extension Points"
        DomainInterface[🔌 Domain Interface<br/>Repository patterns]
        ConfigInterface[🔌 Config Interface<br/>Site-specific settings]
        HandlerInterface[🔌 Handler Interface<br/>Route registration]
        MigrationInterface[🔌 Migration Interface<br/>Schema management]
    end
    
    %% Core connections
    CoreConfig --> ConfigInterface
    CoreAuth --> HandlerInterface
    CoreDB --> DomainInterface
    CoreHTTP --> HandlerInterface
    
    %% Blog connections
    BlogConfig --> ConfigInterface
    BlogDomain --> DomainInterface
    BlogRepository --> DomainInterface
    BlogHandlers --> HandlerInterface
    
    %% Manga connections
    MangaConfig --> ConfigInterface
    MangaDomain --> DomainInterface
    MangaRepository --> DomainInterface
    MangaHandlers --> HandlerInterface
    
    %% Portfolio connections
    PortfolioConfig --> ConfigInterface
    PortfolioDomain --> DomainInterface
    PortfolioRepository --> DomainInterface
    PortfolioHandlers --> HandlerInterface
    
    %% Styling
    classDef core fill:#e1f5fe,stroke:#0277bd,stroke-width:2px
    classDef blog fill:#e8f5e8,stroke:#2e7d32,stroke-width:2px
    classDef manga fill:#fff3e0,stroke:#ef6c00,stroke-width:2px
    classDef portfolio fill:#f3e5f5,stroke:#7b1fa2,stroke-width:2px
    classDef interface fill:#fce4ec,stroke:#c2185b,stroke-width:2px
    
    class CoreConfig,CoreAuth,CoreDB,CoreHTTP core
    class BlogDomain,BlogRepository,BlogHandlers,BlogConfig blog
    class MangaDomain,MangaRepository,MangaHandlers,MangaConfig manga
    class PortfolioDomain,PortfolioRepository,PortfolioHandlers,PortfolioConfig portfolio
    class DomainInterface,ConfigInterface,HandlerInterface,MigrationInterface interface
```

## Database Schema Relationships

This diagram shows the multi-tenant database structure:

```mermaid
erDiagram
    USERS {
        uuid id PK
        string site_id FK
        string email
        string password_hash
        timestamp created_at
        timestamp updated_at
    }
    
    POSTS {
        uuid id PK
        string site_id FK
        uuid author_id FK
        string title
        string slug
        text content
        string status
        timestamp published_at
        timestamp created_at
        timestamp updated_at
    }
    
    CATEGORIES {
        uuid id PK
        string site_id FK
        string name
        string slug
        uuid parent_id FK
        timestamp created_at
    }
    
    TAGS {
        uuid id PK
        string site_id FK
        string name
        string slug
        timestamp created_at
    }
    
    POST_CATEGORIES {
        uuid post_id FK
        uuid category_id FK
    }
    
    POST_TAGS {
        uuid post_id FK
        uuid tag_id FK
    }
    
    PROJECTS {
        uuid id PK
        string site_id FK
        uuid owner_id FK
        string title
        string slug
        text description
        string category
        string status
        timestamp created_at
    }
    
    PROJECT_IMAGES {
        uuid id PK
        uuid project_id FK
        string file_name
        string original_url
        string thumbnail_url
        int sort_order
        timestamp created_at
    }
    
    MANGA_SERIES {
        uuid id PK
        string site_id FK
        string title
        string slug
        text description
        string status
        timestamp created_at
    }
    
    MANGA_CHAPTERS {
        uuid id PK
        uuid series_id FK
        int chapter_number
        string title
        timestamp published_at
        timestamp created_at
    }
    
    MANGA_PAGES {
        uuid id PK
        uuid chapter_id FK
        int page_number
        string image_url
        timestamp created_at
    }
    
    %% Relationships
    USERS ||--o{ POSTS : "authors"
    POSTS ||--o{ POST_CATEGORIES : "belongs_to"
    CATEGORIES ||--o{ POST_CATEGORIES : "contains"
    POSTS ||--o{ POST_TAGS : "tagged_with"
    TAGS ||--o{ POST_TAGS : "applied_to"
    CATEGORIES ||--o{ CATEGORIES : "parent_child"
    
    USERS ||--o{ PROJECTS : "owns"
    PROJECTS ||--o{ PROJECT_IMAGES : "contains"
    
    MANGA_SERIES ||--o{ MANGA_CHAPTERS : "contains"
    MANGA_CHAPTERS ||--o{ MANGA_PAGES : "contains"
    
    %% Site isolation (all tables have site_id)
    USERS ||--|| POSTS : "site_id"
    USERS ||--|| CATEGORIES : "site_id"
    USERS ||--|| TAGS : "site_id"
    USERS ||--|| PROJECTS : "site_id"
    USERS ||--|| MANGA_SERIES : "site_id"
```

## Request Processing Pipeline

This diagram shows how requests flow through the system:

```mermaid
flowchart TD
    Request[📥 Incoming Request] --> LoadBalancer[⚖️ Load Balancer]
    LoadBalancer --> Server1[🖥️ Server Instance 1]
    LoadBalancer --> Server2[🖥️ Server Instance 2]
    LoadBalancer --> Server3[🖥️ Server Instance 3]
    
    Server1 --> Middleware{🔒 Middleware Pipeline}
    Server2 --> Middleware
    Server3 --> Middleware
    
    Middleware --> CORS[🌐 CORS Handler]
    CORS --> RateLimit[🚦 Rate Limiter]
    RateLimit --> SiteResolver[🏷️ Site ID Resolver]
    SiteResolver --> Auth[🔐 Authentication]
    Auth --> Logger[📝 Request Logger]
    
    Logger --> Router[🎯 Route Handler]
    
    Router --> BlogAPI[📝 Blog API]
    Router --> MangaAPI[📖 Manga API]
    Router --> PortfolioAPI[🎨 Portfolio API]
    Router --> UserAPI[👤 User API]
    
    BlogAPI --> BlogUseCase[📝 Blog Use Case]
    MangaAPI --> MangaUseCase[📖 Manga Use Case]
    PortfolioAPI --> PortfolioUseCase[🎨 Portfolio Use Case]
    UserAPI --> UserUseCase[👤 User Use Case]
    
    BlogUseCase --> BlogRepo[📚 Blog Repository]
    MangaUseCase --> MangaRepo[📚 Manga Repository]
    PortfolioUseCase --> PortfolioRepo[📚 Portfolio Repository]
    UserUseCase --> UserRepo[📚 User Repository]
    
    BlogRepo --> Cache{🔄 Check Cache}
    MangaRepo --> Cache
    PortfolioRepo --> Cache
    UserRepo --> Cache
    
    Cache -->|Hit| CacheReturn[💨 Return Cached Data]
    Cache -->|Miss| Database[(🗄️ PostgreSQL)]
    
    Database --> CacheStore[💾 Store in Cache]
    CacheStore --> Response[📤 HTTP Response]
    CacheReturn --> Response
    
    Response --> Client[👤 Client]
    
    %% Error handling
    Router --> ErrorHandler[❌ Error Handler]
    ErrorHandler --> ErrorLog[📋 Error Logging]
    ErrorLog --> ErrorResponse[📤 Error Response]
    ErrorResponse --> Client
    
    %% Styling
    classDef middleware fill:#e8f5e8,stroke:#2e7d32,stroke-width:2px
    classDef api fill:#fff3e0,stroke:#ef6c00,stroke-width:2px
    classDef storage fill:#e1f5fe,stroke:#0277bd,stroke-width:2px
    classDef error fill:#ffebee,stroke:#c62828,stroke-width:2px
    
    class CORS,RateLimit,SiteResolver,Auth,Logger middleware
    class BlogAPI,MangaAPI,PortfolioAPI,UserAPI api
    class Cache,Database,CacheStore storage
    class ErrorHandler,ErrorLog,ErrorResponse error
```

These diagrams provide a comprehensive visual understanding of Echoforge's architecture, showing how the hexagonal pattern enables clean separation of concerns, how multi-tenancy is maintained throughout the system, and how the platform can be extended with new site types while maintaining architectural integrity.