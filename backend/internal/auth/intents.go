package auth

// Intent enumerates the permission strings stored in the user_intents table.
// Values match the legacy Node `intentList` type exactly — do not rename.
type Intent string

const (
	IntentAdmin    Intent = "ADMIN"
	IntentView     Intent = "VIEW"
	IntentOpen     Intent = "OPEN"
	IntentDownload Intent = "DOWNLOAD"
	IntentUpload   Intent = "UPLOAD"
	IntentCopy     Intent = "COPY"
	IntentDelete   Intent = "DELETE"
	IntentRename   Intent = "RENAME"
)
