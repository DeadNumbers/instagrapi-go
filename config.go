package instagrapi

// API domain used for private API requests.
const apiDomain = "i.instagram.com"

// Default device profile (hardware/system) and app version settings.
var defaultDeviceSettings = map[string]any{
	"android_version": 34,
	"android_release": "14",
	"dpi":             "480dpi",
	"resolution":      "1344x2992",
	"manufacturer":    "Google/google",
	"device":          "husky",
	"model":           "Pixel 8 Pro",
	"cpu":             "husky",
}

// Default app version info.
const defaultAppVersion = "428.0.0.47.67"
const defaultVersionCode = "961145276"

// Bloks versioning ID for the default app version.
var bloksVersioningID = "7189b949425f9bf80ea8bd880cf5a3080b292d9b1c4b38a18d112f7c4b71e7a8"

// Supported capabilities sent in requests.
var supportedCapabilities = []map[string]any{
	{
		"name": "SUPPORTED_SDK_VERSIONS",
		"value": "119.0,120.0,121.0,122.0,123.0,124.0,125.0,126.0,127.0,128.0," +
			"129.0,130.0,131.0,132.0,133.0,134.0,135.0,136.0,137.0,138.0," +
			"139.0,140.0,141.0,142.0",
	},
	{"name": "FACE_TRACKER_VERSION", "value": "14"},
	{"name": "COMPRESSION", "value": "ETC2_COMPRESSION"},
	{"name": "gyroscope", "value": "gyroscope_enabled"},
}

// Login experiments sent during registration/login.
var loginExperiments = "ig_android_reg_nux_headers_cleanup_universe," +
	"ig_android_device_detection_info_upload," +
	"ig_android_nux_add_email_device," +
	"ig_android_gmail_oauth_in_reg," +
	"ig_android_device_info_foreground_reporting," +
	"ig_android_device_verification_separate_endpoint," +
	"ig_android_sim_info_upload," +
	"ig_android_smartlock_hints_universe"
