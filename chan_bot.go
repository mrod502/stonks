package stonks

//Board - where all the tendies reside
type Board struct {
	Board           string `json:"board"`
	Title           int64  `json:"title"`
	WSBoard         int64  `json:"ws_board"`
	PerPage         int64  `json:"per_page"`
	Pages           int64  `json:"pages"`
	MaxFilesize     int64  `json:"max_filesize"`
	MaxWebmFilesize int64  `json:"max_webm_filesize"`
	MaxCommentChars int64  `json:"max_comment_chars"`
	MaxWebmDuration int64  `json:"max_webm_duration"`
	BumpLimit       int64  `json:"bump_limit"`
	ImageLimit      int64  `json:"image_limit"`
	MetaDescription int64  `json:"meta_description"`
	Spoilers        int64  `json:"spoilers"`
	CustomSpoilers  int64  `json:"custom_spoilers"`
	IsArchived      uint8  `json:"is_archived"`
	TrollFlags      uint8  `json:"troll_flags"`
	CountryFlags    uint8  `json:"country_flags"`
	UserIDs         uint8  `json:"user_ids"`
	Oekaki          uint8  `json:"oekaki"`
	SJISTags        uint8  `json:"sjis_tags"`
	CodeTags        uint8  `json:"code_tags"`
	TextOnly        uint8  `json:"text_only"`
	ForcedAnon      uint8  `json:"forced_anon"`
	WebmAudio       uint8  `json:"webm_audio"`
	RequireSubject  uint8  `json:"require_subject"`
	MinImageWidth   int64  `json:"min_image_width"`
	MinImageHeight  int64  `json:"min_image_height"`
	//CoolDowns array `json:"cooldowns"`
}

//Thread - a discussion
type Thread struct {
	No           int64 `json:"no"`
	LastModified int64 `json:"last_modified"`
	Replies      int64 `json:"replies"`
}

//ThreadList - a list of threads (page)
type ThreadList struct {
	Page    int64    `json:"page"`
	Threads []Thread `json:"threads"`
}

//Catalog - a catalog
type Catalog struct {
}

//Archive - an archive
type Archive struct {
}

//Indexes - indexes
type Indexes struct {
}

//Post - basic structure
type Post struct{}
