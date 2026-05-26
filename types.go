package instagrapi

import (
	"time"
)

// TypesBaseModel is the base for all types.
type TypesBaseModel struct{}

// User represents a full Instagram user.
type User struct {
	PK                     string `json:"pk,omitempty"`
	Username               string `json:"username,omitempty"`
	FullName               string `json:"full_name,omitempty"`
	ProfilePicURL          string `json:"profile_pic_url,omitempty"`
	ProfilePicID           string `json:"profile_pic_id,omitempty"`
	IsPrivate              bool   `json:"is_private,omitempty"`
	IsVerified             bool   `json:"is_verified,omitempty"`
	Biography              string `json:"biography,omitempty"`
	ExternalURL            string `json:"external_url,omitempty"`
	ExternalLynxURL        string `json:"external_lynx_url,omitempty"`
	HasAnonymousProfilePic bool   `json:"has_anonymous_profile_picture,omitempty"`
	HasHighlightReels      bool   `json:"has_highlight_reels,omitempty"`
	HasGuides              bool   `json:"has_guides,omitempty"`
	IsBusiness             bool   `json:"is_business,omitempty"`
	FollowerCount          int    `json:"follower_count,omitempty"`
	FollowingCount         int    `json:"following_count,omitempty"`
	MediaCount             int    `json:"media_count,omitempty"`
	GEOCityID              int    `json:"geocity_id,omitempty"`
	Zip                    string `json:"zip,omitempty"`
	PublicEmail            string `json:"public_email,omitempty"`
	PublicPhoneCountryCode int    `json:"public_phone_country_code,omitempty"`
	PublicPhoneNumber      string `json:"public_phone_number,omitempty"`
	BusinessContactType    string `json:"business_contact_type,omitempty"`
	Category               string `json:"category,omitempty"`
	IsCallToActionEnabled  bool   `json:"is_call_to_action_enabled,omitempty"`
	TotalIgtvVideos        int    `json:"total_igtvs,omitempty"`
	TotalArEffects         int    `json:"total_ar_effects,omitempty"`
	AccountType            int    `json:"account_type,omitempty"`
	IsPotentialBusiness    bool   `json:"is_potential_business,omitempty"`
	ShowPostCount          int    `json:"show_post_count,omitempty"`
	AllowedCommenterPolicy int    `json:"allowed_commenter_policy,omitempty"`

	// Relationship fields
	FriendshipStatus *FriendshipStatus `json:"friendship_status,omitempty"`
}

// UserShort is a minimal user representation.
type UserShort struct {
	PK            string `json:"pk,omitempty"`
	Username      string `json:"username,omitempty"`
	FullName      string `json:"full_name,omitempty"`
	ProfilePicURL string `json:"profile_pic_url,omitempty"`
	IsPrivate     bool   `json:"is_private,omitempty"`

	// For hash map usage
	hash int64
}

func (u UserShort) Hash() int64      { return u.hash }
func (u *UserShort) SetHash(h int64) { u.hash = h }

// FriendshipStatus represents the relationship between two users.
type FriendshipStatus struct {
	Following            bool `json:"following,omitempty"`
	FollowedBy           bool `json:"followed_by,omitempty"`
	BlockedByMe          bool `json:"blocked_by_me,omitempty"`
	IsBlocking           bool `json:"is_blocking,omitempty"`
	IsMuting             bool `json:"is_muting,omitempty"`
	IsPrivate            bool `json:"is_private,omitempty"`
	NotificationsEnabled bool `json:"notifications_enabled,omitempty"`
	IncomingRequest      bool `json:"incoming_request,omitempty"`
	OutgoingRequest      bool `json:"outgoing_request,omitempty"`
	CloseFriending       bool `json:"is_close_friending,omitempty"`
	IsBestie             bool `json:"is_bestie,omitempty"`
	IsFeedFavorite       bool `json:"is_feed_favorite,omitempty"`
}

// Account represents the authenticated user's account.
type Account struct {
	PK             string `json:"pk,omitempty"`
	Username       string `json:"username,omitempty"`
	FullName       string `json:"full_name,omitempty"`
	Biography      string `json:"biography,omitempty"`
	ProfilePicURL  string `json:"profile_pic_url,omitempty"`
	ProfilePicID   string `json:"profile_pic_id,omitempty"`
	IsPrivate      bool   `json:"is_private,omitempty"`
	IsVerified     bool   `json:"is_verified,omitempty"`
	Email          string `json:"email,omitempty"`
	Phone_number   string `json:"phone_number,omitempty"`
	Gender         int    `json:"gender,omitempty"`
	IsBusiness     bool   `json:"is_business,omitempty"`
	AccountType    int    `json:"account_type,omitempty"`
	MediaCount     int    `json:"media_count,omitempty"`
	FollowerCount  int    `json:"follower_count,omitempty"`
	FollowingCount int    `json:"following_count,omitempty"`
}

// Media represents an Instagram media item (post, reel, etc.).
type Media struct {
	PK                   any     `json:"pk,omitempty"` // can be string or int
	ID                   string  `json:"id,omitempty"`
	Code                 string  `json:"code,omitempty"`
	TakenAt              int64   `json:"taken_at,omitempty"`
	MediaType            int     `json:"media_type,omitempty"`
	CaptionText          string  `json:"caption_text,omitempty"`
	ThumbnailURL         string  `json:"thumbnail_url,omitempty"`
	ImageURL             string  `json:"image_url,omitempty"` // for type 1 (photo)
	VideoURL             string  `json:"video_url,omitempty"` // for type 2 (video)
	ViewCount            any     `json:"view_count,omitempty"`
	PlayCount            any     `json:"play_count,omitempty"`
	CommentCount         int     `json:"comment_count,omitempty"`
	LikeCount            int     `json:"like_count,omitempty"`
	HasLiked             *bool   `json:"has_liked,omitempty"`
	User                 User    `json:"user,omitempty"`
	CanViewerSave        bool    `json:"can_viewer_save,omitempty"`
	OrganicTrackingToken string  `json:"organic_tracking_token,omitempty"`
	Resources            []Media `json:"resources,omitempty"` // for albums/carousels

	// Story-specific fields
	StoryHashtags  []StoryHashtag  `json:"story_hashtags,omitempty"`
	StoryMentions  []StoryMention  `json:"story_mentions,omitempty"`
	StoryPolls     []StoryPoll     `json:"story_polls,omitempty"`
	StoryLinks     []StoryLink     `json:"story_links,omitempty"`
	StoryLocations []StoryLocation `json:"story_locations,omitempty"`

	// Clips/Reels metadata
	ClipsMetadata *ClipsMetadata `json:"clips_metadata,omitempty"`
}

// MediaOembed is a shortened media representation from the oembed endpoint.
type MediaOembed struct {
	Type         string `json:"type"`
	Title        string `json:"title"`
	AuthorName   string `json:"author_name"`
	AuthorURL    string `json:"author_url"`
	AuthorID     string `json:"author_id"`
	URL          string `json:"url"`
	ThumbnailURL string `json:"thumbnail_url"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	CanonicalURL string `json:"canonical_url"`
}

// Comment represents a comment on media.
type Comment struct {
	PK              string    `json:"pk,omitempty"`
	Text            string    `json:"text,omitempty"`
	User            UserShort `json:"user,omitempty"`
	UserID          string    `json:"user_id,omitempty"`
	TypeID          int       `json:"type,omitempty"`
	CreatedAt       int64     `json:"created_at,omitempty"`
	ContentType     string    `json:"content_type,omitempty"`
	Status          string    `json:"status,omitempty"`
	BitFlags        int       `json:"bit_flags,omitempty"`
	IsRankedComment bool      `json:"is_ranked_comment,omitempty"`
	RepliesCount    int       `json:"reply_count,omitempty"`
}

// Hashtag represents an Instagram hashtag.
type Hashtag struct {
	ID                string `json:"id,omitempty"`
	Name              string `json:"name,omitempty"`
	MediaCount        int    `json:"media_count,omitempty"`
	SearchResultTitle string `json:"search_result_title,omitempty"`
	Description       string `json:"description,omitempty"`
}

// Location represents an Instagram location.
type Location struct {
	PK               int     `json:"pk,omitempty"`
	Name             string  `json:"name,omitempty"`
	Address          string  `json:"address,omitempty"`
	City             string  `json:"city,omitempty"`
	PhoneNumber      string  `json:"phone_number,omitempty"`
	Lat              float64 `json:"lat,omitempty"`
	Lng              float64 `json:"lng,omitempty"`
	ExternalID       string  `json:"external_id,omitempty"`
	ExternalIDSource string  `json:"external_id_source,omitempty"`
}

// Collection represents a saved collection.
type Collection struct {
	ID      int64  `json:"id,omitempty"`
	Name    string `json:"name,omitempty"`
	Slug    string `json:"slug,omitempty"`
	Type    string `json:"type,omitempty"`
	Default bool   `json:"default,omitempty"`
}

// Story represents an Instagram story.
type Story struct {
	PK             any             `json:"pk,omitempty"`
	ID             string          `json:"id,omitempty"`
	TakenAt        int64           `json:"taken_at,omitempty"`
	MediaType      int             `json:"media_type,omitempty"`
	ImageURL       string          `json:"image_url,omitempty"`
	VideoURL       string          `json:"video_url,omitempty"`
	VideoDuration  float64         `json:"video_duration,omitempty"`
	User           UserShort       `json:"user,omitempty"`
	CaptionText    string          `json:"caption_text,omitempty"`
	StoryHashtags  []StoryHashtag  `json:"story_hashtags,omitempty"`
	StoryMentions  []StoryMention  `json:"story_mentions,omitempty"`
	StoryPolls     []StoryPoll     `json:"story_polls,omitempty"`
	StoryLinks     []StoryLink     `json:"story_links,omitempty"`
	StoryLocations []StoryLocation `json:"story_locations,omitempty"`
	StoryStickers  []StorySticker  `json:"story_stickers,omitempty"`
	ReelMentions   []StoryMention  `json:"reel_mentions,omitempty"`
}

// StoryBuild holds the result of building a story.
type StoryBuild struct {
	Mentions []StoryMention `json:"mentions,omitempty"`
	Path     string         `json:"path,omitempty"`
	Paths    []string       `json:"paths,omitempty"`
	Stickers []StorySticker `json:"stickers,omitempty"`
}

// StoryHashtag is a hashtag sticker in a story.
type StoryHashtag struct {
	X, Y, Z  float32 `json:"x,omitempty"`
	Width    float32 `json:"width,omitempty"`
	Height   float32 `json:"height,omitempty"`
	Rotation float32 `json:"rotation,omitempty"`
	Hashtag  Hashtag `json:"hashtag,omitempty"`
}

// StoryMention is a user mention in a story.
type StoryMention struct {
	X        float32   `json:"x,omitempty"`
	Y        float32   `json:"y,omitempty"`
	Width    float32   `json:"width,omitempty"`
	Height   float32   `json:"height,omitempty"`
	Rotation float32   `json:"rotation,omitempty"`
	User     UserShort `json:"user,omitempty"`
}

// StoryPoll is a poll sticker in a story.
type StoryPoll struct {
	X, Y, Z        float32      `json:"x,omitempty"`
	Width          float32      `json:"width,omitempty"`
	Height         float32      `json:"height,omitempty"`
	Rotation       float32      `json:"rotation,omitempty"`
	PollID         string       `json:"poll_id,omitempty"`
	Polls          []PollOption `json:"poll_options,omitempty"`
	IsSharedResult bool         `json:"is_shared_result,omitempty"`
}

type PollOption struct {
	ID    string  `json:"id,omitempty"`
	Text  string  `json:"text,omitempty"`
	Count int     `json:"count,omitempty"`
	Pct   float32 `json:"percentage,omitempty"`
}

// StoryLink is a link sticker in a story.
type StoryLink struct {
	X        float32 `json:"x,omitempty"`
	Y        float32 `json:"y,omitempty"`
	Width    float32 `json:"width,omitempty"`
	Height   float32 `json:"height,omitempty"`
	Rotation float32 `json:"rotation,omitempty"`
	LinkType string  `json:"link_type,omitempty"`
	URL      string  `json:"url,omitempty"`
}

// StoryLocation is a location sticker in a story.
type StoryLocation struct {
	X        float32  `json:"x,omitempty"`
	Y        float32  `json:"y,omitempty"`
	Width    float32  `json:"width,omitempty"`
	Height   float32  `json:"height,omitempty"`
	Rotation float32  `json:"rotation,omitempty"`
	Location Location `json:"location,omitempty"`
}

// StorySticker is a generic story sticker.
type StorySticker struct {
	X, Y, Z  float32        `json:"x,omitempty"`
	Width    float32        `json:"width,omitempty"`
	Height   float32        `json:"height,omitempty"`
	Rotation float32        `json:"rotation,omitempty"`
	ID       string         `json:"id,omitempty"`
	Type     string         `json:"type,omitempty"`
	Extra    map[string]any `json:"extra,omitempty"`
}

// Highlight represents an Instagram highlight.
type Highlight struct {
	PK         string     `json:"pk,omitempty"`
	ID         string     `json:"id,omitempty"`
	Title      string     `json:"title,omitempty"`
	CoverMedia CoverMedia `json:"cover_media,omitempty"`
	MediaCount int        `json:"media_count,omitempty"`
	MediaIDs   []string   `json:"item_ids,omitempty"`
	User       UserShort  `json:"user,omitempty"`
}

type CoverMedia struct {
	ThumbnailURL string `json:"thumbnail_url,omitempty"`
}

// DirectThread represents a direct message thread.
type DirectThread struct {
	PK                   string          `json:"pk,omitempty"`        // thread_vanity_id
	ThreadID             string          `json:"thread_id,omitempty"` // thread_id
	Title                string          `json:"title,omitempty"`
	ThreadType           string          `json:"thread_type,omitempty"`
	ViewerID             string          `json:"viewer_id,omitempty"`
	ApprovalRequired     bool            `json:"approval_required,omitempty"`
	ImageURL             string          `json:"image_url,omitempty"`
	Issues               []ThreadIssue   `json:"issues,omitempty"`
	IssuesList           []any           `json:"issues_v2,omitempty"`
	Members              []ThreadMember  `json:"members,omitempty"`
	Pending              bool            `json:"pending,omitempty"`
	PendingScore         string          `json:"pending_score,omitempty"`
	PendingRank          float32         `json:"pending_rank,omitempty"`
	UnseenCount          int             `json:"unseen_count,omitempty"`
	UnseenCountTimestamp int64           `json:"unseen_count_timestamp,omitempty"`
	LastAt               int64           `json:"last_at,omitempty"`
	Muted                bool            `json:"muted,omitempty"`
	IsPin                bool            `json:"is_pin,omitempty"`
	IsChatChannel        bool            `json:"is_channel,omitempty"`
	Name                 string          `json:"name,omitempty"`
	RolledOutFeatures    []string        `json:"rolled_out_features,omitempty"`
	ThreadInvState       string          `json:"thread_inv_state,omitempty"`
	CallTimestamp        int64           `json:"call_timestamp,omitempty"`
	IsGroupInvite        bool            `json:"is_group_invite,omitempty"`
	SkipCmdWrap          bool            `json:"skip_cmd_wrap,omitempty"`
	HideAllItems         bool            `json:"hide_all_items,omitempty"`
	MentionMuted         bool            `json:"mention_muted,omitempty"`
	IsCloseFriend        bool            `json:"is_close_friend,omitempty"`
	IsFBPage             bool            `json:"is_fb_page,omitempty"`
	FBContactID          int64           `json:"fb_contact_id,omitempty"`
	Items                []DirectMessage `json:"items,omitempty"`
	Inviter              UserShort       `json:"inviter,omitempty"`
	V2Members            []UserShort     `json:"v2_members,omitempty"`
}

type ThreadIssue struct {
	ID   string `json:"id,omitempty"`
	Text string `json:"text,omitempty"`
	Type string `json:"type,omitempty"`
}

type ThreadMember struct {
	User UserShort `json:"user,omitempty"`
	Role string    `json:"role,omitempty"`
	Itsc int       `json:"itsc,omitempty"`
}

// DirectMessage represents a message in a direct thread.
type DirectMessage struct {
	ID            string            `json:"item_id,omitempty"`
	UserID        string            `json:"user_id,omitempty"`
	Timestamp     int64             `json:"timestamp,omitempty"`
	ItemType      string            `json:"item_type,omitempty"`
	ItemID        string            `json:"client_context,omitempty"`
	DirectMessage DirectInbox       `json:"message,omitempty"`
	Media         *DirectMedia      `json:"media,omitempty"`
	Reactions     map[string]string `json:"reactions,omitempty"`
	IsSentAsBoost bool              `json:"is_sent_as_boost,omitempty"`
}

type DirectInbox struct {
	Text              string       `json:"text,omitempty"`
	Replies           []any        `json:"replies,omitempty"`
	Link              *MessageLink `json:"link,omitempty"`
	Attachments       []any        `json:"attachments,omitempty"`
	Media             *DirectMedia `json:"media,omitempty"`
	MediaShare        *MediaShare  `json:"media_share,omitempty"`
	StoryShare        *MediaShare  `json:"story_share,omitempty"`
	AnimatedThumbnail string       `json:"animated_thumbnail_url,omitempty"`
}

type MessageLink struct {
	URL      string `json:"url,omitempty"`
	Title    string `json:"title,omitempty"`
	LinkType string `json:"link_type,omitempty"`
	Platform string `json:"platform,omitempty"`
}

type MediaShare struct {
	Type  string    `json:"media_type,omitempty"`
	Share string    `json:"share_url,omitempty"`
	User  UserShort `json:"user,omitempty"`
	Media Media     `json:"media_or_ad,omitempty"`
	Code  string    `json:"code,omitempty"`
}

type DirectMedia struct {
	ID           string  `json:"id,omitempty"`
	VideoURL     string  `json:"video_url,omitempty"`
	ThumbnailURL string  `json:"thumbnail_url,omitempty"`
	MediaType    int     `json:"media_type,omitempty"`
	Width        float32 `json:"width,omitempty"`
	Height       float32 `json:"height,omitempty"`
	Duration     float64 `json:"video_duration,omitempty"`
}

// DirectShortThread is a minimal thread representation.
type DirectShortThread struct {
	ThreadID   string      `json:"thread_id,omitempty"`
	Title      string      `json:"title,omitempty"`
	ThreadType string      `json:"thread_type,omitempty"`
	Users      []UserShort `json:"users,omitempty"`
}

// Guide represents an Instagram guide.
type Guide struct {
	ID         string `json:"id,omitempty"`
	Title      string `json:"title,omitempty"`
	Type       string `json:"type,omitempty"`
	CoverMedia Media  `json:"cover_media,omitempty"`
}

// Track represents a music track.
type Track struct {
	ID             string         `json:"id,omitempty"`
	Title          string         `json:"title,omitempty"`
	DisplayArtist  string         `json:"display_artist,omitempty"`
	ArtistID       string         `json:"artist_id,omitempty"`
	AudioAssetID   string         `json:"audio_asset_id,omitempty"`
	AUDIOClusterID string         `json:"audio_cluster_id,omitempty"`
	Ignored        bool           `json:"is_eligible_for_audio_export,omitempty"`
	Duration       float64        `json:"duration_in_ms,omitempty"`
	AlbumCoverURL  string         `json:"album_cover_art_url,omitempty"`
	Slug           string         `json:"slug,omitempty"`
	OnboardingInfo map[string]any `json:"onboarding_info,omitempty"`
}

// Note represents a direct note.
type Note struct {
	ID             string    `json:"id,omitempty"`
	Text           string    `json:"text,omitempty"`
	UserID         string    `json:"user_id,omitempty"`
	User           UserShort `json:"user,omitempty"`
	Audience       int       `json:"audience,omitempty"`
	CreatedAt      time.Time `json:"created_at,omitempty"`
	ExpiresAt      time.Time `json:"expires_at,omitempty"`
	IsEmojiOnly    bool      `json:"is_emoji_only,omitempty"`
	HasTranslation bool      `json:"has_translation,omitempty"`
	NoteStyle      int       `json:"note_style,omitempty"`
}

// Share represents a shareable object.
type Share struct {
	Type string `json:"type,omitempty"`
	PK   string `json:"pk,omitempty"`
}

// HashtagInfo holds hashtag metadata from the info endpoint.
type HashtagInfo struct {
	ID                string `json:"id,omitempty"`
	Name              string `json:"name,omitempty"`
	MediaCount        int    `json:"media_count,omitempty"`
	SearchResultTitle string `json:"search_result_title,omitempty"`
}

// Broadcast represents a livestream broadcast.
type Broadcast struct {
	ID             string    `json:"id,omitempty"`
	Status         string    `json:"status,omitempty"`
	PublicID       string    `json:"public_id,omitempty"`
	BroadcastOwner UserShort `json:"owner,omitempty"`
	StreamURL      string    `json:"stream_url,omitempty"`
}

// ClipsMetadata holds metadata for clips/reels.
type ClipsMetadata struct {
	AudioAssetID      string             `json:"audio_asset_id,omitempty"`
	Title             string             `json:"title,omitempty"`
	OriginalSoundInfo *OriginalSoundInfo `json:"original_sound_info,omitempty"`
}

type OriginalSoundInfo struct {
	Audience              int    `json:"audience,omitempty"`
	AudioAssetID          string `json:"audio_asset_id,omitempty"`
	AuthorID              string `json:"author_id,omitempty"`
	IsOriginalAudioFromIG bool   `json:"is_original_audio_from_ig,omitempty"`
}

// StoryArchiveDay represents a day of archived stories.
type StoryArchiveDay struct {
	Date     int64  `json:"date,omitempty"`
	Username string `json:"username,omitempty"`
}

// UserAbout contains about info for a user.
type UserAbout struct {
	PK            string `json:"pk,omitempty"`
	Username      string `json:"username,omitempty"`
	FullName      string `json:"full_name,omitempty"`
	Biography     string `json:"biography,omitempty"`
	ProfilePicURL string `json:"profile_pic_url,omitempty"`
	ExternalURL   string `json:"external_url,omitempty"`
}

// AccountSecurityInfo holds security-related account data.
type AccountSecurityInfo struct {
	IsPhoneConfirmed              bool     `json:"is_phone_confirmed,omitempty"`
	IsTwoFactorEnabled            bool     `json:"is_two_factor_enabled,omitempty"`
	IsTOTPTwoFactorEnabled        bool     `json:"is_totp_two_factor_enabled,omitempty"`
	IsTrustedNotificationsEnabled bool     `json:"is_trusted_notifications_enabled,omitempty"`
	HasReachableEmail             bool     `json:"has_reachable_email,omitempty"`
	BackupCodes                   []string `json:"backup_codes,omitempty"`
}

// BioLink represents a link in the user's bio.
type BioLink struct {
	URL      string `json:"url,omitempty"`
	Title    string `json:"title,omitempty"`
	LinkType string `json:"link_type,omitempty"` // "external"
}
