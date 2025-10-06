# Data Flow Diagrams

This document illustrates how data flows through Echoforge's multi-tenant architecture for different site types and operations.

## Overall System Data Flow

This high-level diagram shows the complete data journey from client request to database response:

```mermaid
flowchart TB
    subgraph "Client Layer"
        WebApp[🌐 Web Application]
        MobileApp[📱 Mobile App]
        CLI[💻 CLI Tool]
        PostmanAPI[📮 API Testing]
    end
    
    subgraph "Load Balancer & CDN"
        LB[⚖️ Load Balancer]
        CDN[🌍 CDN<br/>Static Assets]
    end
    
    subgraph "Application Servers"
        direction TB
        subgraph "Server Instance 1"
            App1[🚀 Echoforge App]
            Cache1[🔄 Local Cache]
        end
        subgraph "Server Instance 2"
            App2[🚀 Echoforge App]
            Cache2[🔄 Local Cache]
        end
        subgraph "Server Instance N"
            AppN[🚀 Echoforge App]
            CacheN[🔄 Local Cache]
        end
    end
    
    subgraph "Data Layer"
        Redis[(🔄 Redis Cluster<br/>Distributed Cache)]
        PostgresMain[(🗄️ PostgreSQL Primary<br/>Read/Write)]
        PostgresReplica[(🗄️ PostgreSQL Replica<br/>Read Only)]
        FileStorage[📁 File Storage<br/>S3/MinIO]
    end
    
    subgraph "External Services"
        EmailService[📧 Email Service]
        Analytics[📊 Analytics]
        Monitoring[📈 Monitoring]
    end
    
    %% Data Flow Connections
    WebApp --> LB
    MobileApp --> LB
    CLI --> LB
    PostmanAPI --> LB
    
    LB --> App1
    LB --> App2
    LB --> AppN
    
    CDN <--> FileStorage
    
    App1 <--> Cache1
    App2 <--> Cache2
    AppN <--> CacheN
    
    Cache1 <--> Redis
    Cache2 <--> Redis
    CacheN <--> Redis
    
    App1 --> PostgresMain
    App2 --> PostgresMain
    AppN --> PostgresMain
    
    App1 --> PostgresReplica
    App2 --> PostgresReplica
    AppN --> PostgresReplica
    
    PostgresMain --> PostgresReplica
    
    App1 --> FileStorage
    App2 --> FileStorage
    AppN --> FileStorage
    
    App1 --> EmailService
    App1 --> Analytics
    App1 --> Monitoring
    
    %% Data flow labels
    LB -.->|"Route by site_id<br/>subdomain/header"| App1
    App1 -.->|"Write Operations<br/>CREATE, UPDATE, DELETE"| PostgresMain
    App1 -.->|"Read Operations<br/>SELECT queries"| PostgresReplica
    Redis -.->|"Cache Hit<br/>Fast Response"| App1
    App1 -.->|"Cache Miss<br/>Store Result"| Redis
    
    %% Styling
    classDef client fill:#e1f5fe,stroke:#0277bd,stroke-width:2px
    classDef server fill:#e8f5e8,stroke:#2e7d32,stroke-width:2px
    classDef data fill:#fff3e0,stroke:#ef6c00,stroke-width:2px
    classDef external fill:#f3e5f5,stroke:#7b1fa2,stroke-width:2px
    
    class WebApp,MobileApp,CLI,PostmanAPI client
    class LB,CDN,App1,App2,AppN,Cache1,Cache2,CacheN server
    class Redis,PostgresMain,PostgresReplica,FileStorage data
    class EmailService,Analytics,Monitoring external
```

## Multi-Tenant Request Flow

This diagram shows how site-specific data is isolated and processed:

```mermaid
sequenceDiagram
    participant Client as 👤 Client<br/>blog.example.com
    participant LB as ⚖️ Load Balancer
    participant App as 🚀 Application
    participant SiteMiddleware as 🏷️ Site Resolver
    participant Cache as 🔄 Redis Cache
    participant DB as 🗄️ PostgreSQL
    
    Note over Client,DB: Blog Post Request Flow
    
    Client->>LB: GET /api/v1/posts<br/>Host: blog.example.com
    LB->>App: Route to available instance
    
    App->>SiteMiddleware: Process request
    SiteMiddleware->>SiteMiddleware: Extract site_id from subdomain
    Note over SiteMiddleware: site_id = "blog-001"
    
    SiteMiddleware->>Cache: Check cache<br/>Key: posts:blog-001:page:1
    
    alt Cache Hit
        Cache-->>SiteMiddleware: Return cached posts
        SiteMiddleware-->>Client: HTTP 200 + Posts JSON
    else Cache Miss
        SiteMiddleware->>DB: SELECT * FROM posts<br/>WHERE site_id = 'blog-001'<br/>AND status = 'published'
        DB-->>SiteMiddleware: Return blog posts
        SiteMiddleware->>Cache: Store in cache<br/>TTL: 2 hours
        SiteMiddleware-->>Client: HTTP 200 + Posts JSON
    end
    
    Note over Client,DB: Manga Site Request (Different Tenant)
    
    Client->>LB: GET /api/v1/series<br/>Host: manga.example.com
    LB->>App: Route to available instance
    
    App->>SiteMiddleware: Process request
    SiteMiddleware->>SiteMiddleware: Extract site_id from subdomain
    Note over SiteMiddleware: site_id = "manga-001"
    
    SiteMiddleware->>Cache: Check cache<br/>Key: series:manga-001:page:1
    
    Cache-->>SiteMiddleware: Cache miss
    SiteMiddleware->>DB: SELECT * FROM manga_series<br/>WHERE site_id = 'manga-001'<br/>AND status = 'published'
    DB-->>SiteMiddleware: Return manga series
    SiteMiddleware->>Cache: Store in cache<br/>TTL: 2 hours
    SiteMiddleware-->>Client: HTTP 200 + Series JSON
    
    Note over Client,DB: Data isolation maintained at every layer
```

## Content Publishing Flow

This diagram illustrates how content moves through the publishing pipeline:

```mermaid
flowchart TD
    subgraph "Content Creation"
        Author[✍️ Author/Creator]
        Editor[📝 Rich Text Editor]
        MediaUpload[📸 Media Upload]
    end
    
    subgraph "Content Processing"
        Draft[📄 Draft State]
        Validation[✅ Content Validation]
        MediaProcess[🖼️ Media Processing]
        Preview[👁️ Preview Generation]
    end
    
    subgraph "Publishing Pipeline"
        PublishQueue[📤 Publish Queue]
        SEOOptimization[🔍 SEO Optimization]
        CacheWarming[🔥 Cache Warming]
        IndexUpdate[📊 Search Index Update]
    end
    
    subgraph "Content Delivery"
        CDN[🌍 CDN Distribution]
        Cache[🔄 Edge Caching]
        API[🚀 API Response]
    end
    
    subgraph "Storage & Database"
        Database[(🗄️ PostgreSQL<br/>Content metadata)]
        FileStorage[(📁 File Storage<br/>Media files)]
        SearchDB[(🔍 Search Database<br/>Full-text index)]
    end
    
    %% Content Creation Flow
    Author --> Editor
    Author --> MediaUpload
    
    Editor --> Draft
    MediaUpload --> MediaProcess
    
    Draft --> Validation
    MediaProcess --> Validation
    
    Validation -->|Valid| Preview
    Validation -->|Invalid| Editor
    
    Preview -->|Publish| PublishQueue
    Preview -->|Save Draft| Database
    
    %% Publishing Flow
    PublishQueue --> SEOOptimization
    SEOOptimization --> Database
    
    PublishQueue --> CacheWarming
    CacheWarming --> Cache
    
    PublishQueue --> IndexUpdate
    IndexUpdate --> SearchDB
    
    Database --> CDN
    FileStorage --> CDN
    
    %% Content Delivery
    CDN --> API
    Cache --> API
    SearchDB --> API
    
    %% Data Labels
    Draft -.->|"Auto-save<br/>Every 30s"| Database
    MediaProcess -.->|"Resize, Optimize<br/>Generate Thumbnails"| FileStorage
    SEOOptimization -.->|"Meta tags<br/>Structured data"| Database
    CacheWarming -.->|"Preload popular<br/>content"| Cache
    
    %% Styling
    classDef creation fill:#e8f5e8,stroke:#2e7d32,stroke-width:2px
    classDef processing fill:#fff3e0,stroke:#ef6c00,stroke-width:2px
    classDef publishing fill:#f3e5f5,stroke:#7b1fa2,stroke-width:2px
    classDef delivery fill:#e1f5fe,stroke:#0277bd,stroke-width:2px
    classDef storage fill:#fce4ec,stroke:#c2185b,stroke-width:2px
    
    class Author,Editor,MediaUpload creation
    class Draft,Validation,MediaProcess,Preview processing
    class PublishQueue,SEOOptimization,CacheWarming,IndexUpdate publishing
    class CDN,Cache,API delivery
    class Database,FileStorage,SearchDB storage
```

## User Authentication Flow

This diagram shows how authentication data flows through the system:

```mermaid
sequenceDiagram
    participant User as 👤 User
    participant Client as 💻 Client App
    participant API as 🚀 API Gateway
    participant Auth as 🔐 Auth Service
    participant Cache as 🔄 Redis
    participant DB as 🗄️ Database
    participant Email as 📧 Email Service
    
    Note over User,Email: User Registration Flow
    
    User->>Client: Fill registration form
    Client->>API: POST /api/v1/auth/register<br/>{email, password, site_id}
    
    API->>Auth: Validate registration data
    Auth->>Auth: Hash password with bcrypt
    Auth->>DB: Check if user exists<br/>WHERE email = ? AND site_id = ?
    
    alt User Exists
        DB-->>Auth: User found
        Auth-->>API: HTTP 409 Conflict
        API-->>Client: Registration failed
    else New User
        DB-->>Auth: No user found
        Auth->>DB: INSERT INTO users<br/>(email, password_hash, site_id)
        DB-->>Auth: User created
        
        Auth->>Email: Send verification email
        Auth->>Cache: Store verification token<br/>TTL: 24 hours
        
        Auth-->>API: HTTP 201 Created
        API-->>Client: Registration successful
        Client-->>User: Check email message
    end
    
    Note over User,Email: User Login Flow
    
    User->>Client: Enter credentials
    Client->>API: POST /api/v1/auth/login<br/>{email, password, site_id}
    
    API->>Auth: Validate login data
    Auth->>Cache: Check rate limit<br/>Key: login_attempts:email:site_id
    
    alt Rate Limited
        Cache-->>Auth: Too many attempts
        Auth-->>API: HTTP 429 Too Many Requests
    else Within Limits
        Cache-->>Auth: Attempts within limit
        Auth->>DB: SELECT * FROM users<br/>WHERE email = ? AND site_id = ?
        
        alt User Not Found
            DB-->>Auth: No user found
            Auth-->>API: HTTP 401 Unauthorized
        else User Found
            DB-->>Auth: Return user data
            Auth->>Auth: Verify password with bcrypt
            
            alt Invalid Password
                Auth->>Cache: Increment failed attempts
                Auth-->>API: HTTP 401 Unauthorized
            else Valid Password
                Auth->>Auth: Generate JWT token
                Auth->>Cache: Store session<br/>Key: session:user_id:site_id<br/>TTL: 7 days
                Auth->>DB: UPDATE users SET last_login = NOW()
                
                Auth-->>API: HTTP 200 + JWT token
                API-->>Client: Login successful + token
                Client-->>User: Redirect to dashboard
            end
        end
    end
```

## File Upload and Processing Flow

This diagram shows how media files are processed and stored:

```mermaid
flowchart TD
    subgraph "Client Side"
        User[👤 User]
        FileInput[📁 File Input]
        Progress[📊 Upload Progress]
    end
    
    subgraph "API Layer"
        UploadEndpoint[📤 Upload Endpoint<br/>/api/v1/upload]
        Validation[✅ File Validation<br/>Size, Type, Security]
        TempStorage[📂 Temporary Storage<br/>/tmp/uploads]
    end
    
    subgraph "Processing Pipeline"
        Queue[📋 Processing Queue<br/>Background Jobs]
        
        subgraph "Image Processing"
            Resize[🖼️ Image Resize]
            Thumbnail[🔍 Thumbnail Generation]
            Optimize[⚡ Image Optimization]
            Watermark[💧 Watermark Addition]
        end
        
        subgraph "Video Processing"
            VideoResize[🎬 Video Resize]
            VideoThumbnail[📸 Video Thumbnail]
            VideoCompress[🗜️ Video Compression]
        end
        
        subgraph "Document Processing"
            PDFThumbnail[📄 PDF Thumbnail]
            DocPreview[📋 Document Preview]
            TextExtract[📝 Text Extraction]
        end
    end
    
    subgraph "Storage Layer"
        PrimaryStorage[📁 Primary Storage<br/>S3/MinIO]
        CDN[🌍 CDN Distribution]
        Database[(🗄️ File Metadata<br/>PostgreSQL)]
        SearchIndex[(🔍 Search Index<br/>File content)]
    end
    
    subgraph "Delivery"
        API[🚀 File API<br/>Secure URLs]
        DirectAccess[🔗 Direct CDN Access<br/>Public files]
    end
    
    %% Upload Flow
    User --> FileInput
    FileInput --> UploadEndpoint
    UploadEndpoint --> Validation
    
    Validation -->|Valid| TempStorage
    Validation -->|Invalid| User
    
    TempStorage --> Queue
    
    %% Processing Flows
    Queue --> Resize
    Queue --> VideoResize
    Queue --> PDFThumbnail
    
    Resize --> Thumbnail
    Thumbnail --> Optimize
    Optimize --> Watermark
    
    VideoResize --> VideoThumbnail
    VideoThumbnail --> VideoCompress
    
    PDFThumbnail --> DocPreview
    DocPreview --> TextExtract
    
    %% Storage Flow
    Watermark --> PrimaryStorage
    VideoCompress --> PrimaryStorage
    TextExtract --> PrimaryStorage
    
    PrimaryStorage --> CDN
    PrimaryStorage --> Database
    
    TextExtract --> SearchIndex
    
    %% Delivery Flow
    Database --> API
    CDN --> DirectAccess
    SearchIndex --> API
    
    API --> User
    DirectAccess --> User
    
    %% Progress Updates
    Queue -.->|"Job Status<br/>WebSocket"| Progress
    Resize -.->|"Processing<br/>25%"| Progress
    Thumbnail -.->|"Processing<br/>50%"| Progress
    Optimize -.->|"Processing<br/>75%"| Progress
    PrimaryStorage -.->|"Complete<br/>100%"| Progress
    
    %% Error Handling
    Validation -.->|"Size/Type<br/>Error"| User
    Queue -.->|"Processing<br/>Failed"| User
    PrimaryStorage -.->|"Storage<br/>Error"| User
    
    %% Styling
    classDef client fill:#e1f5fe,stroke:#0277bd,stroke-width:2px
    classDef api fill:#e8f5e8,stroke:#2e7d32,stroke-width:2px
    classDef processing fill:#fff3e0,stroke:#ef6c00,stroke-width:2px
    classDef storage fill:#f3e5f5,stroke:#7b1fa2,stroke-width:2px
    classDef delivery fill:#fce4ec,stroke:#c2185b,stroke-width:2px
    
    class User,FileInput,Progress client
    class UploadEndpoint,Validation,TempStorage api
    class Queue,Resize,Thumbnail,Optimize,Watermark,VideoResize,VideoThumbnail,VideoCompress,PDFThumbnail,DocPreview,TextExtract processing
    class PrimaryStorage,CDN,Database,SearchIndex storage
    class API,DirectAccess delivery
```

## Caching Strategy Flow

This diagram illustrates the multi-level caching architecture:

```mermaid
flowchart TD
    subgraph "Request Path"
        Client[👤 Client Request]
        CDN[🌍 CDN Edge Cache<br/>TTL: 24h]
        LB[⚖️ Load Balancer]
        App[🚀 Application Server]
    end
    
    subgraph "Application Caching"
        LocalCache[💾 Local Cache<br/>In-Memory<br/>TTL: 5min]
        Redis[🔄 Redis Cluster<br/>Distributed Cache<br/>TTL: 2h]
    end
    
    subgraph "Database Layer"
        QueryCache[📊 Query Result Cache<br/>TTL: 15min]
        Database[(🗄️ PostgreSQL<br/>Source of Truth)]
    end
    
    subgraph "Cache Keys Strategy"
        SiteSpecific[🏷️ Site-Specific Keys<br/>posts:blog-001:page:1]
        UserSpecific[👤 User-Specific Keys<br/>user:profile:123:blog-001]
        GlobalCache[🌐 Global Keys<br/>config:blog-001]
    end
    
    %% Request Flow
    Client --> CDN
    CDN -->|Cache Miss| LB
    CDN -->|Cache Hit| Client
    
    LB --> App
    App --> LocalCache
    
    LocalCache -->|Hit| App
    LocalCache -->|Miss| Redis
    
    Redis -->|Hit| LocalCache
    Redis -->|Miss| QueryCache
    
    QueryCache -->|Hit| Redis
    QueryCache -->|Miss| Database
    
    Database --> QueryCache
    QueryCache --> Redis
    Redis --> LocalCache
    LocalCache --> App
    App --> LB
    LB --> CDN
    CDN --> Client
    
    %% Cache Strategy
    App --> SiteSpecific
    App --> UserSpecific
    App --> GlobalCache
    
    SiteSpecific --> Redis
    UserSpecific --> Redis
    GlobalCache --> Redis
    
    %% Cache Invalidation
    subgraph "Cache Invalidation"
        DBUpdate[🔄 Database Update]
        InvalidatePattern[🗑️ Invalidate Pattern<br/>posts:blog-001:*]
        InvalidateUser[🗑️ Invalidate User<br/>user:*:blog-001]
        InvalidateGlobal[🗑️ Invalidate Global<br/>config:blog-001]
    end
    
    Database --> DBUpdate
    DBUpdate --> InvalidatePattern
    DBUpdate --> InvalidateUser
    DBUpdate --> InvalidateGlobal
    
    InvalidatePattern --> Redis
    InvalidateUser --> Redis
    InvalidateGlobal --> Redis
    
    %% Performance Metrics
    subgraph "Cache Performance"
        HitRate[📈 Cache Hit Rate<br/>Target: >90%]
        ResponseTime[⚡ Response Time<br/>Target: <100ms]
        MemoryUsage[💾 Memory Usage<br/>Monitor: Redis]
    end
    
    CDN -.-> HitRate
    LocalCache -.-> ResponseTime
    Redis -.-> MemoryUsage
    
    %% Cache Warming
    subgraph "Cache Warming"
        Scheduler[⏰ Background Scheduler]
        PopularContent[🔥 Popular Content<br/>Preload]
        NewContent[🆕 New Content<br/>Publish Event]
    end
    
    Scheduler --> PopularContent
    NewContent --> PopularContent
    PopularContent --> Redis
    PopularContent --> CDN
    
    %% Styling
    classDef request fill:#e1f5fe,stroke:#0277bd,stroke-width:2px
    classDef cache fill:#e8f5e8,stroke:#2e7d32,stroke-width:2px
    classDef database fill:#fff3e0,stroke:#ef6c00,stroke-width:2px
    classDef strategy fill:#f3e5f5,stroke:#7b1fa2,stroke-width:2px
    classDef invalidation fill:#ffebee,stroke:#c62828,stroke-width:2px
    classDef metrics fill:#fce4ec,stroke:#c2185b,stroke-width:2px
    classDef warming fill:#e0f2f1,stroke:#00695c,stroke-width:2px
    
    class Client,CDN,LB,App request
    class LocalCache,Redis,QueryCache cache
    class Database database
    class SiteSpecific,UserSpecific,GlobalCache strategy
    class DBUpdate,InvalidatePattern,InvalidateUser,InvalidateGlobal invalidation
    class HitRate,ResponseTime,MemoryUsage metrics
    class Scheduler,PopularContent,NewContent warming
```

## Search and Analytics Data Flow

This diagram shows how search queries and analytics data flow through the system:

```mermaid
sequenceDiagram
    participant User as 👤 User
    participant SearchAPI as 🔍 Search API
    participant Cache as 🔄 Redis Cache
    participant SearchEngine as 🔍 Search Engine<br/>Elasticsearch
    participant Analytics as 📊 Analytics Service
    participant Database as 🗄️ PostgreSQL
    participant Dashboard as 📈 Analytics Dashboard
    
    Note over User,Dashboard: Search Query Flow
    
    User->>SearchAPI: GET /api/v1/search?q=golang&site_id=blog-001
    SearchAPI->>Cache: Check search cache<br/>Key: search:blog-001:golang:page:1
    
    alt Cache Hit
        Cache-->>SearchAPI: Return cached results
    else Cache Miss
        SearchAPI->>SearchEngine: Query index<br/>Match: golang, Filter: site_id=blog-001
        SearchEngine-->>SearchAPI: Return search results
        SearchAPI->>Cache: Store results<br/>TTL: 30 minutes
    end
    
    SearchAPI->>Analytics: Track search event<br/>{query: "golang", site_id: "blog-001", results: 15}
    SearchAPI-->>User: Return search results
    
    Note over User,Dashboard: Click Tracking Flow
    
    User->>SearchAPI: Click on result #3
    SearchAPI->>Analytics: Track click event<br/>{query: "golang", position: 3, document_id: "post-123"}
    SearchAPI->>Database: Increment view count<br/>UPDATE posts SET view_count = view_count + 1
    
    Note over User,Dashboard: Analytics Processing Flow
    
    Analytics->>Analytics: Process search events<br/>Batch every 5 minutes
    Analytics->>Database: Store search metrics<br/>INSERT INTO search_analytics
    
    Analytics->>Analytics: Calculate CTR<br/>clicks / impressions
    Analytics->>Database: Store aggregated metrics<br/>INSERT INTO daily_metrics
    
    Note over User,Dashboard: Dashboard Query Flow
    
    Dashboard->>Database: Query analytics data<br/>SELECT * FROM search_analytics<br/>WHERE site_id = 'blog-001'
    Database-->>Dashboard: Return metrics data
    Dashboard-->>User: Display search analytics
    
    Note over User,Dashboard: Real-time Updates
    
    Analytics->>Dashboard: WebSocket: New search event
    Dashboard-->>User: Update dashboard in real-time
```

These data flow diagrams provide a comprehensive view of how information moves through Echoforge's architecture, from simple request-response cycles to complex content processing pipelines and real-time analytics systems. Each diagram emphasizes the multi-tenant nature of the platform and how data isolation is maintained throughout all operations.