erDiagram
  User ||--o{ Profile : has
  User ||--o{ Session : has
  User ||--|| UserModeration : has
  User ||--o{ ExternalIdentity : maps
  User ||--o{ WebAuthnCredential : has
  User ||--o{ Device : owns
  User ||--o{ RecoveryMethod : configured
  User ||--o{ AuthenticationChallenge : receives

  WebAuthnCredential ||--o{ DeviceCredential : "synced to"
  Device ||--o{ DeviceCredential : uses
  Device ||--o{ Session : creates
  Device ||--o{ DeviceVerification : verifies

  Session ||--|| Profile : "operates as"

  Profile ||--|| ProfileSettings : has
  Profile ||--o{ Circle : owns
  Profile ||--o{ CircleMembership : joins
  Profile ||--o{ Post : authors
  Profile ||--o{ Discussion : authors
  Profile ||--o{ Comment : authors
  Profile ||--o{ ChatParticipant : participates
  Profile ||--o{ Message : sends
  Profile ||--o{ Report : files
  Profile ||--o{ ModerationAction : performs
  Profile ||--o{ Attachment : uploads
  Profile ||--o{ Block : blocks
  Profile ||--o{ Mute : mutes
  Profile ||--o{ ExportBundle : exports

  Circle ||--o{ CircleMembership : contains
  Circle ||--o{ Post : contains
  Circle ||--o{ Discussion : hosts
  Circle ||--o{ Event : organizes
  Circle ||--o{ Activity : records
  Circle ||--o{ Role : defines

  Role ||--o{ RolePermission : maps
  Permission ||--o{ RolePermission : maps
  Circle ||--o{ ProfileRole : assigns
  Profile ||--o{ ProfileRole : holds
  Role ||--o{ ProfileRole : grants

  Post ||--o{ PostComment : links
  Discussion ||--o{ DiscussionComment : links
  Comment ||--o{ Comment : "replies to"

  Post ||--o{ Attachment : "attaches via PostAttachment"
  Discussion ||--o{ Attachment : "attaches via custom"
  Comment ||--o{ Reaction : receives
  Post ||--o{ Reaction : receives

  Chat ||--o{ ChatParticipant : includes
  Chat ||--o{ Message : contains
  Message ||--o{ MessageRead : receipts
  Message ||--o{ Attachment : "attaches via MessageAttachment"
  Message ||--o{ Reaction : receives

  Event ||--o{ EventRSVP : receives
  Event ||--o{ Ticket : issues
  Profile ||--o{ Ticket : owns
  Ticket }o--|| Order : "part of"

  Profile ||--o{ Notification : receives
  Activity }o--o{ Circle : about

  User ||--|| UserModeration : "moderation state"
  Report ||--o{ ModerationAction : "results in"

  User {
    string ID PK
    string Username UK
    string Email UK
    string AccountStatus
    datetime CreatedAt
    datetime UpdatedAt
    datetime DeletedAt
  }

  ExternalIdentity {
    string ID PK
    string OwnerType
    string OwnerID FK
    string DID UK
    string PublicKey
    string Provider
    string Proof
    datetime CreatedAt
    datetime UpdatedAt
  }

  WebAuthnCredential {
    string ID PK
    string UserID FK
    string CredentialID UK
    string PublicKey
    string CredentialType
    int SignCount
    string Transports
    string AAGUID
    string AttestationFormat
    blob AttestationObject
    string FriendlyName
    bool IsSynced
    bool IsBackup
    datetime LastUsedAt
    datetime CreatedAt
    datetime RevokedAt
  }

  Device {
    string ID PK
    string UserID FK
    string DeviceFingerprint UK
    string FriendlyName
    string DeviceType
    string OS
    string Browser
    bool IsTrusted
    datetime LastVerifiedAt
    datetime FirstSeenAt
    datetime LastSeenAt
    datetime RevokedAt
  }

  DeviceCredential {
    string ID PK
    string DeviceID FK
    string CredentialID FK
    datetime LinkedAt
    datetime LastUsedAt
  }

  Session {
    string ID PK
    string UserID FK
    string DeviceID FK
    string ActiveProfileID FK
    string AuthenticatedCredentialIDs
    string SessionToken
    string RefreshToken
    string IPAddress
    string UserAgent
    int AuthLevel
    datetime CreatedAt
    datetime LastActivityAt
    datetime ExpiresAt
    datetime RevokedAt
  }

  RecoveryMethod {
    string ID PK
    string UserID FK
    string Type
    string EncryptedSecret
    string Salt
    bool IsVerified
    int RemainingUses
    datetime VerifiedAt
    datetime LastUsedAt
    datetime ExpiresAt
    datetime CreatedAt
    datetime RevokedAt
  }

  AuthenticationChallenge {
    string ID PK
    string UserID FK
    string Challenge UK
    string Type
    string RequiredCredentialID
    jsonb Options
    datetime CreatedAt
    datetime ExpiresAt
    datetime CompletedAt
  }

  DeviceVerification {
    string ID PK
    string DeviceID FK
    string Method
    string VerificationCode
    bool IsVerified
    datetime VerifiedAt
    datetime ExpiresAt
    datetime CreatedAt
  }

  Profile {
    string ID PK
    string UserID FK
    string Handle UK
    string Name
    string DisplayName
    string Bio
    string AvatarURL
    bool IsActive
    datetime CreatedAt
    datetime UpdatedAt
    datetime DeletedAt
  }

  ProfileSettings {
    string ProfileID PK
    bool IsPublic
    string Location
    string Website
    jsonb Interests
    jsonb SocialLinks
    datetime UpdatedAt
  }

  Circle {
    string ID PK
    string OwnerProfileID FK
    string Name
    string Description
    string Visibility
    bool AutoModEnabled
    datetime CreatedAt
    datetime UpdatedAt
    datetime DeletedAt
  }

  CircleMembership {
    string ID PK
    string CircleID FK
    string ProfileID FK
    string State
    datetime JoinedAt
    datetime LeftAt
    datetime CreatedAt
    datetime UpdatedAt
  }

  Role {
    string ID PK
    string CircleID FK
    string Name
    datetime CreatedAt
    datetime UpdatedAt
  }

  Permission {
    string ID PK
    string Key UK
    string Description
  }

  RolePermission {
    string ID PK
    string RoleID FK
    string PermissionID FK
  }

  ProfileRole {
    string ID PK
    string CircleID FK
    string ProfileID FK
    string RoleID FK
    datetime CreatedAt
  }

  Post {
    string ID PK
    string CircleID FK
    string AuthorProfileID FK
    string Body
    string BodyFormat
    string ContentWarning
    string Visibility
    int ReplyCount
    int AttachmentsCount
    string ModerationStatus
    datetime EditedAt
    datetime CreatedAt
    datetime UpdatedAt
    datetime DeletedAt
    string CID
    string Signature
  }

  Discussion {
    string ID PK
    string CircleID FK
    string AuthorProfileID FK
    string Title
    string Body
    bool IsPinned
    bool IsLocked
    int ReplyCount
    datetime EditedAt
    datetime CreatedAt
    datetime UpdatedAt
    datetime DeletedAt
    string CID
    string Signature
  }

  Comment {
    string ID PK
    string AuthorProfileID FK
    string Body
    string BodyFormat
    string ReplyToCommentID FK
    datetime EditedAt
    datetime CreatedAt
    datetime UpdatedAt
    datetime DeletedAt
    string CID
    string Signature
  }

  PostComment {
    string ID PK
    string PostID FK
    string CommentID FK
  }

  DiscussionComment {
    string ID PK
    string DiscussionID FK
    string CommentID FK
  }

  Reaction {
    string ID PK
    string TargetType
    string TargetID
    string ProfileID FK
    string Key
    datetime CreatedAt
  }

  Attachment {
    string ID PK
    string OwnerProfileID FK
    string MimeType
    int Size
    string StorageURL
    string CID
    datetime CreatedAt
    datetime DeletedAt
  }

  PostAttachment {
    string ID PK
    string PostID FK
    string AttachmentID FK
  }

  MessageAttachment {
    string ID PK
    string MessageID FK
    string AttachmentID FK
  }

  Chat {
    string ID PK
    string Type
    string Name
    datetime LastMessageAt
    datetime CreatedAt
    datetime UpdatedAt
    datetime DeletedAt
  }

  ChatParticipant {
    string ID PK
    string ChatID FK
    string ProfileID FK
    bool IsAdmin
    datetime JoinedAt
  }

  Message {
    string ID PK
    string ChatID FK
    string SenderProfileID FK
    string Content
    string MessageType
    string EncryptionScheme
    string Nonce
    string ReplyToMessageID FK
    bool IsFlagged
    datetime EditedAt
    datetime CreatedAt
    datetime UpdatedAt
    datetime DeletedAt
  }

  MessageRead {
    string ID PK
    string MessageID FK
    string ProfileID FK
    datetime ReadAt
  }

  Event {
    string ID PK
    string CircleID FK
    string OrganizerProfileID FK
    string Title
    string Location
    string Timezone
    datetime StartTime
    datetime EndTime
    bool RequiresTicket
    int Capacity
    datetime CreatedAt
    datetime UpdatedAt
    datetime DeletedAt
  }

  EventRSVP {
    string ID PK
    string EventID FK
    string ProfileID FK
    string Status
    datetime CreatedAt
    datetime UpdatedAt
  }

  Order {
    string ID PK
    string BuyerProfileID FK
    float TotalAmount
    string Currency
    string PaymentStatus
    datetime CreatedAt
    datetime UpdatedAt
  }

  Ticket {
    string ID PK
    string EventID FK
    string ProfileID FK
    string OrderID FK
    float Price
    string Currency
    string TicketStatus
    string QRCode
    datetime IssuedAt
    datetime CreatedAt
    datetime UpdatedAt
  }

  Notification {
    string ID PK
    string ProfileID FK
    string ActorProfileID FK
    string Type
    string EntityType
    string EntityID
    datetime DeliveredAt
    datetime ReadAt
    string Channel
    datetime CreatedAt
  }

  Activity {
    string ID PK
    string CircleID FK
    string ActorProfileID FK
    string EventType
    string TargetType
    string TargetID
    datetime OccurredAt
    jsonb Payload
  }

  UserModeration {
    string UserID PK
    float TrustScore
    int WarningCount
    int ViolationCount
    datetime BannedUntil
    string Notes
    datetime UpdatedAt
  }

  Report {
    string ID PK
    string ReporterProfileID FK
    string TargetType
    string TargetID
    string Reason
    string State
    datetime CreatedAt
    datetime UpdatedAt
  }

  ModerationAction {
    string ID PK
    string ModeratorProfileID FK
    string TargetType
    string TargetID
    string Action
    string Reason
    datetime CreatedAt
  }

  Block {
    string ID PK
    string BlockerProfileID FK
    string BlockedProfileID FK
    datetime CreatedAt
  }

  Mute {
    string ID PK
    string MuterProfileID FK
    string MutedProfileID FK
    datetime CreatedAt
  }

  ExportBundle {
    string ID PK
    string ProfileID FK
    string Format
    string Location
    datetime CreatedAt
  }