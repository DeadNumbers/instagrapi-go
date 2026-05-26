package instagrapi

import "fmt"

// ClientError is the base error type for all Instagram API errors.
type ClientError struct {
	Message  string
	Reason   string
	Code     int
	Response any // raw JSON response from Instagram
}

func (e *ClientError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	if e.Reason != "" {
		return fmt.Sprintf("%s (%v)", e.Reason, e.Response)
	}
	return "unknown error"
}

// GenericRequestError - "Sorry, there was a problem with your request"
type GenericRequestError struct{ ClientError }

// ClientGraphqlError — raised due to GraphQL issues
type ClientGraphqlError struct{ ClientError }

// ClientJSONDecodeError — raised due to JSON decoding issues
type ClientJSONDecodeError struct{ ClientError }

// ClientConnectionError — network connectivity issues
type ClientConnectionError struct{ ClientError }

// ClientBadRequestError — HTTP 400
type ClientBadRequestError struct{ ClientError }

// ClientUnauthorizedError — HTTP 401
type ClientUnauthorizedError struct{ ClientError }

// ClientForbiddenError — HTTP 403
type ClientForbiddenError struct{ ClientError }

// ClientNotFoundError — HTTP 404
type ClientNotFoundError struct{ ClientError }

// ClientThrottledError — HTTP 429
type ClientThrottledError struct{ ClientError }

// ClientRequestTimeout — HTTP 408
type ClientRequestTimeout struct{ ClientError }

// ClientLoginRequired — Instagram redirect to login page
type ClientLoginRequired struct{ ClientError }

// ReloginAttemptExceeded — max relogin attempts exceeded
type ReloginAttemptExceeded struct{ ClientError }

// PrivateError — for Private API and last_json logic
type PrivateError struct{ ClientError }

// NotFoundError — resource not found
type NotFoundError struct{ PrivateError }

func (e *NotFoundError) Error() string { return "Not found" }

// FeedbackRequired — action requires user feedback
type FeedbackRequired struct{ PrivateError }

// ChallengeError base type
type ChallengeError struct{ ClientError }

// ChallengeRedirection — challenge needs URL redirect
type ChallengeRedirection struct{ ChallengeError }

// ChallengeRequired — challenge must be resolved
type ChallengeRequired struct{ ChallengeError }

// ChallengeSelfieCaptcha — selfie captcha required
type ChallengeSelfieCaptcha struct{ ChallengeError }

// ChallengeUnknownStep — unknown challenge step
type ChallengeUnknownStep struct{ ChallengeError }

// SelectContactPointRecoveryForm — select contact point form
type SelectContactPointRecoveryForm struct{ ChallengeError }

// RecaptchaChallengeForm — reCAPTCHA required
type RecaptchaChallengeForm struct{ ChallengeError }

// SubmitPhoneNumberForm — submit phone number form
type SubmitPhoneNumberForm struct{ ChallengeError }

// LegacyForceSetNewPasswordForm — force password change
type LegacyForceSetNewPasswordForm struct{ ChallengeError }

// LoginRequired — Instagram requests relogin
type LoginRequired struct{ PrivateError }

// SentryBlock — rate limited by Sentry
type SentryBlock struct{ PrivateError }

// RateLimitError — account-level rate limit
type RateLimitError struct{ PrivateError }

// ProxyAddressIsBlocked — IP blocked by Instagram
type ProxyAddressIsBlocked struct{ PrivateError }

// BadPassword — incorrect password
type BadPassword struct{ PrivateError }

// BadCredentials — bad credentials
type BadCredentials struct{ ClientError }

// PleaseWaitFewMinutes — wait before retrying
type PleaseWaitFewMinutes struct{ PrivateError }

// UnknownError — unknown private error
type UnknownError struct{ PrivateError }

// TrackNotFound — music track not found
type TrackNotFound struct{ NotFoundError }

// MediaError — media-related error
type MediaError struct{ PrivateError }

// MediaNotFound — media not found
type MediaNotFound struct {
	NotFoundError
	MediaError
}

func (e *MediaNotFound) Error() string { return "Media not found" }

// StoryNotFound — story not found
type StoryNotFound struct {
	NotFoundError
	MediaError
}

func (e *StoryNotFound) Error() string { return "Story not found" }

// UserError — user-related error
type UserError struct{ PrivateError }

// UserNotFound — user not found
type UserNotFound struct {
	NotFoundError
	UserError
}

func (e *UserNotFound) Error() string { return "User not found" }

// CollectionError — collection-related error
type CollectionError struct{ PrivateError }

// CollectionNotFound — collection not found
type CollectionNotFound struct {
	NotFoundError
	CollectionError
}

func (e *CollectionNotFound) Error() string { return "Collection not found" }

// DirectError — DM-related error
type DirectError struct{ PrivateError }

// DirectThreadNotFound — thread not found
type DirectThreadNotFound struct {
	NotFoundError
	DirectError
}

func (e *DirectThreadNotFound) Error() string { return "Direct thread not found" }

// DirectMessageNotFound — message not found
type DirectMessageNotFound struct {
	NotFoundError
	DirectError
}

func (e *DirectMessageNotFound) Error() string { return "Direct message not found" }

// VideoTooLongException — video exceeds length limit
type VideoTooLongException struct{ PrivateError }

// VideoNotDownload — failed to download video
type VideoNotDownload struct{ PrivateError }

// VideoNotUpload — failed to upload video
type VideoNotUpload struct{ PrivateError }

// VideoConfigureError — configure step failed for video
type VideoConfigureError struct{ VideoNotUpload }

// PhotoNotUpload — failed to upload photo
type PhotoNotUpload struct{ PrivateError }

// PhotoConfigureError — configure step failed for photo
type PhotoConfigureError struct{ PhotoNotUpload }

// IGTVNotUpload — failed to upload IGTV
type IGTVNotUpload struct{ PrivateError }

// IGTVConfigureError — configure step failed for IGTV
type IGTVConfigureError struct{ IGTVNotUpload }

// ClipNotUpload — failed to upload clip/reel
type ClipNotUpload struct{ PrivateError }

// ClipConfigureError — configure step failed for clip
type ClipConfigureError struct{ ClipNotUpload }

// AlbumNotDownload — failed to download album
type AlbumNotDownload struct{ PrivateError }

// AlbumUnknownFormat — unknown album media format
type AlbumUnknownFormat struct{ PrivateError }

// AlbumConfigureError — configure step failed for album
type AlbumConfigureError struct{ PrivateError }

// HashtagError — hashtag-related error
type HashtagError struct{ PrivateError }

// HashtagNotFound — hashtag not found
type HashtagNotFound struct {
	NotFoundError
	HashtagError
}

func (e *HashtagNotFound) Error() string { return "Hashtag not found" }

// LocationError — location-related error
type LocationError struct{ PrivateError }

// LocationNotFound — location not found
type LocationNotFound struct {
	NotFoundError
	LocationError
}

func (e *LocationNotFound) Error() string { return "Location not found" }

// TwoFactorRequired — 2FA code required
type TwoFactorRequired struct{ PrivateError }

// HighlightNotFound — highlight not found
type HighlightNotFound struct {
	NotFoundError
	PrivateError
}

func (e *HighlightNotFound) Error() string { return "Highlight not found" }

// NoteNotFound — note not found
type NoteNotFound struct{ NotFoundError }

func (e *NoteNotFound) Error() string { return "Not found" }

// PrivateAccount — account is private
type PrivateAccount struct{ PrivateError }

// InvalidTargetUser — invalid target user
type InvalidTargetUser struct{ PrivateError }

// InvalidMediaId — invalid media ID
type InvalidMediaId struct{ PrivateError }

// MediaUnavailable — media is unavailable
type MediaUnavailable struct{ PrivateError }

// CommentUnavailable — comment is unavailable
type CommentUnavailable struct{ PrivateError }

// CommentNotFound — comment not found
type CommentNotFound struct{ PrivateError }

func (e *CommentNotFound) Error() string { return "Comment not found" }

// CommentsDisabled — comments disabled by author
type CommentsDisabled struct{ PrivateError }

func (e *CommentsDisabled) Error() string { return "Comments disabled by author" }

// ValidationError — validation failed
type ValidationError struct{ error }

func (e *ValidationError) Error() string {
	if e.error != nil {
		return e.error.Error()
	}
	return "validation error"
}

// EmailInvalidError — email not valid
type EmailInvalidError struct{ ClientError }

// EmailNotAvailableError — email already in use
type EmailNotAvailableError struct{ ClientError }

// EmailVerificationSendError — failed to send verification email
type EmailVerificationSendError struct{ ClientError }

// AgeEligibilityError — age eligibility check failed
type AgeEligibilityError struct{ ClientError }

// CaptchaChallengeRequired — captcha required, no solver configured
type CaptchaChallengeRequired struct {
	ClientError
	ChallengeDetails map[string]any `json:"challenge_details,omitempty"`
}

func (e *CaptchaChallengeRequired) Error() string {
	return "Captcha challenge required"
}

// SignupSpamError — legacy signup flow rejected as spam
type SignupSpamError struct{ FeedbackRequired }

func (e *SignupSpamError) Error() string { return "Signup spam detected" }
