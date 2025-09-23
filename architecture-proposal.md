# Architecture Proposal: Circles.diy

**Status:** Draft  
**Authors:** [Aaron Lewis](https://github.com/AaronL1011)  
**Reviewers:** [TBD]  
**Date:** 2025-09-23  

## 1. Purpose
This document proposes the core technical architecture for **Circles.diy**, a peer-to-peer–leaning social communication platform. The design aims to maximize:

- **Resilience** – robust against failure and censorship
- **Sovereignty** – user and circle ownership of data and rules
- **Speed** – lightweight, fast-loading UX
- **Accessibility** – minimal client requirements, progressive enhancement

## 2. High-Level Overview
Circles.diy combines a lightweight server-rendered experience with selective peer-to-peer communication:

- **Backend:** Go services, gRPC transport
- **Gateway:** HTTP/HTTPS edge layer bridging browser traffic to gRPC services
- **Client:** Server-rendered HTML templates with htmx for dynamic updates
- **Realtime:** WebRTC for peer-to-peer media/data, SFU for multiparty

The system is **hybrid**: servers handle discovery, federation, policy, and moderation; peers handle real-time presence, messaging, and calls where possible.

## 3. Architecture Components

### 3.1 Gateway and Transport
- Browser communicates via **HTTP(S)** to the gateway.
- Gateway serves **HTML fragments + JSON** for htmx swaps.
- Internal traffic: **gRPC over mTLS**.
- **SSE** provides server-to-browser push for updates.
- **WebSocket/WebTransport** reserved for true bidirectional messaging.
- Optionally: **gRPC-Web or Connect-Web** to expose protobuf contracts directly to browsers.

### 3.2 Realtime and Media
- **WebRTC** for peer-to-peer data channels and media streams.
- Gateway handles **signaling** (SDP + ICE exchange) and issues **short-lived tokens**.
- **STUN/TURN** servers for NAT traversal; TURN fallback expected.
- **SFU (LiveKit, Janus, Jitsi)** used for multiparty sessions.
- Mesh connections only for very small rooms (≤3 participants).

### 3.3 Circles as Federation Units
- A **Circle** is the **trust boundary and federation domain**.
- Each Circle enforces **its own membership, policies, and cryptographic keys**.
- Cross-circle federation is **opt-in** and **policy-gated**.
- Data model aligns with **ActivityPub-like objects** for future interoperability.

### 3.4 Identity and Sessions
- **WebAuthn** for passwordless, phishing-resistant authentication.
- Accounts internally mapped to **Decentralized Identifiers (DIDs)** for portability.
- Gateway issues **HTTP-only session cookies**.
- Realtime joins authenticated via **short-lived tokens** minted at the gateway.

### 3.5 Data Model and Sync
- **Server = source of truth**.
- Clients cache state (timelines, drafts, presence) in **IndexedDB**.
- Endpoints provide **cursor-based “since” reads** for reconciliation after reconnects.
- Collaborative editing and drafts may use **CRDTs (Yjs/Automerge)**.

### 3.6 Security and Encryption
- Mandatory **TLS transport security**.
- **1:1 messaging:** End-to-end encryption (double ratchet style).
- **Groups:** Begin with server-enforced access; migrate toward **Messaging Layer Security (MLS)**.
- **WebRTC media:** Insertable-streams E2EE when possible, with SFU fallback.

### 3.7 Feeds and Delivery
- Feeds composed **server-side** and rendered as HTML partials.
- **Cursor-based APIs** for infinite scroll.
- **SSE nudges** notify clients of new content → client pulls on demand.
- Custom feeds generated **server-side** for consistency and moderation.

### 3.8 Moderation
- Moderation is **per Circle**.
- Gateway enforces:
  - Rate limiting
  - Blocklists
  - Content filters
- Cross-circle federation requires **allowlists** and **quarantine queues**.
- Peer connections rely on **gateway-issued tokens**, ensuring bans/mutes are respected.

### 3.9 Operations
- All external traffic terminates at the **gateway**.
- **SFU and TURN** run as separate, autoscalable components.
- Observability: structured logs + request IDs for end-to-end tracing.
- Traces connect **user interactions → gateway → backend services**.

### 3.10 Data Storage Architecture
- **Primary Store:** PostgreSQL holds all core application data. Multi-tenancy is scoped by `circle_id`, allowing per-circle isolation and potential sharding.  
- **Media Storage:** User uploads (images, audio, video) are stored in S3-compatible object storage, with metadata managed in Postgres.  
- **Caching & Sync:** Clients use IndexedDB for local caching of timelines and drafts. Server endpoints expose cursor-based reads for reconciliation.  
- **Search & Discovery:** External search index (e.g., Meilisearch/OpenSearch) supports full-text queries; Postgres remains authoritative.  
- **Events & Fan-out:** An append-only event log backs timelines, notifications, and federation delivery, ensuring reliable fan-out and replay.  
- **Moderation & Retention:** Content removals use tombstones where possible; moderation data is first-class with explicit retention policies.  
- **Deployment Options:** Same model supports both self-managed FOSS deployments and managed hosting, with scale-out via read replicas and partitioning as needed.

### 3.11 Scaling
- **Database:** PostgreSQL scales by moving large circles (`circle_id`) to dedicated shards with read replicas. Hot read paths (timelines, notifications) use materialized tables fed by the event log.  
- **Media:** Assets live in S3-compatible object storage, fronted by CDN for distribution. Metadata and permissions remain in Postgres.  
- **Realtime Rooms:** Interactive rooms are capped in the low hundreds of participants. Multiparty sessions route through an SFU cluster, scaled horizontally by instance and region. TURN bandwidth is pooled and managed with short-lived credentials.  
- **Live Broadcasts:** Large one-to-many streams use a broadcast mode (single publisher → many subscribers). This leverages WebRTC SFU fan-out or HLS/DASH fallback for scale, similar to Twitch.  
- **Client Load:** Clients cache state locally (IndexedDB) and receive lightweight SSE nudges, avoiding mass fan-out from the server. Cursor-based APIs ensure predictable pagination and reduce stampede effects.  
- **Moderation & Federation:** Gateway enforces per-circle rate limits, queue backpressure, and cross-circle delivery budgets. Larger circles may enable stricter cadence controls or moderation tooling without affecting smaller ones.  
- **Deployment:** Self-managed instances scale up to modest circle sizes; managed hosting provisions dedicated resources (DB shard, SFU pool, CDN) for larger circles transparently.

### 3.12 Privacy
- **Data Ownership:** Circles are the trust boundary. All user data, content, and metadata are scoped by `circle_id` and subject to that circle’s policies.  
- **Minimal Retention:** Only the data needed for core functionality is retained. Drafts, presence, and caches are ephemeral by default; users may delete content and request full erasure.  
- **Encryption:** Transport is always TLS. 1:1 messages support end-to-end encryption; group messaging begins with server-enforced access and may adopt MLS for scalable E2EE. Media storage uses encrypted buckets with short-lived signed URLs.  
- **Local-First Caching:** Client caches (IndexedDB) are controlled by the user’s device. No hidden background sync; reconciliation happens only when the user is online and authenticated.  
- **Federation Boundaries:** Cross-circle delivery is explicit and policy-gated. Inbound events are quarantined until accepted; outbound events carry only the fields required by the receiving host.  
- **Telemetry & Logs:** Self-managed deployments collect only operator-configured metrics. Managed hosting strips or anonymizes personally identifiable information (PII) from logs where possible.  
- **Access Control:** Sessions are bound to WebAuthn credentials. Short-lived tokens are used for realtime joins to limit exposure if compromised.  

### 3.13 Security Threat Model
- **Attack Surface:** Gateway endpoints, federation ingress, WebRTC signaling, and SFU/TURN infrastructure are the primary exposure points.  
- **Common Threats:**  
  - Account takeover → mitigated with WebAuthn, short-lived tokens.  
  - Spam and flooding → gateway-level rate limiting and moderation queues.  
  - Malicious federation → quarantine inbound events, signature checks, allowlist of peers.  
  - Replay and tampering → append-only event log with idempotency keys; signed envelopes for federation traffic.  
- **Isolation:** Each circle is logically partitioned (`circle_id`) in the datastore; large circles may run on dedicated shards to reduce blast radius.  
- **Zero Trust Defaults:** All cross-service traffic uses mTLS; no implicit trust between components.  
- **Monitoring:** Structured logs, request tracing, and anomaly alerts for auth failures, rate spikes, and federation abuse.  

### 3.14 Federation Complexity
- **Trust Boundaries:** Each circle defines its own membership, moderation, and retention rules. Federation is optional and always policy-gated.  
- **Policy Divergence:** Different circles may enforce different standards (e.g., moderation strictness, privacy retention). This introduces interoperability friction.  
- **Certified Circle Framework:** To reduce complexity, circles can align with a **Certified Circle policy framework** — a set of baseline rules for moderation, privacy, and interoperability.  
  - Certified Circles advertise compliance via signed policy descriptors.  
  - Other circles can use this certification to decide trust levels for federation (e.g., auto-accept posts from Certified Circles, quarantine others).  
- **Reputation Standard:** Certification provides a shared reputation layer without requiring global governance. Circles remain sovereign but can opt-in to a recognizable standard for credibility.  
- **Implementation:** Policy descriptors are published and versioned as machine-readable documents; verification happens during federation handshake.

### 3.15 Disaster Recovery & Business Continuity

**Data Protection:**
- **Backup Strategy:** PostgreSQL continuous WAL archiving with point-in-time recovery; object storage replication across availability zones
- **Recovery Objectives:** RTO ≤ 4 hours, RPO ≤ 15 minutes for managed hosting; self-hosted instances responsible for their own backup cadence
- **Geographic Distribution:** Managed hosting maintains cross-region backups; critical data replicated to secondary regions

**Circle Migration & Portability:**
- **Export Functionality:** Complete circle data export in ActivityPub-compatible format including messages, media, member lists, and policy configurations
- **Import Process:** New instances can reconstruct circle state from export packages with cryptographic verification of data integrity
- **Gradual Migration:** Live migration support for moving active circles between instances with minimal downtime

**Service Continuity:**
- **Component Redundancy:** All managed hosting components run with N+1 redundancy; database clusters with automatic failover
- **Graceful Degradation:** System continues operating with reduced functionality during partial outages (e.g., P2P features degrade to server-mediated messaging)
- **Status Communication:** Public status page and in-app notifications during service disruptions

**Self-Hosted Resilience:**
- **Documentation:** Comprehensive disaster recovery runbooks for self-hosted operators
- **Automated Tooling:** Backup and recovery scripts included in deployment packages
- **Community Support:** Shared knowledge base and mutual aid network for operational issues

### 3.16 Developer Experience & API Strategy

**Third-Party Integration:**
- **REST API:** Full-featured HTTP API for circle management, messaging, and federation
- **WebHook System:** Real-time event delivery for external services (moderation bots, analytics, backup systems)
- **GraphQL Gateway:** Unified query interface for complex data fetching patterns
- **SDK Libraries:** Official client libraries for Go, JavaScript, Python, and Rust

**Extension Architecture:**
- **Plugin System:** Server-side extensions for custom moderation, authentication, and content processing
- **Client Hooks:** Browser extension API for custom UI components and workflow automation
- **Federation Adapters:** Pluggable protocols for interoperating with Matrix, ActivityPub, and proprietary systems

**Self-Hosted Tooling:**
- **Admin Dashboard:** Web-based management interface for instance configuration, user management, and system monitoring
- **CLI Tools:** Command-line utilities for deployment, migration, and maintenance tasks
- **Monitoring Integration:** Prometheus metrics, Grafana dashboards, and alerting templates
- **Update Management:** Automated update system with rollback capabilities and compatibility checking

**Documentation & Support:**
- **API Documentation:** Interactive API explorer with authentication sandbox
- **Integration Guides:** Step-by-step tutorials for common integration patterns
- **Community Forums:** Developer community with official support presence
- **Professional Services:** Custom integration and consulting for enterprise deployments

### 3.17 Performance Benchmarks & Targets

**Latency Targets:**
- **Gateway Response Time:** ≤200ms p95 for authenticated API requests
- **Timeline Loading:** ≤500ms p95 for initial feed render (50 messages)
- **Message Delivery:** ≤100ms p95 for same-circle message posting
- **WebRTC Connection:** ≤3 seconds p95 for peer-to-peer call establishment
- **Federation Delivery:** ≤2 seconds p95 for cross-circle message delivery

**Throughput Expectations:**
- **Small Circles (≤100 users):** 10 messages/second sustained, 100 messages/second burst
- **Medium Circles (≤1000 users):** 100 messages/second sustained, 1000 messages/second burst
- **Large Circles (≤10000 users):** 1000 messages/second sustained, 5000 messages/second burst
- **Database Operations:** 10,000 reads/second and 1,000 writes/second per circle on dedicated hardware

**Resource Requirements:**
- **Self-Hosted Small Instance:** 2 vCPU, 4GB RAM supports 100 concurrent users with 95% uptime
- **Self-Hosted Medium Instance:** 8 vCPU, 16GB RAM supports 1000 concurrent users with 99% uptime
- **Managed Hosting:** Guaranteed resource isolation and auto-scaling to handle 3x normal traffic spikes

**Monitoring & SLA:**
- **Managed Hosting Uptime:** 99.9% availability target with 4-hour RTO for critical issues
- **Performance Alerting:** Automated alerts when response times exceed thresholds by 50%
- **Capacity Planning:** Monthly capacity reports and scaling recommendations for self-hosted operators

## 4. Deployment

Circles.diy supports two main deployment modes:

1. **Self-Managed FOSS Instance**
 - Distributed as open-source software.
 - Intended for communities or individuals who want full control over their infrastructure.
 - Deployable on bare metal, cloud VMs, or container platforms (Docker Compose, Kubernetes).
 - Provides sovereignty but requires operational expertise.

2. **Paid Managed Hosting**
 - Fully hosted offering maintained by the Circles.diy team.
 - Provides ease of adoption and turnkey reliability.
 - Includes automatic upgrades, monitoring, and moderation tooling.
 - Targets communities who prefer convenience and support over self-hosting.

This dual-path approach ensures accessibility to both technically proficient operators and general communities who value simplicity.

### 4.1 Economic Model & Resource Allocation

**Self-Managed Deployment:**
- **Cost Structure:** Users bear infrastructure costs directly (cloud VMs, storage, bandwidth)
- **Resource Requirements:**
  - Small circles (≤100 users): 2-4 vCPU, 4-8GB RAM, 50GB storage
  - Medium circles (≤1000 users): 4-8 vCPU, 8-16GB RAM, 200GB storage + CDN
  - Large circles (≤10000 users): Dedicated database shard, SFU cluster, 500GB+ storage
- **Operational Burden:** Circle administrators handle updates, monitoring, backups

**Managed Hosting Tiers:**
- **Starter Circles:** $29/month for up to 100 active users, 10GB storage, basic support
- **Community Circles:** $99/month for up to 1000 active users, 100GB storage, priority support, advanced moderation tools
- **Organization Circles:** $299/month for up to 10000 active users, 1TB storage, dedicated resources, SLA guarantees
- **Enterprise:** Custom pricing for larger deployments, on-premise options, white-labeling

**Resource Allocation Strategy:**
- **Compute:** Gateway and backend services scale horizontally via container orchestration
- **Storage:** PostgreSQL scales through read replicas and circle-based sharding; object storage scales linearly
- **Bandwidth:** Media delivery optimized through CDN; P2P reduces server bandwidth for real-time features
- **Cost Optimization:** Managed hosting leverages multi-tenancy and resource pooling to offer better economics than individual VPS deployment

## 5. Key Tradeoffs
- **Hybrid vs Pure P2P:** Hybrid chosen for practical browser support, NAT traversal, and moderation.
- **Server-rendered HTML vs SPA:** Server-rendered chosen for speed, simplicity, and accessibility.
- **ActivityPub Alignment:** Objects shaped for interoperability without committing to federation scope immediately.
- **Deployment Options:** Balancing sovereignty (self-managed) against accessibility (managed hosting).

## 6. Key Design Decisions & Open Questions

### 6.1 Resolved Design Decisions
- **Hybrid Architecture:** Chosen over pure P2P for browser compatibility, NAT traversal, and moderation capabilities
- **Server-Rendered HTML:** Selected over SPA for accessibility, performance, and simplicity
- **Circle-Centric Model:** Preferred over global federation for clear trust boundaries and scalable governance
- **PostgreSQL + Object Storage:** Balanced approach providing ACID guarantees with cost-effective media handling
- **WebAuthn Authentication:** Passwordless approach reduces security risks and improves user experience

### 6.2 Implementation Timing Questions
- **MLS Adoption:** Should Messaging Layer Security be implemented in MVP or deferred until group messaging reaches enterprise scale?
- **CRDT Integration:** How much local-first functionality (collaborative editing, offline sync) belongs in initial release vs future iterations?
- **ActivityPub Compliance:** Should full federation interoperability be a launch requirement or post-MVP milestone?
- **API Completeness:** Which third-party integration capabilities are essential for community adoption vs nice-to-have features?

### 6.3 Operational Trade-offs
- **Feature Parity:** How to balance advanced features in managed hosting vs keeping self-hosted deployments competitive?
- **Scaling Triggers:** At what circle size should automatic migration to dedicated infrastructure occur?
- **Security Defaults:** How restrictive should default federation policies be for new circles?
- **Economic Sustainability:** What pricing strategy best supports long-term platform development while remaining accessible?

## 7. Implementation Roadmap

### 7.1 MVP (Months 1-6)
**Core Infrastructure:**
- ✅ **Gateway + gRPC Services:** HTTP bridge to Go backend services
- ✅ **Basic Circle Management:** Create, join, leave circles
- ✅ **WebAuthn Authentication:** Passwordless login foundation
- ✅ **Server-rendered UI:** htmx-powered dynamic updates
- ✅ **PostgreSQL Backend:** Multi-tenant data model with `circle_id` scoping

**Essential Features:**
- ✅ **Text Messaging:** Basic circle chat with server-side delivery
- ✅ **Content Feeds:** Timeline rendering and cursor-based pagination
- ✅ **Basic Moderation:** Rate limiting, blocklists, content filters
- ✅ **Media Upload:** Image/file sharing via object storage

**Deployment:**
- ✅ **Docker Compose Stack:** Self-hosted deployment option
- ✅ **Basic Monitoring:** Structured logging and health checks

### 7.2 Scale & Polish (Months 7-12)
**Real-time Features:**
- 🔄 **WebRTC Signaling:** P2P data channels and voice calls
- 🔄 **STUN/TURN Infrastructure:** NAT traversal for P2P connections
- 🔄 **SSE Push Notifications:** Live updates for active users

**Enhanced Security:**
- 🔄 **End-to-End Encryption:** 1:1 message encryption (double ratchet)
- 🔄 **DID Integration:** Portable identity mapping
- 🔄 **Advanced Moderation:** Content quarantine and review queues

**Performance:**
- 🔄 **Read Replicas:** Database scaling for growing circles
- 🔄 **CDN Integration:** Media delivery optimization
- 🔄 **Client Caching:** IndexedDB for offline-capable timelines

### 7.3 Federation & Growth (Months 13-18)
**Inter-Circle Communication:**
- 🔄 **Federation Protocol:** ActivityPub-aligned cross-circle messaging
- 🔄 **Certified Circle Framework:** Policy descriptor verification
- 🔄 **Trust Management:** Allowlists and reputation scoring

**Advanced Media:**
- 🔄 **SFU Integration:** Multi-party video calls (LiveKit/Janus)
- 🔄 **Live Streaming:** Broadcast mode for large audiences
- 🔄 **Collaborative Editing:** CRDT-based document sharing

**Enterprise Features:**
- 🔄 **Managed Hosting Platform:** Turnkey deployment for organizations
- 🔄 **Advanced Analytics:** Circle health and engagement metrics
- 🔄 **API Ecosystem:** Third-party integrations and extensions

### 7.4 Dependencies & Critical Path
1. **Gateway ← WebAuthn ← Circle Management** (foundational)
2. **PostgreSQL ← Media Storage ← Content Feeds** (core content)
3. **WebRTC Signaling ← STUN/TURN ← P2P Features** (real-time)
4. **Federation Protocol ← Trust Framework ← Inter-circle** (growth)

## 8. References

### 8.1 Federation & Protocol Standards
- [ActivityPub W3C Spec](https://www.w3.org/TR/activitypub/) - Foundation for interoperable social networking
- [Matrix Spec](https://spec.matrix.org/) - Real-time communication and federation reference
- [AT Protocol Overview](https://atproto.com/) - Decentralized social networking architecture
- [Messaging Layer Security (MLS)](https://messaginglayersecurity.rocks/) - Scalable group messaging encryption

### 8.2 Technical Implementation
- [WebRTC Specification](https://w3c.github.io/webrtc-pc/) - Peer-to-peer real-time communication
- [WebAuthn Guide](https://webauthn.guide/) - Passwordless authentication implementation
- [htmx Documentation](https://htmx.org/) - Progressive enhancement for server-rendered apps
- [gRPC-Web](https://github.com/grpc/grpc-web) - Browser-compatible gRPC implementation

### 8.3 Operational & Security
- [OWASP Application Security](https://owasp.org/www-project-top-ten/) - Security best practices
- [SRE Best Practices](https://sre.google/sre-book/table-of-contents/) - Site reliability engineering principles
- [PostgreSQL High Availability](https://www.postgresql.org/docs/current/high-availability.html) - Database scaling and backup strategies

### 8.4 Related Projects
- [Mastodon](https://github.com/mastodon/mastodon) - ActivityPub-based social networking
- [Element](https://github.com/vector-im/element-web) - Matrix protocol client implementation
- [PeerTube](https://github.com/Chocobozzz/PeerTube) - Federated video sharing platform
- [Jitsi Meet](https://github.com/jitsi/jitsi-meet) - WebRTC conferencing implementation

**End of Proposal**
